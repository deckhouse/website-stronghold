---
title: "Restoring from a snapshot"
weight: 30
description: "Restore Stronghold from an integrated Raft storage snapshot."
---

Snapshot restore is used when you need to bring Stronghold back to a previously saved state using a backup.

## Before you start

- If you are restoring Stronghold after an incident, bring the cluster back to an operational state: prepare fresh storage and obtain the temporary root token created during re-initialization.
- Make sure the target cluster uses integrated Raft storage as its primary storage backend.
- Copy the snapshot file to the node where you will run the restore.
- Make sure you have the original unseal or recovery keys. After restore, you need those original keys and not only the newly issued initialization keys.

{{< alert level="warning" >}}
Snapshot restore is available only for integrated Raft storage snapshots. If Stronghold runs on top of etcd, PostgreSQL, or another external backend, use that storage system's own restore procedure.
{{< /alert >}}

## Restoring via CLI

Use the `-force` flag when restoring a snapshot:

```shell
d8 stronghold operator raft snapshot restore -force /tmp/snapshots/backup.snap
```

The `-force` flag is required because the current cluster state and the snapshot data belong to different storage states.

## Restoring via API

To restore through the API, use the `POST /sys/storage/raft/snapshot-force` endpoint:

```shell
curl \
  --request POST \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --data-binary @/tmp/snapshots/backup.snap \
  "${VAULT_ADDR}/v1/sys/storage/raft/snapshot-force"
```

## After restore

After the snapshot is loaded, unseal Stronghold with the original keys:

```shell
d8 stronghold operator unseal
```

## Recommendations

- Periodically test both the snapshot restore and creation procedures in a non-production environment.
- Record where backups are stored and who is responsible for key access.
- If you operate multiple cluster domains or replication setups, define the restore order for each of them separately.
