---
title: "Key management"
description: "Runbook for Stronghold key operations: rekeying unseal and recovery keys, rotating the encryption key, generating and revoking the root token, and responding to key compromise or loss."
weight: 50
---

This page describes routine and emergency operations with Stronghold keys. For the key model, see [Seal](../../../concepts/seal/).

| Key | Purpose | Replacement operation |
| --- | --- | --- |
| Unseal key (Shamir shares) | Decrypts the root key during manual unsealing | `d8 stronghold operator rekey` |
| Recovery key (shares) | Authorizes operator actions with auto unseal via HSM/KMS | `d8 stronghold operator rekey -target=recovery` |
| Root key | Decrypts the keyring | Changed by rekey |
| Encryption key (keyring) | Encrypts data in storage | `d8 stronghold operator rotate` |
| Root token | Unrestricted API access | `d8 stronghold operator generate-root`, revocation |

{{< alert level="warning" >}}
In DP `Automatic` mode, initialization creates a single key share with a threshold of 1. The unseal key and root token are stored in the `stronghold-keys` secret (`unsealKey`, `rootToken`) in the `d8-stronghold` namespace. The `stronghold-automatic` Deployment unseals the nodes with them (check period is 60 seconds by default, logs: `d8 k -n d8-stronghold logs deploy/stronghold-automatic`) and configures Stronghold with the root token. In `Manual` mode, there is neither the secret nor the unsealer. Do not rekey or revoke the root token in DP without consulting support: this may break automatic unsealing.
{{< /alert >}}

<!-- TODO(verify): supported rekey and root token revocation procedure for Stronghold in DKP (Automatic mode), updating the stronghold-keys secret. -->

## Rekeying unseal keys

Rekey when key holders change, when you change the number of shares or the threshold, or when a share is suspected to be compromised.

1. Initialize the operation with the new number of shares and threshold. It is recommended to encrypt the new shares with the holders' PGP keys and enable verification:

   ```shell
   d8 stronghold operator rekey -init \
     -key-shares=5 \
     -key-threshold=3 \
     -pgp-keys="alice.asc,bob.asc,carol.asc,dave.asc,erin.asc" \
     -verify
   ```

   The command returns the operation `Nonce`. Share it with the holders of the current shares.

1. Each current share holder runs the command and enters their share:

   ```shell
   d8 stronghold operator rekey -nonce=<NONCE>
   ```

1. Once the threshold is reached, Stronghold outputs the new shares (PGP-encrypted, if keys were provided) and a `Verification Nonce`.
1. Confirm receipt of the new shares: the threshold number of holders enter the **new** shares:

   ```shell
   d8 stronghold operator rekey -verify -nonce=<VERIFICATION_NONCE>
   ```

   The old shares remain valid until verification completes.

1. Distribute the new shares to the holders and destroy the old ones.

Check progress or cancel the operation:

```shell
d8 stronghold operator rekey -status
d8 stronghold operator rekey -cancel
```

## Rekeying recovery keys

With auto unseal via [HSM](../../kms-hsm/hsm/) or KMS, recovery keys are used instead of unseal keys. The procedure is the same; the operation is authorized with the current recovery keys:

```shell
d8 stronghold operator rekey -init -target=recovery -key-shares=5 -key-threshold=3
d8 stronghold operator rekey -target=recovery -nonce=<NONCE>
```

