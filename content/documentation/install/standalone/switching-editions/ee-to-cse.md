---
title: "Switching Stronghold from EE to CSE"
description: "How to upgrade a standalone Stronghold EE 1.15.x installation to Stronghold CSE 1.16.0 on single-node and multi-node clusters."
weight: 20
---

Stronghold Enterprise Edition (EE) can be upgraded to Stronghold Certified Security Edition (CSE) in one of the following ways:

- in a standalone installation;
- [in a DP installation](../../../dkp/platform-management/switching-editions/ee-to-cse/).

{{< alert level="warning" >}}
Upgrading from Stronghold EE 1.15.x to Stronghold CSE 1.16.0 is supported. If you use a Stronghold EE version earlier than 1.15.x, first [update to the latest version of the branch](../../../dkp/update/update/).
{{< /alert >}}

{{< alert level="warning" >}}
The service may be temporarily unavailable during the upgrade to Stronghold CSE.
{{< /alert >}}

## Upgrading a standalone installation

Before starting the upgrade, do the following:

1. Check the current Stronghold version:

   ```shell
   stronghold version
   ```

1. Save the unseal keys and the root token to a secure storage.

1. Create a snapshot of the Stronghold cluster. Example:

   ```shell
   stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ls -lh stronghold-*.snap
   ```

   Check that the snapshot has been created:

   ```shell
   ls -lh ./stronghold-*.snap
   ```

   {{< alert level="info" >}}
   Store the resulting files outside the Stronghold cluster and outside the nodes it runs on.
   {{< /alert >}}

1. Prepare the Stronghold CSE 1.16.0 package or binary on each node.

### Upgrading a single-node cluster

To upgrade a single-node cluster, do the following:

1. Check the cluster state before the upgrade:

   ```shell
   stronghold status
   ```

   The following values must be present:

   ```shell
   Initialized             true
   Sealed                  false
   Version                 1.15.0+ee
   HA Mode                 active
   ```

1. Stop the `stronghold` service:

   ```shell
   sudo systemctl stop stronghold
   ```

1. Replace the Stronghold EE binary with the Stronghold CSE 1.16.0 binary.

1. Make sure the binary permissions have not changed, or run:

   ```shell
   sudo chmod 511 /opt/stronghold/stronghold
   sudo chown stronghold:stronghold /opt/stronghold/stronghold
   ```

1. Start the `stronghold` service:

   ```shell
   sudo systemctl start stronghold
   sudo systemctl status stronghold --no-pager
   ```

1. Check that the node has started correctly:

   ```shell
   stronghold status
   stronghold version
   ```

   Check that:

   - the `stronghold` service is in the `Running` state;
   - there are no startup errors;
   - the version is 1.16.0 (CSE).

1. If necessary, [unseal Stronghold](../raft-lost-quorum-recovery/#unseal-stronghold).

### Upgrading a multi-node cluster

A multi-node cluster is upgraded node by node so that the cluster remains available.

1. Determine the current leader node:

   ```shell
   stronghold status
   ```

   The output contains the address of the current leader node in the `HA Cluster` field:

   ```console
   ...
   HA Cluster              https://10.241.32.36:8201
   ...
   ```

   Note the leader node.

1. Upgrade all follower nodes one by one. On each follower node:

   - Stop the service:

     ```shell
     sudo systemctl stop stronghold
     ```

   - Replace the Stronghold EE binary with Stronghold CSE 1.16.0.
     If necessary, restore the permissions:

     ```shell
     sudo chmod 511 /opt/stronghold/stronghold
     sudo chown stronghold:stronghold /opt/stronghold/stronghold
     ```

   - Start the service:

     ```shell
     sudo systemctl start stronghold
     sudo systemctl status stronghold --no-pager
     ```

   - If necessary, [unseal the Stronghold node](../raft-lost-quorum-recovery/#unseal-stronghold).

   - Check the node state:

     ```shell
     stronghold status
     stronghold version
     ```

   - Check that:

     - the `stronghold` service is in the `Running` state;
     - there are no startup errors;
     - the version is 1.16.0 (CSE).

   - Move on to the next follower node only after the current one has been checked successfully.

1. After upgrading all follower nodes, change the leader node:

   ```shell
   stronghold operator step-down
   # Wait for a while.
   stronghold status
   ```

   Expected result: the leader node has changed, and the `HA Cluster` field shows a different value.

1. Upgrade the former leader node:

   ```shell
   sudo systemctl stop stronghold
   # Replace the Stronghold EE binary with Stronghold CSE 1.16.0.
   # If necessary, restore the permissions:
   # sudo chmod 511 /opt/stronghold/stronghold
   # sudo chown stronghold:stronghold /opt/stronghold/stronghold
   sudo systemctl start stronghold
   sudo systemctl status stronghold --no-pager
   stronghold status
   stronghold version
   ```

   Check that:

   - the `stronghold` service is in the `Running` state;
   - there are no startup errors;
   - the version is 1.16.0 (CSE).

1. If necessary, [unseal the last Stronghold node](../raft-lost-quorum-recovery/#unseal-stronghold).

1. Perform the final cluster check:

   ```shell
   stronghold status
   stronghold version
   ```

   Check that:

   - all cluster nodes are available;
   - there are no replication errors or quorum problems;
   - version 1.16.0 (CSE) is installed on the nodes.
