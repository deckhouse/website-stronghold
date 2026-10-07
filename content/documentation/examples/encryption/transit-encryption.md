---
title: "Encrypting database fields with Transit"
linkTitle: "Transit encryption"
description: "Encrypting individual fields in an application database with the Transit engine, envelope encryption with datakey, key rotation, rewrap, and min_decryption_version."
weight: 10
params:
  relatedLinks:
    - title: "Transit secrets engine"
      url: ../../../user/secrets-engines/transit/
    - title: "Secrets engines API"
      url: ../../../reference/api/secrets/
    - title: "Policies"
      url: ../../../concepts/policy/
---

The application encrypts sensitive fields (document numbers, phone numbers, tokens) through the Stronghold API and stores only ciphertext in the database. The encryption key never leaves Stronghold, so a leaked database dump does not expose the data.

## Goal

Set up a Transit key for an application, encrypt and decrypt a field, apply envelope encryption for large objects, rotate the key, re-encrypt data (rewrap), and prohibit decryption with old key versions.

## Prerequisites

- A Stronghold token with permissions to enable secrets engines and create policies.
- The `base64`, `jq`, and `openssl` utilities for the examples.

## Step 1. Create a key

1. Enable the secrets engine:

   ```bash
   d8 stronghold secrets enable transit
   ```

1. Create a named key for the application:

   ```bash
   d8 stronghold write -f transit/keys/orders type=aes256-gcm96
   ```

   Use a separate key for each application or data set.

## Step 2. Separate permissions

1. Application policy: encryption and decryption only:

   ```bash
   d8 stronghold policy write orders-app - <<'POLICY'
   path "transit/encrypt/orders" {
     capabilities = ["update"]
   }
   path "transit/decrypt/orders" {
     capabilities = ["update"]
   }
   path "transit/datakey/plaintext/orders" {
     capabilities = ["update"]
   }
   POLICY
   ```

1. Re-encryption job policy: rewrap only, with no access to plaintext:

   ```bash
   d8 stronghold policy write orders-rewrap - <<'POLICY'
   path "transit/rewrap/orders" {
     capabilities = ["update"]
   }
   path "transit/keys/orders" {
     capabilities = ["read"]
   }
   POLICY
   ```

Leave key management (`transit/keys/orders/*`) to administrators.

## Step 3. Encrypt and decrypt a field

Data is sent to Stronghold base64-encoded.

1. Encrypt a value:

   ```bash
   d8 stronghold write -field=ciphertext transit/encrypt/orders \
     plaintext=$(echo -n "4509 123456" | base64)
   ```

   Store the `vault:v1:...` result in a table column. The `v1` prefix is the key version used for encryption.

1. Decrypt the value:

   ```bash
   d8 stronghold write -field=plaintext transit/decrypt/orders \
     ciphertext="vault:v1:..." | base64 --decode
   ```

To process many records in one request, use the `batch_input` parameter of the `encrypt`, `decrypt`, and `rewrap` endpoints.

## Step 4. Use envelope encryption for large data

Sending files and large objects through the API is inefficient. Instead, request a data key: Stronghold returns it in plaintext for local encryption and in encrypted form for storage.

1. Get a data key:

   ```bash
   d8 stronghold write -format=json -f transit/datakey/plaintext/orders > datakey.json
   jq -r '.data.plaintext' datakey.json | base64 --decode > dek.bin
   jq -r '.data.ciphertext' datakey.json > dek.enc
   ```

1. Encrypt the file locally and delete the plaintext key:

   ```bash
   openssl enc -aes-256-cbc -pbkdf2 -in report.pdf -out report.pdf.enc -pass file:dek.bin
   shred -u dek.bin
   ```

   Store `report.pdf.enc` together with `dek.enc`.

1. To decrypt, recover the data key through Transit:

   ```bash
   d8 stronghold write -field=plaintext transit/decrypt/orders \
     ciphertext="$(cat dek.enc)" | base64 --decode > dek.bin
   openssl enc -d -aes-256-cbc -pbkdf2 -in report.pdf.enc -out report.pdf -pass file:dek.bin
   shred -u dek.bin
   ```

If the data key is needed only for storage (for example, another service will decrypt it), use `transit/datakey/wrapped/orders`: the response does not contain the plaintext key.

## Step 5. Rotate the key

1. Create a new key version:

   ```bash
   d8 stronghold write -f transit/keys/orders/rotate
   ```

   New data is encrypted with the latest version; old data can still be decrypted.

1. To rotate automatically, set a period:

   ```bash
   d8 stronghold write transit/keys/orders/config auto_rotate_period=720h
   ```

## Step 6. Re-encrypt data (rewrap)

Rewrap re-encrypts ciphertext with the latest key version without exposing plaintext. Run it for all table records, for example as a background job with the `orders-rewrap` policy:

```bash
d8 stronghold write -field=ciphertext transit/rewrap/orders ciphertext="vault:v1:..."
```

Write the `vault:v2:...` result back to the column.

## Step 7. Prohibit decryption with old versions

When the database no longer contains ciphertext with old versions, raise the minimum decryption version:

```bash
d8 stronghold write transit/keys/orders/config min_decryption_version=2
```

Versions below `min_decryption_version` are archived and are not used for decryption. In an emergency, you can lower the value again.

To delete old versions permanently, use `transit/keys/orders/trim` with the `min_available_version` parameter. This action is irreversible.

## Verification

1. View the key state:

   ```bash
   d8 stronghold read transit/keys/orders
   ```

   Check the `latest_version`, `min_decryption_version`, and `auto_rotate_period` values.

1. Try to decrypt `vault:v1:...` ciphertext after step 7: the request must fail.
1. Make sure a token with the `orders-rewrap` policy cannot call `transit/decrypt/orders`.

## Cleanup

1. Delete the policies: `d8 stronghold policy delete orders-app` and `d8 stronghold policy delete orders-rewrap`.
1. Delete the local files `datakey.json`, `dek.enc`, and `report.pdf.enc`.
1. To delete the test key, allow deletion and delete the key:

   ```bash
   d8 stronghold write transit/keys/orders/config deletion_allowed=true
   d8 stronghold delete transit/keys/orders
   ```

   Data encrypted with this key becomes inaccessible.
