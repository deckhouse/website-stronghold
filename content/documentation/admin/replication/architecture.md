---
title: "Architecture: Stronghold and Stronghold EE"
linkTitle: "Architecture"
weight: 12
description: "Stronghold node layers, the difference between Stronghold and Stronghold EE, and how the layer order shapes replication behavior."
---

This page shows the architectural differences between Stronghold and Stronghold EE through the layers of a single node, and explains how the order of those layers shapes replication behavior: why the replication stream carries already-encrypted data and why each cluster owns its own seal.

Base Stronghold corresponds to Vault CE. The diagrams are simplified: they show what a node is made of and the direction of requests, not the network layout. The diagram notation is collected in the [Legend](#legend) section at the end of the page.

## Comparing replication levels and variants

Replication works at several levels — within a cluster, between clusters, and per individual mount (KV). There are five variants in total; they are covered in more detail below.

| Variant | Level | Purpose | Standby nodes | Edition |
| --- | --- | --- | --- | --- |
| HA: storage replication | Within a cluster | Node fault tolerance | Standby nodes forward all requests to the active node | Stronghold and Stronghold EE |
| Performance standby | Within a cluster | Scaling reads | Standby nodes read locally, writes go to the active node | Stronghold EE |
| Disaster Recovery | Between clusters | Hot standby and failover | The secondary does not serve clients and waits for a promote | Stronghold EE |
| Performance | Between clusters | Scaling and distributing reads | Secondaries read locally, writes go to the primary | Stronghold EE |
| KV replication (KV1/KV2) | Cross-cluster, per mount (API) | Copying selected KV stores | the replica (slave) is read-only, writes go to the master | Stronghold EE, except a DR secondary |

## The layers of a single node

A request passes through the node layers top to bottom. The key point is the **order of the layers**: what sits lower is applied later and "sees" the data as the layers above already processed it.

### Stronghold node

![Stronghold node layers](../../../../images/stronghold-node-ce.png "Stronghold node: API → Security Barrier → Raft → Physical storage")

_See the [Legend](#legend) for the diagram notation._

In base Stronghold a node has four layers: `Stronghold API → Security Barrier → Raft → Physical storage`.

- **Stronghold API** accepts a request and routes it to the right backend.
- **Security Barrier** encrypts the data. Everything below the barrier exists only in encrypted form.
- **Raft** handles consensus and storage replication between the nodes of a cluster (fault tolerance).
- **Physical storage** is the on-disk backend; the data there is already encrypted.

Base Stronghold has neither a WAL Backend nor seal wrap. That is why such a cluster does not take part in cross-cluster replication and has no second encryption layer.

### Stronghold EE node

![Stronghold EE node layers](../../../../images/stronghold-node-ee.png "Stronghold EE node: API → Barrier → WAL Backend → Raft → Sealwrap → Physical storage")

_See the [Legend](#legend) for the diagram notation._

Stronghold EE adds two layers, and **where** they sit matters: `Stronghold API → Security Barrier → WAL Backend → Raft → Sealwrap → Physical storage`.

- **WAL Backend** sits right under the barrier. It records every mutation into an ordered log that serves as the data source for replication. Because it is below the barrier, the log receives values that are already barrier-encrypted.
- **Sealwrap** sits at the very bottom, between Raft and physical storage. It is a second encryption layer for sensitive paths; the key is held by the seal (the external `Sealwrapper` box).

| Layer | Stronghold | Stronghold EE |
| --- | --- | --- |
| Stronghold API | yes | yes |
| Security Barrier | yes | yes |
| WAL Backend | — | yes |
| Raft | yes | yes |
| Sealwrap | — | yes |
| Physical storage | yes | yes |

{{< alert level="info" >}}
Two replication properties follow from the Stronghold EE layer order:

1. **WAL Backend is below the barrier** → the log (and therefore the replication stream) contains ciphertext, not plaintext.
2. **Sealwrap is below the WAL Backend** → the seal wrap is applied only after the write has entered the log → the data encrypted by the seal never enters the replication stream, and each cluster seals its data with its own seal.
{{< /alert >}}

{{< alert level="warning" >}}
In a Standalone Stronghold EE installation some layers and the features tied to them can be turned off in the server configuration:

- `disable_sealwrap = true` disables the **Sealwrap** layer; sensitive paths are protected only by the Security Barrier, without a second encryption layer.
- `disable_wal_replication = true` disables the **WAL Backend**, and with it native replication and performance standby; the node boots as an ordinary one.
- `disable_performance_standby = true` keeps replication but disables **performance standby**: standby nodes stop serving reads and forward all requests to the active node.
{{< /alert >}}

## Nodes in a cluster

Several nodes form a cluster on top of shared Raft storage: one node is active, the rest are standby. Nodes sync data over Raft (orange arrows), so every node holds a full encrypted copy of the storage. A cluster has a single seal — one `Sealwrapper` for the whole group.

### HA cluster (Stronghold)

![Stronghold HA cluster](../../../../images/stronghold-cluster-ha-ce.png "Stronghold HA cluster: active node and standby, Raft sync")

_See the [Legend](#legend) for the diagram notation._

The client reads from and writes to the active node (green R/W). In base Stronghold, standby nodes do not serve requests: they forward both reads and writes to the active node (Forward R/W). The active node syncs storage to the standby nodes over Raft — this provides fault tolerance: if the active node fails, one of the standby nodes becomes active.

### Performance standby (Stronghold EE)

![Stronghold EE performance standby cluster](../../../../images/stronghold-cluster-perf-standby-ee.png "Stronghold EE cluster: performance standby and WAL streaming, external Sealwrapper")

_See the [Legend](#legend) for the diagram notation._

The same HA cluster, but in Stronghold EE the standby nodes act as performance standby: they serve reads locally and forward writes to the active node (Forward only Write). To keep the standby nodes serving fresh data, the active node streams log events to them (blue `WAL` arrows). Because the WAL Backend is below the barrier, these events carry already-encrypted data. The cluster still has a single seal — one shared external `Sealwrapper`.

See the [Performance standby](../performance-standby/) page for details.

## Clusters between each other

Cross-cluster replication is available only in Stronghold EE. It links whole clusters: the primary streams its log to the secondary (blue Data WAL Streaming), and the secondary forwards client writes back to the primary. Each cluster has its own `Sealwrapper`, and it may differ.

### Performance

![Performance: primary → secondary](../../../../images/stronghold-clusters-performance.png "Performance replication: filtered WAL; the primary and the secondary have their own Sealwrappers (1 and 2)")

_See the [Legend](#legend) for the diagram notation._

The primary streams only non-local data to the secondary — shown on the diagram as the filtered log `filtered WAL`. Local mounts and data stay on each cluster and never leave it. The secondary serves reads locally and forwards writes to the primary. There can be several secondary clusters — the primary streams to all of them. You can limit the data replicated to a specific secondary with [path filters](../performance/#path-filters). The primary and the secondary each have their own seal (`Sealwrapper 1` and `Sealwrapper 2`).

### Disaster Recovery

![Disaster Recovery: primary → secondary](../../../../images/stronghold-clusters-dr.png "DR replication: WAL (all data); the primary and the secondary have their own Sealwrappers (1 and 2)")

_See the [Legend](#legend) for the diagram notation._

DR copies everything, including local data — shown on the diagram as `WAL`. The secondary is a full copy of the primary, but it does not serve clients and waits for a promote (`Promote Request`). When the primary fails, the secondary is promoted and takes over. The promote steps and how to recover the former primary are on the [Disaster recovery](../disaster-recovery/) page. The primary and the secondary each have their own seal (`Sealwrapper 1` and `Sealwrapper 2`).

### Why the seal can differ between clusters

The replication stream is captured at the WAL Backend level, that is **above** the sealwrap layer. So the seal wrap has not yet been applied to the source data, and the data encrypted by the seal is never sent across the cross-cluster link. Hence:

- the secondary applies seal wrap with its own seal, not the primary's;
- the primary and the secondary may even use different seal types (for example, different KMS);
- seal wrap is configured per cluster and does not depend on replication.

For more, see the [seal wrap](../../kms-hsm/sealwrap/) page. For what exactly is replicated in each mode, see the [Replication overview](../overview/).

## KV replication (at the API level)

![Stronghold EE node with KV replication](../../../../images/stronghold-node-kv-replication.png "KV replication: the KV replicator in the Stronghold API layer, syncing over the API")

_See the [Legend](#legend) for the diagram notation._

KV replication stands apart from cross-cluster WAL replication. It is performed by a separate component — the **KV replicator** — that lives in the topmost node layer, `Stronghold API`, that is **above the Security Barrier**. Because of that it does not work with the encrypted WAL stream but with logical secrets through the regular public API endpoints.

How it differs from perf/DR replication:

- **API level, not WAL.** WAL replication is captured below the barrier (at the WAL Backend layer) and carries the ciphertext of a whole cluster. KV replication works above the barrier, talking to API endpoints, and syncs only `KV1/KV2` mounts.
- **Configured on the slave node.** A `master-slave`, pull model: the consumer (`slave`) reaches out to the source (`master`) and pulls the secrets. The configuration is set when mounting a KV store on the consumer side.
- **Works at the node level.** Since it communicates through a specific node's API endpoints, replication is configured on a specific node of a cluster, not on the cluster as a whole.
- **Compatible with other stores.** The `master` can be a store of a different version or implementation — Vault CE, Vault EE, or any secret store with a compatible Vault KV/KV2 API. This simplifies migration and makes heterogeneous architectures possible.

Hence **where KV replication can be enabled**: on any node that serves API endpoints, regardless of its role in perf/DR replication. The only exception is a **DR secondary**: it does not answer client requests until promoted, so KV replication is not available there.

KV replication is an independent overlay: it is unrelated to the seal, the WAL, and cross-cluster replication, and is configured per mount.

See the [KV1/KV2 replication](../kv-replication/) page for details.

## Combined topologies

The modes are independent and can be combined:

- **One primary, both modes at once.** A cluster can be a performance primary and a DR primary at the same time: streaming data to performance secondaries and keeping a DR standby.
- **Each cluster is HA itself.** Primaries and secondaries in cross-cluster topologies are full HA clusters; in Stronghold EE their standby nodes serve reads as performance standby.
- **Multiple secondaries.** A single primary can have many secondary clusters — both performance and DR.

Below are a few examples of such combinations. The size (HA or a single node) is chosen per cluster, KV replication can be attached to any cluster except a DR secondary, and each cluster owns its own seal — so seal wrap can be configured differently across a topology: every cluster with its own seal somewhere, one and the same seal elsewhere, or none at all (`disable_sealwrap`).

![Combined topology: single-node primary, KV replication and different seals](../../../../images/stronghold-topology-combined-1.png "A single-node DR+Performance primary, HA secondaries, KV replication into the primary; each cluster has its own Sealwrapper")

_See the [Legend](#legend) for the diagram notation._

The primary is a DR and a Performance primary at once and consists of a single node; both secondaries are HA clusters, and a separate cluster feeds the primary over KV replication. Each cluster owns its own seal — `Sealwrapper 1`…`4` are all different in this topology.

![Combined topology: HA primary, KV into a performance secondary and different seals](../../../../images/stronghold-topology-combined-2.png "An HA primary with perf standby, HA secondaries, KV replication into a performance secondary; the clusters use different Sealwrappers")

_See the [Legend](#legend) for the diagram notation._

The same combined primary, but HA with performance standby; both secondaries are HA clusters too. KV replication is attached to a performance secondary — so it can be added not only to the primary. The clusters use different seals — `Sealwrapper 1` and `Sealwrapper 2`.

![Combined topology: HA primary, single-node secondaries and a shared seal](../../../../images/stronghold-topology-combined-3.png "An HA primary with perf standby, single-node secondaries, a single Sealwrapper 1, no KV")

_See the [Legend](#legend) for the diagram notation._

An HA primary with performance standby streams data to single-node DR and Performance secondaries; KV replication is not used, and all clusters are configured with one and the same seal — `Sealwrapper 1`.

## Legend

The diagrams have two levels of objects (node and cluster) and three kinds of links, distinguished by color.

| Object | How it looks | What it is |
| --- | --- | --- |
| Node | a box with a stack of colored layers | a single Stronghold instance; each layer is a processing level inside the node |
| Cluster | an ellipse around several nodes | a group of nodes on shared Raft storage, acting as a single unit |
| Client | a small circle on the left | an external consumer that talks to the cluster |
| Sealwrapper | a separate external box | the seal — the mechanism that holds the key for the seal wrap layer; one per cluster |

| Arrow | Color | Meaning |
| --- | --- | --- |
| Client access | green | client reads and writes (R/W), and forwarding of writes from a standby to the active node |
| Raft | orange | node sync **within** a cluster — storage replication between nodes |
| WAL streaming | blue | the change-log stream: within a cluster — events from the active node to performance standby; **between** clusters — data from the primary to the secondary |

There are two kinds of links, and they must not be confused: **Raft** (orange arrows) syncs nodes within a single cluster, while **replication** (blue arrows between clusters) links clusters to each other. Green arrows are client access and the forwarding of their writes, not replication. In short: Raft links nodes within a cluster; replication links clusters to each other.

The node layer colors are the same on every diagram:

| Layer | Color | Edition | What it does |
| --- | --- | --- | --- |
| Stronghold API | green | Stronghold and Stronghold EE | accepts requests and routes them to backends |
| Security Barrier | light blue | Stronghold and Stronghold EE | encryption: everything below the barrier is stored encrypted |
| WAL Backend | blue | Stronghold EE only | a log of mutations for replication; sits below the barrier |
| Raft | orange | Stronghold and Stronghold EE | consensus and storage replication between nodes (HA) |
| Sealwrap | crimson | Stronghold EE only | a second encryption layer for sensitive paths; the key is held by the seal |
| Physical storage | purple | Stronghold and Stronghold EE | data on disk (already encrypted) |
