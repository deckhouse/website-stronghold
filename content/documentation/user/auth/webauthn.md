---
title: "WebAuthn method"
linkTitle: "WebAuthn"
weight: 85
---

## WebAuthn auth method

The `webauthn` auth method lets users authenticate to Deckhouse Stronghold with `FIDO2` authenticators and `passkeys`. `WebAuthn` support is available in Stronghold starting from version `1.17`.

This method is intended for passwordless sign-in to the Stronghold web UI and for custom web applications that use the Stronghold HTTP API. Make sure the user's browser and device support `WebAuthn` before enabling it for production use.

## How it works

`WebAuthn` uses a two-step flow:

1. Stronghold returns registration or login options to the browser.
1. The browser talks to the user's authenticator.
1. The browser sends the result back to Stronghold, and Stronghold issues a token.

Login uses the `auth/<mount>/login/begin` and `auth/<mount>/login/finish` endpoints. Initial passkey enrollment uses `auth/<mount>/register/begin` and `auth/<mount>/register/finish`.

## Configuration

Enable and configure the method before users can authenticate. The main parameters are:

- `rp_id` is the `Relying Party` identifier. This is typically the Stronghold DNS name. The host in `rp_id` must match the host of the origin used for authentication.
- `rp_display_name` is the human-readable service name shown by the browser and authenticator. If omitted, Stronghold uses `rp_id`.
- `rp_origins` is the list of allowed origins that can perform `WebAuthn` operations.
- `auto_registration` controls whether users can self-register passkeys. If `true` (default), users can enroll themselves. If `false`, users must be pre-created via the `user/` path.

### Enable the method

```shell
d8 stronghold auth enable webauthn
```

By default the method is mounted at `auth/webauthn`. You can mount it at a custom path if needed:

```shell
d8 stronghold auth enable -path=my-passkeys webauthn
```

### Configure the Relying Party

```shell
d8 stronghold write auth/webauthn/config \
  rp_id="stronghold.example.com" \
  rp_display_name="Deckhouse Stronghold" \
  rp_origins="https://stronghold.example.com"
```

Example with self-registration disabled:

```shell
d8 stronghold write auth/webauthn/config \
  rp_id="stronghold.example.com" \
  rp_display_name="Deckhouse Stronghold" \
  rp_origins="https://stronghold.example.com" \
  auto_registration=false
```

### Pre-create a user

If `auto_registration=false`, an administrator must create the user in advance and define the token parameters:

```shell
d8 stronghold write auth/webauthn/user/alice \
  display_name="Alice Doe" \
  token_policies="developers" \
  token_ttl="1h"
```

The `auth/webauthn/user/<name>` path can be used to:

- pre-create a user for registration;
- update `display_name`;
- set token parameters such as `token_policies`, `token_ttl`, `token_max_ttl`, and `token_period`;
- inspect user metadata and the number of registered credentials;
- delete the user together with all enrolled passkeys.

## Sign in through the UI

The Stronghold web UI supports `WebAuthn` directly.

To enroll a passkey for the first time:

1. Select the `WebAuthn` auth method.
1. Disable `Use passkey picker` so the `Username` field is shown.
1. Enter the username and click `Register`.
1. Confirm the operation on the device or in the passkey manager.

For later sign-ins, two modes are available:

- with `Use passkey picker` enabled, the browser shows discoverable credentials and the user does not need to enter a username;
- with it disabled, the user enters `Username` first and then the browser prompts for a matching passkey.

## Registration and login through the API

The following flow is suitable for custom web applications.

### Register a passkey

1. Request registration options:

```shell
curl \
  --request POST \
  --data '{"username":"alice"}' \
  https://stronghold.example.com/v1/auth/webauthn/register/begin
```

1. Pass the returned `publicKey` options to `navigator.credentials.create(...)` in the browser.

1. Finish registration by sending the credential back to Stronghold:

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

Send this payload to `POST /v1/auth/webauthn/register/finish`.

### Login

1. Request login options:

```shell
curl \
  --request POST \
  --data '{"username":"alice"}' \
  https://stronghold.example.com/v1/auth/webauthn/login/begin
```

For passkey picker mode, `username` can be omitted.

1. Pass the returned options to `navigator.credentials.get(...)`.

1. Finish login by sending the authenticator response to `POST /v1/auth/webauthn/login/finish`:

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

Stronghold returns the token in `auth.client_token`.

## Practical notes

- Use `rp_origins` values that exactly match the real Stronghold UI or application URLs.
- In production, it is often better to set `auto_registration=false` and pre-create users with the required policies.
- `WebAuthn` is a good fit for passwordless UI sign-in, while machine-to-machine scenarios are usually better served by `AppRole`, `JWT`, or `Kubernetes`.
