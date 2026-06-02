---
title: "Метод WebAuthn"
linkTitle: "WebAuthn"
weight: 85
description: "Беспарольная аутентификация в Deckhouse Stronghold с помощью WebAuthn."
---

Метод аутентификации `webauthn` позволяет аутентифицироваться в Deckhouse Stronghold с помощью `FIDO2`-совместимых аутентификаторов и passkey.
Поддержка `WebAuthn` появилась в Stronghold, начиная с версии `1.17`.

Этот способ подходит для безпарольного входа в веб-интерфейс Stronghold и для интеграции с собственными веб-приложениями, которые работают с HTTP API Stronghold.

Перед началом работы убедитесь, что браузер и устройство пользователя поддерживают `WebAuthn`.

## Как это работает

`WebAuthn` использует двухшаговый сценарий:

1. Stronghold выдаёт браузеру параметры для регистрации или входа.
1. Браузер обращается к аутентификатору пользователя.
1. Результат проверки отправляется обратно в Stronghold.
1. Stronghold выдаёт токен.

Для входа используются эндпоинты `auth/<mount>/login/begin` и `auth/<mount>/login/finish`.
Для первичной привязки passkey используются эндпоинты `auth/<mount>/register/begin` и `auth/<mount>/register/finish`.

## Конфигурация

Метод нужно предварительно включить и настроить.

Основные параметры:

- `rp_id` — идентификатор `Relying Party`.
  Обычно это DNS-имя Stronghold.
  Хост в `rp_id` должен совпадать с хостом origin, из которого выполняется аутентификация.
- `rp_display_name` — отображаемое имя сервиса для браузера и аутентификатора.
  Если параметр не задан, Stronghold использует значение `rp_id`.
- `rp_origins` — список разрешённых origin, с которых браузер может выполнять операции `WebAuthn`.
- `auto_registration` — если `true` по умолчанию, пользователь может самостоятельно зарегистрировать passkey.
  Если `false`, пользователя нужно заранее создать через путь `user/`.

### Включение метода

```shell
d8 stronghold auth enable webauthn
```

По умолчанию метод будет доступен по пути `auth/webauthn`.
При необходимости можно использовать другой путь монтирования:

```shell
d8 stronghold auth enable -path=my-passkeys webauthn
```

### Настройка `Relying Party`

```shell
d8 stronghold write auth/webauthn/config \
  rp_id="stronghold.example.com" \
  rp_display_name="Deckhouse Stronghold" \
  rp_origins="https://stronghold.example.com"
```

Пример конфигурации, в которой самостоятельная регистрация отключена:

```shell
d8 stronghold write auth/webauthn/config \
  rp_id="stronghold.example.com" \
  rp_display_name="Deckhouse Stronghold" \
  rp_origins="https://stronghold.example.com" \
  auto_registration=false
```

### Предварительное создание пользователя

Если `auto_registration=false`, администратор должен заранее создать пользователя и назначить ему параметры будущего токена:

```shell
d8 stronghold write auth/webauthn/user/alice \
  display_name="Alice Doe" \
  token_policies="developers" \
  token_ttl="1h"
```

Через путь `auth/webauthn/user/<name>` можно:

- заранее создать пользователя для регистрации;
- изменить `display_name`;
- назначить параметры токена, например `token_policies`, `token_ttl`, `token_max_ttl` и `token_period`;
- посмотреть метаданные пользователя и количество привязанных учётных данных `WebAuthn`;
- удалить пользователя вместе с его зарегистрированными passkey.

## Вход через UI

Веб-интерфейс Stronghold поддерживает `WebAuthn` напрямую.

### Первая регистрация passkey

Для первой регистрации passkey выполните следующие шаги:

1. Выберите метод входа `WebAuthn`.
1. Отключите опцию `Use passkey picker`, чтобы в форме появилось поле `Username`.
1. Введите имя пользователя и нажмите `Register`.
1. Подтвердите операцию на устройстве или в менеджере passkey.

### Последующие входы

Для последующих входов доступны два режима:

- с включённой опцией `Use passkey picker` браузер покажет список сохранённых обнаруживаемых учётных данных, и имя пользователя можно не вводить;
- с отключённой опцией пользователь явно вводит `Username`, после чего браузер предлагает подходящий passkey для этого пользователя.

## Регистрация и вход через API

Ниже показан общий сценарий для интеграции с собственным веб-приложением.

### Регистрация passkey

1. Запросите параметры регистрации:

```shell
curl \
  --request POST \
  --data '{"username":"alice"}' \
  https://stronghold.example.com/v1/auth/webauthn/register/begin
```

1. Передайте полученные `publicKey`-параметры в `navigator.учётные данные.create(...)` в браузере.
1. Завершите регистрацию, отправив результат обратно в Stronghold:

```json
{
  "username": "alice",
  "credential": {
    "id": "...",
    "rawId": "...",
    "type": "public-key",
    "response": {
      "clientDataJSON": "...",
      "attestationObject": "..."
    }
  }
}
```

Запрос отправляется на `POST /v1/auth/webauthn/register/finish`.

### Вход

1. Запросите параметры входа:

```shell
curl \
  --request POST \
  --data '{"username":"alice"}' \
  https://stronghold.example.com/v1/auth/webauthn/login/begin
```

Для входа через passkey picker можно не передавать `username`.

1. Передайте полученные параметры в `navigator.учётные данные.get(...)`.
1. Завершите вход, отправив ответ аутентификатора на `POST /v1/auth/webauthn/login/finish`:

```json
{
  "username": "alice",
  "credential": {
    "id": "...",
    "rawId": "...",
    "type": "public-key",
    "response": {
      "clientDataJSON": "...",
      "authenticatorData": "...",
      "signature": "...",
      "userHandle": "..."
    }
  }
}
```

В ответе Stronghold возвращает токен в поле `auth.client_token`.

## Практические замечания

- Используйте `rp_origins`, которые точно соответствуют реальным адресам Stronghold UI или вашего приложения.
- Для production-окружения удобно отключать `auto_registration` и предварительно создавать пользователей с нужными политиками.
- `WebAuthn` особенно полезен для безпарольного входа в UI.
- Для машинных сценариев обычно лучше подходят `AppRole`, `JWT` или `Kubernetes`.
