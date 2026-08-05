---
title: "Overview"
linkTitle: "Overview"
weight: 10
description: "Overview of native Performance and Disaster Recovery replication between Stronghold clusters."
---

Native replication moves the data of one Stronghold cluster to other clusters
over the network. Unlike [KV1/KV2 replication](../../kv-replication/overview/),
which copies individual secrets between mounts, native replication moves the
whole cluster state: mounts, policies, identity, and the rest of the data.

Two independent modes are supported.

| Mode | Purpose | What it moves | Serves clients |
| --- | --- | --- | --- |
| **Performance** | Scaling and distributing reads | Shared cluster data (mounts, policies, identity), except node-local paths | Yes: reads locally, forwards writes to the primary, keeps its own token store |
| **Disaster Recovery (DR)** | Hot standby and failover | All cluster data, including local data (tokens, leases) | No: waits for a promote |

There is also **performance standby** — not a separate cluster, but the
non-active HA nodes inside a single cluster that serve reads locally. See
[Performance standby](../performance-standby/).

## What replicates

Both modes work at the cluster storage level but move a different amount of
data.

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
- Node operational state and seal material.

**Disaster Recovery** replicates everything above plus the local data: local
mounts and auth methods, local policies, tokens, and leases. A DR secondary is
a full copy of the primary, so after a promote it keeps working with the same
tokens and leases. Only what is inherently per-node is left out: seal material,
the change log and replication index, Raft state, and node-discovery caches.

| Entity type | Performance | Disaster Recovery |
| --- | --- | --- |
| Non-local mounts and their secrets | Yes (filterable) | Yes |
| Local mounts (`local`) | No | Yes |
| Non-local auth methods | Yes | Yes |
| Local auth methods (`local`) | No | Yes |
| ACL policies | Yes | Yes |
| Local policies | No | Yes |
| Namespaces | Yes | Yes |
| Identity (entities, groups, aliases) | Yes | Yes |
| Tokens | No (own token store) | Yes |
| Leases | No | Yes |
| Audit devices | No (node-local) | Yes |

{{< alert level="info" >}}
Regardless of the replication mode, any cluster can additionally use
[seal wrap](../../kms-hsm/sealwrap/) — a second encryption layer for sensitive
values on top of the storage barrier. Seal material is never replicated: each
cluster owns its own seal, and a DR secondary may even run a different seal type
than the primary. So seal wrap is configured per cluster and does not depend on
replication.
{{< /alert >}}

## Requirements

- **Enterprise Edition.** Native replication is available only in Stronghold EE.
- **Integrated Raft storage.** Replication keeps a change log that is written
  only on integrated Raft storage. On other backends it does not start, and its
  endpoints return `replication not supported by this physical backend`.
- **Access to the cluster port.** A secondary connects to the primary over the
  cluster port (`cluster_address`, TLS with ALPN), not the API port. That port
  must be reachable between the clusters, and their TLS certificates must match.

## Enabling and disabling

In Stronghold EE replication is enabled by default: the `sys/replication/*`
endpoints are available right away. Enabling a specific mode is a separate
operation, described on the Performance and DR pages.

To turn replication off completely, set the following parameter in the server
configuration:

```hcl
disable_wal_replication = true
```

The node then boots as an ordinary node: replication, performance standby, and
the `sys/replication/*` endpoints are unavailable. To turn off only performance
standby without touching replication, use `disable_performance_standby = true`.

{{< alert level="info" >}}
The `disable_wal_replication` and `disable_performance_standby` parameters are
set in the server configuration of a Standalone Stronghold installation.
{{< /alert >}}

When replication is off, `GET sys/replication/status` returns `404`.

## Checking the state

```shell
d8 stronghold read sys/replication/status
```

The mode status shows `mode`, `cluster_id`, `state` (`running`, `stream-wals`,
`idle`), `connection_state`, `last_wal`, `last_remote_wal`, and the lists of
known primaries and secondaries. A secondary has caught up when its `last_wal`
reaches `last_remote_wal`.

{{< alert level="info" >}}
If replication is turned on for a cluster that already has data (for example,
after an upgrade from a build without replication), or after a period with
replication disabled, the node re-indexes storage by itself on the first boot.
There is no need to call `sys/replication/reindex` manually.
{{< /alert >}}

## Available pages

- [Architecture and diagrams](../architecture/) for what nodes and clusters are,
  plus combined topologies.
- [Performance replication](../performance/) for read scaling, secondary setup,
  and path filters.
- [Disaster recovery](../disaster-recovery/) for a hot standby, failover, and
  the promote ceremony.
- [Performance standby](../performance-standby/) for serving reads from
  non-active HA nodes within a cluster.
