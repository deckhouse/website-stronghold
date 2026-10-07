---
title: "One-time TOTP codes for an application"
linkTitle: "TOTP codes"
description: "The TOTP secrets engine: creating a key for an application user, getting and validating a code, and protection against code reuse."
weight: 70
params:
  relatedLinks:
    - title: "TOTP secrets engine"
      url: ../../../user/secrets-engines/totp/
    - title: "MFA for administrators"
      url: ../../access/admin-mfa/
---

The TOTP secrets engine lets an application validate one-time second-factor codes of its own users without storing the keys itself. This is a separate capability: it is not related to MFA when logging in to Stronghold itself (see [MFA for administrators](../../access/admin-mfa/)).

## Goal

Create a TOTP key for an application user, get a code and validate it, and make sure the same code cannot be used twice.

## Prerequisites

- A Stronghold token with permission to enable secrets engines and write to `totp/*`.

## Step 1. Enable the engine and create a key

```bash
d8 stronghold secrets enable totp

d8 stronghold write -format=json totp/keys/alice \
  generate=true \
  issuer="Example" \
  account_name="alice@example.com" \
  period=30 \
  algorithm=SHA1 \
  digits=6 | jq -r '.data.url'
```

The command returns a URL of the form `otpauth://totp/...`. The application shows it to the user as a QR code (the response also contains a `barcode` field with a base64 image), and the user adds the key to an authenticator app.

Main key parameters:

| Parameter | Purpose |
| --- | --- |
| `generate` | If `true`, Stronghold creates the secret itself. Otherwise, pass the `key` or `url` of an existing key |
| `issuer`, `account_name` | The service name and account shown by the authenticator |
| `period` | The code rotation period in seconds |
| `algorithm` | `SHA1`, `SHA256`, or `SHA512` |
| `digits` | The code length: `6` or `8` |
| `skew` | The allowed time deviation in periods: `0` or `1` |
| `exported` | If `false`, the URL and QR code are not returned after the key is created |

## Step 2. Give the application minimal permissions

The application only needs to validate codes:

```bash
d8 stronghold policy write totp-verifier - <<'POLICY'
path "totp/code/*" {
  capabilities = ["update"]
}
POLICY
```

The `read` capability on `totp/code/*` allows getting the codes themselves, so do not give it to the application.

## Step 3. Get and validate a code

1. Get the current code. This imitates the user's authenticator:

   ```bash
   CODE=$(d8 stronghold read -field=code totp/code/alice)
   ```

1. Validate the code the way the application does:

   ```bash
   d8 stronghold write totp/code/alice code="$CODE"
   ```

   In the response, `valid` is `true`.

1. Validate an incorrect code:

   ```bash
   d8 stronghold write totp/code/alice code=000000
   ```

   In the response, `valid` is `false`.

## Step 4. Make sure the code is single-use

A repeated validation of an already used code is rejected:

```bash
# repeated validation of the same code
d8 stronghold write totp/code/alice code="$CODE"
```

Stronghold responds with the error `code already used; wait until the next time period`. The next code can be validated after the period changes.

## Verification

```bash
d8 stronghold list totp/keys
d8 stronghold read totp/keys/alice
```

The list contains `alice`, and reading the key returns its parameters (`issuer`, `period`, `algorithm`, `digits`) without the secret.

## Cleanup

```bash
d8 stronghold delete totp/keys/alice
d8 stronghold policy delete totp-verifier
d8 stronghold secrets disable totp
```
