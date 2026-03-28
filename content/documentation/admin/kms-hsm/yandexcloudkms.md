---
title: "Yandex Cloud KMS"
weight: 15
description: "Configure Stronghold auto-unseal with the yandexcloudkms seal."
---

Stronghold supports `seal "yandexcloudkms"` for auto-unseal and root key protection through Yandex Cloud KMS.

{{< alert level="warning" >}}
`seal "yandexcloudkms"` is currently supported only for Standalone Stronghold installations.
{{< /alert >}}

Among cloud KMS seal integrations, Stronghold supports `yandexcloudkms`. `awskms` and `gcpckms` configurations are not supported in Stronghold.

## What `seal "yandexcloudkms"` does

The `seal "yandexcloudkms"` configuration allows Stronghold to:

- use Yandex Cloud KMS for encryption and decryption operations related to the root key;
- automatically unseal after restart without manual unseal key entry;
- rely on an external KMS instead of locally managed key material.

If double encryption is also enabled, the external KMS must be available not only during unseal, but during normal Stronghold runtime as well.

## Configuration examples

```hcl
seal "yandexcloudkms" {
  kms_key_id  = "abj1abc23def456ghi78"
  oauth_token = "y0_AQAAAA..."
}
```

Example using a service account key:

```hcl
seal "yandexcloudkms" {
  kms_key_id                = "abj1abc23def456ghi78"
  service_account_key_file = "/etc/stronghold/yc-sa-key.json"
}
```

## `seal "yandexcloudkms"` parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `kms_key_id` | yes | ID of the symmetric key in Yandex Cloud KMS. |
| `oauth_token` | no | OAuth token used for Yandex Cloud authentication. |
| `service_account_key_file` | no | Path to the JSON authorized key file of a service account. |
| `endpoint` | no | Custom Yandex Cloud API endpoint. |
| `disabled` | no | Used during migration from one seal mechanism to another. |

Important details:

- `kms_key_id` is required;
- `oauth_token` and `service_account_key_file` are mutually exclusive;
- if neither `oauth_token` nor `service_account_key_file` is set, Stronghold attempts to use the VM service account through instance metadata;
- if `endpoint` is omitted, the default Yandex Cloud SDK endpoint is used.

## Authentication order

For `yandexcloudkms`, authentication values are resolved in the following order:

1. Environment variables.
2. Stronghold configuration file values.
3. Yandex Cloud VM service account credentials.

This means environment variables take precedence over values from the `seal` block.

## Environment variables

The following environment variables are supported:

- `YANDEXCLOUD_KMS_KEY_ID`
- `YANDEXCLOUD_OAUTH_TOKEN`
- `YANDEXCLOUD_SERVICE_ACCOUNT_KEY_FILE`
- `YANDEXCLOUD_ENDPOINT`

You can use them instead of the corresponding configuration file parameters, or combine them if that fits your operational model.

## Access requirements

During initialization, Stronghold checks that the configured key exists and that the process has permission to perform encryption with it.

In practice, `seal "yandexcloudkms"` requires:

- an existing symmetric key in Yandex Cloud KMS;
- permissions to encrypt and decrypt with that key;
- valid authentication through an OAuth token, a service account key file, or a VM service account.

## Practical recommendations

- For production environments, prefer a VM service account or a dedicated service account with minimum required permissions.
- Do not configure `oauth_token` and `service_account_key_file` at the same time.
- When rotating KMS keys, plan for rewrap operations and verify access to older key material versions.
- If you use double encryption, account for the availability of Yandex Cloud KMS during normal Stronghold runtime.

## See also

- [HSM support](./hsm/)
- [Double encryption](./sealwrap/)
- [Standalone configuration](../standalone/configuration/)
