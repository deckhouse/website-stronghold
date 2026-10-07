---
title: "Auto-unseal with Yandex Cloud KMS"
linkTitle: "Auto-unseal with Yandex Cloud KMS"
description: "A standalone Stronghold server with auto-unseal through Yandex Cloud KMS: key and service account in yc CLI, the seal yandexcloudkms stanza, initialization with recovery keys, migration from Shamir, and verification after a restart."
weight: 20
params:
  edition: ee
  relatedLinks:
    - title: "Yandex Cloud KMS"
      url: ../../../admin/kms-hsm/yandexcloudkms/
    - title: "Seal and unseal"
      url: ../../../concepts/seal/
    - title: "Standalone configuration templates"
      url: ../../../install/standalone/topologies/
    - title: "Key management operations"
      url: ../../../admin/operations/key-management/
---

{{< alert level="warning" >}}
Auto-unseal through Yandex Cloud KMS is available only in Stronghold EE.
{{< /alert >}}

With auto-unseal, the Stronghold root key is encrypted with a Yandex Cloud KMS symmetric key. After a restart, Stronghold decrypts the root key through KMS on its own, and no unseal keys are needed. Instead of unseal keys, the operator receives recovery keys for administrative operations.

{{< alert level="warning" >}}
`seal "yandexcloudkms"` is supported only in a standalone Stronghold installation.
{{< /alert >}}

![Auto-unseal flow with Yandex Cloud KMS](../../../images/ex-yandex-kms-unseal.en.png)

## Goal

Create a KMS key and a service account with minimal permissions, configure `seal "yandexcloudkms"`, initialize a new server with recovery keys or move an existing server from Shamir to KMS, and make sure Stronghold unseals automatically after a restart.

## Prerequisites

