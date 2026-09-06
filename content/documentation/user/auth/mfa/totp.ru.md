---
title: "TOTP"
weight: 20
---

## Настройка TOTP

Stronghold поддерживает проверку дополнительного фактора при аутентификации с использованием
Time-Based One-Time Password (TOTP) — одноразовых короткоживущих кодов.
Проверка TOTP может быть установлена как для конкретного пользователя, так и для метода
аутентификации целиком, в том числе принудительно.

Для настройки TOTP выполните следующие шаги:

1. Включите метод MFA TOTP и получите его идентификатор:

   {{< tabs name="stronghold_cmd_49406" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
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
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   TOTP_METHOD_ID=$(stronghold write identity/mfa/method/totp \
       -format=json \
       generate=true \
       issuer=MyTOTP \
       period=30 \
       key_size=30 \
       algorithm=SHA256 \
       digits=6 | jq -r '.data.method_id')
   echo $TOTP_METHOD_ID
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Если необходимо включить (или пересоздать) TOTP MFA для конкретного пользователя, укажите его идентификатор:

   ```shell
   ENTITY_ID="f0075fa0-89ca-6235-5b90-b4420134cd36"
   ```

1. Сгенерируйте QR-код для настройки OTP:

   {{< tabs name="stronghold_cmd_64039" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold write -field=barcode \
       /identity/mfa/method/totp/admin-generate \
       method_id=$TOTP_METHOD_ID entity_id=$ENTITY_ID \
       | base64 -d > /tmp/qr-code.png
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold write -field=barcode \
       /identity/mfa/method/totp/admin-generate \
       method_id=$TOTP_METHOD_ID entity_id=$ENTITY_ID \
       | base64 -d > /tmp/qr-code.png
   ```
   {{% /tab %}}
   {{< /tabs >}}

Если у пользователя есть доступ к эндпоинту `identity/mfa/method/totp/generate`,
тогда пользователь сам сможет получить настройки TOTP MFA через веб-интерфейс Stronghold,
используя указанный выше идентификатор.

## Включение MFA

В качестве примера разберём проверку MFA для метода аутентификации Userpass.

1. Получите идентификатор метода:

   {{< tabs name="stronghold_cmd_18206" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   LDAP_ACCESSOR=$(d8 stronghold auth list -format=json \
       --detailed | jq -r '."userpass/".accessor')
   echo $LDAP_ACCESSOR
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   LDAP_ACCESSOR=$(stronghold auth list -format=json \
       --detailed | jq -r '."userpass/".accessor')
   echo $LDAP_ACCESSOR
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Включите MFA:

   {{< tabs name="stronghold_cmd_35603" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold write /identity/mfa/login-enforcement/userpass-totp-enforcement \
       mfa_method_ids="$TOTP_METHOD_ID" \
       auth_method_accessors=$LDAP_ACCESSOR
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold write /identity/mfa/login-enforcement/userpass-totp-enforcement \
       mfa_method_ids="$TOTP_METHOD_ID" \
       auth_method_accessors=$LDAP_ACCESSOR
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Выполните вход:

   {{< tabs name="stronghold_cmd_84099" >}}
   {{% tab name="Stronghold в DKP" %}}
   ```shell
   d8 stronghold login -method=userpass username=user password='My-Password-1234'
   Initiating Interactive MFA Validation...
   Enter the passphrase for methodID "22c35aa4-bf37-cf31-4187-c5a676c19aca" of type "totp":
   ```
   {{% /tab %}}
   {{% tab name="Stronghold в Linux" %}}
   ```shell
   stronghold login -method=userpass username=user password='My-Password-1234'
   Initiating Interactive MFA Validation...
   Enter the passphrase for methodID "22c35aa4-bf37-cf31-4187-c5a676c19aca" of type "totp":
   ```
   {{% /tab %}}
   {{< /tabs >}}

Чтобы отключить проверку MFA, выполните:

{{< tabs name="stronghold_cmd_42432" >}}
{{% tab name="Stronghold в DKP" %}}
```shell
d8 stronghold delete identity/mfa/login-enforcement/userpass-totp-enforcement
```
{{% /tab %}}
{{% tab name="Stronghold в Linux" %}}
```shell
stronghold delete identity/mfa/login-enforcement/userpass-totp-enforcement
```
{{% /tab %}}
{{< /tabs >}}
