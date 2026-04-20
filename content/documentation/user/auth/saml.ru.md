---
title: "SAML"
linkTitle: "SAML"
weight: 75
description: "Аутентификация в Deckhouse Stronghold через внешний SAML 2.0 Identity Provider."
---

## SAML

Метод `saml` позволяет аутентифицировать пользователей в **Deckhouse Stronghold** через внешний `SAML 2.0` Identity Provider по профилю `Web SSO`. Stronghold в этом сценарии выступает как `SAML Service Provider`, проверяет ответ от Identity Provider и затем выдаёт токен Stronghold в соответствии с выбранной ролью.

Этот метод подходит:

- для входа в Stronghold UI через браузер;
- для собственных интеграций, которые запускают SAML-сценарий через HTTP API.

## Как это работает

Сценарий входа состоит из трёх этапов:

1. Клиент вызывает `auth/<mount>/sso_service_url` и получает URL для перехода к Identity Provider и временный `token_poll_id`.
2. Пользователь проходит аутентификацию у Identity Provider, после чего тот отправляет подписанный `SAML`-ответ в `auth/<mount>/callback`.
3. Клиент обменивает `token_poll_id` и `client_verifier` на токен через `auth/<mount>/token`.

Stronghold поддерживает два режима клиента:

- `browser` — для UI-сценариев;
- `cli` — для внешних инструментов, которые открывают страницу IdP в браузере и затем опрашивают endpoint выдачи токена.

## Включение метода

```bash
d8 stronghold auth enable saml
```

По умолчанию метод будет смонтирован по пути `auth/saml`. При необходимости можно использовать другой путь:

```bash
d8 stronghold auth enable -path=corp-saml saml
```

## Базовая конфигурация

Перед началом работы Stronghold нужно настроить как `SAML Service Provider`.

Основные параметры:

- `entity_id` — идентификатор `Service Provider`;
- `acs_urls` — список разрешённых callback URL для `Assertion Consumer Service`;
- `default_role` — роль по умолчанию, если она нужна;
- `idp_metadata_url` — URL метаданных Identity Provider;
- `idp_sso_url`, `idp_entity_id`, `idp_cert` — ручная альтернатива метаданным;
- `validate_response_signature` и `validate_assertion_signature` — параметры проверки подписей;
- `verbose_logging` — расширенное логирование SAML-обмена.

### Настройка через метаданные IdP

```bash
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://stronghold.example.com/v1/auth/saml/callback" \
  idp_metadata_url="https://idp.example.com/app/stronghold/sso/saml/metadata" \
  default_role="employees" \
  validate_response_signature=true \
  validate_assertion_signature=true
```

### Ручная настройка

```bash
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://stronghold.example.com/v1/auth/saml/callback" \
  idp_sso_url="https://idp.example.com/sso" \
  idp_entity_id="https://idp.example.com/entity" \
  idp_cert=@idp-signing-cert.pem \
  validate_response_signature=true \
  validate_assertion_signature=true
```

## ACS URL

Параметр `acs_urls` определяет, на какие адреса Identity Provider может вернуть SAML-ответ после успешной аутентификации пользователя.

При настройке `acs_urls` убедитесь, что каждый адрес:

- совпадает с callback URL, разрешённым в SAML-приложении на стороне Identity Provider;
- указывает на endpoint Stronghold вида `.../v1/auth/<mount>/callback`;
- использует `https://` в production.

Если Stronghold доступен по нескольким адресам, можно задать несколько `acs_urls`:

```bash
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://primary.example.com/v1/auth/saml/callback,https://secondary.example.com/v1/auth/saml/callback" \
  idp_metadata_url="https://idp.example.com/app/stronghold/sso/saml/metadata"
```

## Роли SAML

Роли определяют, какие субъекты и атрибуты из SAML-утверждения допускаются к аутентификации, а также какие параметры токена применяются после входа.

Основные параметры роли:

- `bound_subjects`;
- `bound_subjects_type`;
- `bound_attributes`;
- `bound_attributes_type`;
- `groups_attribute`;
- `alias_metadata`;
- параметры токена: `token_policies`, `token_ttl`, `token_max_ttl`, `token_period`, `token_bound_cidrs`.

Пример роли:

```bash
d8 stronghold write auth/saml/role/employees \
  bound_subjects="*@example.com" \
  bound_subjects_type="glob" \
  bound_attributes=department=platform \
  groups_attribute="memberOf" \
  token_policies="default,developers" \
  token_ttl="1h"
```

## Ограничения по subject и атрибутам

После того как Identity Provider аутентифицировал пользователя, Stronghold проверяет ограничения роли по содержимому SAML-утверждения.

- `bound_subjects` сопоставляется с `SAML subject`;
- `bound_attributes` сопоставляется с обязательными атрибутами assertion.

Это позволяет выдавать доступ только пользователям, соответствующим ожидаемым условиям.

## Привязка SAML-групп к Stronghold Identity

Если вы хотите, чтобы членство в SAML-группах выдавало политики Stronghold через `Identity`, задайте `groups_attribute` в роли и создайте соответствующие внешние `Identity groups` и `aliases`.

Общий порядок:

1. создайте внешнюю `Identity group` с нужными политиками;
2. получите `mount accessor` для `SAML auth method`;
3. создайте `group alias`, имя которого совпадает со значением, приходящим в SAML-атрибуте.

## Вход через UI

Stronghold UI поддерживает `SAML` напрямую.

Типовой сценарий:

1. Выберите метод входа `SAML`.
2. Укажите имя роли, если для mount path не задан `default_role`.
3. Нажмите `Sign In`.
4. Пройдите аутентификацию у внешнего Identity Provider.
5. После успешного callback Stronghold выдаст токен и выполнит вход.

## Запуск логина через API

Этот API-сценарий полезен для собственных порталов, обёрток и внешних CLI-инструментов.

### Шаг 1. Сгенерируйте verifier и challenge

```bash
verifier="$(uuidgen)"
challenge="$(printf '%s' "$verifier" | openssl dgst -sha256 -binary | base64)"
```

### Шаг 2. Запросите SSO URL

```bash
curl \
  --request POST \
  --data "{\"role\":\"employees\",\"client_challenge\":\"$challenge\",\"client_type\":\"browser\"}" \
  https://stronghold.example.com/v1/auth/saml/sso_service_url
```

В ответе Stronghold вернёт:

- `sso_service_url`;
- `token_poll_id`.

### Шаг 3. Обменяйте временное состояние на токен

```bash
curl \
  --request POST \
  --data "{\"token_poll_id\":\"<poll-id>\",\"client_verifier\":\"$verifier\"}" \
  https://stronghold.example.com/v1/auth/saml/token
```

Токен Stronghold возвращается в `auth.client_token`.

## Практические рекомендации

- В production используйте только `https://` в `acs_urls`.
- Если IdP умеет подписывать и response, и assertion, включайте обе проверки подписи.
- Не оставляйте `verbose_logging=true` в production, так как журналы могут содержать чувствительные SAML-данные.
- В роли должно быть задано хотя бы одно условие допуска: `bound_subjects` или `bound_attributes`.
- Если используется несколько `acs_urls`, клиент должен передавать корректный `acs_url` при запуске входа.
