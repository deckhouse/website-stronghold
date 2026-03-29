---
title: "Managed Keys in Stronghold"
linkTitle: "Introduction"
weight: 10
description: "Overview of Managed Keys in Stronghold: supported backends and the secrets engines that can use them."
---

`Managed Keys` let Stronghold use cryptographic keys stored in an external trusted system, such as an HSM or external KMS, without storing the private key material inside Stronghold itself.

This approach is useful when:

- private keys must remain outside Stronghold;
- signing or encryption operations must be performed by an external system;
- security or compliance requirements prohibit exporting key material.

In Stronghold, a managed key is a named record managed through the `sys/managed-keys/<type>/<name>` API.

## How it works

Stronghold stores the configuration needed to access an external key, but not the private key itself. When a secrets engine needs to perform signing, signature verification, encryption, or decryption, it calls the corresponding managed key, which then delegates the operation to the external backend.

This means:

- key material remains in the external system;
- Stronghold uses the managed key as an abstraction over the external backend;
- multiple managed keys can be configured for the same backend if you need different keys or different access policies.

## Supported backends

Stronghold supports the following managed key types:

- `pkcs11`
- `yandexcloudkms`

### `pkcs11`

The `pkcs11` type is used for HSMs and PKCS#11-compatible libraries. A managed key of this type references a previously defined `kms_library "pkcs11"` stanza in Stronghold server configuration.

A typical workflow is:

- configure `kms_library "pkcs11"` in the server configuration;
- register a managed key through `sys/managed-keys/pkcs11/<name>`;
- optionally verify key access using `test/sign`;
- allow the key for the required mount path.

Example:

```shell
stronghold write sys/managed-keys/pkcs11/my-hsm-key \
  library=softhsm \
  token_label=managed-keys \
  pin=1234 \
  key_label=my-signing-key \
  usages=sign,verify
```

### `yandexcloudkms`

The `yandexcloudkms` type is used for integration with Yandex Cloud KMS.

Its configuration uses:

- `kms_key_id`
- `oauth_token` or `service_account_key_json`
- `endpoint` when needed

For authentication to Yandex Cloud KMS, Stronghold can use one of the following options:

- `oauth_token`
- `service_account_key_json`
- the service account attached to the virtual machine in Yandex Cloud

If neither `oauth_token` nor `service_account_key_json` is provided, Stronghold attempts to use the virtual machine service account through the standard instance credentials flow. This is similar to common Vault patterns where cloud KMS integrations can rely on the identity of the running VM instance.

Example:

```shell
stronghold write sys/managed-keys/yandexcloudkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  oauth_token=<oauth_token> \
  usages=sign,verify
```

Example using the virtual machine service account:

```shell
stronghold write sys/managed-keys/yandexcloudkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  usages=sign,verify
```

{{< alert level="info" >}}
The `oauth_token` and `service_account_key_json` parameters are mutually exclusive. If one of them is specified, it is used explicitly for authentication to Yandex Cloud.
{{< /alert >}}

{{< alert level="info" >}}
For `yandexcloudkms`, the set of allowed operations depends on the key itself and its configuration. If the managed key is going to be used by `pki` or `ssh`, it must support signing. For `transit`, the required operations depend on the selected usage scenario.
{{< /alert >}}

## Which secrets engines can use Managed Keys

In Stronghold, managed keys can be used by:

- `SSH`
- `PKI`
- `Transit`

### SSH

The SSH secrets engine can use a managed key as its certificate authority key and use it to sign SSH certificates.

This is useful when the SSH CA key must remain in an external HSM or KMS.

### PKI

The PKI secrets engine can use a managed key for root and intermediate CA generation and for certificate signing.

This is one of the primary managed key use cases in Stronghold. For PKI, it is also important to allow the managed key for the specific PKI mount if the key was not created with `any_mount=true`.

Example:

```shell
stronghold secrets tune -allowed-managed-keys=my-hsm-key pki/
```

### Transit

The Transit secrets engine can use a managed key as an external cryptographic key.

Depending on backend capabilities and the key type, this can allow:

- signing;
- signature verification;
- encryption;
- decryption.

For Transit, the managed key is specified when creating or rotating a transit key.

## Namespaces and scope

A managed key is bound to a specific namespace. The secrets engine that uses the key must exist in the same namespace as the managed key itself.

If the key is not created with `any_mount=true`, you must explicitly allow its use for a given mount path.

## Basic operations

List managed keys of a given type:

```shell
stronghold list sys/managed-keys/pkcs11
```

Read key configuration:

```shell
stronghold read sys/managed-keys/pkcs11/my-hsm-key
```

Verify key accessibility with a test signature:

```shell
stronghold write sys/managed-keys/pkcs11/my-hsm-key/test/sign
```

Delete a managed key:

```shell
stronghold delete sys/managed-keys/pkcs11/my-hsm-key
```

## Practical recommendations

- Use `pkcs11` when keys live in a local or network-attached HSM exposed through a PKCS#11 library.
- Use `yandexcloudkms` when keys are managed in Yandex Cloud KMS.
- Restrict `usages` to the minimum required operations.
- If the key should not be available to every mount, do not enable `any_mount` and configure `allowed-managed-keys` explicitly.
- Before attaching a key to `pki`, `ssh`, or `transit`, verify it with `test/sign` or another appropriate validation flow.

## Related sections

- [PKI secrets engine](../secrets-engines/pki/)
- [Transit secrets engine](../secrets-engines/transit/)
- [Signed SSH certificates](../secrets-engines/signed-ssh-certificates/)
