---
title: "Seal/Unseal"
description: "Stronghold must be unsealed before it can access its data. Likewise, it can be sealed to lock it down."
weight: 25
---

When you start a Stronghold server, it starts in a sealed state. In this state, Stronghold can access the physical storage, but it cannot decrypt any of the data on it.

Unsealing is the process of obtaining the plaintext root key that is necessary to read the decryption key. Prior to unsealing, the only possible Stronghold operations are to unseal Stronghold and check the status of the server.

## Why?

Stronghold encrypts the data using an encryption key (in the keyring) and stores them in its [storage backend](../../install/standalone/configuration/#storage). To protect this encryption key, Stronghold encrypts it using another encryption key known as the root key and stores it with the data.

To decrypt the data, Stronghold needs the root key so that it can decrypt the encryption key. Unsealing is the process of getting access to this root key. Stronghold encrypts the root key using the unseal key, and stores it alongside all other Stronghold data.

To summarize, Stronghold encrypts most data using the encryption key in the keyring. To get the keyring, Stronghold uses the root key to decrypt it. The root key itself requires the unseal key to decrypt it.

## Shamir seals

Instead of distributing the unseal key to an operator as a single key, the default Stronghold configuration uses an algorithm known as [Shamir's Secret Sharing](https://en.wikipedia.org/wiki/Shamir%27s_Secret_Sharing) to split the key into shares.

Stronghold requires a certain threshold of shares to reconstruct the unseal key. Stronghold operators add shares one at a time in any order until Stronghold has enough shares to reconstruct the key. Then, Stronghold uses the unseal key to decrypt the root key. This is the Stronghold unseal process.

## Unsealing

You can run the unseal process using the `d8 stronghold operator unseal` command or the API.
The unseal process is stateful. You can enter each key using multiple mechanisms from
multiple client machines. The process allows each share of the root
key to reside on a distinct client machine for better security.

When you use the Shamir seal with multiple nodes, you must unseal each node with the required threshold of shares. Stronghold does not distribute partially unsealed nodes across the cluster.

An exception is the [`seal "inner-cluster"`](#auto-unseal-with-inner-cluster) mechanism: unsealed cluster nodes can unseal other nodes listed in the configuration.

Once you unseal a Stronghold node, it remains unsealed until one of the following happens:

1. You reseal it using the API.
1. You restart the server.
1. Stronghold's storage layer encounters an unrecoverable error.

Unsealing nodes can make automating a Stronghold installation
difficult. Automated tools can easily install, configure, and start Stronghold,
but unsealing it using Shamir seals without additional automation is a manual process. For most users,
[auto unseal](#auto-unseal) via an external KMS or HSM, or the built-in [`inner-cluster`](#auto-unseal-with-inner-cluster) mechanism, provides a better experience.

## Sealing

When you seal Stronghold through the API, Stronghold deletes the root
key from memory and requires another unseal process to restore it. Sealing Stronghold requires only a single operator with root privileges.

If there is a detected intrusion, you can seal the Stronghold data quickly to try to minimize damages. Users cannot access the data from a sealed Stronghold without access to the root key shares.

## Auto unseal

Auto unseal aids in reducing the operational complexity of
keeping the unseal key secure. Stronghold supports two approaches:

1. An external trusted service or device (KMS, HSM): at startup, Stronghold connects to it and prompts it to decrypt the root key from storage.
1. The built-in hybrid [`seal "inner-cluster"`](#auto-unseal-with-inner-cluster) mechanism: cluster nodes unseal each other based on Shamir, without an external KMS.

There are certain Stronghold operations besides unsealing that
require a quorum of users to perform, for example, generating a root token. When
you use a Shamir seal, you must provide the unseal keys to authorize these
operations. When you use auto unseal via an external KMS or HSM, these operations require recovery
keys instead.

Just as the Shamir seal initialization process yields unseal keys,
initializing with auto unseal based on an external KMS or HSM yields recovery keys.

It is still possible to seal a Stronghold node using the API. When you use the API, Stronghold remains sealed until you restart it or use the unseal API. Auto
unseal via KMS or HSM requires the recovery key fragments instead of the unseal key fragments
that the Shamir seal provides. The process remains the same.

For a list of examples and supported providers of external auto unseal, refer to
[KMS and HSM](../../admin/kms-hsm/).

{{< alert level="warning" >}}
Recovery keys cannot decrypt the root key and therefore are not sufficient to unseal
Stronghold if the KMS- or HSM-based auto unseal mechanism isn't working.
Using such auto unseal creates a strict Stronghold lifecycle dependency on the underlying seal mechanism.
If a seal mechanism such as the Cloud KMS key becomes unavailable
or is deleted before you migrate the seal, you cannot recover
access to the Stronghold cluster until the mechanism is available again.

**If the seal mechanism or its keys are permanently deleted, then the Stronghold cluster cannot be recovered, even from backups.**
To mitigate this risk, we recommend careful controls around managing the seal mechanism.

Secondary replication clusters can have the seal configuration independently of the primary cluster. When you properly configure the secondary cluster's seal, it guards against some of the risk. However, you could still lose unreplicated items such as local mounts.
{{< /alert >}}

## Auto unseal with inner-cluster

In Enterprise Edition (EE) and Certified Security Edition (CSE), Stronghold supports built-in automatic cluster unsealing without external services or KMS.

The `seal "inner-cluster"` mechanism is hybrid: it relies on Shamir, but unsealed cluster nodes unseal other nodes listed in the configuration. When a node is unsealed, it periodically polls the nodes specified in the configuration and submits Shamir key shares to them.

### Configuration

To enable the mechanism, add a `seal "inner-cluster"` section to the configuration file.
Inside it, describe the nodes that participate in unsealing in `node` blocks.

`node` block parameters:

| Parameter | Description |
| --- | --- |
| `name` | Node name. We recommend using the host name or the name of the pod where Stronghold runs. |
| `address` | Cluster node address. |
| `token` | Node token. Optional. |
| `tls_ca_cert` | Path to the CA certificate file used to verify the node certificate. |
| `tls_skip_verify` | If `true`, certificate verification is skipped. |

Configuration example:

```hcl
seal "inner-cluster" {
  node {
    name            = "stronghold-1"
    address         = "http://127.0.0.1:8200"
    token           = ""
    tls_ca_cert     = ""
    tls_skip_verify = true
  }
  node {
    name            = "stronghold-2"
    address         = "http://127.0.0.2:8200"
    token           = ""
    tls_ca_cert     = ""
    tls_skip_verify = true
  }
  node {
    name            = "stronghold-3"
    address         = "http://127.0.0.3:8200"
    token           = ""
    tls_ca_cert     = ""
    tls_skip_verify = true
  }
}
```

## Recovery key

When you initialize Stronghold using an HSM or KMS, Stronghold returns recovery keys to the operator rather than unseal keys.
Stronghold generates recovery keys from an internal recovery key that is split via Shamir's Secret Sharing, similar
to Stronghold's treatment of unseal keys when you run it without an HSM or KMS.

When you perform an operation such as `generate-root`
that uses recovery keys, Stronghold automatically selects the recovery
keys for this purpose, rather than the barrier unseal keys.

### Initialization

When you initialize Stronghold, it performs the split according to the following CLI flags
and their API equivalents in the [`/sys/init`](../../reference/api/system/#post-sysinit) endpoint:

- `recovery-shares`: The number of shares into which to split the recovery
  key. This value is equivalent to the `recovery_shares` value in the API
  endpoint.
- `recovery-threshold`: The threshold of shares required to reconstruct the
  recovery key. This value is equivalent to the `recovery_threshold` value in
  the API endpoint.
- `recovery-pgp-keys`: The PGP keys to use to encrypt the returned recovery
  key shares. This value is equivalent to the `recovery_pgp_keys` value in the
  API endpoint, although as with `pgp_keys` the object in the API endpoint is
  an array, not a string.

If you don't set these options, Stronghold does not initialize; therefore, it will not generate the keys. Refer to
[HSM](../../admin/kms-hsm/hsm/) for more details.

### Rekeying

#### Unseal key

You can rekey Stronghold's unseal keys using a `d8 stronghold operator rekey` operation from the CLI or the matching API calls. Stronghold authorizes the rekey operation when you meet the threshold of recovery keys. After rekeying, Stronghold encrypts the new barrier key with the configured HSM or KMS, and stores it in its storage backend.

{{< alert level="info" >}}
Seal wrap is available in Stronghold Enterprise editions.
{{< /alert >}}

#### Recovery key

You can rekey the recovery key to change the number of shares/threshold or to
target different key holders using different PGP keys. When you use the Stronghold CLI,
it performs the rekey operation by using the `-target=recovery` flag to `d8 stronghold operator rekey`.

Using the API, you perform the rekey operation with the same parameters as the
[normal `/sys/rekey` endpoint](../../reference/api/system/#post-sysrekeyinit); however, the
API prefix for the rekey operation is at `/sys/rekey-recovery-key` rather than
`/sys/rekey`.

## Seal migration

You cannot perform the seal migration process without downtime. Due to the
technical underpinnings of the seal implementations, the process requires that
you briefly take the whole cluster down. While experiencing some downtime may
be unavoidable, switching seals is a rare event and the
inconvenience of the downtime is an acceptable trade-off.

You should make sure to create a backup before you start the seal migration process, in case something goes wrong.

The seal migration operation requires both the old and new seals to be available during the migration. For example, if you migrate from auto unseal to Shamir seals, the service backing the auto unseal must be available during the migration.

### General migration procedure

The following steps are common for seal migrations between any supported kinds and for
any storage backend.

1. Take a standby node down and update the [seal configuration](../../admin/kms-hsm/).

   - If the migration is from Shamir seal to auto seal, add the new auto
     seal block to the configuration.
   - If the migration is from auto seal to Shamir seal, add `disabled = true`
     to the old seal block.
   - If the migration is from auto seal to another auto seal, add `disabled = true` to the old seal block and add the new auto seal block.

   Now, bring the standby node back up and run the unseal command on each key, by
   supplying the `-migrate` flag.

   - Supply Shamir unseal keys if the old seal was Shamir, which is migrated
     as the recovery keys for the auto seal.
   - Supply recovery keys if the old seal is one of auto seals, which is
     migrated as the recovery keys of the new auto seal, or as Shamir unseal
     keys if the new seal is Shamir.

1. Perform step 1 for all the standby nodes, one at a time. You must
   bring back each downed standby node before moving on to the other standby nodes,
   specifically when you use integrated storage, because it helps to retain the
   quorum.

1. [Step down](../../reference/api/system/#post-sysstep-down) the
   active node. One of the standby nodes will become the new active node.
   When you use integrated storage, ensure that Stronghold reaches a quorum and elects a leader.

1. The new active node performs the migration. Monitor the server log in
   the active node to witness the completion of the seal migration process.
   Wait for a little while for the migration information to replicate to all the
   nodes, if you use integrated storage. In Enterprise editions, switching an auto seal
   implies that the seal wrapped storage entries get rewrapped. Monitor the log
   and wait until this process is complete. Look for `seal re-wrap completed` in the log.

{{< alert level="warning" >}}
Any change to the `seal` stanza in your Stronghold configuration invokes the seal-rewrap process,
including migrations from the same auto-unseal type like `pkcs11` to `pkcs11`.
{{< /alert >}}

1. Seal migration is now completed. Take down the old active node, update its
   configuration to use the new seal blocks (completely unaware of the old seal type),
   and bring it back up. It will be auto-unsealed if the new seal is one of the
   auto seals, or will require unseal keys if the new seal is Shamir.

1. You can now update the configuration files of all the nodes to have only the
   new seal information. You can restart standby nodes right away, and you can restart the active
   node upon a leadership change.

### Additional migration scenarios

#### Migration from Shamir to auto unseal

To migrate from Shamir keys to auto unseal, take your server cluster offline and
update the [seal](../../admin/kms-hsm/) with the appropriate
seal configuration. Bring your server back up and leave the rest of the nodes
offline if you use multi-server mode, then run the unseal process with the
`-migrate` flag and bring the rest of the cluster online.

All unseal commands must specify the `-migrate` flag. Once you enter the required
threshold of unseal keys, Stronghold migrates the unseal keys to recovery
keys.

```bash
d8 stronghold operator unseal -migrate
```

#### Migration from auto unseal to Shamir

To migrate from auto unseal to Shamir keys, take your server cluster offline
and update the [seal configuration](../../admin/kms-hsm/) and add `disabled = true` to the seal block. This flag allows the migration to use this information
to decrypt the key, but will not unseal Stronghold.

When you bring your server back
up, run the unseal process with the `-migrate` flag and use the recovery keys
to perform the migration. All unseal commands must specify the `-migrate` flag.
Once you enter the required threshold of recovery keys, Stronghold migrates the recovery keys
that it will use as unseal keys.

#### Migration from auto unseal to auto unseal

{{< alert level="info" >}}
Migration between the same auto unseal types is supported.
In the steps below, you can only migrate from
one type of auto unseal to a different type, for example Transit to AWSKMS.
{{< /alert >}}

To migrate from auto unseal to a different auto unseal configuration, take your
server cluster offline and update the existing [seal
configuration](../../admin/kms-hsm/) and add `disabled = true` to the seal
block. Then add another seal block to describe the new seal.

When you bring your server back up, run the unseal process with the `-migrate`
flag and use the recovery keys to perform the migration. All unseal commands
must specify the `-migrate` flag. Once you enter the required threshold of recovery keys, Stronghold keeps the recovery keys and uses them as recovery keys in the new
seal.

#### Migration with integrated storage

Integrated storage uses the Raft protocol, which requires a quorum of
servers to be online before the cluster is functional. Therefore, bringing the
cluster back up one node at a time with the updated seal configuration does not
work in this case.

Follow the same steps for each kind of migration described
above with the exception that after you take the cluster offline, you should update the
seal configurations of all the nodes appropriately and bring them all back up.
When the quorum of nodes are back up, Raft elects a leader and the leader
node performs the migration. The migrated information is replicated to
all other cluster peers. When the peers eventually become the leader,
migration does not happen again on the peer nodes.
