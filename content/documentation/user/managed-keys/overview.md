---
title: "Managed Keys"
linkTitle: "Overview"
weight: 10
description: "Overview of managed keys in Stronghold: supported backends and secret engines that can use them."
---

`Managed Keys` allow Stronghold to use cryptographic keys stored in an external trusted system, such as an HSM or external KMS, without storing private key material inside Stronghold itself.

This approach is useful in the following scenarios:

- private keys must be stored outside Stronghold.
- signing or encryption operations must be performed by an external system.
- security or compliance requirements prohibit exporting key material.

In Stronghold, a managed key is a named record managed through the `sys/managed-keys/<type>/<name>` API.

## How it works

Stronghold stores the configuration required to access an external key, but not the private key itself.
When a secret engine needs to perform signing, signature verification, encryption, or decryption, it calls the corresponding managed key.
The managed key then delegates the operation to the external backend.

This means the following:

- key material remains in the external system.
- Stronghold uses the managed key as an abstraction layer over the external backend.
- multiple managed keys can be configured for the same backend if different keys or different access policies are required.

## Supported backends

Stronghold supports the following managed key types:

- `pkcs11`.
- `yandexcloudkms`.

### `pkcs11`

The `pkcs11` type is used to work with HSMs and `PKCS#11`-compatible libraries.
A managed key of this type refers to a previously declared `kms_library "pkcs11"` in the Stronghold server configuration.

This scenario typically requires the following:

- configure `kms_library "pkcs11"` in the server configuration.
- register a managed key of type `pkcs11`.
- allow the required secret engines to use the key.

### `yandexcloudkms`

The `yandexcloudkms` type is used to integrate with Yandex Cloud KMS.

The following parameters are used to configure this type of managed key:

- `kms_key_id`.
- `oauth_token` or `service_account_key_json`.
- `endpoint`, if required.

To authenticate Stronghold in Yandex Cloud KMS, you can use one of the following methods:

- `oauth_token`.
- `service_account_key_json`.
- the virtual machine service account in Yandex Cloud.

If neither `oauth_token` nor `service_account_key_json` is specified, Stronghold attempts to use the virtual machine service account through the standard instance credentials mechanism.

{{< alert level="info" >}}
For `yandexcloudkms`, the `oauth_token` and `service_account_key_json` parameters are mutually exclusive.
If one of them is specified, it is used for authentication in Yandex Cloud.
The set of supported operations depends on the key itself and its configuration.
If the managed key is intended for use with `PKI` or `SSH`, it must support signing.
For `Transit`, the supported operations depend on the selected usage scenario.
{{< /alert >}}

## Support in secret engines

In Stronghold, managed keys can be used by the following secret engines:

- `SSH`.
- `PKI`.
- `Transit`.

### SSH

The `SSH` secrets engine can use a managed key as a certificate authority (CA) key and use it to sign SSH certificates.

This is useful when the SSH CA must remain in an external HSM or external KMS.

### PKI

The `PKI` secrets engine can use a managed key to generate and maintain root and intermediate CAs, and to sign certificates.

This is one of the main managed key use cases in Stronghold.

If the key is not declared with `any_mount=true`, its use must be explicitly allowed for a specific `mount path`.

### Transit

The `Transit` secrets engine can use a managed key as an external cryptographic key.

Depending on the capabilities of the backend and the key type, this allows the following operations:

- signing.
- signature verification.
- encryption.
- decryption.

For `Transit`, the managed key is specified when creating or rotating a transit key.

## Namespaces and scope

A managed key is bound to a specific namespace.
The secret engine that uses this key must be located in the same namespace as the managed key itself.

If the key is not declared with `any_mount=true`, its use must be explicitly allowed for a specific `mount path`.

## Practical recommendations

- Use `pkcs11` if your keys are stored in a local or network HSM with a `PKCS#11` library.
- Use `yandexcloudkms` if your keys are managed in Yandex Cloud KMS.
- Restrict `usages` to only the operations required by your scenario.
- If the key should not be available to all `mount path` values, do not enable `any_mount` and configure `allowed-managed-keys` explicitly.
- Before connecting a key to `PKI`, `SSH`, or `Transit`, verify that the external key is accessible and supports the required operations.

Detailed commands for registering, checking, viewing, and deleting managed keys are available on the [Basic operations](../basic-operations/) page.
