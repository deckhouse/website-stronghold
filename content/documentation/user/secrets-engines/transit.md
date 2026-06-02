---
title: "Transit secrets engine"
description: "Information about Transit secrets engine in Deckhouse Stronghold."
weight: 50
---

The Transit secrets engine performs cryptographic operations on data in transit.
Stronghold does not store the data sent to this engine.
It can also be viewed as “cryptography as a service” or “encryption as a service”.

The Transit engine lets you:

- encrypt and decrypt data;
- sign data and verify signatures;
- compute hashes and HMAC values;
- generate random bytes.

The primary Transit use case is encrypting data in applications and then storing the encrypted data in external storage.
This removes the burden of implementing encryption and decryption correctly from application developers and shifts it to Stronghold.

Transit supports derived keys.
This makes it possible to use the same key for different purposes by deriving a new key from user-provided context.
In this mode, convergent encryption can also be enabled so that identical input values produce identical ciphertext.

Data key generation allows processes to request a high-entropy key of a given length.
That key is returned encrypted with the specified key.
Usually, the key is also returned in plain text for immediate use, but this behavior can be disabled to meet audit requirements.

## Working set management

The Transit engine supports key versioning.
Key versions older than the value of `min_decryption_version` are archived, while newer versions remain in the working set.

This matters for two reasons:

- performance — keys in the working set load faster;
- security — decryption with older key versions can be disallowed.

If older data must be decrypted in an emergency, you can temporarily lower `min_decryption_version`.

At present, the archive is stored in a single storage entry.
In some backends, especially those that use Raft or Paxos for high availability (HA), frequent key rotation can cause the archive entry to exceed the allowed size.

For scenarios with frequent rotation, a good alternative is to use named keys tied to time intervals.
For example, you can use 5-minute periods rounded to the nearest multiple of five.
This makes it possible to use several keys at the same time and deterministically choose the correct key at any given moment.

## NIST recommendations for key rotation

Periodic rotation of encryption keys is recommended even if a compromise has not occurred.

For AES-GCM keys, rotation should happen before approximately 2<sup>32</sup> encryption operations are performed with a single key version, in accordance with NIST 800-38D.
Operators should estimate the key usage rate and configure the rotation frequency so that this limit is not exceeded.

For example, if a key is used about 40 million times per day, rotating it once every three months is sufficient.

## Key types

The Transit secrets engine currently supports the following key types.
All key types also generate separate HMAC keys.

- `aes128-gcm96` — AES-GCM with a 128-bit AES key and a 96-bit `nonce`; supports encryption, decryption, derived keys, and convergent encryption.
- `aes256-gcm96` — AES-GCM with a 256-bit AES key and a 96-bit `nonce`; supports encryption, decryption, derived keys, and convergent encryption; this is the default.
- `chacha20-poly1305` — ChaCha20-Poly1305 with a 256-bit key; supports encryption, decryption, derived keys, and convergent encryption.
- `ed25519` — Ed25519; supports signing, signature verification, and derived keys.
- `ecdsa-p256` — ECDSA using the P-256 curve; supports signing and signature verification.
- `ecdsa-p384` — ECDSA using the P-384 curve; supports signing and signature verification.
- `ecdsa-p521` — ECDSA using the P-521 curve; supports signing and signature verification.
- `rsa-2048` — a 2048-bit RSA key; supports encryption, decryption, signing, and signature verification.
- `rsa-3072` — a 3072-bit RSA key; supports encryption, decryption, signing, and signature verification.
- `rsa-4096` — a 4096-bit RSA key; supports encryption, decryption, signing, and signature verification.
- `hmac` — HMAC; supports HMAC generation and verification.

{{< alert level="info" >}}
All key types support HMAC operations by using a second randomly generated key created during key initialization or rotation.
The `hmac` key type supports only HMAC operations and behaves the same as the other algorithms for HMAC operations, but it also supports key import.
By default, the `hmac` key type uses a 256-bit key.
{{< /alert >}}

RSA operations use one of the following methods:

- OAEP — for encryption and decryption, with the SHA-256 hash function and MGF;
- PSS — for signing and signature verification, with a configurable hash function that is also used for MGF;
- PKCS#1 v1.5 — for signing and signature verification, with a configurable hash function.

## Convergent encryption

Convergent encryption is a mode in which the same combination of plain text and context always results in the same ciphertext.
This is achieved by deriving a key with a key derivation function and deterministically deriving the `nonce`.

Because these values differ for any combination of plain text and context in a 2^256 key space, the risk of `nonce` reuse is effectively zero.

This approach has many practical applications.
One common scenario is storing ciphertext in a database with limited search and query support.
In this case, rows with the same value for a specific field can be returned by a query.

To allow algorithm upgrades when needed, several versions of convergent encryption have been supported over time:

- Version 1 required the client to provide the `nonce`.
  This offered high flexibility, but it could be dangerous if implemented incorrectly.
  Keys that use this version cannot be upgraded.
- Version 2 used an algorithmic approach to compute the parameters.
  However, that algorithm was vulnerable to offline plain-text confirmation attacks.
  Because of this, an attacker could brute-force decryption when the plain text was small.
  Version 2 keys can be upgraded by rotating to a new key version.
  After that, existing values can be rewrapped with the new key version, and the version 3 algorithm will be used.
- Version 3 uses a different algorithm designed to resist offline plain-text confirmation attacks.
  It is similar to AES-SIV in that it uses a pseudorandom function (PRF) to generate the `nonce` from the plain text.

## Configuration

Most secrets engines require initial configuration before use.
These steps are usually performed by an operator or a configuration management system.

Enable the Transit secrets engine:

```shell
stronghold secrets enable transit
```

Example output:

```console
Success! Enabled the transit secrets engine at: transit/
```

