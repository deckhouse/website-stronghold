---
title: "WebAuthn"
linkTitle: "WebAuthn"
weight: 85
description: "Безпарольная аутентификация в Deckhouse Stronghold с помощью WebAuthn."
---

## WebAuthn

Метод `webauthn` позволяет аутентифицироваться в **Deckhouse Stronghold** с помощью `FIDO2`-совместимых аутентификаторов и `passkeys`. Поддержка `WebAuthn` появилась в Stronghold, начиная с версии `1.17`.

Этот способ подходит:

- для безпарольного входа в веб-интерфейс Stronghold;
- для интеграции с собственными веб-приложениями, работающими с HTTP API Stronghold.

Перед началом работы убедитесь, что браузер и устройство пользователя поддерживают `WebAuthn`.

## Как это работает

`WebAuthn` использует двухшаговый сценарий:

1. Stronghold выдаёт браузеру параметры для регистрации или входа.
2. Браузер обращается к аутентификатору пользователя.
3. Результат проверки отправляется обратно в Stronghold.
4. Stronghold выдаёт токен.

Для входа используются endpoint'ы:

- `auth/<mount>/login/begin`
- `auth/<mount>/login/finish`

Для первичной привязки `passkey` используются:

- `auth/<mount>/register/begin`
- `auth/<mount>/register/finish`

## Конфигурация

Метод нужно предварительно включить и настроить.

Основные параметры:

- `rp_id` — идентификатор `Relying Party`;
- `rp_display_name` — отображаемое имя сервиса;
- `rp_origins` — список разрешённых origin;
- `auto_registration` — разрешена ли самостоятельная регистрация `passkey`.

### Включение метода

```bash
d8 stronghold auth enable webauthn
```

По умолчанию метод будет доступен по пути `auth/webauthn`. При необходимости можно использовать другой путь:

```bash
d8 stronghold auth enable -path=my-passkeys webauthn
```

### Настройка Relying Party

```bash
d8 stronghold write auth/webauthn/config \
  rp_id="stronghold.example.com" \
  rp_display_name="Deckhouse Stronghold" \
  rp_origins="https://stronghold.example.com"
```

Пример с отключённой самостоятельной регистрацией:

```bash
d8 stronghold write auth/webauthn/config \
  rp_id="stronghold.example.com" \
  rp_display_name="Deckhouse Stronghold" \
  rp_origins="https://stronghold.example.com" \
  auto_registration=false
```

## Предварительное создание пользователя

Если `auto_registration=false`, администраторам нужно заранее создать пользователя и назначить ему параметры токена.

Пример:

```bash
d8 stronghold write auth/webauthn/user/alice \
  display_name="Alice Doe" \
  token_policies="developers" \
  token_ttl="1h"
```

Через путь `auth/webauthn/user/<name>` можно:

- заранее создать пользователя;
- изменить `display_name`;
- назначить параметры токена;
- посмотреть метаданные пользователя;
- удалить пользователя вместе с зарегистрированными `passkeys`.

## Вход через UI

Stronghold UI поддерживает `WebAuthn` напрямую.

### Первая регистрация passkey

1. Выберите метод входа `WebAuthn`.
2. Отключите опцию `Use passkey picker`, чтобы появилось поле `Username`.
3. Введите имя пользователя и нажмите `Register`.
4. Подтвердите операцию на устройстве или в менеджере `passkeys`.

### Последующие входы

Доступны два режима:

- с включённой опцией `Use passkey picker` браузер покажет список обнаруживаемых учётных данных;
- с отключённой опцией пользователь явно вводит `Username`, после чего браузер предлагает соответствующий `passkey`.

## Регистрация и вход через API

Этот сценарий полезен для собственных веб-приложений.

### Регистрация passkey

1. Запросите параметры регистрации:

```bash
curl \
  --request POST \
  --data '{"username":"alice"}' \
  https://stronghold.example.com/v1/auth/webauthn/register/begin
```

1. Передайте полученные `publicKey`-параметры в `navigator.credentials.create(...)`.
2. Завершите регистрацию, отправив результат обратно в Stronghold на `POST /v1/auth/webauthn/register/finish`.

### Вход

1. Запросите параметры входа:

```bash
curl \
  --request POST \
  --data '{"username":"alice"}' \
  https://stronghold.example.com/v1/auth/webauthn/login/begin
```

Для входа через `passkey picker` можно не передавать `username`.

1. Передайте полученные параметры в `navigator.credentials.get(...)`.
2. Отправьте результат аутентификатора на `POST /v1/auth/webauthn/login/finish`.

В ответе Stronghold возвращает токен в поле `auth.client_token`.

## Практические замечания

- Используйте `rp_origins`, которые точно соответствуют реальным адресам Stronghold UI или вашего приложения.
- Для production удобно отключать `auto_registration` и заранее создавать пользователей с нужными политиками.
- `WebAuthn` особенно полезен для безпарольного входа в UI.
- Для машинных сценариев обычно лучше подходят `AppRole`, `JWT` или `Kubernetes`.