- A standalone Stronghold server installed as described in [Installation](../../../install/standalone/installation/). In the examples, the configuration is in `/opt/stronghold/config.hcl` and the service is named `stronghold`.
- Network access from the Stronghold nodes to the Yandex Cloud KMS API on port `443`/TCP.
- The [yc CLI](https://yandex.cloud/en/docs/cli/) configured for a Yandex Cloud folder, with permissions to create KMS keys and service accounts in it.
- The `jq` utility.
- To migrate an existing server: the threshold number of Shamir unseal keys and a data backup.

## Step 1. Create a KMS key

1. Create a symmetric key:

   ```bash
   yc kms symmetric-key create \
     --name stronghold-unseal \
     --default-algorithm aes-256 \
     --rotation-period 8760h
   ```

1. Save the key ID:

   ```bash
   KMS_KEY_ID=$(yc kms symmetric-key get stronghold-unseal --format json | jq -r '.id')
   echo "$KMS_KEY_ID"
   ```

{{< alert level="danger" >}}
Deleting the key or its versions makes the root key impossible to decrypt, and Stronghold data is lost. Enable key deletion protection (`--deletion-protection`) and restrict permissions to manage the key.
{{< /alert >}}

## Step 2. Create a service account

1. Create a service account:

   ```bash
   yc iam service-account create --name stronghold-unseal
   ```

1. Grant the account the role to encrypt and decrypt with this key only:

   ```bash
   yc kms symmetric-key add-access-binding stronghold-unseal \
     --role kms.keys.encrypterDecrypter \
     --service-account-name stronghold-unseal
   ```

   <!-- TODO(verify): whether kms.keys.encrypterDecrypter is enough to check that the key exists during initialization or kms.keys.viewer is also required -->

## Step 3. Choose an authentication method

Choose one of the options.

### Option A. VM service account (recommended)

If Stronghold runs on a VM in Yandex Cloud, attach the service account to the VM. Stronghold gets a token through the metadata service, and no secrets are needed in the configuration:

```bash
yc compute instance update <vm-name> --service-account-name stronghold-unseal
```

### Option B. Authorized service account key

If Stronghold runs outside Yandex Cloud, create an authorized key and place it on the node:

```bash
yc iam key create \
  --service-account-name stronghold-unseal \
  --output yc-sa-key.json
sudo install -o stronghold -g stronghold -m 0400 yc-sa-key.json /etc/stronghold/yc-sa-key.json
rm yc-sa-key.json
```

The `oauth_token` parameter is also supported, but an OAuth token is tied to a user and is not suitable for production.

## Step 4. Add the seal stanza to the configuration

Add a `seal` stanza to `/opt/stronghold/config.hcl` on each node.

For option A:

```hcl
seal "yandexcloudkms" {
  kms_key_id = "<KMS_KEY_ID>"
}
```

For option B:

```hcl
seal "yandexcloudkms" {
  kms_key_id               = "<KMS_KEY_ID>"
  service_account_key_file = "/etc/stronghold/yc-sa-key.json"
}
```

Instead of configuration parameters, you can use the `YANDEXCLOUD_KMS_KEY_ID`, `YANDEXCLOUD_SERVICE_ACCOUNT_KEY_FILE`, `YANDEXCLOUD_OAUTH_TOKEN`, and `YANDEXCLOUD_ENDPOINT` environment variables. Environment variables take precedence over the configuration. The full parameter list is in [Yandex Cloud KMS](../../../admin/kms-hsm/yandexcloudkms/).

For a new server, go to step 5. For a server already initialized with Shamir, go to step 6.

## Step 5. Initialize a new server

1. Start Stronghold:

   ```bash
   sudo systemctl start stronghold
   ```

1. Initialize Stronghold. With auto-unseal, recovery keys are created instead of unseal keys:

   ```bash
   d8 stronghold operator init \
     -recovery-shares=5 \
     -recovery-threshold=3
   ```

   Distribute the recovery keys among key holders and store the root token. Recovery keys do not unseal Stronghold, but they are required to generate a root token, rekey, and migrate the seal.

1. For an HA cluster, start the other nodes with the same `seal` stanza. They join the cluster through `retry_join` and unseal automatically.

## Step 6. Move an existing server from Shamir to KMS

Seal migration requires brief downtime. Both the Shamir keys and KMS must be available during the migration.

1. Take a Raft snapshot:

   ```bash
   d8 stronghold operator raft snapshot save pre-seal-migration.snap
   ```

1. Add the `seal "yandexcloudkms"` stanza to the configuration (step 4) and restart Stronghold:

   ```bash
   sudo systemctl restart stronghold
   ```

   The log shows a migration mode message:

   ```bash
   journalctl -u stronghold.service | grep "seal migration"
   ```

   ```text
   core: entering seal migration mode; Stronghold will not automatically unseal even if using an autoseal: from_barrier_type=shamir to_barrier_type=yandexcloudkms
   ```

1. Unseal with the `-migrate` flag, entering the threshold number of unseal keys:

   ```bash
   d8 stronghold operator unseal -migrate
   ```

   Once the threshold is reached, the Shamir unseal keys become recovery keys.

For an HA cluster, migrate node by node starting with standby nodes, and step down the active node as described in [Seal migration](../../../concepts/seal/#seal-migration).

## Verification

1. Check the status:

   ```bash
   d8 stronghold status
   ```

   Expected values:

   ```text
   Seal Type                yandexcloudkms
   Recovery Seal Type       shamir
   Initialized              true
   Sealed                   false
   ```

1. Restart the service and make sure Stronghold unseals without entering keys:

   ```bash
   sudo systemctl restart stronghold
   sleep 10
   d8 stronghold status -format=json | jq '{type, sealed, recovery_seal}'
   ```

   The `sealed` field must be `false`.

1. If Stronghold remains sealed, check the log for Yandex Cloud authentication or key access errors:

   ```bash
   journalctl -u stronghold.service --since "10 minutes ago"
   ```

1. Check that the recovery keys work by starting and cancelling root token generation:

   ```bash
   d8 stronghold operator generate-root -init
   d8 stronghold operator generate-root -cancel
   ```

## Operations

- During scheduled KMS key rotation, old versions are kept and remain usable for decryption. Do not delete key versions. Re-encryption of data after rotation is described in [Double encryption](../../../admin/kms-hsm/sealwrap/).
- KMS unavailability does not affect an unsealed server without double encryption, but a restarted node does not unseal while KMS is unavailable.
- Rekeying and storing recovery keys are described in [Key management operations](../../../admin/operations/key-management/).

## Cleanup

For a test environment:

1. Stop Stronghold and delete its data.
1. Delete the authorized key: `yc iam key list --service-account-name stronghold-unseal`, then `yc iam key delete <id>`.
1. Delete the service account: `yc iam service-account delete stronghold-unseal`.
1. Delete the KMS key: `yc kms symmetric-key delete stronghold-unseal`. Do this only after deleting the Stronghold data encrypted with this key.
