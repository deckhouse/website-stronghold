---
title: "TOTP"
linkTitle: "TOTP"
weight: 20
description: "Configure TOTP as an additional authentication factor in Deckhouse Stronghold."
---

Stronghold supports validating an additional factor during authentication by using **Time-Based One-Time Password (TOTP)** — short-lived one-time codes. You can configure TOTP validation:

- for a specific user;
- for an entire authentication method;
- including in enforced mode.

## When to use

TOTP is suitable if:

- you need to add a second factor to an existing authentication method;
- you need to improve the protection of user sign-in;
- you need a standard MFA scenario that uses an authenticator application and a QR code.

## Configure TOTP

To configure TOTP, follow these steps:

1. Enable the `TOTP MFA` method and get its ID:

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

1. If you need to enable or recreate `TOTP MFA` for a specific user, specify that user's ID in the `entity_id` parameter:

   ```shell
   ENTITY_ID="f0075fa0-89ca-6235-5b90-b4420134cd36"
   ```

1. Generate a QR code to configure OTP in the authenticator application:

   ```shell
   d8 stronghold write -field=barcode \
     /identity/mfa/method/totp/admin-generate \
     method_id="$TOTP_METHOD_ID" entity_id="$ENTITY_ID" \
     | base64 -d > /tmp/qr-code.png
   ```

After that, open the QR code and scan it in an application that supports TOTP.

{% alert level="info" %}
If a user has access to the `identity/mfa/method/totp/generate` endpoint, they can get the `TOTP MFA` settings themselves through the Stronghold web interface by using the method ID.
{% endalert %}

## Enable MFA

The following example shows how to enable MFA for the `userpass` authentication method.

1. Get the method accessor:

   ```shell
   USERPASS_ACCESSOR=$(d8 stronghold auth list -format=json \
     --detailed | jq -r '."userpass/".accessor')
   echo "$USERPASS_ACCESSOR"
   ```

1. Enable MFA validation:

   ```shell
   d8 stronghold write /identity/mfa/login-enforcement/userpass-totp-enforcement \
     mfa_method_ids="$TOTP_METHOD_ID" \
     auth_method_accessors="$USERPASS_ACCESSOR"
   ```

After that, second-factor validation through TOTP will be enabled for the `userpass` authentication method.

1. Log in:

   ```console
   $ d8 stronghold login -method=userpass username=user password='My-Password-1234'
   Initiating Interactive MFA Validation...
   Enter the passphrase for methodID "22c35aa4-bf37-cf31-4187-c5a676c19aca" of type "totp":
   ```

After the user enters a valid TOTP code, they will receive a Stronghold token.

## Disable MFA

To disable MFA validation, run the following command:

```shell
d8 stronghold delete identity/mfa/login-enforcement/userpass-totp-enforcement
```

## Important considerations

- TOTP is an additional factor, not a separate primary sign-in method.
- A primary authentication method, such as `userpass`, must be configured first.
- You can apply TOTP both to individual users and to the entire authentication method.
- For a user self-service scenario, you must grant permissions for TOTP settings generation separately.

## Best practices

- Use a separate `issuer` so that users can more easily distinguish the Stronghold entry in the authenticator application.
- Store the QR code and the TOTP secret as sensitive data until the initial binding is complete.
- Before enabling enforced MFA, make sure that users have had time to register the second factor.
- Test MFA sign-in before enabling the enforcement policy for all users.
