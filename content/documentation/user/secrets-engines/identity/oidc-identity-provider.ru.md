---
title: "OIDC identity provider"
linkTitle: "OIDC identity provider"
description: "Использование Deckhouse Stronghold как OIDC identity provider"
weight: 10
---

Deckhouse Stronghold может работать как **OIDC identity provider**. Это позволяет клиентским приложениям, которые поддерживают OpenID Connect (OIDC), использовать Stronghold как источник идентичности и аутентифицировать пользователей через встроенные методы аутентификации Stronghold.

Такой сценарий полезен, если нужно:

- использовать Stronghold как единый OIDC identity provider для внутренних приложений;
- аутентифицировать пользователей через уже настроенные методы аутентификации Stronghold;
- передавать приложениям данные об идентичности пользователей через OIDC.

## Как это работает

Stronghold выступает посредником между клиентским приложением и методами аутентификации Stronghold. Пользователь проходит аутентификацию в Stronghold, а приложение получает OIDC-совместимые данные для входа.

Для работы нужны:

- включённый метод аутентификации Stronghold;
- клиентское приложение, зарегистрированное в Stronghold;
- параметры `client_id`, `client_secret` и `issuer`, которые потом используются в OIDC relying party.

## Что нужно подготовить

Система OIDC provider в Stronghold построена на механизме идентичности. Этот механизм включён по умолчанию и не может быть отключён или перемещён.

В каждом namespace Stronghold по умолчанию уже есть OIDC provider и ключ. Это позволяет начать работу с минимальной настройкой.

## Минимальная настройка

Ниже показан базовый сценарий, который позволяет клиентскому приложению использовать Stronghold как OIDC identity provider.

### Шаг 1. Включите метод аутентификации

Для примера используем метод `userpass`:

```text
$ d8 stronghold auth enable userpass
Success! Enabled userpass auth method at: userpass/
```

В режиме OIDC можно использовать любой метод аутентификации Stronghold. В примере выбран `userpass`, потому что его проще всего показать в документации.

### Шаг 2. Создайте пользователя

```text
$ d8 stronghold write auth/userpass/users/end-user password="securepassword"
Success! Data written to: auth/userpass/users/end-user
```

Этот пользователь будет проходить аутентификацию в Stronghold через клиентское приложение.

### Шаг 3. Создайте клиентское приложение

```text
$ d8 stronghold write identity/oidc/client/my-webapp \
  redirect_uris="https://localhost:9702/auth/oidc-callback" \
  assignments="allow_all"
Success! Data written to: identity/oidc/client/my-webapp
```

Эта команда создаёт клиентское приложение, которое можно использовать в OIDC-сценарии.

Параметр `assignments` определяет, каким сущностям и группам Stronghold разрешено проходить аутентификацию через это приложение. По умолчанию доступ не разрешён никому. Значение `allow_all` открывает доступ всем сущностям Stronghold.

### Шаг 4. Прочитайте учётные данные клиента

```text
$ d8 stronghold read identity/oidc/client/my-webapp

Key                 Value
---                 -----
access_token_ttl    24h
assignments         [allow_all]
client_id           GSDTnn3KaOrLpNlVGlYLS9TVsZgOTweO
client_secret       hvo_secret_gBKHcTP58C4aq7FqPWsuqKgpiiegd7ahpifGae9WGkHRCwFEJTZA9KGdNVpzE0r8
client_type         confidential
id_token_ttl        24h
key                 default
redirect_uris       [https://localhost:9702/auth/oidc-callback]
```

Из этого ответа вам понадобятся:

- `client_id` — идентификатор клиентского приложения;
- `client_secret` — секрет клиентского приложения.

Обычно именно эти значения нужно указать при настройке OIDC relying party.

### Шаг 5. Прочитайте OIDC discovery configuration

```text
$ curl -s http://127.0.0.1:8200/v1/identity/oidc/provider/default/.well-known/openid-configuration
{
  "issuer": "http://127.0.0.1:8200/v1/identity/oidc/provider/default",
  "jwks_uri": "http://127.0.0.1:8200/v1/identity/oidc/provider/default/.well-known/keys",
  "authorization_endpoint": "http://127.0.0.1:8200/ui/vault/identity/oidc/provider/default/authorize",
  "token_endpoint": "http://127.0.0.1:8200/v1/identity/oidc/provider/default/token",
  "userinfo_endpoint": "http://127.0.0.1:8200/v1/identity/oidc/provider/default/userinfo",
  "request_parameter_supported": false,
  "request_uri_parameter_supported": false,
  "id_token_signing_alg_values_supported": [
    "RS256",
    "RS384",
    "RS512",
    "ES256",
    "ES384",
    "ES512",
    "EdDSA"
  ],
  "response_types_supported": [
    "code"
  ],
  "scopes_supported": [
    "openid"
  ],
  "subject_types_supported": [
    "public"
  ],
  "grant_types_supported": [
    "authorization_code"
  ],
  "token_endpoint_auth_methods_supported": [
    "none",
    "client_secret_basic",
    "client_secret_post"
  ]
}
```

Каждый OIDC provider в Stronghold публикует discovery metadata. На практике здесь чаще всего нужен параметр `issuer` — его тоже нужно передать в клиентское приложение.

## Что указать в клиентском приложении

После базовой настройки у вас есть три основных параметра для интеграции приложения с Stronghold:

- `client_id`;
- `client_secret`;
- `issuer`.

Дальше настройка зависит уже от конкретного приложения, которое выступает как OIDC relying party.

## Поддерживаемый поток аутентификации

Сейчас Stronghold в режиме OIDC provider поддерживает:

- [Authorization Code Flow](https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth).

## Практические рекомендации

Чтобы клиенту было проще настроить интеграцию, придерживайтесь такого порядка:

- сначала включите и проверьте метод аутентификации Stronghold;
- затем создайте клиентское приложение;
- после этого получите `client_id`, `client_secret` и `issuer`;
- только потом переходите к настройке OIDC relying party;
- если нужно быстро проверить сценарий, используйте `userpass` как самый простой метод аутентификации для теста.

## Что дальше

- Если вам нужно выпускать OIDC-совместимые токены с данными идентичности, используйте [OIDC identity tokens](./oidc-identity-tokens/).
- Если вам нужна аутентификация пользователей через внешний OIDC-провайдер, используйте [OIDC](../../auth/oidc/overview/).
