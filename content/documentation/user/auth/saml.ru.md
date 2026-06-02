---
title: "SAML"
linkTitle: "SAML"
weight: 75
params:
  edition: ee
---

Метод `saml` позволяет аутентифицировать пользователей в Deckhouse Stronghold через внешний SAML 2.0 Identity Provider по профилю Web SSO.

В этом сценарии Stronghold выступает как SAML Service Provider. Он валидирует ответ от Identity Provider и затем выдаёт токен Stronghold в соответствии с выбранной ролью.

Этот метод подходит в следующих случаях:

- для входа в веб-интерфейс Stronghold через браузер;
- для собственных интеграций, которые запускают SAML-сценарий через HTTP API.

## Как это работает

Сценарий входа состоит из трёх этапов:

1. Клиент вызывает `auth/<mount>/sso_service_url` и получает URL для перехода к Identity Provider, а также временный `token_poll_id`.
1. Пользователь проходит аутентификацию у Identity Provider. После этого провайдер отправляет подписанный SAML-ответ в `auth/<mount>/callback`.
1. Клиент обменивает `token_poll_id` и `client_verifier` на токен через `auth/<mount>/token`.

Stronghold поддерживает два режима клиента:

- `browser` — для сценариев с веб-интерфейсом;
- `cli` — для внешних инструментов, которые открывают страницу IdP в браузере, а затем опрашивают эндпоинт выдачи токена.

## Включение метода

Включите метод аутентификации SAML:

```shell
d8 stronghold auth enable saml
```

По умолчанию метод будет смонтирован по пути `auth/saml`.

При необходимости используйте другой путь:

```shell
d8 stronghold auth enable -path=corp-saml saml
```

## Конфигурация

Перед началом работы настройте Stronghold как SAML Service Provider.

Основные параметры конфигурации:

- `entity_id` — идентификатор Service Provider, который должен совпадать с настройками приложения на стороне Identity Provider;
- `acs_urls` — список разрешённых callback URL для Assertion Consumer Service;
- `default_role` — необязательная роль по умолчанию. Если она задана, при запуске входа роль можно не передавать явно;
- `idp_metadata_url` — URL метаданных Identity Provider;
- `idp_sso_url`, `idp_entity_id` и `idp_cert` — ручная альтернатива `idp_metadata_url`;
- `validate_response_signature` и `validate_assertion_signature` — параметры проверки подписей. В production-окружении рекомендуется включать оба параметра, если IdP умеет подписывать и response, и assertion;
- `verbose_logging` — расширенное логирование SAML-обмена. Используйте только для диагностики.

### Настройка через метаданные IdP

```shell
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://stronghold.example.com/v1/auth/saml/callback" \
  idp_metadata_url="https://idp.example.com/app/stronghold/sso/saml/metadata" \
  default_role="employees" \
  validate_response_signature=true \
  validate_assertion_signature=true
```

### Ручная настройка

Если метаданные IdP недоступны, задайте параметры вручную:

```shell
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://stronghold.example.com/v1/auth/saml/callback" \
  idp_sso_url="https://idp.example.com/sso" \
  idp_entity_id="https://idp.example.com/entity" \
  idp_cert=@idp-signing-cert.pem \
  validate_response_signature=true \
  validate_assertion_signature=true
```

Если в `acs_urls` задано несколько адресов, при запуске сценария входа клиент должен явно передать нужный `acs_url`.

## Assertion Consumer Service URLs

Параметр `acs_urls` определяет, на какие адреса Identity Provider может вернуть SAML-ответ после успешной аутентификации пользователя.

При настройке `acs_urls` убедитесь, что каждый адрес:

- совпадает с одним из callback URL, разрешённых в SAML-приложении на стороне Identity Provider;
- указывает на callback-эндпоинт Stronghold вида `.../v1/auth/<mount>/callback` для выбранного пути монтирования;
- использует `https://` в production-окружении.

Если Stronghold доступен по нескольким публичным адресам, можно указать несколько ACS URL:

```shell
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://primary.example.com/v1/auth/saml/callback,https://secondary.example.com/v1/auth/saml/callback" \
  idp_metadata_url="https://idp.example.com/app/stronghold/sso/saml/metadata"
```

Если вы используете `namespaces`, учитывайте путь пространства имён в API URL либо передавайте заголовок `X-Vault-Namespace`, чтобы callback URL соответствовал реальному расположению auth mount.

## Роли

Роли SAML определяют, какие субъекты и атрибуты из SAML-утверждения допускаются к аутентификации, а также какие параметры токена будут применены после входа.

Основные параметры роли:

- `bound_subjects` — ограничивает допустимые SAML subject;
- `bound_subjects_type` — тип сопоставления: `string` или `glob`;
- `bound_attributes` — список обязательных атрибутов assertion и ожидаемых значений;
- `bound_attributes_type` — тип сопоставления значений атрибутов: `string` или `glob`;
- `groups_attribute` — атрибут, из которого Stronghold создаст алиасы групп Identity;
- `alias_metadata` — статические метаданные, которые будут записаны в entity alias;
- параметры токена: `token_policies`, `token_ttl`, `token_max_ttl`, `token_period`, `token_bound_cidrs`.

