---
title: "Inspecting a snapshot"
weight: 25
description: "Locally inspect and validate a Stronghold integrated Raft snapshot."
---

You can check and analyze an existing snapshot file without restoring it into a cluster using the `stronghold operator raft snapshot inspect` command.

Unlike `snapshot save` and `snapshot restore`, this command works **locally** and does not require access to a running Stronghold server. That makes it useful for backup validation, initial troubleshooting, and snapshot analysis before a restore.

{{< alert level="warning" >}}
The `inspect` command applies to integrated Raft storage snapshots. If Stronghold uses etcd, PostgreSQL, or another external backend, validate backups with the tools and procedures of that storage system.
{{< /alert >}}

## Basic usage

To check a snapshot, run the following command specifying a path to the snapshot file instead of `<SNAPSHOT_FILE>`:

```shell
d8 stronghold operator raft snapshot inspect <SNAPSHOT_FILE>
```

Example:

```shell
d8 stronghold operator raft snapshot inspect raft.snap
```

The command prints snapshot metadata and a table that shows the number of keys and the amount of space used by each key group.

### Command output

The default output includes:

- Snapshot metadata such as `ID`, `Size`, `Index`, `Term`, and `Version`.
- Grouped storage keys.
- The number of keys in each group.
- The total size of data in each group.

Example abbreviated output:

```text
ID           bolt-snapshot
Size         93290
Index        8957
Term         11
Version      1

Key Name                                          Count      Size
--------                                          -----      ----
wal/logs                                          54         9 KB
index/pages                                       14         26.1 KB
sys/policy                                        3          3.4 KB
core/cluster                                      2          236 B
```

### Main flags

| Flag | Description |
|------|-------------|
| `-details` | Enables detailed key analysis and prints the grouped key table. Enabled by default. |
| `-depth` | Controls grouping depth by path segments. |
| `-filter` | Limits output to keys matching a specific prefix. |
| `-format` | Output format: `table` or `json`. |
| `-validate` | Runs additional snapshot consistency checks. |

## Common scenarios

### Quick validation after backup creation

You can inspect the file immediately after creating it:

```shell
d8 stronghold operator raft snapshot save /backup/raft.snap
d8 stronghold operator raft snapshot inspect /backup/raft.snap
```

### Validation of snapshot consistency

Use the `-validate` flag for a thorough backup check:

```shell
d8 stronghold operator raft snapshot inspect -validate raft.snap
```

This helps confirm that:

- The snapshot is not empty.
- Contents of `state.bin` can be parsed correctly.
- Critical paths such as `core` and `sys` are present.

This is useful for automated backup checks, but it does not replace a real restore test.

### Analysis of selected prefixes

To inspect a specific part of the stored data, use `-filter` together with `-depth`:

```shell
d8 stronghold operator raft snapshot inspect -depth 3 -filter=core raft.snap
```

This is useful when troubleshooting storage growth or identifying unusually large key groups.

### Output in JSON

For scripts and monitoring workflows, use JSON output:

```shell
d8 stronghold operator raft snapshot inspect -format=json raft.snap
```

When combined with `-validate`, JSON output is convenient for automated processing with tools such as `jq`.

## Important considerations

- The `inspect` command does not restore a snapshot or modify cluster state.
- Checksum and structural validation do not guarantee that the snapshot is fully suitable for restore into a specific cluster.
- For full confidence, periodically run a test restore in a separate environment.