In the API, use the [`/sys/rekey-recovery-key`](../../../reference/api/system/#post-sysrekey-recovery-keyinit) prefix.

## Rotating the encryption key

Rotation adds a new encryption key to the keyring. New data is encrypted with the new key; older keys stay in the keyring to decrypt previously written data. The operation causes no downtime and requires a token with `sudo` on `sys/rotate`.

```shell
d8 stronghold operator rotate
```

Check the current key term and install time:

```shell
d8 stronghold operator key-status
```

Configure automatic rotation via [`/sys/rotate/config`](../../../reference/api/system/#post-sysrotateconfig):

```shell
d8 stronghold write sys/rotate/config enabled=true interval=720h max_operations=3000000000
```

- `interval` — how long after the active key is installed to rotate it;
- `max_operations` — how many encryption operations to perform before rotating.

`interval` is a duration string (`720h`) or a number of seconds, with a minimum of 24 hours. `max_operations` accepts values from 1,000,000 to 3,865,470,566; if not set, the maximum applies.

## Generating a root token

A root token is needed only for emergency operations when administrative access through auth methods is impossible. Generation requires the threshold number of unseal or recovery key shares.

### With a one-time password (OTP)

1. Initialize the operation. Stronghold outputs a `Nonce` and an `OTP`. Save the OTP: it is required to decode the token:

   ```shell
   d8 stronghold operator generate-root -init
   ```

1. Each holder enters their share:

   ```shell
   d8 stronghold operator generate-root -nonce=<NONCE>
   ```

1. After the last share, Stronghold outputs an `Encoded Token`. Decode it:

   ```shell
   d8 stronghold operator generate-root -decode=<ENCODED_TOKEN> -otp=<OTP>
   ```

### With a PGP key

Instead of an OTP, the token can be encrypted with the recipient's PGP key:

```shell
d8 stronghold operator generate-root -init -pgp-key=operator.asc
```

After the shares are entered, decrypt the `Encoded Token`:

```shell
echo "<ENCODED_TOKEN>" | base64 -d | gpg --decrypt
```

Check progress or cancel the operation:

```shell
d8 stronghold operator generate-root -status
d8 stronghold operator generate-root -cancel
```

## Revoking the root token

Revoke the root token after emergency operations:

```shell
d8 stronghold token revoke <ROOT_TOKEN>
```

Or, if the root token is used in the current session:

```shell
d8 stronghold token revoke -self
```

Make sure working administrative access through auth methods is configured before revocation.

## Responding to compromise

### Compromised unseal or recovery key share

One share is not enough to unseal when the threshold is greater than one, but the safety margin is reduced.

1. [Rekey](#rekeying-unseal-keys) with the remaining holders.
1. If the threshold number of shares may be exposed and storage data may be accessed, rekey and [rotate the encryption key](#rotating-the-encryption-key), then review audit logs for suspicious operations.
1. Assess whether Raft snapshots may have been copied: a snapshot together with the threshold number of shares gives access to the data.

### Compromised root token

1. Revoke the token immediately. If the token value is unknown, use its accessor from the audit log:

   ```shell
   d8 stronghold token revoke -accessor <ACCESSOR>
   ```

   Child tokens and leases created by this token are revoked automatically. Orphan tokens are not revoked: find them in the audit log and revoke them separately.

1. Review audit logs for operations performed with this token: changes to policies, auth methods, audit devices.
1. Rotate static secrets that may have been read.

### Compromised regular token

1. Revoke the token:

   ```shell
   d8 stronghold token revoke <TOKEN>
   ```

1. If secrets issued through a specific path are compromised, revoke leases by prefix:

   ```shell
   d8 stronghold lease revoke -prefix database/creds/app-role
   ```

   The `-force` flag removes leases from Stronghold even if revocation fails in the external system. Use it only if the credentials have been removed from the external system manually.

1. For auth methods, revoke all tokens they issued by prefix, for example `d8 stronghold lease revoke -prefix auth/approle/`, and rotate the credentials (for example, the AppRole `secret_id`).

For more on leases, see [Leases](../../../concepts/lease/); for tokens, see [Tokens](../../../concepts/tokens/).

## Lost keys

### Lost unseal key shares

If the holders still have the threshold number of shares, rekey to restore the required number of shares.

If fewer shares than the threshold remain, the data cannot be decrypted after the next seal or restart:

1. Do not restart or seal the running nodes.
1. While the cluster is unsealed, move secrets and configuration to a new cluster with new keys.

<!-- TODO(verify): supported ways to move data from a cluster with lost unseal keys (a Raft snapshot cannot be unsealed without the original keys). -->

### Lost recovery keys

With auto unseal, the cluster keeps unsealing automatically, but operations requiring recovery keys become impossible: root token generation, recovery key rekey, [seal migration](../../../concepts/seal/#seal-migration).

1. Keep working administrative access through auth methods.
1. Plan migrating the data to a new cluster with new recovery keys.

<!-- TODO(verify): whether Stronghold offers a way to recover or reissue recovery keys without the threshold of current shares. -->
