---
title: "Overview"
linkTitle: "Overview"
weight: 10
params:
  edition: ee
description: "Overview of native Performance and Disaster Recovery replication between Stronghold clusters."
---

Native replication copies the state of one Stronghold cluster to one or more other clusters over the cluster network. Unlike [KV1/KV2 replication](../../kv-replication/overview/), which copies individual secrets between mounts, native replication works at the level of the whole cluster keyspace and is built on the internal write-ahead log (WAL) and a Merkle index for fast change detection and convergence.

Stronghold supports two independent replication modes. A single cluster can be a primary in one mode and a secondary in another at the same time.

| Mode | Purpose | What is replicated | Does the secondary serve clients? |
| --- | --- | --- | --- |
| **Performance** | Scaling reads and geographically distributing them | The shared keyspace (mounts, policies, identity, and so on), **except** node-local paths | Yes: reads are served locally, writes are forwarded to the primary; it keeps its own token store |
| **Disaster Recovery (DR)** | Hot standby and failover | The entire non-ignored keyspace, **including** local data (tokens, leases) | No: it waits for a promote |

There is also **performance standby** — this is not a separate cluster but the non-active HA nodes **inside a single** cluster that serve reads locally. See [Performance standby](../performance-standby/).

## Requirements

- **Enterprise Edition.** Native replication is available only in Stronghold EE.
- **Integrated Raft storage.** The WAL is written atomically only on a transactional backend, so integrated Raft storage is required as the primary storage backend. On other backends the replication pipeline does not start and replication endpoints return `replication not supported by this physical backend`.
- **Connectivity over the cluster address.** A secondary connects to the primary over the cluster port (`cluster_address`, TLS with ALPN), not over the API port. Make sure the cluster port is reachable between clusters and that TLS certificates match.

## How replication is enabled

In Stronghold EE the replication pipeline is **enabled by default** — no additional configuration is needed to make the `sys/replication/*` endpoints available. Enabling a specific mode (Performance or DR) is a separate runtime operation described on the mode pages.

The pipeline can be turned off explicitly. The switches are evaluated in the following order of priority:

1. `disable_wal_replication = true` in the server configuration turns off the **entire** pipeline (WAL, Merkle index, log shippers, performance standby, and the `sys/replication/*` endpoints). The node then boots as an ordinary node. This is an explicit off switch and overrides everything else.
1. The `STRONGHOLD_ENABLE_REPLICATION` environment variable overrides the default: a falsey value (`0`, `false`, `no`, `off`, `disabled`) turns replication off; a truthy value turns it on. It does not override `disable_wal_replication = true`.
1. If nothing is set, the default is **on**.

Separately, `disable_performance_standby = true` turns off **only** performance standby without touching replication itself.

{{< alert level="info" >}}
The `disable_wal_replication` and `disable_performance_standby` parameters and the `STRONGHOLD_ENABLE_REPLICATION` environment variable apply to the server configuration of a Standalone Stronghold installation.
{{< /alert >}}

When replication is disabled, `GET sys/replication/status` returns `404` because the endpoints are not registered. This is a convenient indicator of the feature state.

## Status and convergence

Check the overall state at any time:

```shell
d8 stronghold read sys/replication/status
```

The mode-specific status contains the fields `mode`, `cluster_id`, `state` (for example, `running`, `stream-wals`, `idle`), `connection_state`, `last_wal`, `last_remote_wal`, `merkle_root`, and the lists of known secondaries and primaries. A secondary is caught up when the Merkle root of its subtree matches the primary.

{{< alert level="info" >}}
On the first boot with replication enabled after an upgrade from a build without replication (data exists but there is no Merkle index), or after a period with replication disabled, the node automatically rebuilds the Merkle index from storage. No manual `sys/replication/reindex` call is required in these cases.
{{< /alert >}}

## Available pages

- [Performance replication](../performance/) for read scaling, secondary setup, and path filters.
- [Disaster recovery](../disaster-recovery/) for a hot standby and failover, including the promote ceremony.
- [Performance standby](../performance-standby/) for serving reads from non-active HA nodes within a cluster.
