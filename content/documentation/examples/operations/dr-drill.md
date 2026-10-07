---
title: "Regular disaster recovery drills"
linkTitle: "DR drills"
description: "Running regular drills: restoring a snapshot into an isolated cluster, a test promotion of a DR secondary, a checklist, and frequency."
weight: 110
params:
  relatedLinks:
    - title: "Backups"
      url: ../../../admin/backups/overview/
    - title: "Inspecting a snapshot"
      url: ../../../admin/backups/inspect/
    - title: "Restoring from a snapshot"
      url: ../../../admin/backups/restore/
    - title: "Automated snapshots"
      url: ../../../admin/backups/automated-snapshots/
    - title: "Disaster recovery"
      url: ../../../admin/replication/disaster-recovery/
---

A backup that has never been restored does not guarantee recovery. Regular drills confirm that snapshots are usable, keys are available, the procedure is documented, and recovery time meets targets.

## Goal

Regularly test two recovery scenarios:

- restoring a Raft snapshot into an isolated cluster;
- failing over to a DR secondary (Stronghold EE);

and record the actual recovery time (RTO) and data loss (RPO).

## Frequency

| Check | Frequency |
| --- | --- |
| `snapshot inspect -validate` for a fresh snapshot | After each snapshot or daily |
| Restoring a snapshot into an isolated cluster | Quarterly and after a Stronghold version upgrade |
| Planned failover to a DR secondary (Stronghold EE) | Every six months |
| Checking availability of unseal or recovery key share holders | Quarterly |
| [Break-glass access](../../access/break-glass-access/) drills | Every six months |

## Prerequisites

- A Stronghold cluster with integrated Raft storage.
- An isolated restore environment: a separate network with no access to production databases, LDAP, and other systems managed by secrets engines.
- The unseal or recovery key share holders of the production cluster. After a snapshot restore, the original keys are required.
- For part 2: Stronghold EE and configured [DR replication](../../../admin/replication/disaster-recovery/).

## Part 1. Restoring a snapshot into an isolated cluster

{{< alert level="warning" >}}
The restored cluster contains production secrets and secrets engine configuration. If it gets network access to production systems, it starts revoking expired leases and rotating passwords in real databases and LDAP. Isolate the environment at the network level.
{{< /alert >}}

1. Get a snapshot. Use the latest automated snapshot (Stronghold EE):

   ```bash
   d8 stronghold read sys/storage/raft/snapshot-auto/status/<config_name>
   ```

   The `last_snapshot_url` field points to the latest successful snapshot. Alternatively, save a snapshot manually:

   ```bash
   d8 stronghold operator raft snapshot save drill.snap
   ```

1. Check the snapshot integrity:

   ```bash
   d8 stronghold operator raft snapshot inspect -validate drill.snap
   ```

1. In the isolated environment, deploy a Stronghold cluster of the same or a newer version with integrated Raft storage (see [Installing Stronghold on Linux](../../../install/standalone/installation/)). Initialize and unseal it, and log in with the temporary root token:

   ```bash
   d8 stronghold operator init
   d8 stronghold operator unseal
   d8 stronghold login <temporary_root_token>
   ```

   <!-- TODO(verify): restore procedure when the production cluster uses KMS auto-unseal or Stronghold EE built-in auto-unseal: whether the test environment needs access to the same KMS -->

1. Record the restore start time and restore the snapshot:

   ```bash
   d8 stronghold operator raft snapshot restore -force drill.snap
   ```

1. Unseal the cluster with the original keys of the production cluster:

   ```bash
   d8 stronghold operator unseal
   ```

1. Check the data and record the end time:

   ```bash
   d8 stronghold status
   d8 stronghold secrets list
   d8 stronghold auth list
   d8 stronghold policy list
   d8 stronghold kv get -mount=secret <known_test_secret>
   ```

   Create a control secret with a timestamp in the production cluster in advance: it makes it easy to determine the RPO.

1. Destroy the environment and all snapshot copies in it.

## Part 2. Planned failover to a DR secondary

{{< alert level="info" >}}
DR replication is available only in Stronghold EE.
{{< /alert >}}

Perform the failover in an agreed maintenance window. Do not promote the DR secondary while the primary is active: you would end up with two primaries.

1. Make sure the secondary has caught up with the primary: `last_wal` on the secondary equals `last_remote_wal`:

   ```bash
   d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/dr/status
   ```

1. Demote the current primary:

   ```bash
   d8 stronghold write -address="${PRIMARY_ADDR}" -force sys/replication/dr/primary/demote
   ```

1. Generate a DR operation token on the secondary using the key share procedure and promote the secondary. The commands are given in [Promote a DR secondary](../../../admin/replication/disaster-recovery/#promote-a-dr-secondary).

1. Switch clients to the new primary (DNS record or load balancer) and record the time.

1. Check operation: OIDC login, reading secrets, issuing dynamic credentials and certificates.

1. Return the former primary to the secondary role. On the new primary, issue an activation token and connect the former primary:

   ```bash
   d8 stronghold write -address="${SECONDARY_ADDR}" sys/replication/dr/primary/secondary-token id=dr-old-primary
   d8 stronghold write -address="${PRIMARY_ADDR}" sys/replication/dr/secondary/update-primary token=<activation_token>
   ```

1. If needed, fail back with the same steps.

## Drill checklist

- [ ] The snapshot passed `inspect -validate`.
- [ ] Key share holders are available, and the quorum was gathered in acceptable time.
- [ ] The restored cluster was unsealed with the original keys.
- [ ] The control secret was read, and the RPO was measured.
- [ ] Auth methods, policies, and mounts are present as expected.
- [ ] For DR: the secondary was promoted, clients were switched, and the former primary was connected as a secondary.
- [ ] The recovery time (RTO) was measured and compared with the target.
- [ ] The isolated environment and snapshot copies were destroyed.
- [ ] Issues found were added to the recovery runbook.

## Verification

A drill is successful if the actual RTO and RPO do not exceed the targets and all checklist items are done. Keep a drill record: date, participants, Stronghold version, snapshot used, measured RTO and RPO, and issues found.

## Cleanup

1. Delete the isolated cluster and its storage.
1. Delete local snapshot copies: `shred -u drill.snap`.
1. Revoke tokens issued for the drill on the production cluster.
