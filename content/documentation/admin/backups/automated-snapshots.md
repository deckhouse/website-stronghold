---
title: "Automated snapshots"
weight: 40
description: "Configure automated backups for Stronghold integrated storage."
params:
  edition: ee
---

Automated snapshots let Stronghold create backups of integrated Raft storage on a schedule and write them either to local disk or to S3-compatible object storage.

{{< alert level="warning" >}}
Automated snapshots are available only when Stronghold uses integrated Raft storage. For etcd, PostgreSQL, and other external backends, configure backup procedures provided by the storage system itself.
{{< /alert >}}

Automated snapshots are designed for recurring backups without manual command execution. They are especially useful for production clusters where backup retention must be predictable and copies should be stored outside the cluster itself.

Consider the following when using automated snapshots:

- You can create multiple named snapshot configurations.
- Each configuration defines the schedule, retention policy, and storage type.
- Supported storage types are `local` and `aws-s3`.
- For production environments, local storage is usually less suitable than external object storage: the active node can change over time, and backups are safer when stored outside the system they protect.

Snapshot names use the format `<file_prefix>_<RFC3339_timestamp>.tar.gz`. Retention cleanup only considers files or objects with this naming format under the configured path or prefix.

The snapshot manager checks configurations once a minute. After creating or updating a configuration, the next snapshot is scheduled immediately; subsequent snapshots are scheduled from the start time of the previous snapshot plus the configured `interval`.

## Server configuration (S3 upload pool)

Optional top-level keys in the Stronghold **server** HCL configuration tune background S3 uploads for automated snapshots of Raft storage. These are set in the server configuration file used to start a specific node: `stronghold server -config config.hcl`.

<div class="table__styling--container"></div>

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `snapshot_auto_s3_upload_workers` | Integer | `3` | Maximum concurrent background S3 uploads. Parallelism applies only **across different** snapshot configurations. When the limit is reached, new uploads wait in the queue. Allowed range: `1`–`32`. |
| `snapshot_auto_upload_pool_shutdown` | Duration string | `60s` | Time from the start of the automated snapshot loop shutdown on a specific node (due to process termination or loss of leadership) until background upload processes are forcibly terminated. Allowed range: `1s`–`15m`. |

Example:

```hcl
snapshot_auto_s3_upload_workers = 4
snapshot_auto_upload_pool_shutdown = "90s"
```

Increase the number of background upload processes if you have many S3 snapshot configurations and sufficient network and CPU capacity; decrease it if you hit S3 rate limits or have limited bandwidth.

## Create or update a configuration

| Method | Path |
|--------|------|
| `POST`   | `/sys/storage/raft/snapshot-auto/config/:name` |

The endpoint requires `sudo` privileges.

### Core parameters

<div class="table__styling--container"></div>

| Parameter | Type | Required | Default | Description                                                                                                                                                                                                                                                                  |
|-----------|------|----------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `name` | String | Yes | — | Name of the configuration to create or update                                                                                                                                                                                                                                |
| `interval` | Integer or string | Yes | — | Time between backups. You can specify seconds or Go duration format such as `24h`, `1h30m`. A standalone day value such as `1d` is also accepted; to combine days with smaller units use equivalent hours (e.g. `25h30m` instead of `1d1h30m`). The minimum interval is `3m` |
| `retain` | Integer | No | `3` | Number of backups to keep. Retention cleanup runs only after a snapshot has been successfully saved or uploaded. Older backups are deleted when the limit is exceeded. A value of `0` disables deletion of old snapshots                                                     |
| `storage_type` | Immutable string | Yes | — | Storage type: `local` or `aws-s3`                                                                                                                                                                                                                                            |
| `path_prefix` | String | Yes | — | For `local`, the directory where snapshots are stored. For `aws-s3`, the object prefix inside the bucket; a leading `/` is ignored. For local storage, `path_prefix` cannot be changed after creation.                                                                       |
| `file_prefix` | String | No | `stronghold-snapshot` | Prefix for the file or object name stored in `path_prefix`. For local storage, `file_prefix` cannot be changed after creation.                                                                                                                                               |

### Additional parameters for local

<div class="table__styling--container"></div>

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `local_max_space` | Integer | No | `0` | Maximum number of bytes backup files with the specified `file_prefix` may use in the `path_prefix` directory. A value of `0` disables the check |

### Additional parameters for `aws-s3`

<div class="table__styling--container"></div>

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `aws_s3_bucket` | String | Yes | — | Bucket name used for backup storage |
| `aws_s3_region` | String | No | — | Bucket region |
| `aws_access_key_id` | String | No | — | Access key ID for the bucket |
| `aws_secret_access_key` | String | No | — | Secret access key for the bucket |
| `aws_s3_endpoint` | String | No | — | S3 service endpoint |
| `aws_s3_disable_tls` | Boolean | No | `false` | Disables TLS for the S3 endpoint. Use only for testing |
| `aws_s3_ca_certificate` | String | No | — | CA certificate for the S3 endpoint in PEM format |
| `retry_enabled` | Boolean | No | `false` | Enables retry attempts if uploading a snapshot to S3 fails |
| `retry_max_attempts` | Integer | No | `5` | Maximum number of retry upload attempts after the initial upload failure. Backoff schedule: 1 minute, 5 minutes, 15 minutes, then every 30 minutes. A value of `0` disables the retry limit (unlimited retries) |

