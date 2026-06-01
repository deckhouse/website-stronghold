---
title: "WebAuthn Method"
linkTitle: "WebAuthn"
weight: 85
description: "Passwordless authentication in Deckhouse Stronghold using WebAuthn."
---

The `webauthn` authentication method lets users authenticate to Deckhouse Stronghold using `FIDO2`-compatible authenticators and passkeys.
Support for `WebAuthn` is available in Stronghold starting from version `1.17`.

This method is suitable for passwordless sign-in to the Stronghold web interface and for integration with custom web applications that work with the Stronghold HTTP API.

Before you begin, make sure that the user's browser and device support `WebAuthn`.

## How it works

`WebAuthn` uses a two-step flow:

1. Stronghold returns registration or login parameters to the browser.
1. The browser communicates with the user's authenticator.
1. The verification result is sent back to Stronghold.
1. Stronghold issues a token.

The following endpoints are used for login: `auth/<mount>/login/begin` and `auth/<mount>/login/finish`.
The following endpoints are used for the initial passkey registration: `auth/<mount>/register/begin` and `auth/<mount>/register/finish`.

## Configuration

Enable and configure the method before using it.

Main parameters:

- `rp_id` — the `Relying Party` identifier.
  This is usually the Stronghold DNS name.
  The host in `rp_id` must match the host of the origin from which authentication is performed.
- `rp_display_name` — the display name of the service shown to the browser and authenticator.
  If this parameter is not set, Stronghold uses the value of `rp_id`.
- `rp_origins` — the list of allowed origins from which the browser can perform `WebAuthn` operations.
- `auto_registration` — if `true` by default, the user can register a passkey independently.
  If `false`, create the user in advance using the `user/` path.

### Enable the method

```shell
d8 stronghold auth enable webauthn
```

By default, the method is available at `auth/webauthn`.
If necessary, you can use a different mount path:

```shell
d8 stronghold auth enable -path=my-passkeys webauthn
```

### Configure the `Relying Party`

```shell
d8 stronghold write auth/webauthn/config \
  rp_id="stronghold.example.com" \
  rp_display_name="Deckhouse Stronghold" \
  rp_origins="https://stronghold.example.com"
```

An example configuration with self-registration disabled:

```shell
d8 stronghold write auth/webauthn/config \
  rp_id="stronghold.example.com" \
  rp_display_name="Deckhouse Stronghold" \
  rp_origins="https://stronghold.example.com" \
  auto_registration=false
```

### Pre-create a user

If `auto_registration=false`, an administrator must create the user in advance and assign future token parameters:

```shell
d8 stronghold write auth/webauthn/user/alice \
  display_name="Alice Doe" \
  token_policies="developers" \
  token_ttl="1h"
```

Using the `auth/webauthn/user/<name>` path, you can:

- pre-create a user for registration .
- update `display_name` .
- assign token parameters such as `token_policies`, `token_ttl`, `token_max_ttl`, and `token_period` .
- view user metadata and the number of registered `WebAuthn` credentials .
- delete a user together with all registered passkeys.

## Sign in through the UI

The Stronghold web interface supports `WebAuthn` directly.

### First passkey registration

To register a passkey for the first time, follow these steps:

1. Select the `WebAuthn` login method.
1. Disable the `Use passkey picker` option so that the `Username` field appears in the form.
1. Enter the username and click `Register`.
1. Confirm the operation on the device or in the passkey manager.

### Subsequent sign-ins

Two modes are available for subsequent sign-ins:

- with the `Use passkey picker` option enabled, the browser shows a list of saved discoverable credentials, and the username is optional .
- with the option disabled, the user explicitly enters `Username`, after which the browser offers a suitable passkey for that user.

## Registration and login through the API

The following is a general flow for integrating with a custom web application.

### Register a passkey

1. Request registration parameters:

```shell
curl \
  --request POST \
  --data '{"username":"alice"}' \
  https://stronghold.example.com/v1/auth/webauthn/register/begin
```

1. Pass the returned `publicKey` parameters to `navigator.credentials.create(...)` in the browser.
1. Complete the registration by sending the result back to Stronghold:

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

Send the request to `POST /v1/auth/webauthn/register/finish`.

### Log in

1. Request login parameters:

```shell
curl \
  --request POST \
  --data '{"username":"alice"}' \
  https://stronghold.example.com/v1/auth/webauthn/login/begin
```

For login through the passkey picker, you can omit `username`.

1. Pass the returned parameters to `navigator.credentials.get(...)`.
1. Complete the login by sending the authenticator response to `POST /v1/auth/webauthn/login/finish`:

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

Stronghold returns a token in the `auth.client_token` field.

## Practical notes

- Use `rp_origins` values that exactly match the actual addresses of the Stronghold web interface or your application.
- For production environments, it is convenient to disable `auto_registration` and pre-create users with the required policies.
- `WebAuthn` is especially useful for passwordless sign-in to the UI.
- For machine-to-machine scenarios, `AppRole`, `JWT`, or `Kubernetes` are usually a better fit.
