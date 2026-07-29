---
title: "Overview"
linkTitle: "Overview"
weight: 10
params:
  edition: ee
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

## Requirements

- **Enterprise Edition.** Native replication is available only in Stronghold EE.
- **Integrated Raft storage.** Replication keeps a change log that is written
  only on integrated Raft storage. On other backends it does not start, and its
  endpoints return `replication not supported by this physical backend`.
- **Access to the cluster port.** A secondary connects to the primary over the
  cluster port (`cluster_address`, TLS with ALPN), not the API port. That port
  must be reachable between the clusters, and their TLS certificates must match.

{{< alert level="warning" >}}
Native Performance and DR replication is available only in a Standalone
Stronghold installation. Cross-cluster replication is not supported in the DKP
module.
{{< /alert >}}

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

- [Topologies and scenarios](../topologies/) for diagrams and typical cases
  across all three modes.
- [Performance replication](../performance/) for read scaling, secondary setup,
  and path filters.
- [Disaster recovery](../disaster-recovery/) for a hot standby, failover, and
  the promote ceremony.
- [Performance standby](../performance-standby/) for serving reads from
  non-active HA nodes within a cluster.
