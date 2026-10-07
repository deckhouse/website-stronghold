---
title: "Upgrade"
description: "Upgrading a Stronghold HA cluster with Raft storage in Linux without downtime: backup, upgrading standby nodes, handing over the active role, and checking versions."
weight: 30
---

A Stronghold HA cluster with Raft storage is upgraded one node at a time. Standby nodes are upgraded first; then the active node hands over its role to one of the upgraded nodes and is upgraded last. This order keeps the quorum and the cluster available during the upgrade.

<!-- TODO(verify): supported upgrade paths (whether versions can be skipped) and whether rollback to the previous version is possible after an upgrade. -->

## Preparation

1. Read the [release notes](../../../release-notes/) for all versions between the current and target ones. Pay attention to configuration changes and breaking changes.
1. Check the cluster version and state:

   ```shell
   d8 stronghold status
   d8 stronghold operator raft list-peers
   d8 stronghold operator raft autopilot state
   ```

   All nodes must be `voter`, and Autopilot must report `Healthy: true`.

1. Take a storage snapshot and inspect it (see [Creating a snapshot](../../../admin/backups/save/) and [Inspecting a snapshot](../../../admin/backups/inspect/)):

   ```shell
   d8 stronghold operator raft snapshot save pre-upgrade.snap
   d8 stronghold operator raft snapshot inspect pre-upgrade.snap
   ```

1. Back up the configuration file, systemd unit, and current binary on each node:

   ```shell
   cp -a /opt/stronghold/stronghold /opt/stronghold/stronghold.bak
   cp -a /opt/stronghold/config.hcl /opt/stronghold/config.hcl.bak
   ```

1. If Shamir seal is used, make sure unseal key holders are available: every node must be unsealed after a restart.
1. Copy the new binary to all nodes, for example to `/tmp/stronghold`, and check its version:

   ```shell
   /tmp/stronghold version
   ```

1. Identify the active node: in the `d8 stronghold operator raft list-peers` output, its state is `leader`.

## Upgrading standby nodes

Perform the steps on each standby node in turn. Do not move to the next node until the current one has rejoined the cluster.

1. Replace the binary:

   ```shell
   install -o root -g root -m 0755 /tmp/stronghold /opt/stronghold/stronghold
   ```

1. Restart the service:

   ```shell
   systemctl restart stronghold
   ```

1. If Shamir seal is used, unseal the node. Point to the node being upgraded:

   ```shell
   export STRONGHOLD_ADDR=https://raft-node-2.demo.tld:8200
   d8 stronghold operator unseal
   ```

1. Check that the node has rejoined the cluster and is unsealed:

   ```shell
   d8 stronghold status
   d8 stronghold operator raft autopilot state
   ```

   Wait until Autopilot shows the node as `healthy` and the cluster as `Healthy: true`.

1. Check the node [server logs](../../../admin/operations/logs/) for errors:

   ```shell
   journalctl -u stronghold.service --since "10 minutes ago"
   ```

## Upgrading the active node

1. Hand over the active role to one of the upgraded nodes. Run the command against the current active node:

   ```shell
   export STRONGHOLD_ADDR=https://raft-node-1.demo.tld:8200
   d8 stronghold operator step-down
   ```

1. Make sure another node has become the leader:

   ```shell
   d8 stronghold operator raft list-peers
   ```

1. Upgrade the former active node the same way as standby nodes: replace the binary, restart the service, unseal the node, and wait until it rejoins the cluster.

## Post-upgrade checks

1. Check the version on each node:

   ```shell
   for node in raft-node-1 raft-node-2 raft-node-3; do
     STRONGHOLD_ADDR=https://${node}.demo.tld:8200 d8 stronghold status | grep -E 'Version|HA Mode|Sealed'
   done
   ```

1. Check the cluster state:

   ```shell
   d8 stronghold operator raft list-peers
   d8 stronghold operator raft autopilot state
   ```

1. Check the main scenarios: login via the auth methods in use, reading and writing a test secret, issuing dynamic credentials.
1. Check that metrics and alerts work (see [Monitoring](../../../admin/operations/monitoring/)).

The `autopilot state` output shows each node version (`Version`), which lets you confirm that all nodes are upgraded. Autopilot automated upgrades (upgrade migration) are marked as an Enterprise-edition feature in the source code and are not used in this procedure.

## Rollback

If a node fails to start after the upgrade or the cluster misbehaves:

1. Stop the upgraded node and restore the previous binary from `/opt/stronghold/stronghold.bak`.
1. If the new version has already modified the storage data and the previous version does not start, restore the cluster from the `pre-upgrade.snap` snapshot on the previous version (see [Restoring from a snapshot](../../../admin/backups/restore/)).

{{< alert level="warning" >}}
Restoring from a snapshot returns data to the moment the snapshot was taken. Changes made after the snapshot are lost.
{{< /alert >}}
