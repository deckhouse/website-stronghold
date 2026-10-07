---
title: "Cryptographic algorithms"
linkTitle: "Introduction"
weight: 10
description: "Administrator overview of TLS, storage encryption, HSM, PKI, and Transit cryptographic algorithms in Stronghold."
---

This page summarizes the cryptographic mechanisms of Stronghold:

- which algorithms are used for TLS;
- how data in the storage is encrypted;
- how HSM complements the built-in encryption;
- which algorithms are available in `PKI`;
- which algorithms are available in `Transit`.

## Summary

| Area | Main algorithms and mechanisms |
| --- | --- |
| TLS | Configurable versions `TLS 1.0`–`TLS 1.3`, cipher suites for `TLS 1.2` and earlier, server and client certificate verification |
| Storage encryption | Built-in `AES-256-GCM` cryptographic barrier |
| Additional protection | `seal "pkcs11"` for HSM and `seal wrap` for additional encryption of critical data |
| PKI | `RSA`, `EC`, `Ed25519`, and supported `GOST` variants |
| Transit | `AES-GCM`, `ChaCha20-Poly1305`, `RSA`, `ECDSA`, `Ed25519`, `HMAC`, and supported `GOST` variants |

## GOST algorithms

Stronghold supports scenarios that require GOST cryptography:

- `TLS` supports `TLS 1.3` with verification of certificates based on `GOST 34.10` keys and with the `Magma` and `Kuznyechik` GOST ciphers;
- the `PKI` engine supports issuing certificates with `GOST 34.10` keys;
- the `Transit` engine supports signing with asymmetric `GOST 34.10` keys, as well as encryption with the symmetric `GOST 34.12` (`Magma`, `Kuznyechik`) and `GOST 28147-89` algorithms;
- when `Managed Keys` and HSMs that support GOST cryptography are used, you can work not only with `RSA`, but also with keys and certificates based on the supported GOST algorithms.

## TLS: versions, cipher suites, and certificate verification

