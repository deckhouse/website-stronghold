---
title: "OIDC identity provider"
description: "Сведения о разделе \"OIDC identity provider\" в Deckhouse Stronghold."
weight: 20
---

## Обзор

Stronghold может выступать как провайдер идентификации OpenID Connect (OIDC) для клиентских приложений, которые используют этот протокол.
Это позволяет приложениям использовать Stronghold как источник [идентификации](../../../../concepts/identity/) и применять доступные [методы аутентификации](../../../../concepts/auth/) для аутентификации конечных пользователей.

После включения этой возможности Stronghold может выступать посредником между клиентским приложением и внешними провайдерами идентификации через уже настроенные методы аутентификации.
Кроме того, клиентские приложения могут получать данные об идентификации конечных пользователей из Stronghold.

Система OIDC-провайдеров в Stronghold построена на механизме секретов идентификации.
Этот механизм включён по умолчанию, его нельзя отключить или переместить.

В каждом неймспейсе Stronghold по умолчанию доступны OIDC-провайдер и ключ.
Эта встроенная конфигурация позволяет начать использование Stronghold в качестве провайдера идентификации с минимальной начальной настройкой.

Чтобы клиентское приложение могло использовать Stronghold как OIDC-провайдер, обычно требуется:

- Включить метод аутентификации.
- Создать пользователя.
- Создать клиентское приложение.
- Получить `client_id` и `client_secret`.
- Получить значение `issuer` из OIDC discovery-конфигурации.

## Настройка

Ниже приведён минимальный пример настройки, который позволяет клиентскому приложению использовать Stronghold в качестве OIDC-провайдера.

1. Включите метод аутентификации `userpass`.

   ```console
   $ d8 stronghold auth enable userpass
   Success! Enabled userpass auth method at: userpass/
   ```

   В режиме OIDC можно использовать любой метод аутентификации Stronghold.
   В этом примере для простоты используется метод `userpass`.

1. Создайте пользователя.

   ```console
   $ d8 stronghold write auth/userpass/users/end-user password="securepassword"
   Success! Data written to: auth/userpass/users/end-user
   ```

   Этот пользователь будет проходить аутентификацию в Stronghold через клиентское приложение, то есть через OIDC relying party.

1. Создайте клиентское приложение.

   ```console
   $ d8 stronghold write identity/oidc/client/my-webapp \
     redirect_uris="https://localhost:9702/auth/oidc-callback" \
     assignments="allow_all"
   Success! Data written to: identity/oidc/client/my-webapp
   ```

   Эта команда создаёт клиентское приложение, которое можно использовать при настройке OIDC relying party.

   Параметр `assignments` ограничивает сущности и группы Stronghold, которым разрешена аутентификация через это клиентское приложение.
   По умолчанию аутентификация не разрешена ни одной сущности.
   Чтобы разрешить её всем сущностям Stronghold, используйте встроенное назначение `allow_all`.

1. Прочитайте учётные данные клиента.

   ```console
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

   Параметры `client_id` и `client_secret` — это учётные данные клиентского приложения.
   Обычно они требуются при настройке OIDC relying party.

1. Прочитайте OIDC discovery-конфигурацию.

   ```console
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

   Каждый Stronghold OIDC-провайдер публикует discovery-метаданные.
   Значение `issuer` обычно требуется при настройке OIDC relying party.

## Использование

После настройки метода аутентификации Stronghold и клиентского приложения используйте следующие значения для настройки OIDC relying party:

- `client_id` — идентификатор клиентского приложения;
- `client_secret` — секрет клиентского приложения;
- `issuer` — эмитент OIDC-провайдера Stronghold.

Дальнейшая настройка зависит от конкретного клиентского приложения.
Подробности смотрите в документации соответствующей OIDC relying party.

## Поддерживаемые процессы

Функция OIDC-провайдера в Stronghold в настоящее время поддерживает следующий процесс аутентификации:

- [Authorization Code Flow](https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth).
