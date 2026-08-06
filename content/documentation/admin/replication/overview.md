---
title: "Overview"
linkTitle: "Overview"
weight: 10
description: "Overview of replication in Stronghold: Performance, Disaster Recovery, performance standby, and KV1/KV2 replication."
---

Replication in Stronghold moves secrets and data between clusters and nodes. It
solves different tasks: scaling reads, keeping a hot standby for a disaster, and
copying individual secrets. There are four replication mechanisms below; they can
be used together.

## Replication types

### Performance

Scales and distributes reads across clusters. A Performance secondary receives
the shared cluster state — non-local mounts, auth methods, policies, namespaces,
and identity — and serves clients locally: it handles reads itself and forwards
writes to the primary. Each secondary keeps its own token store, so you log in on
it through a replicated auth method. Available only in Stronghold EE on
integrated Raft storage.

### Disaster Recovery (DR)

A hot standby for a primary failure. A DR secondary is a full copy of the
primary, including local data (tokens and leases), but it does not serve clients
and waits for a promote. After a failover it keeps working with the same tokens
and leases as the former primary. Available only in Stronghold EE on integrated
Raft storage.

### Performance standby

Not a separate cluster, but the standby (non-leader) nodes inside a single
cluster. Without replication the standby nodes only wait for leadership and do
not serve requests; with replication enabled they serve reads locally, forwarding
writes to the active node. This scales reads inside a cluster without deploying a
second one. See [Performance standby](../performance-standby/).

### KV1/KV2 replication

Copies individual secrets of selected `KV1/KV2` mounts between clusters at the API
level, per mount. It works in a master-slave (pull) model: the consumer cluster
periodically pulls changes from the source over ordinary API methods
(`list`/`read`), and the local replicated mount is read-only. Unlike native
replication it does not move the whole cluster state, does not require access to
the cluster port, and can be configured on almost any node — including against an
external store with a compatible Vault KV/KV2 API. For the full description see
the [KV1/KV2 replication](../kv-replication/) page.

## Comparing replication levels and variants

Replication works at several levels — within a cluster, between clusters, and per
individual mount (KV). There are five variants in total:

| Variant | Level | Purpose | Standby nodes | Edition |
| --- | --- | --- | --- | --- |
| HA: storage replication | Within a cluster | Node fault tolerance | Standby nodes forward all requests to the active node | Stronghold and Stronghold EE |
| Performance standby | Within a cluster | Scaling reads | Standby nodes read locally, writes go to the active node | Stronghold EE |
| Disaster Recovery | Between clusters | Hot standby and failover | The secondary does not serve clients and waits for a promote | Stronghold EE |
| Performance | Between clusters | Scaling and distributing reads | Secondaries read locally, writes go to the primary | Stronghold EE |
| KV replication (KV1/KV2) | Cross-cluster, per mount (API) | Copying selected KV stores | the replica (slave) is read-only, writes go to the master | Stronghold EE, except a DR secondary |

Three variants move data between clusters: Performance, DR, and KV1/KV2 (HA and
performance standby work inside a single cluster). Below is what exactly crosses
between clusters (performance standby does not cross the cluster boundary):

| Object type | Performance | Disaster Recovery | KV1/KV2 |
| --- | --- | --- | --- |
| Non-local mounts and their secrets | Yes (filterable) | Yes | Secrets of selected KV1/KV2 mounts |
| Local mounts (`local`) | No | Yes | Secrets of selected local KV1/KV2 mounts |
| Non-local auth methods | Yes | Yes | No |
| Local auth methods (`local`) | No | Yes | No |
| ACL policies | Yes | Yes | No |
| Local policies | No | Yes | No |
| Namespaces | Yes | Yes | No |
| Identity (entities, groups, aliases) | Yes | Yes | No |
| Tokens | No (own token store) | Yes | No |
| Leases | No | Yes | No |
| Audit devices | No (node-local) | Yes | No |

KV1/KV2 replication works at the level of an individual store and does not depend
on the `local` flag: any `KV1/KV2` mount can be a source or a target, including a
local one. It copies only the secrets, not the mount configuration itself.

{{< alert level="info" >}}
Regardless of the replication mode, any cluster can additionally use
[seal wrap](../../kms-hsm/sealwrap/) — a second encryption layer for sensitive
values on top of the storage barrier. The data encrypted by the seal is never
replicated: each cluster owns its own seal, and a secondary may even run a
different seal type than the primary. So seal wrap is configured per cluster and
does not depend on replication.
{{< /alert >}}

## What replicates in Performance and DR

This section details what exactly the Performance and DR cluster modes move.
Their setup is on the [Performance](../performance/) and
[Disaster recovery](../disaster-recovery/) pages; KV1/KV2 setup is on a
[separate page](../kv-replication/).

**Performance** replicates the shared cluster state — everything that must be
identical on every secondary:

- **Non-local mounts** and their data: secrets under those mounts are copied to
  the secondary (you can narrow them with
  [path filters](../performance/#path-filters)).
- **Non-local auth methods**: their configuration and data. Auth always
  replicates and is never subject to filters, so a secret filter cannot lock
  operators out.
- **ACL policies** (global).
- **Namespaces** with all their contents.
- **Identity**: entities, groups, and aliases. Always replicated, regardless of
  filters.
- **System and configuration storage** (`core/`, `sys/`), except node-local
  sections.

**Performance does not replicate** what is local to each cluster:

- **Local mounts and local auth methods** (created with the `local` flag) and
  their data — they exist only on the cluster where they were created.
- **Local policies**.
- **Tokens**: each cluster runs its own token store. A secondary issues its own
  service tokens, and the primary's tokens are not valid on it — which is why
  you log in on a secondary through a replicated auth method.
- **Leases** and their expiration.
- **Audit devices**: their configuration is node-local.
- Node operational state and the data encrypted by the seal.

**Disaster Recovery** replicates everything above plus the local data: local
mounts and auth methods, local policies, tokens, and leases. A DR secondary is
a full copy of the primary, so after a promote it keeps working with the same
tokens and leases. Only what is inherently per-node is left out: the data
encrypted by the seal, the change log and replication index, Raft state, and
node-discovery caches.

## Available pages

- [Architecture and diagrams](../architecture/) for what nodes and clusters are,
  plus combined topologies.
- [Performance replication](../performance/) for read scaling, secondary setup,
  and path filters.
- [Disaster recovery](../disaster-recovery/) for a hot standby, failover, and
  the promote ceremony.
- [Performance standby](../performance-standby/) for serving reads from standby
  nodes within a cluster.
- [KV1/KV2 replication](../kv-replication/) for copying individual KV mounts
  between clusters (master-slave, pull, at the API level).
