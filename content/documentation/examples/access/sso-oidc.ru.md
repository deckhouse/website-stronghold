---
title: "Единый вход через корпоративный IdP (OIDC)"
linkTitle: "SSO через OIDC"
description: "Как организовать единый вход в Stronghold через корпоративного провайдера идентификации по OIDC: общая схема, полный пример с Keycloak и сопоставлением групп с политиками, вход через Dex в DKP и ссылки на руководства по провайдерам."
weight: 30
params:
  relatedLinks:
    - title: "Метод OIDC"
      url: ../../../user/auth/oidc/
    - title: "OIDC-провайдер Keycloak"
      url: ../../../user/auth/oidc/keycloak/
    - title: "OIDC-провайдер GitLab"
      url: ../../../user/auth/oidc/gitlab/
    - title: "Настройка Stronghold в DKP"
      url: ../../../install/dkp/configuration/
    - title: "Идентификационные данные (Identity)"
      url: ../../../concepts/identity/
    - title: "Аварийный доступ (break-glass)"
      url: ../break-glass-access/
---

Единый вход (SSO) позволяет пользователям входить в Stronghold под корпоративной учётной записью, а администраторам — управлять доступом через группы в провайдере идентификации (IdP). Stronghold поддерживает любой IdP, совместимый с OpenID Connect.

## Цель

- Пользователи входят в веб-интерфейс и CLI Stronghold через корпоративный IdP.
- Политики назначаются по группам IdP через внешние группы Identity.
- Вход разрешён только членам определённых групп.

## Общая схема

![Схема входа через OIDC](../../../images/ex-sso-oidc.png)

Что нужно настроить в любом IdP:

1. Клиент (приложение) OIDC с типом `confidential` и секретом клиента.
1. URI перенаправления: для веб-интерфейса — `https://<адрес Stronghold>/ui/stronghold/auth/<путь метода>/oidc/callback`, для CLI — `http://localhost:8250/oidc/callback`.
1. Утверждение (claim) со списком групп пользователя в ID-токене.

## Выбор провайдера