TLS for the Stronghold API is configured in the `listener "tcp"` section using the [standalone configuration](../../install/standalone/configuration/#listener) parameters.

### Supported TLS versions

The `tls_min_version` and `tls_max_version` parameters accept the following values:

- `tls10`;
- `tls11`;
- `tls12`;
- `tls13`.

For production environments, use only `TLS 1.2` and `TLS 1.3`.

### Supported cipher suites

The `tls_cipher_suites` parameter lets you explicitly set the list of ciphers for `TLS 1.2` and earlier versions.

Keep in mind:

- `tls_cipher_suites` does not affect `TLS 1.3`; for `TLS 1.3`, the cipher order is set by the global `tls_13_cipher_policy` parameter (see the [configuration](../../install/standalone/configuration/#parameter-overview));
- the names of allowed ciphers correspond to the list from Go `crypto/tls`;
- if you specify only cipher suites that are forbidden by the `HTTP/2` specification, the CLI and other `net/http`-based clients may not work correctly.

A practical approach:

- limit `tls_min_version` to `tls12`;
- set `tls_cipher_suites` only when required by security or compatibility requirements;
- after changing the cipher suites, check that the CLI, UI, and load balancers work.

### Certificate verification

Stronghold supports several levels of TLS certificate verification:

- `tls_cert_file` and `tls_key_file` set the server certificate and private key;
- `tls_cert_file` can contain a full PEM chain, in which the node certificate must come first;
- `tls_require_and_verify_client_cert = true` enables mandatory client certificate verification;
- `tls_client_ca_file` sets the CA file that Stronghold uses to verify client certificates;
- `tls_disable_client_certs = true` fully disables client certificate verification.

In environments that use GOST cryptography, Stronghold supports `TLS 1.3` with verification of `GOST 34.10` certificates, as well as TLS encryption using the `Magma` and `Kuznyechik` algorithms.

## Storage encryption

Stronghold protects data in the storage with a built-in cryptographic barrier based on `AES-256-GCM`.

This means:

- `AES-GCM` with a 256-bit key is used to encrypt data in the storage;
- the barrier provides not only confidentiality, but also data integrity control;
- a separate `nonce` is used for each entry; the standard `GCM nonce` size is 96 bits (12 bytes);
- Stronghold supports barrier key rotation, so new entries can be encrypted with a new key without losing access to previously stored data.

Barrier encryption works on top of the selected storage backend. Therefore, regardless of whether `raft`, `filesystem`, `etcd`, or `postgresql` is used, already encrypted Stronghold values are written to the physical storage.

It is important to distinguish between:

- **cryptographic protection of Stronghold data**, provided by the `AES-256-GCM` barrier;
- **operational backup procedures for the storage backend**, which remain the administrator's responsibility.

## Additional encryption via HSM

HSM support does not replace the built-in `AES-256-GCM` barrier; it complements it.

In a typical scenario:

- data in the storage backend is still protected by the Stronghold barrier;
- the root key and unseal operations can be protected via `seal "pkcs11"` and an external HSM;
- additional encryption via `seal wrap` can be used for some critical internal data.

Thus, Stronghold uses several levels of protection:

1. built-in storage encryption with `AES-256-GCM`;
1. protection of the root key and auto-unseal via HSM;
1. if necessary, an additional `seal wrap` layer for the most sensitive values.

If the HSM or external key provider supports GOST cryptography, this approach also applies to scenarios with GOST keys: the external module protects the key material, and Stronghold continues to use it in `PKI`, `Transit`, and related cryptographic operations.

{{< alert level="warning" >}}
Support for `seal "pkcs11"` applies to the standalone Stronghold installation.
{{< /alert >}}

More details:

- [HSM support](../kms-hsm/hsm/)
- [Double encryption](../kms-hsm/sealwrap/)

## Algorithms in PKI

The `PKI` secrets engine issues and signs `X.509` certificates.

When configuring roles and generating keys, the following key types are available:

- `rsa`;
- `ec`;
- `ed25519`;
- `gost3410`.

The `any` value can be used in roles as a service mode when you need to allow any supported key type. It is not a separate cryptographic algorithm.

Additionally, you can set key size parameters for PKI:

- for `rsa`: `2048`, `3072`, `4096`;
- for `ec`: `224`, `256`, `384`, `521`;
- for `ed25519`, the `key_bits` parameter is not used;
- for GOST parameters, the key type already defines the corresponding parameter set.

For the signature algorithm, the `signature_bits` parameter supports:

- `256` for `SHA-2-256`;
- `384` for `SHA-2-384`;
- `512` for `SHA-2-512`.

If `signature_bits` is not set, Stronghold selects the value automatically:

- `SHA-2-256` is used for `RSA`;
- for `NIST P-Curves`, the hash matching the curve size is selected.

In practice, this means that PKI in Stronghold covers typical corporate scenarios with `RSA` and `EC`, modern scenarios with `Ed25519`, and infrastructures that require the supported `GOST` algorithms.

In particular, PKI supports issuing certificates with `GOST 34.10` keys, including the supported `256`-bit and `512`-bit parameter sets.

For details on configuring PKI, see [PKI secrets engine](../../user/secrets-engines/pki/).

## Algorithms in Transit

`Transit` provides cryptographic operations as a service: encryption, decryption, signing, signature verification, HMAC, hashing, and data key operations.

### Main Transit key types

Symmetric encryption algorithms:

- `aes128-gcm96`;
- `aes256-gcm96` (default);
- `chacha20-poly1305`;
- `gost28147`;
- `gost341264`;
- `gost3412128`.

Asymmetric algorithms for signing and, in some cases, encryption:

- `ecdsa-p256`;
- `ecdsa-p384`;
- `ecdsa-p521`;
- `ed25519`;
- `rsa-2048`;
- `rsa-3072`;
- `rsa-4096`;
- `gost3410`.

Specialized types:

- `hmac` for HMAC operations;
- `managed_key` for operations via an external managed key, if it has been configured beforehand.

The `managed_key` type is not a separate encryption algorithm: in this mode, Transit uses a preconfigured external key source.

If `managed_key` points to an HSM or another external provider that supports GOST cryptography, Transit can also use external GOST keys.

### Available operations

- `aes128-gcm96`, `aes256-gcm96`, `chacha20-poly1305`, and the supported symmetric `GOST` types are suitable for `encrypt` / `decrypt`;
- `RSA` is suitable for `encrypt` / `decrypt`, as well as `sign` / `verify`;
- `ECDSA`, `Ed25519`, and the supported asymmetric `GOST` types are used for `sign` / `verify`;
- all key types can be used in HMAC scenarios via a separate HMAC key that Transit maintains together with the main key;
- the `hmac` type is intended only for HMAC operations.

For GOST scenarios, this means:

- `gost3410-*` is used for `sign` / `verify` operations;
- `gost3412128` and `gost341264` correspond to the symmetric `GOST 34.12` algorithms (`Kuznyechik` and `Magma`);
- `gost28147` covers `GOST 28147-89` scenarios.

For RSA, Transit uses:

- `OAEP` for encryption and decryption;
- `PSS` for signing and signature verification;
- `PKCS#1 v1.5` for signing and signature verification.

Derived keys are supported for symmetric keys, and convergent encryption is supported for some types. For `AES-GCM`, Stronghold recommends regular key rotation depending on the load.

For details on Transit capabilities, see [Transit secrets engine](../../user/secrets-engines/transit/).

## Practical recommendations

- For public and service-to-service APIs, set `tls_min_version = "tls12"` or higher.
- Do not restrict `tls_cipher_suites` without a clear reason: it may unexpectedly affect client compatibility.
- Treat HSM as an additional level of protection, not as a replacement for the built-in Stronghold encryption.
- For PKI, standardize the allowed key types and key lengths at the role level in advance.
- For Transit, define the allowed key types separately for each scenario: symmetric encryption, signing, HMAC, BYOK, managed keys.
