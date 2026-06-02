---
title: "JWT-аутентификация"
linkTitle: "JWT"
weight: 30
description: "JWT-аутентификация в Deckhouse Stronghold."
---

JWT-аутентификация подходит для сценариев, в которых клиент уже получает JSON Web Token (JWT) от доверенного поставщика токенов и передаёт его в Deckhouse Stronghold для проверки. В этом сценарии Deckhouse Stronghold не запускает интерактивный вход через браузер, а только проверяет уже выданный токен.

{{< alert level="info" >}}
Если нужен интерактивный вход через браузер и OIDC-провайдер, используйте обзор OIDC-аутентификации.
{{< /alert >}}

## Как работает JWT-аутентификация

При использовании метода `jwt`
клиент передаёт в Deckhouse Stronghold JWT и имя роли.
После этого Deckhouse Stronghold выполняет следующие проверки:

- проверяет подпись JWT;
- проверяет срок действия токена;
- проверяет параметры, связанные с ролью;
- при успешной проверке выдаёт токен Deckhouse Stronghold.

В результате пользователь или приложение получают обычный токен Deckhouse Stronghold, с которым можно выполнять дальнейшие операции.

## Проверка JWT

Подписи JWT проверяются по открытым ключам издателя токенов. Для одного бэкенда можно выбрать один из следующих способов проверки:

- **Статические ключи** — набор открытых ключей хранится в конфигурации бэкенда.
- **JWKS** — используется URL JSON Web Key Set, из которого ключи извлекаются при аутентификации.
- **OIDC Discovery** — используется URL OIDC Discovery, из которого извлекаются ключи и применяются дополнительные проверки OIDC, например `iss` и `aud`.

Если нужно использовать несколько способов проверки, создайте несколько бэкендов JWT-аутентификации.

## Аутентификация через CLI

Для аутентификации через CLI используйте следующую команду:

```shell
d8 stronghold write auth/<path-to-jwt-backend>/login role=demo jwt=...
```

Путь по умолчанию для бэкенда JWT-аутентификации — `/jwt`. Если используется бэкенд по умолчанию, команда будет выглядеть так:

```shell
d8 stronghold write auth/jwt/login role=demo jwt=...
```

Если бэкенд JWT смонтирован по другому пути, используйте этот путь вместо `jwt`.

## Аутентификация через API

По умолчанию используется эндпоинт `auth/jwt/login`. Если метод аутентификации включён по другому пути, замените `jwt` на нужное значение.

Пример запроса:

```shell
curl \
  --request POST \
  --data '{"jwt": "your_jwt", "role": "demo"}' \
  http://127.0.0.1:8200/v1/auth/jwt/login
```

В ответе Deckhouse Stronghold вернёт токен в поле `auth.client_token`.

Пример ответа:

```json
{
  "auth": {
    "client_token": "38fe9691-e623-7238-f618-c94d4e7bc674",
    "accessor": "78e87a38-84ed-2692-538f-ca8b9f400ab3",
    "policies": ["default"],
    "metadata": {
      "role": "demo"
    },
    "lease_duration": 2764800,
    "renewable": true
  }
}
```

## Включение метода

Перед аутентификацией включите и настройте бэкенд JWT-аутентификации. Обычно это делает администратор или средство управления конфигурацией.

Чтобы включить метод аутентификации, выполните команду:

```shell
d8 stronghold auth enable jwt
```

Также можно смонтировать этот же метод по другому пути. Например, по пути `oidc`:

```shell
d8 stronghold auth enable -path=oidc jwt
```

Бэкенд будет смонтирован по выбранному пути.

## Настройка бэкенда

Для настройки Deckhouse Stronghold используется эндпоинт `/config`. Чтобы поддерживать роли типа `jwt`, укажите один из следующих источников ключей:

- локальные ключи;
- URL JWKS;
- URL OIDC Discovery.