| Провайдер | Когда использовать | Руководство |
| --- | --- | --- |
| Dex в DP | Stronghold установлен в DP, пользователи уже входят в кластер через модуль `user-authn` | [Вход через Dex в DP](#вход-через-dex-в-dp) на этой странице |
| Keycloak | Корпоративный IdP на Keycloak, в том числе с федерацией из LDAP или AD | [OIDC-провайдер Keycloak](../../../user/auth/oidc/keycloak/), пример ниже |
| GitLab | Учётные записи и группы разработчиков ведутся в GitLab | [OIDC-провайдер GitLab](../../../user/auth/oidc/gitlab/) |
| Другой OIDC-провайдер | Любой IdP с discovery-документом `/.well-known/openid-configuration` | [Метод OIDC](../../../user/auth/oidc/) |

## Предварительные требования

- Stronghold доступен пользователям по HTTPS, например `https://stronghold.example.com`.
- IdP доступен узлам Stronghold по HTTPS, сертификат IdP выдан доверенным УЦ.
- Токен Stronghold с правами на настройку методов аутентификации, политик и Identity.
- Подготовленный [аварийный доступ](../break-glass-access/) на случай недоступности IdP.

## Пример: Keycloak

В примере пользователи из группы Keycloak `stronghold-admins` получают политику `admin`, а из группы `stronghold-users` — политику `user-read`.

### Шаг 1. Настройте клиент в Keycloak

1. В Realm `corp` создайте клиент `stronghold` с протоколом `openid-connect`, включите **Client authentication** и **Standard flow**.
1. В **Valid Redirect URIs** укажите:

   ```text
   https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback
   http://localhost:8250/oidc/callback
   ```

1. Добавьте в клиент маппер **Group Membership** с именем утверждения `groups` и отключите **Full group path**.
1. На вкладке **Credentials** скопируйте секрет клиента.

Подробности настройки — в разделе [«OIDC-провайдер Keycloak»](../../../user/auth/oidc/keycloak/).

### Шаг 2. Настройте метод OIDC

1. Включите метод и настройте подключение к Realm:

   ```bash
   d8 stronghold auth enable oidc
   d8 stronghold write auth/oidc/config \
     oidc_discovery_url="https://keycloak.example.com/realms/corp" \
     oidc_client_id="stronghold" \
     oidc_client_secret="<секрет клиента>" \
     default_role="corp"
   ```

1. Создайте роль. `bound_claims` пропускает только участников двух групп, `groups_claim` передаёт группы в Identity:

   ```bash
   d8 stronghold write auth/oidc/role/corp -<<EOF
   {
     "role_type": "oidc",
     "allowed_redirect_uris": [
       "https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback",
       "http://localhost:8250/oidc/callback"
     ],
     "oidc_scopes": ["openid", "profile", "email"],
     "user_claim": "preferred_username",
     "groups_claim": "groups",
     "bound_claims": { "groups": ["stronghold-users", "stronghold-admins"] },
     "token_policies": ["default"],
     "token_ttl": "1h",
     "token_max_ttl": "8h"
   }
   EOF
   ```

### Шаг 3. Сопоставьте группы с политиками

Для каждой группы создайте внешнюю группу Identity и псевдоним. Значение `name` псевдонима должно совпадать с именем группы в утверждении `groups`:

```bash
OIDC_ACCESSOR=$(d8 stronghold auth list -format=json | jq -r '."oidc/".accessor')

ADMINS_ID=$(d8 stronghold write -field=id identity/group \
  name="corp-admins" type="external" policies="admin")
d8 stronghold write identity/group-alias \
  name="stronghold-admins" mount_accessor="$OIDC_ACCESSOR" canonical_id="$ADMINS_ID"

USERS_ID=$(d8 stronghold write -field=id identity/group \
  name="corp-users" type="external" policies="user-read")
d8 stronghold write identity/group-alias \
  name="stronghold-users" mount_accessor="$OIDC_ACCESSOR" canonical_id="$USERS_ID"
```

Политики `admin` и `user-read` должны существовать заранее. Примеры политик приведены в разделе [«Примеры политик»](../../../concepts/policy-examples/).

### Шаг 4. Проверьте вход

1. Войдите через CLI — откроется браузер со страницей входа Keycloak:

   ```bash
   d8 stronghold login -method=oidc role=corp
   ```

1. Проверьте политики токена:

   ```bash
   d8 stronghold token lookup
   ```

   У члена `stronghold-admins` в `identity_policies` должна быть политика `admin`.

1. В веб-интерфейсе выберите метод **OIDC**, укажите роль `corp` и войдите.

## Вход через Dex в DP

Если Stronghold установлен как модуль DP в режиме `Automatic`, после инициализации Stronghold модуль сам настраивает вход через [Dex](/products/kubernetes-platform/documentation/v1/modules/user-authn/) — провайдер аутентификации DP, который получает пользователей и группы из подключённого внешнего IdP или LDAP. <!-- TODO(verify): путь к документации модуля user-authn -->

- Метод OIDC включается по пути `oidc_deckhouse`, для администраторов создаётся роль `deckhouse_administrators`.
- Администраторы задаются в параметре `settings.management.administrators` ресурса ModuleConfig `stronghold` — группами (`type: Group`) или отдельными пользователями (`type: User`). Пользователь должен входить хотя бы в одну группу.

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: stronghold
spec:
  enabled: true
  version: 1
  settings:
    management:
      mode: Automatic
      administrators:
      - type: Group
        name: admins
```

Вход пользователя:

```bash
d8 stronghold login -path=oidc_deckhouse -method=oidc -no-print
```

Чтобы выдать права не только администраторам, создайте внешние группы Identity с псевдонимами на accessor метода `oidc_deckhouse/` так же, как в шаге 3 примера с Keycloak. Имена псевдонимов должны совпадать с именами групп, которые передаёт Dex.

<!-- TODO(verify): перезаписывает ли модуль stronghold ручные изменения в auth/oidc_deckhouse; какие роли, кроме deckhouse_administrators, можно использовать для входа обычных пользователей -->

Подробнее — в разделах [«Настройка Stronghold»](../../../install/dkp/configuration/#управление-доступами) и [«Настройка доступа и первый вход»](../../../user/get-started/access/).

## Устранение неполадок

| Симптом | Вероятная причина | Что сделать |
| --- | --- | --- |
| `redirect_uri` не принимается IdP или Stronghold | URI в клиенте IdP и в `allowed_redirect_uris` роли не совпадают | Укажите одинаковые URI в обоих местах, включая путь метода. |
| Отказ во входе из-за `bound_claims` или пустой список групп | IdP не передаёт группы в ID-токене | Проверьте маппер групп в IdP и значение `groups_claim`. |
| Вход проходит, но нет политик групп | Имя псевдонима не совпадает с именем группы в утверждении (например, передаётся полный путь `/stronghold-admins`) | Сравните `groups` в ID-токене с именами псевдонимов. В Keycloak отключите **Full group path**. |
| `x509: certificate signed by unknown authority` при записи `auth/oidc/config` | Stronghold не доверяет сертификату IdP | Передайте цепочку УЦ в параметре `oidc_discovery_ca_pem`. |
| IdP недоступен, никто не может войти | Сбой IdP или ошибка конфигурации | Воспользуйтесь [аварийным доступом](../break-glass-access/). |

## Очистка

```bash
d8 stronghold auth disable oidc
```

Удалите внешние группы `corp-admins` и `corp-users` через `identity/group/name/<имя>` и клиент `stronghold` в Keycloak.