Retries reuse the temporary copy from the failed upload. They stop when the configured maximum is reached, when retries are disabled, when the temporary file is no longer available, or when the next scheduled snapshot time is reached. A retry attempt must also finish at least one minute before the next scheduled snapshot; otherwise the attempt is considered failed and the next one is scheduled according to the backoff schedule. Retry failures update `last_snapshot_error` in the status, but do not increment `consecutive_errors` beyond the original failed scheduled snapshot.

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

{{< alert level="info" >}}
Before applying the configuration, make sure the directory from `path_prefix` exists and is writable. The `failed to create snapshot directory at destination` error usually means the directory is missing or unavailable.
{{< /alert >}}

Apply the configuration from `local-snapshot.json` using the following command:

{{< tabs name="stronghold_cmd_6865" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-local-snapshots @local-snapshot.json
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold write sys/storage/raft/snapshot-auto/config/my-local-snapshots @local-snapshot.json
```

{{% /tab %}}
{{< /tabs >}}

{{< alert level="info" >}}
In the example, the path `/stronghold/data/` is a path inside the container, and by default it is mounted at `/var/lib/deckhouse/stronghold/` on the master nodes of the DKP cluster. Created snapshots can be found and downloaded from the master nodes at `/var/lib/deckhouse/stronghold/backups/`. Snapshots will remain unchanged when pods are restarted.
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
  "aws_secret_access_key": "<SECRET_ACCESS_KEY>",
  "retry_enabled": true,
  "retry_max_attempts": 5
}
```

{{< alert level="info" >}}
Before applying the configuration, make sure the bucket already exists and the provided credentials have read and write permissions.
{{< /alert >}}

Apply the configuration from `minio-snapshot.json` using the following command:

{{< tabs name="stronghold_cmd_45767" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-remote-snapshots @minio-snapshot.json
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold write sys/storage/raft/snapshot-auto/config/my-remote-snapshots @minio-snapshot.json
```

{{% /tab %}}
{{< /tabs >}}

### Updating an existing configuration

To modify only selected fields, provide a partial JSON document:

```json
{
  "interval": "3m",
  "retain": "10"
}
```

Apply the modified configuration from `local-snapshot-update.json` using the following command:

{{< tabs name="stronghold_cmd_53206" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold write sys/storage/raft/snapshot-auto/config/my-local-snapshots @local-snapshot-update.json
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold write sys/storage/raft/snapshot-auto/config/my-local-snapshots @local-snapshot-update.json
```

{{% /tab %}}
{{< /tabs >}}

After any successful update, the next snapshot is scheduled immediately.

## List configurations

| Method | Path |
|--------|------|
| `LIST`   | `/sys/storage/raft/snapshot-auto/config` |

Example command:

{{< tabs name="stronghold_cmd_33978" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold list sys/storage/raft/snapshot-auto/config
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold list sys/storage/raft/snapshot-auto/config
```

{{% /tab %}}
{{< /tabs >}}

## Read configuration parameters

| Method | Path |
|--------|------|
| `GET`    | `/sys/storage/raft/snapshot-auto/config/:name` |

Example command:

{{< tabs name="stronghold_cmd_67199" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold read sys/storage/raft/snapshot-auto/config/my-remote-snapshots
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold read sys/storage/raft/snapshot-auto/config/my-remote-snapshots
```

{{% /tab %}}
{{< /tabs >}}

For `aws-s3`, the response does not expose `aws_access_key_id` or `aws_secret_access_key`.

## Delete a configuration

| Method | Path |
|--------|------|
| `DELETE` | `/sys/storage/raft/snapshot-auto/config/:name` |

Example command:

{{< tabs name="stronghold_cmd_50803" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold delete sys/storage/raft/snapshot-auto/config/my-remote-snapshots
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold delete sys/storage/raft/snapshot-auto/config/my-remote-snapshots
```

{{% /tab %}}
{{< /tabs >}}

{{< alert level="info" >}}
Deleting an automated snapshot configuration does not remove existing snapshot files from local or object storage.
{{< /alert >}}

## Read backup status

| Method | Path |
|--------|------|
| `GET`    | `/sys/storage/raft/snapshot-auto/status/:name` |

Example command:

{{< tabs name="stronghold_cmd_25509" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold read sys/storage/raft/snapshot-auto/status/my-remote-snapshots
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold read sys/storage/raft/snapshot-auto/status/my-remote-snapshots
```

{{% /tab %}}
{{< /tabs >}}

Main status fields:

- `consecutive_errors`: Number of backup errors in a row. Resets after a successfully completed snapshot.
- `last_snapshot_end`: End time of the last successful snapshot.
- `last_snapshot_error`: Text of the most recent error during snapshot.
- `last_snapshot_start`: Start time of the last completed snapshot.
- `last_snapshot_url`: Location of the last successful snapshot.
- `last_rotation_error`: Text of the last error during retention cleanup, or `n/a` if cleanup succeeded. A cleanup error does not mark the snapshot itself as failed.
- `next_snapshot_start`: Next scheduled start time.
- `snapshot_start`: Start time of the current backup job.
- `snapshot_url`: Location of the currently written snapshot.

Additional fields for `aws-s3`

- `retry_in_progress`: Flag indicating that a retry of a failed S3 upload is in progress.
- `retry_attempt`: Number of the current retry attempt. `0` if retries are not active.
