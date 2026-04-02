---
title: "Automated snapshots"
weight: 40
description: "Configure automated backups for Stronghold integrated storage."
---

Automated snapshots let Stronghold create scheduled backups of integrated Raft storage and write them either to local disk or to S3-compatible object storage.

{{< alert level="warning" >}}
Automated snapshots are available only when Stronghold uses integrated Raft storage. For `etcd`, `postgresql`, and other external backends, configure backup procedures provided by the storage system itself.
{{< /alert >}}

## When to use them

Automated snapshots are designed for recurring backups without manual command execution. They are especially useful for production clusters where backup retention must be predictable and copies should be stored outside the cluster itself.

## How they work

- You can create multiple named snapshot configurations.
- Each configuration defines the schedule, retention policy, and storage type.
- Supported storage types are `local` and `aws-s3`.
- For production environments, local storage is usually less suitable than external object storage: the active node can change over time, and backups are safer when stored outside the system they protect.

## Create or update a configuration

| Method | Path |
|--------|------|
| POST   | `/sys/storage/raft/snapshot-auto/config/:name` |

The endpoint requires `sudo` privileges.

### Core parameters

<div class="table__styling--container"></div>

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | String | Yes | — | Name of the configuration to create or update. |
| `interval` | Integer or string | Yes | — | Time between backups. You can specify seconds or Go duration format such as `24h`. |
| `retain` | Integer | No | `3` | Number of backups to keep. Older backups are deleted when the limit is exceeded. |
| `storage_type` | Immutable string | Yes | — | Storage type: `local` or `aws-s3`. |
| `path_prefix` | Immutable string | Yes | — | For `local`, the directory where snapshots are stored. For `aws-s3`, the object prefix inside the bucket. |
| `file_prefix` | Immutable string | No | `stronghold-snapshot` | Prefix for the file or object name. |

### Additional parameters for `local`

<div class="table__styling--container"></div>

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `local_max_space` | Integer | No | `0` | Maximum number of bytes backup files with the specified `file_prefix` may use in the `path_prefix` directory. A value of `0` disables the check. |

### Additional parameters for `aws-s3`

<div class="table__styling--container"></div>

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `aws_s3_bucket` | String | Yes | — | Bucket name used for backup storage. |
| `aws_s3_region` | String | No | — | Bucket region. |
| `aws_access_key_id` | String | No | — | Access key ID for the bucket. |
| `aws_secret_access_key` | String | No | — | Secret access key for the bucket. |
| `aws_s3_endpoint` | String | No | — | S3 service endpoint. |
| `aws_s3_disable_tls` | Boolean | No | — | Disables TLS for the S3 endpoint. Use only for testing. |
| `aws_s3_ca_certificate` | String | No | — | CA certificate for the S3 endpoint in PEM format. |

## Configuration examples

### Local disk

The following `local-snapshot.json` file creates a configuration that saves a snapshot every 5 minutes into `/stronghold/data/backups`, keeps 4 copies, and uses the `main_stronghold` file prefix:

```json
{
  "interval": "5m",
  "path_prefix": "/stronghold/data/backups",
  "file_prefix": "main_stronghold",
  "retain": "4",
  "storage_type": "local"
}
```

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-local-snapshots @local-snapshot.json
```

{{< alert level="info" >}}
Before applying the configuration, make sure the directory from `path_prefix` exists and is writable. The `failed to create snapshot directory at destination` error usually means the directory is missing or unavailable.
{{< /alert >}}

### S3-compatible storage

The following `minio-snapshot.json` file stores snapshots in S3-compatible object storage:

```json
{
  "interval": "3m",
  "path_prefix": "snapshots",
  "file_prefix": "stronghold_backup",
  "retain": "15",
  "storage_type": "aws-s3",
  "aws_s3_bucket": "my_bucket",
  "aws_s3_endpoint": "minio.domain.ru",
  "aws_access_key_id": "<ACCESS_KEY>",
  "aws_secret_access_key": "<SECRET_ACCESS_KEY>"
}
```

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-remote-snapshots @minio-snapshot.json
```

{{< alert level="info" >}}
Before applying the configuration, make sure the bucket already exists and the provided credentials have read and write permissions.
{{< /alert >}}

### Update an existing configuration

To modify only selected fields, provide a partial JSON document:

```json
{
  "interval": "3m",
  "retain": "10"
}
```

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-local-snapshots @local-snapshot-update.json
```

## List configurations

| Method | Path |
|--------|------|
| LIST   | `/sys/storage/raft/snapshot-auto/config` |

```shell
d8 stronghold list sys/storage/raft/snapshot-auto/config
```

## Read configuration parameters

| Method | Path |
|--------|------|
| GET    | `/sys/storage/raft/snapshot-auto/config/:name` |

```shell
d8 stronghold read sys/storage/raft/snapshot-auto/config/my-remote-snapshots
```

For `aws-s3`, the response does not expose `aws_access_key_id` or `aws_secret_access_key`.

## Delete a configuration

| Method | Path |
|--------|------|
| DELETE | `/sys/storage/raft/snapshot-auto/config/:name` |

```shell
d8 stronghold delete sys/storage/raft/snapshot-auto/config/my-remote-snapshots
```

{{< alert level="info" >}}
Deleting an automated snapshot configuration does not remove existing snapshot files from local or object storage.
{{< /alert >}}

## Read backup status

| Method | Path |
|--------|------|
| GET    | `/sys/storage/raft/snapshot-auto/status/:name` |

```shell
d8 stronghold read sys/storage/raft/snapshot-auto/status/my-remote-snapshots
```

Important status fields:

- `consecutive_errors`: number of backup errors in a row;
- `last_snapshot_end`: end time of the last successful snapshot;
- `last_snapshot_error`: text of the most recent error;
- `last_snapshot_start`: start time of the last completed backup;
- `last_snapshot_url`: location of the last successful snapshot;
- `next_snapshot_start`: next scheduled start time;
- `snapshot_start`: start time of the current backup job;
- `snapshot_url`: location of the currently written snapshot.

## See also

- [Save a storage snapshot](./save/)
- [Restore from a snapshot](./restore/)
