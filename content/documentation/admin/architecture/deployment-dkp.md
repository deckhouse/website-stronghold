---
title: "Deployment in DKP"
description: "How the stronghold module is placed in a Deckhouse Kubernetes Platform cluster: pods on master nodes, Raft, Ingress, access through Dex, and automatic unsealing."
weight: 20
---

In Deckhouse Platform (DP), Stronghold is shipped as the `stronghold` module. The module deploys a Stronghold cluster, initializes and unseals the storage, publishes the web UI and API through the selected inlet (Ingress by default), and, in the `Automatic` mode, configures login through Dex. How to enable and configure the module is described in [Stronghold configuration](../../../install/dkp/configuration/).

## Deployment layout

![layout_plan_dkp.en.png](../../../images/layout_plan_dkp.en.png)

The diagram is illustrative: the active node is elected by Raft and is not necessarily `stronghold-0`. The Ingress, `ingress-tls` secret, Dex, and `stronghold-keys` elements correspond to the default configuration (the `Ingress` inlet with `CertManager`, and the `Automatic` mode).

## Module components

- **`d8-stronghold` namespace**: all Stronghold containers run here, and the module's service secrets are stored here.
- **Stronghold pods**: each Stronghold node runs in its own pod. Pods are managed by a StatefulSet and spread one per node (pod anti-affinity). If `storageClass` is empty, the pods run on control-plane (master) nodes and on nodes with the `node.deckhouse.io/etcd-arbiter` label. If `storageClass` is set, the pods are placed by the module's `nodeSelector` parameter (by default, `node-role.kubernetes.io/control-plane=""`).
  The number of replicas is the number of master nodes plus the number of arbiter (etcd-arbiter) nodes; if the sum is even, one more replica is added. Later it can be changed manually.
- **Storage**: integrated Raft storage in HA mode is enabled by default. If `storageClass` is empty, each node keeps its data on the local disk of its node in the `/var/lib/deckhouse/stronghold` directory (a subdirectory named by the module's `localPathUUID` may be used). If `storageClass` is set, the data is stored in the `data-stronghold-N` PersistentVolumeClaims of the StatefulSet.
- **Internal service**: nodes reach each other by names such as `stronghold-0.stronghold-internal`. The internal API listens on port `8300` (Raft peer addresses use `8301`); `retry_join` between nodes and the inner-cluster unsealing go to port `8300`. The external API listens on `8200` (cluster address `8201`). See [Ports](../ports/).
- **Inlet**: the `inlet` parameter sets how the service is exposed: `Ingress` (default), `GatewayAPI`, `LoadBalancer`, `NodePort`, or `None`. For `LoadBalancer`, `NodePort`, and `None`, you must set `https.mode: CustomCertificate`. With the `Ingress` inlet, the web UI and API are published at an address built from the [`publicDomainTemplate`](/products/kubernetes-platform/documentation/v1/reference/api/global.html#parameters-modules-publicdomaintemplate) template by replacing `%s` with `stronghold`, for example `stronghold.mycompany.tld`.
- **TLS certificate**: the `ingress-tls` secret in the `d8-stronghold` namespace. With the `Ingress` inlet and `https.mode: CertManager`, it is issued by cert-manager using the ClusterIssuer set in the DP settings; for the other inlets, or when `https.mode: CustomCertificate` is used, the certificate from `customCertificate` is used. Without this secret, the pods stay in `ContainerCreating`.

## Operating mode and unsealing

The mode is set by the `management.mode` parameter. The default is `Automatic`; `Manual` is also supported. In the `Manual` mode, automatic initialization is disabled, the Dex and Kubernetes integrations are not configured, there is no `stronghold-keys` secret, and the `management.administrators` parameter is unavailable. In the `Automatic` mode:

1. On first start, the module initializes the storage automatically.
1. The unseal key and the root token are saved to the `stronghold-keys` secret in the `d8-stronghold` namespace.
1. The module unseals the Stronghold nodes automatically after initialization and on every pod restart.

<!-- TODO(verify): whether the DKP module uses the seal "inner-cluster" mechanism (v1.18 release notes: "automatic unsealing with keys kept in Stronghold node memory") or unseals nodes using the stronghold-keys secret. -->

{{< alert level="warning" >}}
The `stronghold-keys` secret contains the root token and the unseal key. Anyone who can read secrets in the `d8-stronghold` namespace gets full access to Stronghold. Restrict access to this namespace with DP RBAC and keep a copy of the secret in a protected location: when the module is disabled, the secret is deleted, and without it access to the data cannot be restored.
{{< /alert >}}

HSM (`seal "pkcs11"`) and Yandex Cloud KMS (`seal "yandexcloudkms"`) are not supported in DP. They are available only in a [standalone deployment](../deployment-standalone/).

## User access

- After initialization, the module creates the `deckhouse_administrators` role in Stronghold and enables web UI login through [Dex](/modules/user-authn/) OIDC authentication (the `user-authn` module).
- Stronghold administrators are set in the module ModuleConfig in the `management.administrators` parameter (only for `management.mode: Automatic`), as groups (`Group`) or individual users (`User`). A user must belong to at least one group, otherwise OIDC login fails.
- Configure other users and their permissions with built-in Stronghold tools: auth methods and policies.

## Cluster integration

The module automatically connects the current DP cluster to Stronghold for the [`secrets-store-integration`](/modules/secrets-store-integration/stable/) module, which delivers secrets to application pods.

<!-- TODO(verify): which auth method (kubernetes auth) and which mount path the module configures for secrets-store-integration. -->

## Quorum and fault tolerance

The Raft quorum is `floor(n/2)+1`, where `n` is the number of Stronghold nodes (the module's PodDisruptionBudget uses `minAvailable = replicas/2 + 1`). By default, `n` equals the number of master nodes plus arbiter nodes, rounded up to an odd number. In a cluster with three master nodes, at least two running pods are required for reads and writes. A cluster with a single master node is not fault tolerant. What to do when quorum is lost is described in [Recover from lost quorum](../../../install/dkp/raft-lost-quorum-recovery/).

## Disabling the module and data

When the module is disabled, all Stronghold containers in the `d8-stronghold` namespace and the `stronghold-keys` secret are deleted. If `storageClass` is empty, data on the nodes in `/var/lib/deckhouse/stronghold` is kept. If `storageClass` is set, the PVCs are deleted together with the StatefulSet (`persistentVolumeClaimRetentionPolicy.whenDeleted: Delete`), so the data is not kept. To restore access, enable the module and put the saved copy of the `stronghold-keys` secret into the namespace.

## Requirements

- Stronghold `1.19` requires DP `1.76` or later.
- Stronghold EE is enabled with a license key in ModuleConfig and is available only in commercial DP editions. See [Editions](../../../about/editions/).
