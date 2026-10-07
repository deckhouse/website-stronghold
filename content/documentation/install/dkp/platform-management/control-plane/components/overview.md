---
title: "Overview"
description: "Overview of Kubernetes control plane components and how the control-plane-manager module manages them in Deckhouse Kubernetes Platform."
weight: 10
---

## Control plane components

Control plane components are responsible for the main cluster operations (for example, scheduling) and handle cluster events (for example, start a new pod when the number of replicas in a Deployment configuration does not match the number of running replicas).

Control plane components can run on any machine in the cluster. However, to simplify cluster configuration and maintenance, all control plane components are run on dedicated nodes where user containers are not allowed to run. If the components run on a single node, this is a `single-master` configuration. Running the components on multiple nodes switches the control plane to the `multi-master` high availability mode.

### kube-apiserver

The API server is a Kubernetes control plane component that exposes the Kubernetes API to clients.

The main implementation of the Kubernetes API server is kube-apiserver. kube-apiserver can be scaled horizontally, that is, deployed as multiple instances. You can run several kube-apiserver instances and balance traffic between these instances on different nodes.

The `control-plane-manager` module simplifies configuring kube-apiserver to enable [audit](auditing/) mode.

### etcd

A distributed and highly reliable key-value data store used as the primary store for all Kubernetes cluster data.

Like kube-apiserver, it supports running as multiple instances to ensure high availability, forming an etcd cluster.

### kube-scheduler

A component that watches for pods without an assigned node and selects the node they should run on.

When scheduling pods to nodes, many factors are taken into account, including resource requirements, hardware or software policy constraints, node/pod affinity and anti-affinity, data locality, and deadlines.

When the general algorithm is not enough (for example, you need to take into account the storageClass of attached PVCs), the scheduler algorithm can be extended with plugins.

For more details on the scheduler algorithm and connecting plugins, see the [Scheduler](scheduler/) section.

### kube-controller manager

A component that runs the controller processes for built-in resources. Each controller is a separate process, but to reduce complexity, all controllers are compiled into a single binary and run in a single process.

These controllers include:

- Node Controller: notices and responds when nodes go down.
- Replication Controller: maintains the correct number of pods for every replication controller object in the system.
- Endpoints Controller: populates Endpoints objects, that is, joins Services and pods.
- Account & Token Controllers: create default accounts and API access tokens for new namespaces.

<!-- TODO add something about configuration here; is it required from the administrator at all? -->

### cloud-controller manager

cloud-controller manager runs controllers that interact with the underlying cloud providers.
It is not covered in the DVP documentation.

TODO a link about cloud-controller-manager is needed, for example, to some module?

## Managing control plane components

The listed control plane components are managed by the `control-plane-manager` module, which runs on all cluster master nodes (nodes with the `node-role.kubernetes.io/control-plane: ""` label).

Functions of the `control-plane-manager` module:

