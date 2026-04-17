---
title: "TOTP"
linkTitle: "TOTP"
weight: 20
description: "Настройка TOTP как дополнительного фактора аутентификации в Deckhouse Stronghold."
---

Stronghold поддерживает проверку дополнительного фактора при аутентификации с использованием **Time-Based One-Time Password (TOTP)** — одноразовых короткоживущих кодов.

Проверка `TOTP` может быть настроена:

- для конкретного пользователя;
- для метода аутентификации целиком;
- в том числе в принудительном режиме.

## Когда использовать

`TOTP` подходит, если:

- требуется добавить второй фактор к уже настроенному методу аутентификации;
- нужно повысить защиту пользовательского входа;
- требуется стандартный сценарий MFA с использованием приложения-аутентификатора и QR-кода.

## Настройка TOTP

Для настройки TOTP выполните следующие действия.

### Шаг 1. Включите метод MFA TOTP

Включите метод `TOTP MFA` и получите его идентификатор:

```bash
TOTP_METHOD_ID=$(d8 stronghold write identity/mfa/method/totp \
    -format=json \
    generate=true \
    issuer=MyTOTP \
    period=30 \
    key_size=30 \
    algorithm=SHA256 \
    digits=6 | jq -r '.data.method_id')
echo $TOTP_METHOD_ID
```

Этот идентификатор потребуется на следующих шагах.

### Шаг 2. Укажите идентификатор пользователя

Если нужно включить или пересоздать `TOTP MFA` для конкретного пользователя, укажите `entity_id` этого пользователя:

```bash
ENTITY_ID="f0075fa0-89ca-6235-5b90-b4420134cd36"
```

### Шаг 3. Сгенерируйте QR-код

Сгенерируйте QR-код для настройки OTP в приложении-аутентификаторе:

```bash
d8 stronghold write -field=barcode \
    /identity/mfa/method/totp/admin-generate \
    method_id=$TOTP_METHOD_ID entity_id=$ENTITY_ID \
    | base64 -d > /tmp/qr-code.png
```

После этого QR-код можно открыть и отсканировать в приложении, поддерживающем `TOTP`.

> Примечание  
> Если у пользователя есть доступ к endpoint'у `identity/mfa/method/totp/generate`, пользователь сможет сам получить настройки `TOTP MFA` через веб-интерфейс Stronghold, используя идентификатор метода.

## Включение MFA

В качестве примера ниже показано включение `MFA` для метода аутентификации `Userpass`.

### Шаг 1. Получите accessor метода

```bash
LDAP_ACCESSOR=$(d8 stronghold auth list -format=json \
    --detailed | jq -r '."userpass/".accessor')
echo $LDAP_ACCESSOR
```

### Шаг 2. Включите MFA-проверку

```bash
d8 stronghold write /identity/mfa/login-enforcement/userpass-totp-enforcement \
    mfa_method_ids="$TOTP_METHOD_ID" \
    auth_method_accessors=$LDAP_ACCESSOR
```

После этого для метода `userpass` будет включена проверка второго фактора через TOTP.

### Шаг 3. Выполните вход

```bash
d8 stronghold login -method=userpass username=user password='My-Password-1234'
Initiating Interactive MFA Validation...
Enter the passphrase for methodID "22c35aa4-bf37-cf31-4187-c5a676c19aca" of type "totp":
```

После ввода корректного TOTP-кода пользователь получит токен Stronghold.

## Отключение MFA

Чтобы отключить проверку MFA, выполните:

```bash
d8 stronghold delete identity/mfa/login-enforcement/userpass-totp-enforcement
```

## Что важно учитывать

- `TOTP` — это дополнительный фактор, а не отдельный базовый метод входа;
- сначала должен быть настроен основной метод аутентификации, например `Userpass`;
- `TOTP` можно применять как к отдельным пользователям, так и ко всему методу аутентификации;
- для пользовательского self-service-сценария нужно отдельно предоставить права на генерацию настроек `TOTP`.

## Практические рекомендации

- Используйте отдельный `issuer`, чтобы пользователю было проще различать запись Stronghold в приложении-аутентификаторе.
- Храните QR-код и секрет TOTP как чувствительные данные до момента первичной привязки.
- Перед включением принудительного MFA убедитесь, что пользователи успели зарегистрировать второй фактор.
- Тестируйте вход с MFA до массового включения enforcement-политики.
