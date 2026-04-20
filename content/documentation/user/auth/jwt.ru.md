---
title: "JWT"
linkTitle: "JWT"
weight: 30
description: "Аутентификация в Deckhouse Stronghold по JWT."
---

## JWT

Способ аутентификации для ролей типа `jwt` проще, чем в `OIDC`, поскольку **Deckhouse Stronghold** требуется только проверить предоставленный JWT [4].

Этот метод подходит, если клиент уже может получить JWT от доверенного issuer и передать его в Stronghold без запуска браузерного OIDC-сценария [4].

> Примечание  
> Если вам нужен интерактивный вход через браузер и OIDC-провайдера, используйте раздел [OIDC](../OIDC/). В `JWT`-сценарии Stronghold не инициирует browser-based flow, а только валидирует уже выданный токен [4].

## Как работает JWT-аутентификация

При использовании метода `JWT` клиент передаёт в Stronghold JWT и имя роли. Stronghold:

- проверяет подпись JWT;
- проверяет срок действия токена;
- проверяет связанные параметры роли;
- при успешной проверке выдаёт токен Stronghold [4].

В результате пользователь или приложение получают обычный токен Stronghold, с которым можно выполнять дальнейшие операции.

## Проверка JWT

Подписи JWT проверяются по открытым ключам issuer. Для одного backend'а можно выбрать один из следующих способов проверки [4]:

- **статические ключи** — набор открытых ключей хранится в конфигурации backend'а;
- **JWKS** — используется URL [JSON Web Key Set](https://tools.ietf.org/html/rfc7517), из которого ключи извлекаются при аутентификации;
- **OIDC Discovery** — используется URL OIDC Discovery, из которого извлекаются ключи и применяются критерии проверки OIDC, например `iss` и `aud` [4].

Если требуется использовать несколько способов проверки, необходимо поднять несколько экземпляров backend'а `JWT auth` [4].

## Аутентификация через CLI

Для аутентификации через CLI используйте:

```bash
d8 stronghold write auth/<path-to-jwt-backend>/login role=demo jwt=...
```

Путь по умолчанию для backend'а JWT-аутентификации — `/jwt`, поэтому если используется backend по умолчанию, команда будет выглядеть так:

```bash
d8 stronghold write auth/jwt/login role=demo jwt=...
```

Если backend JWT смонтирован по другому пути, используйте его вместо `jwt` [4].

## Аутентификация через API

По умолчанию используется endpoint `auth/jwt/login`. Если метод аутентификации включён по другому пути, замените `jwt` на нужное значение [4].

Пример запроса:

```bash
curl \
  --request POST \
  --data '{"jwt": "your_jwt", "role": "demo"}' \
  http://127.0.0.1:8200/v1/auth/jwt/login
```

В ответе Stronghold вернёт токен по адресу `auth.client_token` [4].

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

Перед тем как аутентифицироваться, необходимо включить и настроить backend `JWT auth`. Это обычно делает администратор или средство управления конфигурацией [4].

Включите метод аутентификации:

```bash
d8 stronghold auth enable jwt
```

Также можно смонтировать backend по имени `oidc`:

```bash
d8 stronghold auth enable oidc
```

Backend будет смонтирован по выбранному имени [4].

## Настройка backend'а

Для настройки Stronghold используется endpoint `/config`. Для поддержки ролей `JWT` необходимо наличие:

- локальных ключей;
- либо URL `JWKS`;
- либо URL `OIDC Discovery` [4].

Для ролей `OIDC` дополнительно требуются `oidc_client_id` и `oidc_client_secret`, но для сценария `JWT` они могут оставаться пустыми [4].

### Пример конфигурации через OIDC Discovery

```bash
d8 stronghold write auth/jwt/config \
   oidc_discovery_url="https://myco.auth0.com/" \
   oidc_client_id="m5i8bj3iofytj" \
   oidc_client_secret="f4ubv72nfiu23hnsj" \
   default_role="demo"
```

### Пример конфигурации только для проверки JWT

Если Stronghold должен только валидировать JWT, оставьте `oidc_client_id` и `oidc_client_secret` пустыми [4]:

```bash
d8 stronghold write auth/jwt/config \
   oidc_discovery_url="https://MYDOMAIN.eu.auth0.com/" \
   oidc_client_id="" \
   oidc_client_secret=""
```

## Создание роли

После настройки backend'а создайте именованную роль:

```bash
d8 stronghold write auth/jwt/role/demo \
  allowed_redirect_uris="http://localhost:8250/oidc/callback" \
  bound_subject="r3qX9DljwFIWhsiqwFiu38209F10atW6@clients" \
  bound_audiences="https://vault.plugin.auth.jwt.test" \
  user_claim="https://vault/user" \
  groups_claim="https://vault/groups" \
  policies=webapps \
  ttl=1h
```

Такая роль:

- авторизует JWT с заданными `subject` и `audience`;
- задаёт политику `webapps`;
- использует указанные claims пользователя и групп для настройки identity-псевдонимов [4].

## Связанные параметры

После того как JWT подтверждён как корректно подписанный и неистёкший, Stronghold проверяет все настроенные связанные параметры роли [4].

### bound_subject

Параметр `bound_subject` должен совпадать со значением `sub` в JWT.

### bound_claims

Параметр `bound_claims` позволяет задавать произвольные ограничения по claims. Это карта вида:

```json
{
  "division": "Europe",
  "department": "Engineering"
}
```

Будут авторизованы только JWT, содержащие claims `division` и `department` со значениями `Europe` и `Engineering` [4].

Если значение представляет собой список, claim должен совпадать с одним из элементов списка. Например:

```json
{
  "email": ["fred@example.com", "julie@example.com"]
}
```

## Claims как метаданные

Данные claims можно копировать в метаданные токена аутентификации и alias с помощью `claim_mappings` [4].

Пример:

```json
{
  "division": "organization",
  "department": "department"
}
```

Это означает:

- значение claim `division` копируется в ключ метаданных `organization`;
- значение claim `department` копируется в одноимённый ключ `department` [4].

> Примечание  
> Имя ключа метаданных `role` зарезервировано и не может использоваться для сопоставления claims [4].

## Claims и JSON Pointer

Некоторые параметры, такие как:

- `bound_claims`;
- `groups_claim`;
- `claim_mappings`;
- `user_claim`;

могут ссылаться на данные внутри JWT [4].

Если нужный claim расположен на верхнем уровне JWT, можно указать его имя напрямую. Если он вложен глубже, можно использовать [JSON Pointer](https://tools.ietf.org/html/rfc6901) [4].

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
- `/groups/primary` указывает на `Engineering` [4].

## Практические рекомендации

- Используйте `JWT`, если клиент уже получает JWT от доверенного issuer.
- Выбирайте один способ проверки подписи для одного backend'а.
- Для сложных сценариев используйте несколько backend'ов, а не перегружайте один.
- Сначала добейтесь успешной базовой аутентификации, а затем добавляйте ограничения по claims.
- Перед включением жёстких `bound_claims` и `bound_subject` проверьте реальные значения claims, которые приходят от issuer.
