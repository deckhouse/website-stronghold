---
title: "FAQ"
description: "Answers to frequently asked questions about operating Stronghold: unsealing, keys, backups, monitoring, upgrades, performance, and DKP."
weight: 90
---

## Unsealing and keys

### Why is a node sealed after a restart?

With Shamir seal, this is expected: the decryption key is kept only in memory. Unseal the node with `d8 stronghold operator unseal` or configure auto unseal via [HSM or KMS](../../kms-hsm/hsm/). See [Troubleshooting](../troubleshooting/#node-is-sealed-after-a-restart).

### Do I need to unseal every cluster node?

Yes. With Shamir seal, each node is unsealed separately after every start. With auto unseal, nodes are unsealed automatically. See [Seal](../../../concepts/seal/).

### How do I change unseal key holders?

Rekey with a new number of shares and the new holders' PGP keys. See [Key management](../key-management/#rekeying-unseal-keys).

### Do I need to keep the root token?

No. Revoke the root token after the initial setup and generate a new one only for emergency operations. In DP `Automatic` mode, the root token is stored in the `stronghold-keys` secret and used by `stronghold-automatic`, so do not revoke it (see [Production checklist](../production-checklist/#access-and-authentication)). See [Key management](../key-management/#generating-a-root-token).

### What if the root token is lost?

Generate a new one with `d8 stronghold operator generate-root` using the threshold number of unseal or recovery keys. See [Generating a root token](../key-management/#generating-a-root-token).

### How often should the encryption key be rotated?

Configure automatic rotation via `sys/rotate/config` by interval or number of operations, and rotate out of schedule when compromise is suspected. See [Rotating the encryption key](../key-management/#rotating-the-encryption-key).

## Backup and restore

### How do I back up Stronghold?

For integrated Raft storage, take a snapshot with `d8 stronghold operator raft snapshot save`. See [Creating a snapshot](../../backups/save/) and [Automated snapshots](../../backups/automated-snapshots/).

### Is a snapshot enough to restore?

No. To unseal a restored cluster, you need the unseal or recovery keys (or HSM/KMS access) that were valid when the snapshot was taken. See [Restoring from a snapshot](../../backups/restore/).

### What if most Raft nodes are lost?

Recover the quorum following the guide for [Linux](../../../install/standalone/raft-lost-quorum-recovery/) or [DP](../../../install/dkp/raft-lost-quorum-recovery/).

## Monitoring and logs

### How do I connect Prometheus?

Enable telemetry with `prometheus_retention_time` and scrape `/v1/sys/metrics?format=prometheus`. See [Monitoring](../monitoring/).

### Which endpoint should load balancer checks use?

Use `/v1/sys/health`: it requires no authentication and returns different codes for active, standby, and sealed nodes. See [Health checks via sys/health](../monitoring/#health-checks-via-syshealth).

### How are server logs different from audit logs?

Server logs describe the Stronghold process itself; audit logs record client requests and responses. See [Server logs](../logs/) and [Audit in Stronghold](../../audit/overview/).

### How do I enable debug logs without a restart?

Use the `/sys/loggers` endpoint or send `SIGHUP` after changing `log_level`. See [Changing the log level](../logs/#changing-the-log-level).

## Performance and load

### Why does response time grow?

A frequent cause is a large number of leases and tokens or a slow Raft disk. See [Slow responses and lease explosion](../troubleshooting/#slow-responses-and-lease-explosion).

### How do I protect the cluster from an application that generates excessive load?

Configure rate limit [quotas](../quotas/) for an auth method, mount, or role.

### How many nodes does a cluster need?

For production, 3 or 5 nodes. See [Sizing](../sizing/#number-of-nodes).

### Why do requests hang when audit has problems?

Stronghold does not serve a request if no audit device can record it. Use at least two audit devices. See [Audit device blocks requests](../troubleshooting/#audit-device-blocks-requests).

## Upgrades and certificates

### How do I upgrade a Stronghold cluster in Linux?

Take a snapshot, upgrade standby nodes one at a time, then hand over the active role with `d8 stronghold operator step-down` and upgrade the former active node. See [Upgrade](../../../install/standalone/update/).

### How do I replace a TLS certificate without downtime?

Replace the files at the same paths and send `SIGHUP` to the process. See [TLS certificates](../tls-certificates/).

### How is Stronghold updated in DP?

The module is updated together with Deckhouse according to the [release channel](../../../install/dkp/release-channels/). See [Platform update](../../../install/dkp/update/update/).

## Security

### What should I check before going to production?

Go through the [production checklist](../production-checklist/).

### What should I do if a token is compromised?

Revoke the token and leases by prefix, and review the audit logs. See [Responding to compromise](../key-management/#responding-to-compromise).
