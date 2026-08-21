---
title: "Multifactor"
description: "Configure Stronghold multifactor authentication with the Multifactor service"
weight: 10
---

## Configuring Multifactor

Stronghold supports an additional authentication factor through the [Multifactor](https://multifactor.ru/docs/intro/) service.
After a successful primary login, Stronghold sends a request to Multifactor and waits for the user to confirm it.

Confirmation uses one of the methods configured for the user in Multifactor:

- Push notification in the Multifactor mobile app;
- Telegram;
- Phone call — answer the call and press `#`.

{{< alert level="info" >}}
Stronghold does not accept one-time passwords (OTP) for Multifactor MFA.
The request is treated as an access confirmation: the user approves or denies it in the selected channel.
{{< /alert >}}

### Supported authentication methods

Login MFA, including Multifactor, does not restrict authentication method types in code.
You can enable the check for any method that has an entity at login, for example (but not limited to):

- `userpass`
- `ldap`
- `oidc`
- `jwt`
- `saml`
- `approle`
- `kubernetes`

Multifactor MFA is designed for interactive confirmation.
For machine methods such as `approle` and `kubernetes`, an automated client cannot confirm a push, Telegram message,
or phone call on its own.

For such cases, you can use a shared confirmation scenario: in `username_format`, set a static Multifactor
account name, for example an administrator account.
Then all authentication requests go to that account,
and an operator confirms the login using the machine identifier in the request.

### Prerequisites

Before configuration, complete the following steps:

1. In the Multifactor personal account, create a resource for server authentication, for example `Linux`.
   Do not use the `WebSite` resource type.

1. From the resource settings, copy the **NAS Identifier** and **Shared Secret** values.

1. Make sure users exist in Multifactor or that automatic registration is enabled.
   Stronghold does not synchronize users with Multifactor, but it sends the request with inline registration support.
   If the Multifactor resource setting **Register new users** is enabled,
   Multifactor creates an account on the first successful connection and records this in the access log.
   Access is granted according to Multifactor policies.
   If the option is disabled, Multifactor denies access and does not register new accounts.

   The identity value in Multifactor must match the value produced by the `username_format` parameter.

You can also create users manually or through the Multifactor API, or synchronize them from AD/LDAP using Directory Sync in Multifactor.

### Creating an MFA method

To enable the Multifactor MFA method and obtain its ID, run:

```shell
d8 stronghold write identity/mfa/method/multifactor method_name=my-mfa nas_identifier="rs_nas_id" shared_secret="secret"

Key          Value
---          -----
method_id    93964fd0-dd7e-e22a-74d0-0880ca5e0398

```

Method parameters:

| Parameter | Description |
| --- | --- |
| `method_name` | Unique MFA method name. |
| `nas_identifier` | NAS Identifier from the Multifactor resource settings. |
| `shared_secret` | Shared Secret from the Multifactor resource settings. |
| `api_url` | Base URL of the Multifactor API. Default is `https://api.multifactor.ru`. |
| `username_format` | Template that maps a Stronghold entity to a Multifactor identity. If unset, the entity name is used. |
| `timeout_seconds` | Maximum wait time for confirmation in seconds. Default is `90`, minimum is `65`. |

{{< alert level="warning" >}}
The `shared_secret` parameter is written only when the method is created or updated.
When you read the method configuration, the secret value is not returned.
{{< /alert >}}

### Enabling MFA

The following example enables MFA verification for the Userpass authentication method.

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

After the primary factor is verified, Stronghold initiates a request to Multifactor and waits for confirmation.
Confirm the request in the Multifactor mobile app, Telegram, or by phone call — depending on the user settings in Multifactor.
After the `Granted` status, Stronghold issues a token.

To disable MFA verification, run:

```shell
d8 stronghold delete identity/mfa/login-enforcement/userpass-multifactor-enforcement
```
