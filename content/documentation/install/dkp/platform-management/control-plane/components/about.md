---
title: "Description"
description: "Features of the control-plane-manager module in Deckhouse Kubernetes Platform: certificate management, scaling, version management, and audit."
weight: 5
---

The cluster control plane components are managed by the `control-plane-manager` module, which runs on all cluster master nodes (nodes with the `node-role.kubernetes.io/control-plane: ""` label).

Control plane management features:

- **Management of certificates** required for the control plane to operate, including renewal, issuance when the configuration changes, etc. Allows you to automatically maintain a secure control plane configuration and quickly add extra SANs for secure access to the Kubernetes API.
- **Component configuration**. Automatically creates the required configurations and manifests of control plane components.
- **Component upgrade or downgrade**. Keeps component versions consistent across the cluster.
- **Management of the etcd cluster configuration** and its members. Scales master nodes and performs migration from `single-master` to `multi-master` and back.
- **kubeconfig configuration**. Ensures an always up-to-date configuration for kubectl. Generates, renews, and updates a kubeconfig with cluster-admin permissions and creates a symlink for the root user so that this kubeconfig is used by default.
- **Scheduler extension** by connecting external plugins via webhooks. Managed by the [KubeSchedulerWebhookConfiguration](/modules/control-plane-manager/cr.html#kubeschedulerwebhookconfiguration) resource. Allows you to use more complex logic for workload scheduling in the cluster. For example:
  - placing pods of data storage applications closer to the data itself;
  - prioritizing nodes depending on their state (network load, storage subsystem state, etc.);
  - dividing nodes into zones.

## Certificate management

Management of SSL certificates of control plane components:

- Server certificates for `kube-apiserver` and `etcd`. They are stored in the `d8-pki` secret of the `kube-system` namespace:
  - Kubernetes root CA (`ca.crt` and `ca.key`);
  - etcd root CA (`etcd/ca.crt` and `etcd/ca.key`);
  - RSA certificate and key for signing Service Accounts (`sa.pub` and `sa.key`);
  - root CA for extension API servers (`front-proxy-ca.key` and `front-proxy-ca.crt`).
- Client certificates for connecting `control-plane` components to each other. The module issues, renews, and reissues them if something changes (for example, the SAN list). The following certificates are stored only on nodes:
  - API server server certificate (`apiserver.crt` and `apiserver.key`);
  - client certificate for connecting `kube-apiserver` to `kubelet` (`apiserver-kubelet-client.crt` and `apiserver-kubelet-client.key`);
  - client certificate for connecting `kube-apiserver` to `etcd` (`apiserver-etcd-client.crt` and `apiserver-etcd-client.key`);
  - client certificate for connecting `kube-apiserver` to extension API servers (`front-proxy-client.crt` and `front-proxy-client.key`);
  - `etcd` server certificate (`etcd/server.crt` and `etcd/server.key`);
  - client certificate for connecting `etcd` to other cluster members (`etcd/peer.crt` and `etcd/peer.key`);
  - client certificate for connecting `kubelet` to `etcd` for health checks (`etcd/healthcheck-client.crt` and `etcd/healthcheck-client.key`).

It also allows you to add extra SANs to certificates, which makes it quick and easy to add extra "entry points" to the Kubernetes API.

When certificates change, the corresponding kubeconfig configuration is also updated automatically.

## Scaling

The control plane can run in both `single-master` and `multi-master` configurations.

In the `single-master` configuration:

- `kube-apiserver` uses only the `etcd` instance located on the same node;
- A proxy server responding on localhost is configured on the node; `kube-apiserver` responds on the master node IP address.

In the `multi-master` configuration, control plane components are automatically deployed in fault-tolerant mode:

- `kube-apiserver` is configured to work with all `etcd` instances.
- An additional proxy server responding on localhost is configured on each master node. By default, the proxy server accesses the local `kube-apiserver` instance, but if it is unavailable, it queries the other `kube-apiserver` instances sequentially.

### Scaling master nodes

`control-plane` nodes are scaled automatically using the `node-role.kubernetes.io/control-plane=""` label:

- Setting the `node-role.kubernetes.io/control-plane=""` label on a node deploys `control-plane` components on it, connects the new `etcd` member to the etcd cluster, and regenerates the required certificates and configuration files.
- Removing the `node-role.kubernetes.io/control-plane=""` label from a node deletes all `control-plane` components, regenerates the required configuration files and certificates, and correctly removes the node from the etcd cluster.

> **Warning.** Scaling nodes from 2 to 1 requires [manual actions](/modules/control-plane-manager/faq.html#what-if-the-etcd-cluster-fails) with `etcd`. In all other cases, all required actions are performed automatically. Note that when scaling from any number of master nodes down to 1, sooner or later, at the last step, you will need to scale nodes from 2 to 1.

## Version management

**Patch version** updates of control plane components (that is, within a minor version, for example, from `1.27.3` to `1.27.5`) happen automatically together with the Deckhouse version update. You cannot control patch version updates.

**Minor version** updates of control plane components (for example, from `1.26.*` to `1.28.*`) can be controlled with the [kubernetesVersion](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#clusterconfiguration-kubernetesversion) parameter, where you can choose automatic update mode (the `Automatic` value) or specify the desired minor version of the control plane. The control plane version used by default (with `kubernetesVersion: Automatic`), as well as the list of supported Kubernetes versions, can be found in the [documentation](/products/kubernetes-platform/documentation/v1/supported_versions.html#kubernetes).

The control plane update is performed safely both for clusters with a single master node and for multi-master clusters. The API server may be briefly unavailable during the update. The update does not affect applications running in the cluster and can be performed without scheduling a maintenance window.

If the version specified for the update (the [kubernetesVersion](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#clusterconfiguration-kubernetesversion) parameter) does not match the current control plane version in the cluster, a smart strategy for changing component versions is launched:

- General notes:
  - Updates in different NodeGroups are performed in parallel. Within each NodeGroup, nodes are updated sequentially, one at a time.
- When upgrading:
  - The upgrade is performed in **sequential stages**, one minor version at a time: 1.26 -> 1.27, 1.27 -> 1.28, 1.28 -> 1.29.
  - At each stage, the control plane version is updated first, and then kubelet is updated on the cluster nodes.
- When downgrading:
  - A successful downgrade is guaranteed only one version down from the highest control plane minor version ever used in the cluster.
  - kubelet is downgraded on the cluster nodes first, and then the control plane components are downgraded.

## Audit

If you need to log API operations or debug unexpected behavior, Kubernetes provides [Auditing](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/). You can configure it by creating [Audit Policy](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/#audit-policy) rules, and the result of the audit will be the `/var/log/kube-audit/audit.log` log file with all the operations of interest.

Deckhouse installations have basic policies created by default that log events:

- related to resource creation, deletion, and modification operations;
- performed on behalf of service accounts from the system namespaces `kube-system`, `d8-*`;
- performed on resources in the system namespaces `kube-system`, `d8-*`.

To disable the basic policies, set the [basicAuditPolicyEnabled](/modules/control-plane-manager/configuration.html#parameters-apiserver-basicauditpolicyenabled) flag to `false`.

Audit policy configuration is described in detail in the [corresponding FAQ section](/modules/control-plane-manager/faq.html#how-do-i-configure-additional-audit-policies).
