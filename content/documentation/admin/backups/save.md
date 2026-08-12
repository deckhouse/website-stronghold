---
title: "Creating a snapshot"
weight: 20
description: "Manually create a snapshot of Stronghold integrated Raft storage."
---

A manual snapshot lets you save the current state of Stronghold integrated storage to a file and use it as a backup.

## Before you start

- Make sure Stronghold uses integrated Raft storage.
- Make sure Raft is used as the primary storage backend, not only as `ha_storage`.
- Confirm that you have administrative access to the cluster.
- Prepare a secure location for storing the snapshot file outside the production environment.
- Make sure you can contact the holders of the unseal or recovery keys if restore becomes necessary.

{{< alert level="warning" >}}
The `d8 stronghold operator raft snapshot save` command applies only to clusters that use integrated Raft storage. If Stronghold uses etcd, PostgreSQL, or another external backend, back up that storage system with its own native backup mechanism.
{{< /alert >}}

## Saving a snapshot via CLI

To create a snapshot, run:

```shell
d8 stronghold operator raft snapshot save backup.snap
```

As a result, Stronghold writes the snapshot to the local `backup.snap` file.

After creating the backup, it is recommended that you [inspect the snapshot](./inspect/) to confirm that the file can be read and contains the expected data structure.

## Saving a snapshot via API

To create a snapshot through the API, use the `GET /sys/storage/raft/snapshot` endpoint:

```shell
curl \
  --request GET \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  "${VAULT_ADDR}/v1/sys/storage/raft/snapshot" > backup.snap
```

## Recommendations

- Store snapshot files in a restricted and secure location.
- After creating a snapshot, inspect it using `d8 stronghold operator raft snapshot inspect`. Refer to ["Inspecting a snapshot"](./inspect/) for backup validation details.
- Use [automated snapshots](./automated-snapshots/) for scheduled backups.
- If your environment uses replication, define a separate backup strategy for each cluster domain.
