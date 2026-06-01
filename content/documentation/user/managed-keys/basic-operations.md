---
title: "Basic operations"
linkTitle: "Basic operations"
weight: 20
description: "Basic operations for working with managed key in Deckhouse Stronghold."
---

## Basic operations

This section describes the basic operations for working with `managed key` in Deckhouse Stronghold:

- registering a managed key;
- viewing the list of keys and their configuration;
- checking key availability;
- deleting a managed key;
- allowing a key to be used in secrets engines.

{% alert level="info" %}
Before performing these operations, make sure that the corresponding backend is already prepared. For example, for `pkcs11`, the `kms_library "pkcs11"` setting must be configured in the Stronghold server configuration.
{% endalert %}

## Registering a `pkcs11` managed key

The `pkcs11` type is used to work with HSMs and PKCS#11-compatible libraries.

Use the following command to register a key:

```shell
stronghold write sys/managed-keys/pkcs11/my-hsm-key \
  library=softhsm \
  token_label=managed-keys \
  pin=1234 \
  key_label=my-signing-key \
  usages=sign,verify
```

In this example:

- `my-hsm-key` is the name of the managed key in Stronghold;
- `library` is the name of a previously declared `kms_library`;
- `token_label` is the token label in the external HSM;
- `pin` is the PIN for access;
- `key_label` is the key name in the external backend;
- `usages` is the list of allowed operations.

## Registering a `yandexcloudkms` managed key

The `yandexcloudkms` type is used to work with Yandex Cloud KMS [1].

Use the following command to register a key with `oauth_token`:

```shell
stronghold write sys/managed-keys/yandexcloudkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  oauth_token=<oauth_token> \
  usages=sign,verify
```

Use the following command to register a key using the virtual machine ServiceAccount [1]:

```shell
stronghold write sys/managed-keys/yandexcloudkms/my-yc-key \
  kms_key_id=<kms_key_id> \
  usages=sign,verify
```

If neither `oauth_token` nor `service_account_key_json` is specified, Stronghold tries to use the virtual machine ServiceAccount [1].

## Viewing the list of managed keys

To get the list of managed keys of a specific type, use the following command:

```shell
stronghold list sys/managed-keys/pkcs11
```

For `yandexcloudkms`, use the following path:

```shell
stronghold list sys/managed-keys/yandexcloudkms
```

## Reading the managed key configuration

To read the configuration of a specific key, use the following command:

```shell
stronghold read sys/managed-keys/pkcs11/my-hsm-key
```

This is useful when you need to:

- verify that the key is registered;
- make sure that the backend is configured correctly;
- compare the parameters and scope of use.

## Checking key availability

Before connecting a key to `PKI`, `SSH`, or `Transit`, it is recommended to verify that Stronghold can actually use the external key.

Use the following command for a test signature:

```shell
stronghold write sys/managed-keys/pkcs11/my-hsm-key/test/sign
```

This operation helps confirm the following:

- Stronghold can access the backend;
- the key is found;
- the access permissions and authentication parameters are configured correctly;
- the signing operation is supported.

## Deleting a managed key

If a managed key is no longer needed, delete it with the following command:

```shell
stronghold delete sys/managed-keys/pkcs11/my-hsm-key
```

Before deleting a key, make sure that it is no longer used by active mounts.

## Allowing the key to be used for `PKI`

For the `PKI` secrets engine, a managed key must be allowed for a specific mount path unless the key was declared with `any_mount=true`.

Use the following command:

```shell
stronghold secrets tune -allowed-managed-keys=my-hsm-key pki/
```

This command allows the `my-hsm-key` managed key to be used in the `PKI` mount at `pki/`.

## Using a managed key with `Transit`

For `Transit`, the managed key is specified when creating or rotating a transit key.

The exact scenario depends on the capabilities of the selected backend and the key type. In general, `Transit` uses an external key backend through the managed key and does not store the private key inside Stronghold.

## Using a managed key with `SSH`

For the `SSH` secrets engine, a managed key can be used as an external CA key for signing SSH certificates.

This makes it possible to move the certificate authority key to an external HSM or KMS and avoid storing it inside Stronghold.

## Practical recommendations

- Run a test check with `test/sign` immediately after registering a managed key.
- Do not assign broader `usages` than your scenario requires.
- Explicitly restrict key usage by mount path if you do not need the `any_mount` mode.
- Before deleting a managed key, make sure that it is no longer used by `PKI`, `SSH`, or `Transit`.
- For production environments, document the mapping between the managed key in Stronghold and the external key in the HSM or KMS.