By default, the engine is mounted at its own name.
To specify a different path, use the `-path` argument.

Create a named encryption key:

```shell
stronghold write -f transit/keys/my-key
```

Example output:

```console
Success! Data written to: transit/keys/my-key
```

Usually, each application uses its own encryption key.

## Usage

After the engine is configured and a user or system receives a Stronghold token with the required permissions, cryptographic operations can be performed.

Encrypt data using the `/encrypt` endpoint and the specified key.

{{< alert level="info" >}}
All data must be Base64-encoded.
This is because Stronghold does not require the provided `plaintext` to be text.
For example, it can be a PDF file, an image, or other binary data.
The safest way to pass such data in JSON is to use Base64 encoding.
{{< /alert >}}

```shell
stronghold write transit/encrypt/my-key plaintext=$(echo "my secret data" | base64)
```

Example output:

```console
Key           Value
---           -----
ciphertext    vault:v1:8SDd3WHDOjf7mq69CyCqYjBXAiQQAVZRkFM13ok481zoCmHnSeDX9vyf7w==
```

The returned ciphertext begins with the prefix `vault:v1:`.
The `vault` prefix indicates that the data was encrypted by Stronghold.
The `v1` value indicates that the first key version was used.

This matters during rotation because Stronghold uses the correct key version for decryption.

Stronghold does not store the ciphertext.
The caller is responsible for storing it.
To decrypt the data, send the ciphertext to Stronghold again.

{{< alert level="warning" >}}
The Stronghold HTTP API limits the request size to 32 MB to protect against DoS attacks.
You can change this limit in the `listener` block of the [Stronghold configuration](../../../install/standalone/configuration/#listener).
{{< /alert >}}

Decrypt data using the `/decrypt` endpoint:

```shell
stronghold write transit/decrypt/my-key ciphertext=vault:v1:8SDd3WHDOjf7mq69CyCqYjBXAiQQAVZRkFM13ok481zoCmHnSeDX9vyf7w==
```

Example output:

```console
Key          Value
---          -----
plaintext    bXkgc2VjcmV0IGRhdGEK
```

The result is a Base64 string.
Decode it to get the original data:

```shell
base64 --decode <<< "bXkgc2VjcmV0IGRhdGEK"
```

Example output:

```console
my secret data
```

You can do the same in a single command:

```shell
stronghold write -field=plaintext transit/decrypt/my-key ciphertext=... | base64 --decode
```

Example output:

```console
my secret data
```

Using ACL, you can restrict access to the Transit engine.
For example, trusted operators can be allowed to manage keys, while applications can be limited to encryption and decryption with specific keys.

Rotate the base encryption key.
This generates a new key and adds it to the keyring of the named key:

```shell
stronghold write -f transit/keys/my-key/rotate
```

Example output:

```console
Success! Data written to: transit/keys/my-key/rotate
```

All subsequent encryption operations use the new key.
Older data can still be decrypted because Stronghold uses the keyring.

Update previously encrypted data to use the new key.
Stronghold decrypts the value with the appropriate key version and then encrypts the plain text again with the newest key:

```shell
stronghold write transit/rewrap/my-key ciphertext=vault:v1:8SDd3WHDOjf7mq69CyCqYjBXAiQQAVZRkFM13ok481zoCmHnSeDX9vyf7w==
```

Example output:

```console
Key           Value
---           -----
ciphertext    vault:v2:...
```

This process does not expose the original data.
Because of this, a Stronghold policy can allow even untrusted processes to perform `rewrap` without granting access to the data itself.

## Bring your own key (BYOK)

{{< alert level="warning" >}}
Key import is needed when you must migrate a key from an HSM or another system.
However, it is safer to generate and manage the key inside Stronghold.
{{< /alert >}}

First, get the wrapping key from Transit:

```shell
stronghold read transit/wrapping_key
```

This key is a 4096-bit RSA public key.
Then use it to create the encrypted value for `import`.

Below, target key means the key being imported.

### HSM

If the key is imported from an HSM that supports PKCS#11, there are two possible scenarios:

- If the HSM supports the `CKM_RSA_AES_KEY_WRAP` mechanism, use it to wrap the target key with the wrapping key.
- Otherwise, use two mechanisms.
  First, generate a 256-bit AES key.
  Then use it to wrap the target key with the `CKM_AES_KEY_WRAP_KWP` mechanism.
  After that, wrap the AES key with the wrapping key using the `CKM_RSA_PKCS_OAEP` mechanism with MGF1 and one of the following hash functions: SHA-1, SHA-224, SHA-256, SHA-384, or SHA-512.

The ciphertext is created by concatenating the wrapped target key and the wrapped AES key.
The ciphertext bytes must be Base64-encoded.

### Manual process

If the target key is not stored in an HSM or KMS, perform the following steps to create ciphertext for the `import` endpoint:

1. Generate an ephemeral 256-bit AES key.
1. Wrap the target key with the ephemeral AES key using AES-KWP.
1. Wrap the AES key with the Stronghold wrapping key using RSAES-OAEP with MGF1 and one of the following hash functions: SHA-1, SHA-224, SHA-256, SHA-384, or SHA-512.
1. Delete the ephemeral AES key.
1. Concatenate the wrapped target key and the wrapped AES key.
1. Encode the result in Base64.

{{< alert level="warning" >}}
When wrapping a symmetric key, such as an AES or ChaCha20 key, wrap the raw key bytes.
For example, for a 128-bit AES key, this is a 16-byte array that must be wrapped without Base64 or any other encoding.

When wrapping an asymmetric key, such as an RSA or ECDSA key, wrap the key in PKCS8 format in raw DER binary form.
Do not apply PEM encoding before encryption, and do not Base64-encode the key.
{{< /alert >}}
