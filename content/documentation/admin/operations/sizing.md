---
title: "Sizing"
description: "Guidance on choosing CPU, memory, and disks for a Stronghold cluster, the number of Raft nodes, and the factors that drive load."
weight: 80
---

Secrets storage workloads vary a lot between organizations, so treat the values below as a starting point. Check actual resource usage with metrics (see [Monitoring](../monitoring/)) and adjust the cluster size.

Baseline system requirements are listed in [Requirements](../../../about/requirements/).

## Cluster profiles

The small and large cluster values come from [Requirements](../../../about/requirements/). The Medium profile and disk size are recommendations to validate against your workload.

| Parameter | Small | Medium (recommendation) | Large |
| --- | --- | --- | --- |
| Purpose | Development, testing, initial deployments | Production with moderate load | Production with constant high load |
| CPU | 4–8 cores | 8 cores | 8–16 cores |
| Memory | 8–16 GB | 16 GB | 16–32 GB |
| Disk IOPS | 3000+ | 3000+ | 3000+ |
| Disk throughput | 70+ MB/s | 100+ MB/s | 200+ MB/s |
| Disk size for Raft data | 25 GB | 50 GB | 100+ GB |

<!-- TODO(verify): Medium profile values and recommended Raft data disk size. -->

Approximate performance depending on the number of cores (from [Requirements](../../../about/requirements/)):

| Operation | 4 cores | 16 cores |
| --- | --- | --- |
| Authentication (token issue) | Up to 20 ops/s | Up to 100 ops/s |
| Reading a key up to 1 KB | Up to 500 ops/s | Up to 7000 ops/s |
| Writing a key | Up to 30 ops/s | Up to 150 ops/s |

## Raft disk

Every write in Stronghold is committed to the Raft log on a majority of nodes with a disk sync, so disk latency directly affects write time and leader election stability.

Recommendations:

- use local SSD or NVMe disks; avoid network disks with unpredictable latency;
- use a dedicated partition for Raft data so that OS and audit logs cannot fill it up;
- monitor disk sync latency: aim for single-digit milliseconds at most;
- keep in mind that storage size grows with the number of leases and tokens, and Raft snapshots need space for temporary files.

<!-- TODO(verify): recommended maximum fsync latency for Raft disks. -->

You can check disk sync latency with `fio`:

```shell
fio --name=raft-fsync --directory=/opt/stronghold/data --rw=write \
  --ioengine=sync --fdatasync=1 --bs=4k --size=100m
```

In the output, look at the `fsync/fdatasync/sync_file_range` percentiles. Remove the test file after the check.

## Number of nodes

| Nodes | Tolerated failures | When to use |
| --- | --- | --- |
| 1 | 0 | Development and testing only |
| 3 | 1 | Standard production cluster |
| 5 | 2 | Higher availability requirements, placement across three failure domains |

Recommendations:

- use an odd number of voting nodes: an even number does not increase fault tolerance;
- do not grow beyond five voting nodes: each additional node increases commit time;
- place nodes in different failure domains with low network latency between them;
- to scale reads, use [performance standby](../../replication/performance-standby/) (Stronghold EE); for disaster recovery, use [DR replication](../../replication/disaster-recovery/).

## In DP

{{< alert level="info" >}}
- Resource requests of the `stronghold` Pod are `50m` CPU and `256Mi` memory; the VerticalPodAutoscaler (`Initial` mode) can raise them up to `500m` and `1Gi`. The `stronghold-automatic` Pod requests `10m` CPU and `50Mi` memory.
- The default disk size (`diskSizeGigabytes`) is 1 GB; set a value sufficient for your data.
- The number of replicas equals the number of master nodes plus arbiter nodes; if the result is even, one more replica is added.
- The PodDisruptionBudget has `minAvailable` equal to `replicas/2 + 1` (rounded down).
- Destructive storage migrations require at least 3 replicas.
{{< /alert >}}

## What drives load

| Factor | Impact | How to reduce |
| --- | --- | --- |
| Leases | Each lease is stored and processed on expiry. A large number increases memory usage, startup time, and leader change time | Reduce TTLs, revoke unused leases, apply [quotas](../quotas/) |
| Tokens | Each login creates a token and a storage entry | Reuse tokens, use [Stronghold Agent](../../../user/agent/overview/), reduce TTLs |
| Authentication | Login operations are much more expensive than reading a secret | Cache tokens on the client side |
| Writes | Each write is replicated via Raft with a disk sync | Use fast disks, avoid frequently overwriting the same secrets |
| Audit | Every request and response is written to all audit devices | Use fast local disks or reliable network receivers, [filtering](../../audit/filtering/) |
| Cryptographic operations | The [Transit](../../../user/secrets-engines/transit/) and [PKI](../../../user/secrets-engines/pki/) engines are CPU-intensive, especially with large RSA keys | Add cores, choose less expensive algorithms |

## Revisiting cluster size

Add resources if:

- the 99th percentile of `stronghold_core_handle_request` keeps growing;
- CPU usage on the active node regularly exceeds 70–80%;
- memory usage approaches the node RAM;
- `stronghold_raft_commitTime` grows and leader elections become more frequent.

<!-- TODO(verify): CPU usage thresholds for revisiting cluster size. -->
