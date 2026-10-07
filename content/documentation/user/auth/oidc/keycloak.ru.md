---
title: "OIDC провайдер Keycloak"
linkTitle: "Keycloak"
description: "Настройка входа в Stronghold через Keycloak по OIDC: клиент и маппер групп в Keycloak, конфигурация метода oidc, роль, сопоставление групп Keycloak с группами Identity и проверка входа."
weight: 30
---

На этой странице описана настройка входа в Stronghold через Keycloak по протоколу OIDC. Общие сведения о методе приведены в разделе [«Метод OIDC»](../../oidc/).

## Настройка клиента в Keycloak

1. Выберите или создайте Realm и клиент (Client). Откройте настройки клиента (**Settings**).
1. Протокол клиента (**Client Protocol**): `openid-connect`.
1. Тип доступа (**Access Type**): `confidential`. В новых версиях Keycloak включите **Client authentication**.
1. Включите **Standard Flow Enabled**.
1. Укажите допустимые URI перенаправления (**Valid Redirect URIs**):

   ```text
   https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback
   http://localhost:8250/oidc/callback
   ```

   Первый URI используется для входа через веб-интерфейс, второй — для входа через CLI. Если метод включён по другому пути, замените `oidc` в первом URI на этот путь.

1. Нажмите **Save**.
1. Перейдите на вкладку **Credentials** и сохраните идентификатор (Client ID) и секрет клиента (Client Secret).
1. Чтобы Keycloak передавал группы пользователя, добавьте в клиент (или в его client scope) маппер типа **Group Membership** с именем утверждения `groups`. Отключите **Full group path**, если в Stronghold нужны имена групп без пути.

## Настройка Stronghold

1. Включите метод аутентификации OIDC:

   ```bash
   d8 stronghold auth enable oidc
   ```

   По умолчанию метод включается по пути `auth/oidc/`. Чтобы использовать другой путь, укажите `-path`.

1. Настройте подключение к Keycloak. В `oidc_discovery_url` укажите адрес Realm без `/.well-known/openid-configuration`:

   ```bash
   d8 stronghold write auth/oidc/config \
     oidc_discovery_url="https://keycloak.example.com/realms/myrealm" \
     oidc_client_id="<Client ID>" \
     oidc_client_secret="<Client Secret>" \
     default_role="keycloak"
   ```

   В старых версиях Keycloak адрес Realm содержит префикс `/auth`, например `https://keycloak.example.com/auth/realms/myrealm`.

1. Создайте роль. В `allowed_redirect_uris` перечислите те же URI, что и в клиенте Keycloak:

   ```bash
   d8 stronghold write auth/oidc/role/keycloak -<<EOF
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
     "token_ttl": "1h"
   }
   EOF
   ```

   Основные параметры роли:

   | Параметр | Описание |
   |----------|----------|
   | `allowed_redirect_uris` | Разрешённые URI перенаправления. Должны точно совпадать с URI в Keycloak. |
   | `user_claim` | Утверждение (claim), значение которого становится именем псевдонима сущности, например `preferred_username` или `sub`. |
   | `groups_claim` | Утверждение со списком групп пользователя (`groups` из маппера Group Membership). |
   | `bound_claims` | Утверждения и значения, которые должны совпасть для успешного входа. В примере вход разрешён только участникам групп `stronghold-users` или `stronghold-admins`. |
   | `oidc_scopes` | Запрашиваемые области действия OIDC. |
   | `token_policies`, `token_ttl` | Политики и TTL выдаваемого токена. |

## Сопоставление групп Keycloak с группами Stronghold

Чтобы назначать политики по членству в группах Keycloak, создайте внешние группы Identity и псевдонимы групп. Подробнее о внешних группах — в разделе [«Идентичность»](../../../../concepts/identity/).

1. Создайте внешнюю группу с нужными политиками и сохраните её идентификатор:

   ```bash
   d8 stronghold write -field=id identity/group \
     name="keycloak-admins" \
     type="external" \
     policies="admin"
   ```

1. Получите accessor метода аутентификации `oidc/`:

   ```bash
   d8 stronghold auth list -format=json | jq -r '."oidc/".accessor'
   ```

1. Создайте псевдоним группы. Значение `name` должно совпадать с именем группы в утверждении `groups`:

   ```bash
   d8 stronghold write identity/group-alias \
     name="stronghold-admins" \
     mount_accessor="<accessor метода oidc>" \
     canonical_id="<идентификатор группы>"
   ```

При каждом входе и продлении токена Stronghold обновляет членство сущности во внешних группах по данным Keycloak.

## Проверка входа

1. Выполните вход через CLI:

   ```bash
   d8 stronghold login -method=oidc -path=oidc role=keycloak
   ```

   CLI откроет браузер для входа в Keycloak и примет ответ на `http://localhost:8250/oidc/callback`.

1. Проверьте политики и метаданные полученного токена:

   ```bash
   d8 stronghold token lookup
   ```

1. Для проверки входа через веб-интерфейс выберите метод **OIDC**, при необходимости укажите роль `keycloak` и завершите вход в Keycloak.

Если вход не удаётся, проверьте совпадение URI перенаправления и воспользуйтесь советами из раздела [«Метод OIDC»](../../oidc/).

## Примеры использования

Готовые примеры с этим механизмом:

- [Единый вход через корпоративный IdP (OIDC)](../../../../examples/access/sso-oidc/)

Все примеры собраны в разделе [«Примеры использования»](../../../../examples/).