Для ролей типа `oidc` дополнительно требуются параметры `oidc_client_id` и `oidc_client_secret`. Для сценария `jwt` эти параметры можно оставить пустыми.

### Пример конфигурации через OIDC Discovery

```shell
d8 stronghold write auth/jwt/config \
  oidc_discovery_url="https://myco.auth0.com/" \
  oidc_client_id="m5i8bj3iofytj" \
  oidc_client_secret="f4ubv72nfiu23hnsj" \
  default_role="demo"
```

### Пример конфигурации только для проверки JWT

Если Deckhouse Stronghold должен только проверять JWT, оставьте `oidc_client_id` и `oidc_client_secret` пустыми:

```shell
d8 stronghold write auth/jwt/config \
  oidc_discovery_url="https://MYDOMAIN.eu.auth0.com/" \
  oidc_client_id="" \
  oidc_client_secret=""
```

## Создание роли

После настройки бэкенда создайте именованную роль:

```shell
d8 stronghold write auth/jwt/role/demo \
  bound_subject="r3qX9DljwFIWhsiqwFiu38209F10atW6@clients" \
  bound_audiences="https://vault.plugin.auth.jwt.test" \
  user_claim="https://vault/user" \
  groups_claim="https://vault/groups" \
  policies=webapps \
  ttl=1h
```

Такая роль:

- разрешает аутентификацию по JWT с указанными значениями `sub` и `aud`;
- назначает политику `webapps`;
- использует указанные claims пользователя и групп
  для настройки алиасов identity.

## Связанные параметры роли

После того как JWT успешно проходит проверку подписи и срока действия, Deckhouse Stronghold проверяет все параметры,
связанные с ролью.

### `bound_subject`

Параметр `bound_subject` должен совпадать со значением `sub` в JWT.

### `bound_claims`

Параметр `bound_claims` позволяет задавать произвольные ограничения по claims. Это JSON-файл в виде карты ключей и значений.

Пример:

```json
{
  "division": "Europe",
  "department": "Engineering"
}
```

Будут авторизованы только JWT, содержащие claims `division` и `department` со значениями `Europe` и `Engineering`.

Если значение представляет собой список, claim должен совпадать с одним из элементов списка.

Например:

```json
{
  "email": ["fred@example.com", "julie@example.com"]
}
```

## Claims как метаданные

Данные claims можно копировать в метаданные токена аутентификации и алиасов с помощью параметра `claim_mappings`.

Пример:

```json
{
  "division": "organization",
  "department": "department"
}
```

Это означает следующее:

- значение claim `division` копируется в ключ метаданных `organization`;
- значение claim `department` копируется в ключ метаданных `department`.

{{< alert level="info" >}}
Имя ключа метаданных `role` зарезервировано и не может использоваться для сопоставления claims.
{{< /alert >}}

## Claims и JSON Pointer

Параметры `bound_claims`, `groups_claim`, `claim_mappings` и `user_claim` могут ссылаться как на claims верхнего уровня, так и на вложенные данные внутри JWT.

Если нужный claim расположен на верхнем уровне JWT,укажите его имя напрямую. Если claim вложен глубже,используйте JSON Pointer.

Пример JWT:

```json
{
  "division": "North America",
  "groups": {
    "primary": "Engineering",
    "secondary": "Software"
  }
}
```

В этом случае:

- `division` указывает на `North America`;
- `/groups/primary` указывает на `Engineering`.

## Практические рекомендации

Используйте следующие рекомендации:

- Используйте JWT-аутентификацию, если клиент уже получает JWT от доверенного поставщика токенов.
- Выбирайте для одного бэкенда только один способ проверки подписи.
- Для сложных сценариев создавайте несколько бэкендов, а не перегружайте один.
- Сначала добейтесь успешной базовой аутентификации, а затем добавляйте ограничения по claims.
- Перед включением жёстких ограничений `bound_claims` и `bound_subject` проверьте реальные значения claims, которые приходят от издателя токенов.
