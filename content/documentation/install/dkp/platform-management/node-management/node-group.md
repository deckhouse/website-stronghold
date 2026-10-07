---
title: "Node groups"
description: "Managing Deckhouse Kubernetes Platform cluster nodes with node groups: deployment, updates, static nodes, Cluster API Provider Static, and dedicated nodes."
weight: 15
---

## Managing cluster nodes

Nodes are managed by the `node-manager` module, whose main functions are:

1. Managing multiple nodes as a related group (NodeGroup):
   * Ability to define metadata that will be applied to all nodes in the group.
   * Monitoring a node group as a single entity (grouping nodes on graphs, aggregating node unavailability alerts, and notifications when a certain number or percentage of nodes in the group become unavailable).

1. Installing, updating, and configuring node software (containerd, kubelet, etc.), and connecting the node to the cluster:
   * Installing the operating system (see the [list of supported OS](../../../../about/requirements/#supported-os)) regardless of the infrastructure type used (any cloud or bare-metal hardware).
   * Basic operating system configuration (disabling automatic updates, installing required packages, configuring logging parameters, etc.).
   * Configuring Nginx to balance requests from nodes (kubelet) across API servers, including automatic updates of the upstream server list.
   * Installing and configuring the containerd CRI and Kubernetes, and adding the node to the cluster.
   * Managing node updates and their downtime (disruptions):
     * Automatically determining the acceptable Kubernetes minor version for a node group based on its configuration (`kubernetesVersion`), the default version for the whole cluster, and the current control plane version. Nodes are not allowed to be updated ahead of the control plane.
     * Only one node in a group is updated at a time, and only if all nodes in the group are available.
     * Two types of node updates:
       * regular — always performed automatically;
       * disruptive, for example: kernel update, containerd version change, significant kubelet version change, etc. When automatic disruptive updates are allowed, the node is drained before the update (this can be disabled).
   * Monitoring the update state and progress.

1. Cluster scaling.
   * Within the virtualization platform, maintaining the desired number of nodes in a group is available when using [Cluster API Provider Static](#working-with-static-nodes).

1. Managing Linux users on nodes.

## Node types

The virtualization platform is intended to run on bare-metal servers, so the following sections cover managing `Static` nodes.

You can learn about other node types and cloud provider capabilities in the Deckhouse Platform documentation.

## Node group

Nodes are managed using node groups, which are described by [NodeGroup](/modules/node-manager/cr.html#nodegroup) resources. Each node group performs its own specific tasks, for example:

- a group for Kubernetes control plane components;
- a group for monitoring components;
- a group for virtualization platform control plane components;
- a group of nodes with virtual machines (vm-worker nodes);
- a group of nodes with containerized applications (worker nodes), etc.

How nodes are divided into groups and how components are distributed across node groups depends on the cluster tasks. Examples of virtualization platform cluster configurations are available in the [Platform installation](../../../../install/dkp/install/steps/install/) section.

Nodes in a group share common metadata and parameters, which allows them to be configured automatically according to the group configuration. Deckhouse also tracks the number of nodes in the group and updates the software on them.

The following monitoring functions are available for such node groups:

- grouping of node parameters on group graphs;
- grouping of node unavailability alerts;
- alerts about the unavailability of a certain number or percentage of nodes in the group.

## Deploying, configuring, and updating Kubernetes nodes

### Deploying Kubernetes nodes

Deckhouse automatically performs the following immutable operations to deploy cluster nodes:

1. Configuring and optimizing the operating system for containerd and Kubernetes:
   - required packages are installed from the repositories of the corresponding distribution;
   - kernel parameters, logging parameters, log rotation, and other system parameters are configured.
1. Installing the required versions of containerd and kubelet, and adding the node to the Kubernetes cluster.
1. Configuring Nginx and updating the upstream list for balancing requests from the node to the Kubernetes API.

### Keeping nodes up to date

Two types of updates can be applied to keep cluster nodes up to date:

- **Regular** — such updates are always applied automatically and do not cause the node to stop or reboot.
- **Disruptive** — for example, updating the kernel or containerd version, a significant kubelet version change, etc. For this type of update, you can choose manual or automatic mode (the [`disruptions`](/modules/node-manager/cr.html#nodegroup-v1-spec-disruptions) parameter section). In automatic mode, the node is gracefully drained before the update, and only then is the update performed.

Only one node in a group is updated at any given time, and this is only possible when all nodes in the group are available.

The `node-manager` module has a set of built-in monitoring metrics that allow you to track update progress and receive notifications about problems during the update or about the need to grant permission for the update (manual update approval).

## Working with static nodes

### Limitations

When working with static nodes, the `node-manager` module functions have the following limitations:

- **No node provisioning** — resources (bare-metal servers, virtual machines, related resources) are allocated manually. Further resource configuration (connecting the node to the cluster, configuring monitoring, etc.) is performed automatically or partially automatically.
- **No automatic node scaling** — maintaining the specified number of nodes in a group is available when using [Cluster API Provider Static](#working-with-static-nodes) (the [`staticInstances.count`](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances-count) parameter). Deckhouse will try to maintain the specified number of nodes in the group, cleaning up excess nodes and configuring new ones as needed (selecting them from [StaticInstance](/modules/node-manager/cr.html#staticinstance) resources in the *Pending* state).

### Manual static node management

Configuring or cleaning up a node, as well as connecting it to and disconnecting it from the cluster, can be performed using prepared scripts.

To configure a server (VM) and add the node to the cluster, download and run a special bootstrap script. Such a script is generated for each static node group (each `NodeGroup` resource). It is located in the `d8-cloud-instance-manager/manual-bootstrap-for-<NODEGROUP-NAME>` secret. An example of adding a static node to the cluster is available [in the "Adding a node" section](./adding-node/#adding-a-static-node-manually).

To disconnect a cluster node and clean up the server (virtual machine), run the `/var/lib/bashible/cleanup_static_node.sh` script, which is already present on every static node. An example of disconnecting a cluster node and cleaning up the server is available [in the "Adding a node" section](./adding-node/#how-to-clean-up-a-node-for-adding-it-to-a-cluster-later).

### Automatic node management

A static node is managed automatically using [Cluster API Provider Static](#working-with-static-nodes).

Cluster API Provider Static (CAPS) connects to the server (VM) using [StaticInstance](/modules/node-manager/cr.html#staticinstance) and [SSHCredentials](/modules/node-manager/cr.html#sshcredentials) resources, configures it, and adds the node to the cluster.

When necessary (for example, if the [StaticInstance](/modules/node-manager/cr.html#staticinstance) resource corresponding to the server is deleted or the [number of nodes in the group](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances-count) is decreased), Cluster API Provider Static connects to the cluster node, cleans it up, and disconnects it from the cluster.

### Automatic management of an existing node

{{< alert level="info" >}}
Supported in Deckhouse 1.63 and later.
{{< /alert >}}

To transfer an existing cluster node under CAPS management, prepare [StaticInstance](/modules/node-manager/cr.html#staticinstance) and [SSHCredentials](/modules/node-manager/cr.html#sshcredentials) resources for this node, as for automatic management described above. However, the [StaticInstance](/modules/node-manager/cr.html#staticinstance) resource must additionally be annotated with `static.node.deckhouse.io/skip-bootstrap-phase: ""`.

### Configuring a node via CAPS

Cluster API Provider Static (CAPS) is an implementation of a provider for declarative management of static nodes (bare-metal servers or virtual machines) for the Kubernetes [Cluster API](https://cluster-api.sigs.k8s.io/) project. Essentially, CAPS is an additional abstraction layer over the existing Deckhouse functionality for automatically configuring and cleaning up static nodes using scripts generated for each node group (see the [Working with static nodes](#working-with-static-nodes) section).

CAPS performs the following functions:

- configuring a bare-metal server (or virtual machine) for connection to the Kubernetes cluster;
- connecting the node to the Kubernetes cluster;
- disconnecting the node from the Kubernetes cluster;
- cleaning up the bare-metal server (or virtual machine) after the node is disconnected from the Kubernetes cluster.

CAPS uses the following resources (CustomResource):

- **[StaticInstance](/modules/node-manager/cr.html#staticinstance).** Each `StaticInstance` resource describes a specific host (server, VM) managed by CAPS.
- **[SSHCredentials](/modules/node-manager/cr.html#sshcredentials)**. Contains the SSH data required to connect to the host (`SSHCredentials` is specified in the [`credentialsRef`](/modules/node-manager/cr.html#staticinstance-v1alpha2-spec-credentialsref) parameter of the `StaticInstance` resource).
- **[NodeGroup](/modules/node-manager/cr.html#nodegroup)**. The [`staticInstances`](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances) parameter section defines the required number of nodes in the group and a filter for the set of `StaticInstance` resources that can be used in the group.

CAPS is enabled automatically if the [`staticInstances`](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances) parameter section is filled in the NodeGroup. If the `staticInstances` parameter section is not filled in the `NodeGroup`, nodes for this group are configured and cleaned up manually (see the examples of [adding a static node to the cluster](../adding-node/#adding-a-static-node-manually) and [cleaning up a node](../adding-node/#how-to-clean-up-a-node-for-adding-it-to-a-cluster-later)), not with CAPS.

Workflow for static nodes when using CAPS:

1. **Preparing resources.**

Before transferring a bare-metal server or virtual machine under CAPS management, some preparation may be required, for example:

- Preparing the storage system, adding mount points, etc.;
- Installing OS-specific packages. For example, installing the `ceph-common` package if CEPH volumes are used on the server;
- Configuring the required network connectivity. For example, between the server and the cluster nodes;
- Configuring SSH access to the server and creating a management user with root access via `sudo`. It is good practice to create a separate user and unique keys for each server.

**Creating an [SSHCredentials](/modules/node-manager/cr.html#sshcredentials) resource.**

The `SSHCredentials` resource specifies the parameters CAPS needs to connect to the server via SSH. A single `SSHCredentials` resource can be used to connect to multiple servers, but it is good practice to create unique users and access keys for each server. In this case, there will be a separate `SSHCredentials` resource for each server.

**Creating a [StaticInstance](/modules/node-manager/cr.html#staticinstance) resource.**

A separate `StaticInstance` resource is created in the cluster for each server (VM). It specifies the IP address for the connection and a reference to the `SSHCredentials` resource whose data should be used for the connection.

Possible states of `StaticInstances` and the related servers (VMs) and cluster nodes:

- `Pending`. The server is not configured, and there is no corresponding node in the cluster.
- `Bootstraping`. The server (VM) is being configured and the node is being connected to the cluster.
- `Running`. The server is configured, and the corresponding node has been added to the cluster.
- `Cleaning`. The server is being cleaned up and the node is being disconnected from the cluster.

> You can transfer an existing cluster node that was previously added to the cluster manually under CAPS management by annotating its StaticInstance with `static.node.deckhouse.io/skip-bootstrap-phase: ""`.

**Creating a [NodeGroup](/modules/node-manager/cr.html#nodegroup) resource.**

In the context of CAPS, pay attention to the [`nodeType`](/modules/node-manager/cr.html#nodegroup-v1-spec-nodetype) parameter (must be `Static`) and the [`staticInstances`](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances) parameter section of the `NodeGroup` resource.

The [`staticInstances.labelSelector`](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances-labelselector) parameter section defines a filter that CAPS uses to select the `StaticInstance` resources to be used in the group. The filter allows you to use only specific `StaticInstance` resources for different node groups, and also to use one `StaticInstance` in different node groups. You can leave the filter undefined to use any available `StaticInstance` in the node group.

The [`staticInstances.count`](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances-count) parameter defines the desired number of nodes in the group. When the parameter changes, CAPS starts adding or removing the required number of nodes, running this process in parallel.

According to the [`staticInstances`](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances) parameter section, CAPS will try to maintain the specified number of nodes in the group (the [`count`](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances-count) parameter). When a node needs to be added to the group, CAPS selects a StaticInstance resource in the `Pending` status that matches the [filter](/modules/node-manager/cr.html#nodegroup-v1-spec-staticinstances-labelselector), configures the server (VM), and adds the node to the cluster. When a node needs to be removed from the group, CAPS selects a StaticInstance in the `Running` status, cleans up the server, and removes the node from the cluster (after which the corresponding StaticInstance switches to the `Pending` state and can be used again).

[Example of adding a node using CAPS](./adding-node/#adding-a-static-node-using-cluster-api-provider-static).

## How to interpret the node group state?

**Ready** — the node group contains the minimum required number of scheduled nodes in the `Ready` state for all zones.

Example 1. A node group in the `Ready` state:

```yaml
apiVersion: deckhouse.io/v1
kind: NodeGroup
metadata:
  name: ng1
spec:
  nodeType: CloudEphemeral
  cloudInstances:
    maxPerZone: 5
    minPerZone: 1
status:
  conditions:
  - status: "True"
    type: Ready
---
apiVersion: v1
kind: Node
metadata:
  name: node1
  labels:
    node.deckhouse.io/group: ng1
status:
  conditions:
  - status: "True"
    type: Ready
```

Example 2. A node group in the `Not Ready` state:

```yaml
apiVersion: deckhouse.io/v1
kind: NodeGroup
metadata:
  name: ng1
spec:
  nodeType: CloudEphemeral
  cloudInstances:
    maxPerZone: 5
    minPerZone: 2
status:
  conditions:
  - status: "False"
    type: Ready
---
apiVersion: v1
kind: Node
metadata:
  name: node1
  labels:
    node.deckhouse.io/group: ng1
status:
  conditions:
  - status: "True"
    type: Ready
```

**Updating** — the node group contains at least one node with an annotation prefixed with `update.node.deckhouse.io` (for example, `update.node.deckhouse.io/waiting-for-approval`).

**WaitingForDisruptiveApproval** — the node group contains at least one node that has the `update.node.deckhouse.io/disruption-required` annotation and
does not have the `update.node.deckhouse.io/disruption-approved` annotation.

**Scaling** — calculated only for node groups of the `CloudEphemeral` type. The `True` state is possible in two cases:

1. When the number of nodes is less than the desired number of nodes in the group, that is, when the number of nodes in the group needs to be increased.
1. When a node is marked for deletion or the number of nodes is greater than the desired number, that is, when the number of nodes in the group needs to be decreased.

The desired number of nodes is the sum of all replicas in the node group.

Example. The desired number of nodes is 2:

```yaml
apiVersion: deckhouse.io/v1
kind: NodeGroup
metadata:
  name: ng1
spec:
  nodeType: CloudEphemeral
  cloudInstances:
    maxPerZone: 5
    minPerZone: 2
status:
...
  desired: 2
...
```

**Error** — contains the last error that occurred when creating a node in the node group.

## How NodeGroup parameters take effect

| NG parameter                          | Disruption update          | Node re-provisioning | kubelet restart |
|---------------------------------------|----------------------------|----------------------|-----------------|
| chaos                                 | -                          | -                    | -               |
| cloudInstances.classReference         | -                          | +                    | -               |
| cloudInstances.maxSurgePerZone        | -                          | -                    | -               |
| cri.containerd.maxConcurrentDownloads | -                          | -                    | +               |
| cri.type                              | - (NotManaged) / + (other) | -                    | -               |
| disruptions                           | -                          | -                    | -               |
| kubelet.maxPods                       | -                          | -                    | +               |
| kubelet.rootDir                       | -                          | -                    | +               |
| kubernetesVersion                     | -                          | -                    | +               |
| nodeTemplate                          | -                          | -                    | -               |
| static                                | -                          | -                    | +               |
| update.maxConcurrent                  | -                          | -                    | -               |

For details on all parameters, see the description of the [NodeGroup](/modules/node-manager/cr.html#nodegroup) custom resource.

If the `instanceClass` or `instancePrefix` parameters are changed in the Deckhouse configuration, no `RollingUpdate` will occur. Deckhouse will create new `MachineDeployment` objects and delete the old ones. The number of `MachineDeployment` objects provisioned simultaneously is defined by the `cloudInstances.maxSurgePerZone` parameter.

During an update that requires node disruption (disruption update), pods are evicted from the node. If some pods cannot be evicted, eviction attempts are repeated every 20 seconds for up to 5 minutes. After this time, pods that could not be evicted are deleted forcibly.

## How to allocate nodes for specific workloads?

{{< alert level="warning" >}}
You cannot use the `deckhouse.io` domain in NodeGroup label and taint keys. It is reserved for **Deckhouse** components. Prefer the `dedicated` or `dedicated.client.com` keys.
{{< /alert >}}

There are two mechanisms for solving this task:

1. Setting labels in the NodeGroup `spec.nodeTemplate.labels` for later use in [spec.nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) or [spec.affinity.nodeAffinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity). This specifies which nodes the scheduler will select to run the target application.
1. Setting taints in the NodeGroup `spec.nodeTemplate.taints` and then tolerating them in [spec.tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/). This prevents applications that are not explicitly allowed from running on these nodes.

{{< alert level="info">}}
By default, Deckhouse tolerates taints with the `dedicated` key. Therefore, it is recommended to use this key with any value for taints on dedicated nodes.
If you need to use custom taint keys (for example, `dedicated.client.com`), add them to the [`.spec.settings.modules.placement.customTolerationKeys`](/products/kubernetes-platform/documentation/v1/reference/api/global.html#parameters-modules-placement-customtolerationkeys) array. This will allow system components such as `cni-flannel` to run on these nodes.
{{< /alert >}}

For more details, see the [article on Habr](https://habr.com/ru/company/flant/blog/432748/) (in Russian).

### System nodes

Deckhouse components use labels and taints to select nodes. System components can be assigned to dedicated nodes using the following `NodeGroup`:

```yaml
nodeTemplate:
  labels:
    node-role.deckhouse.io/system: ""
  taints:
    - effect: NoExecute
      key: dedicated.deckhouse.io
      value: system
```

<!-- TODO link to details somewhere in DKP? Or add a section to these docs? -->

<!-- ### Virtualization control plane components

TODO Need to come up with a group for virtualization control plane components. Or do they just run on system? -->

## How to allocate nodes for virtual machines?

For virtual machines to run on nodes of a specific group, in addition to creating the group itself, you need a VirtualMachineClass resource with a nodeSelector.

For example, for the vm-workers group, it may look like this:

```yaml
apiVersion: deckhouse.io/v1
kind: NodeGroup
metadata:
  name: vm-worker
spec:
  nodeType: Static
```

VirtualMachineClass with a nodeSelector for the vm-worker group:

```yaml
apiVersion: virtualization.deckhouse.io/v1alpha2
kind: VirtualMachineClass
metadata:
  name: vm-worker
spec:
  nodeSelector:
    matchExpressions:
    - key: node.deckhouse.io/group
      operator: In
      values:
        - vm-worker
```

A fragment of a virtual machine manifest that will run on nodes of the vm-worker group:

```yaml
apiVersion: virtualization.deckhouse.io/v1alpha2
kind: VirtualMachine
metadata:
  name: vm-name
spec:
  virtualMachineClassName: vm-workers
  # more VM fields ...
```
