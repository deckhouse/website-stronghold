---
title: "KV1/KV2 replication"
description: "What KV1/KV2 replication means for users of a replicated KV store and where to find the administrator guide."
weight: 40
---

KV1/KV2 replication automatically copies secrets from a KV store in one Stronghold cluster (the source, `master`) to a KV store in another cluster (the consumer, `slave`) using a pull model. Data is synchronized periodically according to the settings of the replicated store.

What this means for you as a user of a replicated store:

- a local KV store with replication enabled works only in `read-only` mode: you can read and list secrets, but you cannot create, modify, or delete them;
- make all changes in the source store; they appear in the local store after the next synchronization run;
- the local and remote `mount` paths and namespaces may differ, so check with your administrator which store is the source.

Replication is configured by an administrator when mounting a new KV store. For the setup procedure, parameters, and limitations, see the administrator guide [KV1/KV2 replication](../../../admin/replication/kv-replication/).
