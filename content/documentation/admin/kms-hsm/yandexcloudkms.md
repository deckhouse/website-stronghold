---
title: "Yandex Cloud KMS"
weight: 15
description: "Configure Stronghold auto-unseal with the yandexcloudkms seal."
params:
  relatedLinks:
    - title: "Configuring standalone Stronghold server"
      url: ../../../install/standalone/configuration/
---

Stronghold supports automated unsealing and root key protection through [Yandex Cloud KMS](https://yandex.cloud/en/docs/kms/).
To integrate with Yandex Cloud KMS, use the `seal "yandexcloudkms"` configuration section.

{{< alert level="warning" >}}
Currently, `seal "yandexcloudkms"` is only supported in standalone Stronghold installations.
{{< /alert >}}

Among external cloud KMS integrations, Stronghold supports only Yandex Cloud KMS. The `awskms` and `gcpckms` configurations are not supported.

## Yandex Cloud KMS features

The `seal "yandexcloudkms"` configuration allows:

- Using Yandex Cloud KMS for data encryption and decryption related to the root key.
- Automatically unsealing Stronghold after restarting without a manual unseal key entry.
- Using an external KMS instead of managing key material locally.

If double encryption is enabled in configuration, Yandex Cloud KMS must be available not only during unseal, but during normal Stronghold operation as well.

## Configuration example

```hcl
seal "yandexcloudkms" {
  kms_key_id  = "abj1abc23def456ghi78"
  oauth_token = "y0_AQAAAA..."
}
```

Example using a service account:

```hcl
seal "yandexcloudkms" {
  kms_key_id                = "abj1abc23def456ghi78"
  service_account_key_file = "/etc/stronghold/yc-sa-key.json"
}
```

## Yandex Cloud KMS parameters

| Parameter | Required | Description |
| --- | --- | --- |
| `kms_key_id` | Yes | ID of the symmetric key in Yandex Cloud KMS |
| `oauth_token` | No | OAuth token used for Yandex Cloud authentication. Can't be specified together with `service_account_key_file` |
| `service_account_key_file` | No | Path to the JSON authorized key file of a service account. Can't be specified together with `oauth_token` |
| `endpoint` | No | Custom Yandex Cloud API endpoint. If this parameter is not set, a standard Yandex Cloud SDK endpoint is used instead |
| `disabled` | No | Used during migration from one seal mechanism to another |

{{< alert level="info" >}}
If neither `oauth_token` nor `service_account_key_file` is set, Stronghold attempts to use the VM service account through instance metadata.
{{< /alert >}}

## Authentication credentials order

For Yandex Cloud KMS, authentication values are resolved in the following order:

1. Environment variables.
2. Stronghold configuration file values.
3. Yandex Cloud VM service account credentials.

Therefore, environment variables take precedence over values from the `seal "yandexcloudkms"` section.

## Environment variables

The following environment variables are supported:

- `YANDEXCLOUD_KMS_KEY_ID`
- `YANDEXCLOUD_OAUTH_TOKEN`
- `YANDEXCLOUD_SERVICE_ACCOUNT_KEY_FILE`
- `YANDEXCLOUD_ENDPOINT`

You can use them instead of the corresponding configuration file parameters, or combine them if that fits your operational model.

## Access requirements

During initialization, Stronghold checks that the configured key exists and that the process has permission to perform encryption with it.

ENsure that you meet the following requirements for `seal "yandexcloudkms"` configuration to work:

- A symmetric key is present in Yandex Cloud KMS.
- Permissions to encrypt and decrypt with that key have been obtained.
- There is a valid authentication through an OAuth token, a service account key file, or a VM service account.

## Recommendations

- For production environments, prefer a VM service account or a dedicated service account with minimum required permissions.
- When rotating KMS keys, plan for rewrap operations and verify access to older key material versions.
- If you use double encryption, account for the availability of Yandex Cloud KMS during normal Stronghold runtime.
