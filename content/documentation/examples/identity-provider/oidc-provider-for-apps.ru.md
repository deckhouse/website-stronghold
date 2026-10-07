---
title: "Stronghold как OIDC-провайдер для внутренних приложений"
linkTitle: "OIDC-провайдер для приложений"
description: "Настройка единого входа во внутренние приложения через OIDC-провайдер Stronghold: ключ подписи, назначения, области действия с шаблонами, клиент и провайдер."
weight: 120
params:
  relatedLinks:
    - title: "OIDC identity provider"
      url: ../../../user/secrets-engines/identity/oidc-provider/
    - title: "Токены Identity"
      url: ../../../user/secrets-engines/identity/token/
    - title: "Идентификационные данные (Identity)"
      url: ../../../concepts/identity/
    - title: "API Identity"
      url: ../../../reference/api/identity/
---

Stronghold может выступать OIDC-провайдером (OpenID Provider) для приложений, которые поддерживают вход через OpenID Connect. Пользователь входит в Stronghold любым настроенным методом аутентификации, а приложение получает ID-токен с данными сущности Stronghold: имя, адрес почты, группы.

## Цель

Подключить внутреннее приложение (в примере — Grafana) к OIDC-провайдеру Stronghold так, чтобы входить могли только участники группы `grafana-users`, а приложение получало адрес почты и список групп пользователя.

## Предварительные требования

- Токен Stronghold с правами на `identity/*`.
- Настроенный метод аутентификации для пользователей (например, OIDC через Dex в DP или [`userpass`](../../../user/auth/userpass/)).
- Группа Identity `grafana-users`, в которую входят сущности пользователей, и метаданные `email` у сущностей.
- Внешний адрес Stronghold, доступный браузерам пользователей и приложению, — в примерах `https://stronghold.example.com`.

Сущности, группы и псевдонимы описаны в разделе [«Идентификационные данные (Identity)»](../../../concepts/identity/).

## Шаг 1. Создайте ключ подписи

```bash
d8 stronghold write identity/oidc/key/apps \
  algorithm=RS256 \
  rotation_period=24h \
  verification_ttl=24h
```

Ключ автоматически ротируется раз в сутки, прежняя открытая часть доступна для проверки подписи ещё `verification_ttl`.

## Шаг 2. Ограничьте круг пользователей

Назначение (assignment) задаёт сущности и группы, которым разрешён вход через клиентское приложение:

```bash
GROUP_ID=$(d8 stronghold read -field=id identity/group/name/grafana-users)

d8 stronghold write identity/oidc/assignment/grafana-users group_ids="$GROUP_ID"
```

Встроенное назначение `allow_all` разрешает вход всем сущностям — не используйте его для рабочих приложений.

## Шаг 3. Создайте области действия

Области действия (scopes) добавляют в ID-токен утверждения (claims) по шаблону. Синтаксис шаблонов тот же, что у [шаблонных политик](../../../concepts/policy/).

```bash
d8 stronghold write identity/oidc/scope/email \
  description="Адрес почты пользователя" \
  template='{"email": {{identity.entity.metadata.email}}}'

d8 stronghold write identity/oidc/scope/groups \
  description="Группы пользователя" \
  template='{"groups": {{identity.entity.groups.names}}}'
```

Параметр `template` принимает как строку JSON, так и значение в кодировке base64.

## Шаг 4. Создайте клиентское приложение

```bash
d8 stronghold write identity/oidc/client/grafana \
  redirect_uris="https://grafana.example.com/login/generic_oauth" \
  assignments="grafana-users" \
  key="apps" \
  id_token_ttl=30m \
  access_token_ttl=1h
```

Разрешите клиенту использовать ключ `apps`:

```bash
CLIENT_ID=$(d8 stronghold read -field=client_id identity/oidc/client/grafana)

d8 stronghold write identity/oidc/key/apps allowed_client_ids="$CLIENT_ID"
```

## Шаг 5. Создайте провайдер

```bash
d8 stronghold write identity/oidc/provider/internal \
  issuer="https://stronghold.example.com" \
  allowed_client_ids="$CLIENT_ID" \
  scopes_supported="email,groups"
```

Получите параметры для настройки приложения:

```bash
d8 stronghold read identity/oidc/client/grafana
curl -s https://stronghold.example.com/v1/identity/oidc/provider/internal/.well-known/openid-configuration | jq
```

Из ответов понадобятся `client_id`, `client_secret`, `issuer`, `authorization_endpoint`, `token_endpoint` и `userinfo_endpoint`.

## Шаг 6. Настройте приложение

Пример настройки Grafana (раздел `auth.generic_oauth` в `grafana.ini`):

```ini
[auth.generic_oauth]
enabled = true
name = Stronghold
client_id = <client_id>
client_secret = <client_secret>
scopes = openid email groups
auth_url = https://stronghold.example.com/ui/stronghold/identity/oidc/provider/internal/authorize
token_url = https://stronghold.example.com/v1/identity/oidc/provider/internal/token
api_url = https://stronghold.example.com/v1/identity/oidc/provider/internal/userinfo
email_attribute_path = email
groups_attribute_path = groups
```

Храните `client_secret` в Stronghold и доставляйте в приложение так же, как другие секреты, например через [Stronghold Agent](../../../user/agent/overview/) или [интеграцию с Kubernetes](../../delivery/kubernetes-workloads/).

<!-- TODO(verify): совместимость Grafana и других конкретных приложений с OIDC-провайдером Stronghold подтверждена тестами -->

## Проверка

1. Откройте приложение и выберите вход через Stronghold. Браузер перенаправит вас на страницу входа Stronghold, после входа — обратно в приложение.
1. Войдите пользователем, не входящим в группу `grafana-users`, — вход должен быть отклонён.
1. Проверьте содержимое ID-токена в журнале приложения или декодируйте его: должны присутствовать `iss`, `aud` (равный `client_id`), `email` и `groups`.
1. Проверьте, что открытые ключи доступны:

   ```bash
   curl -s https://stronghold.example.com/v1/identity/oidc/provider/internal/.well-known/keys | jq '.keys[].kid'
   ```

## Очистка

```bash
d8 stronghold delete identity/oidc/provider/internal
d8 stronghold delete identity/oidc/client/grafana
d8 stronghold delete identity/oidc/scope/email
d8 stronghold delete identity/oidc/scope/groups
d8 stronghold delete identity/oidc/assignment/grafana-users
d8 stronghold delete identity/oidc/key/apps
```
