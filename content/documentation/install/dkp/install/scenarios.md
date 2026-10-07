---
title: "Installation scenarios"
description: "Typical scenarios for installing Stronghold in Deckhouse Kubernetes Platform: a test cluster with one master node, a production cluster with three master nodes, dedicated nodes, and air-gapped environments."
hidden: true
weight: 40
---

<!-- TODO(verify): the page is hidden until the placement parameters of the stronghold module on dedicated nodes are verified; remove hidden: true after verification. -->

The `stronghold` module runs Stronghold nodes on the master nodes of the Deckhouse Platform (DP) cluster and stores data in the `/var/lib/deckhouse/stronghold` directory on those nodes (see [Stronghold configuration](../../configuration/)). The scenario depends on the purpose of the cluster and the availability requirements.

| Scenario | Purpose | Stronghold fault tolerance |
| --- | --- | --- |
| [One master node](#test-cluster-with-one-master-node) | Testing, demo environments | None |
| [Three master nodes](#production-cluster-with-three-master-nodes) | Production | Failure of one node |
| [Dedicated nodes](#dedicated-nodes-for-stronghold) | Isolating Stronghold from other workloads | Depends on the number of nodes |
| [Air-gapped environment](#installing-in-an-air-gapped-environment) | Infrastructure without internet access | Depends on the chosen topology |

The general platform installation procedure is described in [Prepare the environment](../steps/prepare/), [Platform installation](../steps/install/), and [Initial access configuration](../steps/access/).

## Test cluster with one master node

Suitable for getting to know the product and for functional testing.

1. Install DP with one master node following [Platform installation](../steps/install/).
1. Enable the `stronghold` module (see [Stronghold configuration](../../configuration/#enabling-the-module)):

   ```shell
   d8 system module enable stronghold
   ```

1. Check the state:

   ```shell
   d8 k get modules stronghold
   d8 k -n d8-stronghold get pods
   ```

{{< alert level="warning" >}}
With one master node, Stronghold runs without fault tolerance: if the node is unavailable, the secret store is unavailable too. Do not use this scenario in production.
{{< /alert >}}

## Production cluster with three master nodes

The typical production configuration. Use an odd number of master nodes to preserve quorum (see [Master nodes](../../platform-management/control-plane/masters/)).

1. Install DP with three master nodes or [add master nodes](../../platform-management/control-plane/masters/#adding-a-master-node) to an existing cluster. Before adding the next node, wait until all master nodes are `Ready`:

   ```shell
   d8 k get no -l node-role.kubernetes.io/control-plane=
   ```

1. Enable the `stronghold` module and configure administrator access (see [Stronghold configuration](../../configuration/)).
1. Check that all Stronghold nodes have joined the Raft cluster:

   ```shell
   d8 stronghold operator raft list-peers
   ```

By default, the number of Stronghold replicas equals the number of master nodes plus the number of arbiter (etcd-arbiter) nodes; if the sum is even, one more replica is added. The Raft quorum is `floor(n/2)+1`, where `n` is the number of replicas.

Set up regular backups: [manual snapshots](../../../../admin/backups/save/) or [automated snapshots](../../../../admin/backups/automated-snapshots/) (Stronghold EE).

## Dedicated nodes for Stronghold

Starting with version `1.19`, the `stronghold` module supports `nodeSelector` and `storageClass` configuration, as well as storage migration for HA installations (see the [release notes](../../../../release-notes/)). This lets you run Stronghold on nodes dedicated to it. If `nodeSelector` is not set, the pods are placed on nodes with the `node-role.kubernetes.io/control-plane=""` label. Note that when `storageClass` is empty, the pods run on control-plane and etcd-arbiter nodes regardless of `nodeSelector`, so set both parameters.

1. Create a NodeGroup for Stronghold nodes with a label and a taint so that other workloads do not run on these nodes (see [Node groups](../../platform-management/node-management/node-group/)):

   ```yaml
   apiVersion: deckhouse.io/v1
   kind: NodeGroup
   metadata:
     name: stronghold
   spec:
     nodeType: Static
     nodeTemplate:
       labels:
         node-role.deckhouse.io/stronghold: ""
       taints:
         - effect: NoExecute
           key: dedicated.deckhouse.io
           value: stronghold
   ```

1. Add an odd number of nodes to the group, for example three (see [Adding a node](../../platform-management/node-management/adding-node/)).
1. Specify the node selector and the StorageClass in the `stronghold` ModuleConfig:

   ```yaml
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleConfig
   metadata:
     name: stronghold
   spec:
     enabled: true
     version: 1
     settings:
       nodeSelector:
         node-role.deckhouse.io/stronghold: ""
       storageClass: <STORAGE_CLASS_NAME>
   ```

<!-- TODO(verify): whether tolerations for the dedicated.deckhouse.io taint are required, the procedure for migrating storage from master nodes to dedicated nodes, and a link to the module parameter reference (/modules/stronghold/stable/configuration.html). -->

## Installing in an air-gapped environment

In an air-gapped environment, cluster nodes cannot reach `registry.deckhouse.ru`. The platform and `stronghold` module images are uploaded to your own container registry.

1. On a machine with internet access, download the DP and module images with `d8 mirror`, transfer them to the air-gapped environment, and push them to the container registry:

   ```shell
   d8 mirror push modules <REGISTRY_HOST:PORT>/<PATH_TO_DKP_REPO> -u <USERNAME> -p <PASSWORD>
   ```

   <!-- TODO(verify): d8 mirror pull/push commands for the complete DKP and stronghold module delivery, and a link to the DKP guide for installing in an air-gapped environment. -->

1. Install DP and specify the access parameters of your container registry in InitConfiguration (see [Platform installation](../steps/install/)).
1. Make sure the container registry is reachable from every master node. If the module is delivered separately, create a ModuleSource resource with the address of the modules in the registry (see the example in [Switching Stronghold from EE to CSE](../../platform-management/switching-editions/ee-to-cse/)).
1. Enable the `stronghold` module and check that the pods are not in the `ImagePullBackOff` or `ErrImagePull` state:

   ```shell
   d8 k -n d8-stronghold get pods
   ```

To access the web interface in an air-gapped environment without a public certificate authority, use a ClusterIssuer with a self-signed CA or your own certificate (see [Ways to organize access via the Ingress inlet](../../configuration/#ways-to-organize-access-via-the-ingress-inlet)).