Пример роли:

```shell
d8 stronghold write auth/saml/role/employees \
  bound_subjects="*@example.com" \
  bound_subjects_type="glob" \
  bound_attributes=department=platform \
  groups_attribute="memberOf" \
  token_policies="default,developers" \
  token_ttl="1h"
```

Если IdP возвращает multivalue-атрибуты, `bound_attributes` может совпасть с любым из ожидаемых значений. Имена атрибутов сопоставляются без учёта регистра.

### Ограничения по атрибутам

После того как Identity Provider аутентифицировал пользователя, Stronghold проверяет ограничения роли по содержимому SAML-утверждения:

- `bound_subjects` сопоставляется с SAML subject;
- `bound_attributes` сопоставляется с обязательными атрибутами assertion и допустимыми значениями.

Это позволяет выдавать доступ только выбранным пользователям или группам из Identity Provider.

Например:

```shell
d8 stronghold write auth/saml/role/support \
  bound_subjects="*@example.com" \
  bound_subjects_type="glob" \
  bound_attributes=groups="support,engineering" \
  token_policies="support-ro"
```

Такая роль разрешит вход только пользователям с subject, оканчивающимся на `@example.com`, и с атрибутом `groups`, содержащим `support` или `engineering`.

Для платформ Microsoft group membership часто приходит в атрибуте `http://schemas.microsoft.com/ws/2008/06/identity/claims/groups`. В этом случае используйте именно это имя атрибута в `bound_attributes` и, при необходимости, в `groups_attribute`.

### Привязка SAML-групп к Identity

Если требуется выдавать политики Stronghold через Identity на основе членства в SAML-группах, задайте `groups_attribute` в роли и создайте соответствующие внешние группы Identity и алиасы.

Общий сценарий выглядит так:

1. Создайте внешнюю группу Identity с нужными политиками.
1. Получите `mount accessor` для SAML auth method.
1. Создайте `group alias`, имя которого совпадает со значением, приходящим в SAML-атрибуте.

Пример команд:

```shell
d8 stronghold write identity/group \
  name="SamlDevelopers" \
  type="external" \
  policies="developers"
```

```shell
d8 stronghold auth list -format=json
```

```shell
d8 stronghold write identity/group-alias \
  name="engineering" \
  mount_accessor="<saml-mount-accessor>" \
  canonical_id="<identity-group-id>"
```

При такой настройке SAML-логин, который возвращает значение `engineering` в атрибуте, указанном в `groups_attribute`, будет связан с внешней группой Identity.

## Вход через веб-интерфейс

Веб-интерфейс Stronghold поддерживает аутентификацию через SAML.

Типовой сценарий выглядит так:

1. Выберите метод входа SAML на форме входа.
1. Укажите имя роли, если для пути монтирования не задан `default_role`.
1. Нажмите `Sign In`.
1. Пройдите аутентификацию у внешнего Identity Provider.
1. После успешного callback Stronghold выдаст токен и выполнит вход.

## Запуск входа через API

Этот сценарий полезен для собственных порталов, обёрток и внешних CLI-инструментов.

### Сгенерируйте verifier и challenge

Параметр `client_challenge` должен быть Base64-кодированным SHA-256-хешем от `client_verifier`.

```shell
verifier="$(uuidgen)"
challenge="$(printf '%s' "$verifier" | openssl dgst -sha256 -binary | base64)"
```

### Запросите SSO URL

```shell
curl \
  --request POST \
  --data "{\"role\":\"employees\",\"client_challenge\":\"$challenge\",\"client_type\":\"browser\"}" \
  https://stronghold.example.com/v1/auth/saml/sso_service_url
```

В ответе Stronghold вернёт:

- `sso_service_url` — URL для перенаправления пользователя к Identity Provider;
- `token_poll_id` — идентификатор для финального обмена на токен.

Если настроено несколько ACS URL, добавьте в запрос параметр `acs_url`.

### Обменяйте временное состояние на токен

После того как IdP вернёт пользователя обратно и Stronghold примет SAML-ответ, обменяйте `token_poll_id` на токен:

```shell
curl \
  --request POST \
  --data "{\"token_poll_id\":\"<poll-id>\",\"client_verifier\":\"$verifier\"}" \
  https://stronghold.example.com/v1/auth/saml/token
```

Токен возвращается в `auth.client_token`.

## Практические замечания

Учитывайте следующие рекомендации:

- В production-окружении используйте только `https://` в ACS URL;
- если IdP умеет подписывать и response, и assertion, включайте обе проверки подписи;
- не оставляйте `verbose_logging=true` в production-окружении, так как в логах могут оказаться чувствительные SAML-данные;
- если у роли задан `token_bound_cidrs`, финальный запрос на `token` должен приходить с разрешённого клиентского адреса;
- в роли SAML должно быть задано хотя бы одно условие допуска: `bound_subjects` или `bound_attributes`;
- если настроено несколько `acs_urls`, клиент должен передавать корректный `acs_url` при запуске сценария входа.
