---
title: "TOTP"
linkTitle: "TOTP"
weight: 20
description: "Настройка TOTP как дополнительного фактора аутентификации в Deckhouse Stronghold."
---

Stronghold поддерживает проверку дополнительного фактора при аутентификации с использованием **Time-Based One-Time Password (TOTP)** — одноразовых короткоживущих кодов.

Проверку TOTP можно настроить:

- для конкретного пользователя;
- для метода аутентификации целиком;
- в том числе в принудительном режиме.

## Когда использовать

TOTP подходит, если:

- требуется добавить второй фактор к уже настроенному методу аутентификации;
- нужно повысить защиту пользовательского входа;
- требуется стандартный сценарий MFA с использованием приложения-аутентификатора и QR-кода.

## Настройка TOTP

Для настройки TOTP выполните следующие действия:

1. Включите метод `TOTP MFA` и получите его идентификатор:

   ```shell
   TOTP_METHOD_ID=$(d8 stronghold write identity/mfa/method/totp \
     -format=json \
     generate=true \
     issuer=MyTOTP \
     period=30 \
     key_size=30 \
     algorithm=SHA256 \
     digits=6 | jq -r '.data.method_id')
   echo "$TOTP_METHOD_ID"
   ```

1. Если нужно включить или пересоздать `TOTP MFA` для конкретного пользователя, укажите идентификатор этого пользователя в параметре `entity_id`:

   ```shell
   ENTITY_ID="f0075fa0-89ca-6235-5b90-b4420134cd36"
   ```

1. Сгенерируйте QR-код для настройки OTP в приложении-аутентификаторе:

   ```shell
   d8 stronghold write -field=barcode \
     /identity/mfa/method/totp/admin-generate \
     method_id="$TOTP_METHOD_ID" entity_id="$ENTITY_ID" \
     | base64 -d > /tmp/qr-code.png
   ```

После этого откройте QR-код и отсканируйте его в приложении, поддерживающем TOTP.

{% alert level="info" %}
Если у пользователя есть доступ к эндпоинту `identity/mfa/method/totp/generate`, он сможет сам получить настройки `TOTP MFA` через веб-интерфейс Stronghold,
используя идентификатор метода.
{% endalert %}

## Включение MFA

В качестве примера ниже показано включение MFA для метода аутентификации `userpass`.

1. Получите accessor метода:

   ```shell
   USERPASS_ACCESSOR=$(d8 stronghold auth list -format=json \
     --detailed | jq -r '."userpass/".accessor')
   echo "$USERPASS_ACCESSOR"
   ```

1. Включите MFA-проверку:

   ```shell
   d8 stronghold write /identity/mfa/login-enforcement/userpass-totp-enforcement \
     mfa_method_ids="$TOTP_METHOD_ID" \
     auth_method_accessors="$USERPASS_ACCESSOR"
   ```

После этого для метода аутентификации `userpass` будет включена проверка второго фактора через TOTP.

1. Выполните вход:

   ```console
   $ d8 stronghold login -method=userpass username=user password='My-Password-1234'
   Initiating Interactive MFA Validation...
   Enter the passphrase for methodID "22c35aa4-bf37-cf31-4187-c5a676c19aca" of type "totp":
   ```

После ввода корректного TOTP-кода пользователь получит токен Stronghold.

## Отключение MFA

Чтобы отключить проверку MFA, выполните следующую команду:

```shell
d8 stronghold delete identity/mfa/login-enforcement/userpass-totp-enforcement
```

## Что важно учитывать

- TOTP — это дополнительный фактор, а не отдельный базовый метод входа;
- сначала должен быть настроен основной метод аутентификации, например `userpass`;
- TOTP можно применять как к отдельным пользователям, так и ко всему методу аутентификации;
- для пользовательского self-service-сценария нужно отдельно предоставить права
  на генерацию настроек TOTP.

## Практические рекомендации

- Используйте отдельный `issuer`, чтобы пользователю было проще различать запись Stronghold
  в приложении-аутентификаторе.
- Храните QR-код и секрет TOTP как чувствительные данные до момента первичной привязки.
- Перед включением принудительного MFA убедитесь, что пользователи успели зарегистрировать
  второй фактор.
- Тестируйте вход с MFA до массового включения enforcement-политики.
