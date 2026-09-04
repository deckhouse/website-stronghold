---
title: "Multifactor"
description: "Configuring Stronghold multifactor authentication with the Multifactor service"
weight: 10
---

Stronghold supports verification of an additional authentication factor through the [Multifactor](https://multifactor.pro/docs/intro/) service.
After a successful authentication via a primary method, Stronghold sends a request to Multifactor and waits for the user to confirm it.

Depending on the user settings in Multifactor, a request can be confirmed in one of the following ways:

- Using a push notification in the Multifactor mobile app.
- In Telegram.
- Via a phone call. To confirm a request, answer a call and press `#`.

{{< alert level="info" >}}
Stronghold does not accept one-time passwords (OTP) for Multifactor MFA.

The user approves or denies an access request in the selected channel.
{{< /alert >}}

## Supported authentication methods

MFA via Multifactor can be enabled for any authentication method that creates a Stronghold entity on use. For example:

- `userpass`
- `ldap`
- `oidc`
- `jwt`
- `saml`
- `webauthn`
- `approle`
- `kubernetes`

Multifactor MFA is designed for interactive confirmation.
For machine methods such as `approle` and `kubernetes`, an automated client cannot confirm a request via a push notification, Telegram message, or phone call on its own.

For such cases, you can configure a request confirmation via a shared Multifactor account.
To do this, in the `username_format` parameter, set a static account name, for example, of an administrator.
In this case, all authentication requests will be directed to that account's owner, who can confirm them using the machine identifier specified in the request.

## Prerequisites

Before configuring Multifactor MFA, complete the following steps:

1. In the Multifactor personal account, create a resource for server authentication, for example, Linux.
   Do not use the WebSite resource type.

1. Copy the "NAS Identifier" and "Shared Secret" values from the created resource settings.

1. Make sure user accounts exist in Multifactor or an automatic registration of these accounts is enabled.

   Stronghold does not synchronize users with Multifactor, but it allows registering them on initial request.
   If the Multifactor resource setting "Register new users" is enabled,
   Multifactor creates an account on the first successful connection and records this information to the access log.
   Access is granted according to Multifactor policies.
   If the automatic registration is disabled, Multifactor denies access and does not register new accounts.

   The identity value in Multifactor must match the value produced by the `username_format` parameter.

You can also create users manually, through the Multifactor API, or synchronize them from AD/LDAP using Directory Sync in Multifactor.

## Creating an MFA method

To create a Multifactor MFA method and obtain its ID, run the following command:

```shell
d8 stronghold write identity/mfa/method/multifactor \
  method_name=my-mfa \
  nas_identifier="rs_nas_id" \
  shared_secret="secret"
```

Example output:

```console
Key          Value
---          -----
method_id    93964fd0-dd7e-e22a-74d0-0880ca5e0398
```

Method parameters:

| Parameter | Description |
| --- | --- |
| `method_name` | Unique MFA method name |
| `nas_identifier` | NAS Identifier from the Multifactor resource settings |
| `shared_secret` | Shared Secret from the Multifactor resource settings |
| `api_url` | Base URL of the Multifactor API. The default is `https://api.multifactor.ru` |
| `username_format` | Template that maps a Stronghold entity to a Multifactor identity. Available parameters are listed in ["Templated policies"](../../../concepts/policy/#templated-policies). If the template is not set, the entity name is used |
| `timeout_seconds` | Maximum wait time for confirmation in seconds. The default value is `90`, the minimum value is `65` |

{{< alert level="warning" >}}
The `shared_secret` parameter is written only when the method is created or updated and is not returned when the configuration is read.
{{< /alert >}}

## Enabling MFA

The following is an MFA configuration example for the Userpass authentication method.

1. Get the authentication method accessor:

   ```shell
   USERPASS_ACCESSOR=$(d8 stronghold auth list -format=json \
       --detailed | jq -r '."userpass/".accessor')
   echo $USERPASS_ACCESSOR
   ```

1. Enable MFA:

   ```shell
   d8 stronghold write /identity/mfa/login-enforcement/userpass-multifactor-enforcement \
       mfa_method_ids="93964fd0-dd7e-e22a-74d0-0880ca5e0398" \
       auth_method_accessors=$USERPASS_ACCESSOR
   ```

1. Log in:

   ```shell
   d8 stronghold login -method=userpass username=mfa-user
   Password (will be hidden):
   ```

   Example output:

   ```console
   Initiating Interactive MFA Validation...
   Asking Stronghold to perform MFA validation with upstream service. You should receive a push notification in your authenticator app shortly
   Success! You are now authenticated. The token information displayed below is
   already stored in the token helper. You do NOT need to run "stronghold login"
   again. Future Stronghold requests will automatically use this token.

   Key                    Value
   ---                    -----
   token                  hv.....
   token_accessor         uH1voyZljOCbttJSICUtom17
   token_duration         768h
   token_renewable        true
   token_policies         ["default"]
   policies               ["default"]
   token_meta_username    mfa-user
   ```

After the primary factor is verified, Stronghold sends a request to Multifactor and waits for confirmation.
Confirm the request in the Multifactor mobile app, Telegram, or by a phone call, depending on the user settings.
After the "Granted" status is obtained, Stronghold issues a token.

To disable the MFA verification for the Userpass method, run the following command:

```shell
d8 stronghold delete identity/mfa/login-enforcement/userpass-multifactor-enforcement
```
