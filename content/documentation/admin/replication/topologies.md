---
title: "Topologies and scenarios"
linkTitle: "Topologies and scenarios"
weight: 15
params:
  edition: ee
description: "Differences between Stronghold replication variants, diagrams, and typical scenarios."
---

Replication in Stronghold works at two levels: **within a single cluster** and **between clusters**. There are four variants in total — from basic fault tolerance in CE to cross-cluster Performance and DR in EE.

The diagrams below are simplified and show the direction of requests and replication, not the network layout.

## Architecture objects

To read the diagrams unambiguously, it helps to separate two levels of objects and two kinds of links between them.

- **Node** — a single Stronghold instance: a process on its own server or pod, with its own copy of storage. At any moment exactly one node in a cluster is active (handles writes); the rest are standby.
- **Cluster** — a group of nodes on top of shared integrated Raft storage, acting as a single unit. Clients talk to the cluster, not to a specific node. A cluster is the smallest unit of cross-cluster replication: the whole cluster is replicated, not an individual node.

There are also two kinds of links, and they must not be confused.

- **Between nodes of one cluster — Raft.** The active node continuously replicates storage changes to the standby nodes over the Raft protocol. This is intra-cluster sync: it provides fault tolerance and needs no separate replication setup.
- **Between clusters — replication.** Separate clusters (usually in different locations) exchange data through cross-cluster replication: the primary streams changes to the secondary, and the secondary forwards client writes back to the primary. This is exactly what Performance or DR is, and it is configured by an administrator.

In short: **Raft links nodes within a cluster; replication links clusters to each other.**

### Legend

| Symbol | Meaning |
| --- | --- |
| `A` | Active node — handles writes |
| `S` | Standby node — spare only (does not serve requests in CE) |
| `PS` | Performance standby — a standby node in EE that serves local reads |
| Framed box with a title | A whole cluster (all of its nodes) |
| Thin arrow inside a box | Intra-cluster sync between nodes (Raft) |
| Thick arrow between boxes | Cross-cluster replication: change stream primary → secondary |
| Dotted arrow between boxes | Write forwarding secondary → primary |

## Comparing the variants

| Variant | Level | Purpose | Non-active nodes or standby | Edition |
| --- | --- | --- | --- | --- |
| HA: storage replication | Within a cluster | Node fault tolerance | Standby nodes forward all requests to the active node | CE and EE |
| Performance standby | Within a cluster | Scaling reads | Standby nodes read locally, writes go to the active node | EE |
| Disaster Recovery | Between clusters | Hot standby and failover | The secondary cluster does not serve clients and waits for a promote | EE, Standalone only |
| Performance | Between clusters | Scaling and distributing reads | Secondary clusters read locally, writes go to the primary | EE, Standalone only |

## Within a single cluster

### HA: storage replication (CE)

The baseline variant. In an HA cluster on integrated Raft storage, data is replicated across all nodes: one node is active and the rest are standby. Standby nodes provide fault tolerance — if the active node fails, one of them becomes active. In CE, standby nodes do not serve requests; they forward everything to the active node.

```text
   ┌───────────────── CLUSTER (HA, Raft) ──────────────────┐
   │                                                       │
   │  write/read ────▶ ┌──────────────┐                    │
   │                   │ ACTIVE node  │                    │
   │                   └──────┬───────┘                    │
   │                          │ storage replication (Raft) │
   │                   ┌──────┴───────┐                    │
   │                   ▼              ▼                     │
   │              ┌─────────┐    ┌─────────┐                │
   │              │ standby │    │ standby │                │
   │              │ (spare) │    │ (spare) │                │
   │              └─────────┘    └─────────┘                │
   │           all requests are forwarded to the active     │
   └───────────────────────────────────────────────────────┘
```

### Performance standby (EE)

The same HA cluster, but in EE the non-active nodes act as performance standby: they serve reads locally and forward writes to the active node. This is an addition on top of HA — it spreads read load within the cluster without creating a separate cluster.

```text
   ┌───────────────────── CLUSTER (HA, EE) ─────────────────┐
   │                                                        │
   │   write ──────▶ ┌──────────────┐                       │
   │                 │  ACTIVE node │ ◀── write (forwarded) │
   │                 │ reads+writes │                       │
   │                 └──────┬───────┘                       │
   │                        │  sync (Raft)                  │
   │                 ┌──────┴───────┐                       │
   │                 ▼              ▼                       │
   │            ┌─────────┐    ┌─────────┐                  │
   │   read ───▶│   PS    │    │   PS    │◀─ read           │
   │            │ locally │    │ locally │                  │
   │            └─────────┘    └─────────┘                  │
   └────────────────────────────────────────────────────────┘
```

