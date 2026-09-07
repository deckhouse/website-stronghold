---
title: "Stronghold backups"
linkTitle: "Overview"
weight: 10
description: "Overview of manual and automated backups for Stronghold storage."
params:
  relatedLinks:
    - title: "Creating a snapshot"
      url: ../save/
    - title: "Inspecting a snapshot"
      url: ../inspect/
    - title: "Restoring from a snapshot"
      url: ../restore/
    - title: "Automated snapshots"
      url: ../automated-snapshots/
---

Stronghold backups are based on snapshots of integrated Raft storage. A snapshot lets you save cluster state and later use it to restore data after a failure or during planned maintenance operations.

Stronghold supports two main scenarios:

- **Manual snapshots**: An administrator explicitly saves a snapshot file and restores it manually if needed.
- **Automated snapshots**: Stronghold saves backups on a schedule to the local disk or S3-compatible storage.

## Using snapshots

Snapshots are useful for:

- Protection against data corruption or operator mistakes.
- Preparation for upgrades and other risky operations.
- Disaster recovery.
- Storing backups outside the cluster or hosting site.

Consider the following when working with snapshots:

- Snapshots apply to Stronghold integrated Raft storage.
- Snapshots work only for clusters that use integrated Raft storage as the primary storage backend.
- If Stronghold uses etcd, PostgreSQL, or another external system as a backend, you have to use the corresponding backup mechanisms of those systems, rather than the built-in Stronghold commands for Raft snapshots (`d8 stronghold operator raft snapshot`).
- Data remains encrypted in the backup.
- To regain access after restore, you need the correct unseal or recovery keys according to your cluster configuration.
- Automated snapshots are an administrative feature and require a deliberate retention policy.

## Recommendations

Follow these recommendations when using Stronghold backups:

- Keep backups outside the same environment they are intended to protect.
- Periodically verify snapshot creation with `d8 stronghold operator raft snapshot inspect` and the backup restore procedure in a test environment.
- For production environments, prefer external object storage over local disk whenever possible.