- **Management of certificates** required for the components to operate, including renewal, issuance when the configuration changes, etc. Allows you to automatically maintain a secure control plane configuration and quickly add extra names (SANs) for secure access to the Kubernetes API.
- **Component configuration**. Automatically creates the required configurations and manifests of control plane components.
- **Component upgrade or downgrade**. Keeps component versions consistent across the cluster.
- **Management of the etcd cluster configuration** and its members. Scales etcd according to the number of master nodes and migrates the cluster from a single-master configuration to a multi-master one and vice versa.
- **kubeconfig configuration**. Ensures an always up-to-date configuration for kubectl. Generates, renews, and updates a kubeconfig with cluster-admin permissions and creates a symlink for the root user so that this kubeconfig is used by default.
- **Scheduler extension** by connecting external plugins via webhooks. Managed by the [KubeSchedulerWebhookConfiguration](/modules/control-plane-manager/cr.html#kubeschedulerwebhookconfiguration) resource. Allows you to use more complex logic for workload scheduling in the cluster. For example:
  - placing pods of data storage applications closer to the data itself,
  - prioritizing nodes depending on their state (network load, storage subsystem state, etc.),
  - dividing nodes into zones, etc.
- **Configuration backups** are saved to the `/etc/kubernetes/deckhouse/backup` directory.

### Certificate management

The `control-plane-manager` module manages the lifecycle of control plane SSL certificates:

- Root certificates for `kube-apiserver` and `etcd`. They are stored in the `d8-pki` secret of the `kube-system` namespace:
  - Kubernetes root CA (`ca.crt` and `ca.key`);
  - etcd root CA (`etcd/ca.crt` and `etcd/ca.key`);
  - RSA certificate and key for signing Service Accounts (`sa.pub` and `sa.key`);
  - root CA for extension API servers (`front-proxy-ca.key` and `front-proxy-ca.crt`).
- Client and server certificates for connecting control plane components to each other. The certificates are issued, renewed, and reissued if something changes (for example, the SAN list). The following certificates are stored only on nodes:
  - API server server certificate (`apiserver.crt` and `apiserver.key`);
  - client certificate for connecting `kube-apiserver` to `kubelet` (`apiserver-kubelet-client.crt` and `apiserver-kubelet-client.key`);
  - client certificate for connecting `kube-apiserver` to `etcd` (`apiserver-etcd-client.crt` and `apiserver-etcd-client.key`);
  - client certificate for connecting `kube-apiserver` to extension API servers (`front-proxy-client.crt` and `front-proxy-client.key`);
  - `etcd` server certificate (`etcd/server.crt` and `etcd/server.key`);
  - client certificate for connecting `etcd` to other cluster members (`etcd/peer.crt` and `etcd/peer.key`);
  - client certificate for connecting `kubelet` to `etcd` for health checks (`etcd/healthcheck-client.crt` and `etcd/healthcheck-client.key`).

You can add an extra list of SANs to the certificates, which makes it quick and easy to create extra "entry points" to the Kubernetes API.

When certificates change, the corresponding kubeconfig configuration is updated automatically.

### Component scaling

The module configures control plane components to run in both `single-master` and `multi-master` configurations.

In the `single-master` configuration:

- `kube-apiserver` uses only the `etcd` instance located on the same node;
- A proxy server responding on localhost is configured on the node; `kube-apiserver` responds on the master node IP address.

In the `multi-master` configuration, control plane components are automatically deployed in high availability mode:

- `kube-apiserver` is configured to work with all `etcd` instances.
- An additional proxy server responding on localhost is configured on each master node. By default, the proxy server accesses the local `kube-apiserver` instance, but if it is unavailable, it queries the other `kube-apiserver` instances sequentially.

### Scaling master nodes

Control plane nodes are scaled automatically using the `node-role.kubernetes.io/control-plane=""` label:

- Setting the `node-role.kubernetes.io/control-plane=""` label on a node deploys `control-plane` components on it, connects the new `etcd` member to the etcd cluster, and regenerates the required certificates and configuration files.
- Removing the `node-role.kubernetes.io/control-plane=""` label from a node deletes all `control-plane` components, regenerates the required configuration files and certificates, and correctly removes the node from the etcd cluster.

> **Warning.** Scaling nodes from 2 to 1 requires [manual actions](./etcd/#rebuilding-the-etcd-cluster) with `etcd`. In all other cases, all required actions are performed automatically. Note that when scaling from any number of master nodes down to 1, sooner or later, at the last step, you will need to scale nodes from 2 to 1.

Other operations with master nodes are described in the [Master nodes](../masters/) section.

### Version management

**Patch version** updates of control plane components (that is, within a minor version, for example, from `1.27.3` to `1.27.5`) happen automatically together with the Deckhouse version update. You cannot control patch version updates.

**Minor version** updates of control plane components (for example, from `1.26.*` to `1.28.*`) can be controlled with the [`kubernetesVersion`](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#clusterconfiguration-kubernetesversion) parameter, where you can choose automatic update mode (the `Automatic` value) or specify the desired minor version. The version used by default (with `kubernetesVersion: Automatic`), as well as the list of supported Kubernetes versions, can be found in the [documentation](/products/kubernetes-platform/documentation/v1/supported_versions.html#kubernetes).

The control plane update is performed safely for both `multi-master` and `single-master` configurations. The API server may be briefly unavailable during the update. The update does not affect applications running in the cluster and can be performed without scheduling a maintenance window.

If the version specified for the update (the [`kubernetesVersion`](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#clusterconfiguration-kubernetesversion) parameter) does not match the current control plane version in the cluster, a smart strategy for changing component versions is launched:

- General notes:
  - Updates in different NodeGroups are performed in parallel. Within each NodeGroup, nodes are updated sequentially, one at a time.
- When upgrading:
  - The upgrade is performed in **sequential stages**, one minor version at a time: 1.26 -> 1.27, 1.27 -> 1.28, 1.28 -> 1.29.
  - At each stage, the control plane version is updated first, and then kubelet is updated on the cluster nodes.
- When downgrading:
  - A successful downgrade is guaranteed only one version down from the highest control plane minor version ever used in the cluster.
  - kubelet is downgraded on the cluster nodes first, and then the control plane components are downgraded.

### Audit

To diagnose API operations, for example, in case of unexpected behavior of control plane components, Kubernetes provides an API operation logging mode. You can configure this mode by creating [Audit Policy](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/#audit-policy) rules, and the result of the audit will be the `/var/log/kube-audit/audit.log` log file with all the operations of interest. For more details, see the [Auditing](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/) section of the Kubernetes documentation.

Deckhouse clusters have the following basic audit policies created by default:

- logging of resource creation, deletion, and modification operations;
<!-- TODO which resources are meant here? Needs clarification. -->
- logging of actions performed on behalf of service accounts from the system namespaces: `kube-system`, `d8-*`;
- logging of actions performed on resources in the system namespaces: `kube-system`, `d8-*`.

You can disable log collection based on the basic policies by setting the [`basicAuditPolicyEnabled`](/modules/control-plane-manager/configuration.html#parameters-apiserver-basicauditpolicyenabled) flag to `false`.

Audit policy configuration is described in detail in the [Audit](auditing/) section.