See the [Performance standby](../performance-standby/) page for details.

## Between clusters

Cross-cluster replication is available only in EE and only in a Standalone installation. Each cluster in such a topology is a separate HA cluster. In the diagrams below, `A` is the active node at the top and `PS` are the standby nodes below it: since the clusters are EE, these nodes act as performance standby and serve local reads. The arrows inside a cluster show the intra-cluster sync over Raft.

### Disaster Recovery

A DR primary serves clients and copies its entire storage to a DR secondary. The secondary does not serve clients and waits for a promote. When the primary fails, the secondary is promoted and takes over.

```text
┌───── DR PRIMARY (HA) ─────┐
│           ┌──┐            │
│           │A │            │
│           └┬─┘            │
│         ┌──┴──┐           │
│         ▼     ▼           │
│        ┌──┐  ┌──┐         │
│        │PS│  │PS│         │
│        └──┘  └──┘         │
│ serves clients            │
└─────────────┬─────────────┘
              │ change stream (entire storage)
              ▼
┌──── DR SECONDARY (HA) ────┐
│           ┌──┐            │
│           │A │            │
│           └┬─┘            │
│         ┌──┴──┐           │
│         ▼     ▼           │
│        ┌──┐  ┌──┐         │
│        │PS│  │PS│         │
│        └──┘  └──┘         │
│ waits for promote         │
└───────────────────────────┘

Primary fails ── promote ──▶ DR SECONDARY becomes PRIMARY
```

The promote steps and how to recover the former primary are on the [Disaster recovery](../disaster-recovery/) page.

### Performance

A single primary accepts writes and streams changes to several performance secondaries. Clients read from the nearest secondary locally, and the secondary forwards writes to the primary. This fits geo-distribution and lower read latency. There can be several secondary clusters — the primary streams changes to all of them.

```text
               ┌────── PRIMARY (HA) ───────┐
               │           ┌──┐            │
               │           │A │            │
               │           └┬─┘            │
               │         ┌──┴──┐           │
               │         ▼     ▼           │
               │        ┌──┐  ┌──┐         │
               │        │PS│  │PS│         │
               │        └──┘  └──┘         │
               │ reads and writes          │
               └─────────────┬─────────────┘
                             │ change stream (to all secondaries)
              ┌──────────────┴───────────────┐
              ▼                               ▼
┌──── SECONDARY (perf) ─────┐  ┌──── SECONDARY (perf) ─────┐
│           ┌──┐            │  │           ┌──┐            │
│           │A │            │  │           │A │            │
│           └┬─┘            │  │           └┬─┘            │
│         ┌──┴──┐           │  │         ┌──┴──┐           │
│         ▼     ▼           │  │         ▼     ▼           │
│        ┌──┐  ┌──┐         │  │        ┌──┐  ┌──┐         │
│        │PS│  │PS│         │  │        │PS│  │PS│         │
│        └──┘  └──┘         │  │        └──┘  └──┘         │
│ reads locally             │  │ reads locally             │
│ writes → primary          │  │ writes → primary          │
└───────────────────────────┘  └───────────────────────────┘
```

You can limit the data replicated to a specific secondary with [path filters](../performance/#path-filters).

## Combined topologies

The modes are independent and can be combined:

- **One primary, both modes at once.** A cluster can be a performance primary and a DR primary at the same time: streaming data to performance secondaries and keeping a DR standby.
- **Each cluster is HA itself.** Primaries and secondaries in cross-cluster topologies are full HA clusters; in EE their standby nodes serve reads as performance standby.
- **Multiple secondaries.** A single primary can have many secondary clusters — both performance and DR.

```text
                        ┌───────────────────────┐
       clients ───────▶ │        PRIMARY        │
                        │  performance primary  │
                        │      + DR primary     │
                        └───────┬───────┬───────┘
          performance           │       │           DR
       (scaling reads)          │       │      (hot standby)
                                ▼       ▼
                   ┌──────────────┐   ┌──────────────┐
                   │ SECONDARY    │   │ DR SECONDARY │
                   │ (perf):      │   │ waits for    │
                   │ reads,       │   │ promote      │
                   │ forwards     │   │              │
                   └──────────────┘   └──────────────┘
```
