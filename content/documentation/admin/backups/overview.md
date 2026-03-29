---
title: "Stronghold backups"
linkTitle: "Introduction"
weight: 10
description: "Overview of manual and automated backups for Stronghold storage."
---

Stronghold backups are based on snapshots of integrated Raft storage. A snapshot lets you save cluster state and later use it to restore data after a failure or during planned administrative operations.

Stronghold supports two main scenarios:

- **manual snapshots**, where an administrator explicitly saves a snapshot file and restores it manually if needed;
- **automated snapshots**, where Stronghold saves backups on a schedule to local disk or S3-compatible storage.

## When to use snapshots

Snapshots are useful for:

- protection against data corruption or operator mistakes;
- preparation for upgrades and other risky operations;
- disaster recovery;
- storing backups outside the cluster or the physical environment that the cluster protects.

## Important considerations

- Snapshots apply to Stronghold integrated Raft storage.
- Snapshots work only for clusters that use integrated Raft storage as the primary storage backend.
- If Stronghold uses `etcd`, `postgresql`, or another external storage backend, use the native backup procedures of that backend instead of `stronghold operator raft snapshot` commands.
- Backup data remains encrypted inside the snapshot.
- To regain access after restore, you need the correct unseal or recovery keys according to your cluster configuration.
- Automated snapshots are an administrative feature and require a deliberate retention and storage policy.

## Available pages

- [Save a storage snapshot](./save/) for manual snapshot creation through CLI and API.
- [Inspect a snapshot](./inspect/) for local snapshot validation and content analysis without restoring it.
- [Restore from a snapshot](./restore/) for cluster restoration from an existing snapshot.
- [Automated snapshots](./automated-snapshots/) for schedule, storage, and status management of automated backups.

## Recommendations

- Keep backups outside the same environment they are intended to protect.
- Periodically verify snapshot creation, inspect snapshots with `stronghold operator raft snapshot inspect`, and test restore procedures.
- For production environments, prefer external object storage over local disk whenever possible.
