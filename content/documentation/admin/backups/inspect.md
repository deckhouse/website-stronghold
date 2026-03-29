---
title: "Inspect a snapshot"
weight: 25
description: "Locally inspect and validate a Stronghold integrated Raft snapshot."
---

The `stronghold operator raft snapshot inspect` command lets you inspect an existing snapshot file without restoring it into a cluster.

Unlike `snapshot save` and `snapshot restore`, this command works **locally** and does not require access to a running Stronghold server. That makes it useful for backup validation, initial troubleshooting, and snapshot analysis before a restore.

{{< alert level="warning" >}}
The `inspect` command applies to integrated Raft storage snapshots. If Stronghold uses `etcd`, `postgresql`, or another external backend, validate backups with the tools and procedures of that storage system.
{{< /alert >}}

## Basic usage

```shell
stronghold operator raft snapshot inspect <snapshot_file>
```

Example:

```shell
stronghold operator raft snapshot inspect raft.snap
```

As described for Vault 1.21, the command prints snapshot metadata and a table that shows the number of keys and the amount of space used by each key group.

## What the command shows

The default output includes:

- snapshot metadata such as `ID`, `Size`, `Index`, `Term`, and `Version`;
- grouped storage keys;
- the number of keys in each group;
- the total size of data in each group.

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

## Main flags

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
stronghold operator raft snapshot save /backup/raft.snap
stronghold operator raft snapshot inspect /backup/raft.snap
```

### Validate snapshot consistency

Use the `-validate` flag for a stronger backup check:

```shell
stronghold operator raft snapshot inspect -validate raft.snap
```

This helps confirm that:

- the snapshot is not empty;
- `state.bin` can be parsed correctly;
- critical paths such as `core` and `sys` are present.

This is useful for automated backup checks, but it does not replace a real restore test.

### Analyze selected prefixes

To inspect only a specific part of the stored data, use `-filter` together with `-depth`:

```shell
stronghold operator raft snapshot inspect -depth 3 -filter=core raft.snap
```

This is useful when troubleshooting storage growth or identifying unusually large key groups.

### Export JSON for scripts

For scripts and monitoring workflows, use JSON output:

```shell
stronghold operator raft snapshot inspect -format=json raft.snap
```

When combined with `-validate`, JSON output is convenient for automated processing with tools such as `jq`.

## Important considerations

- `inspect` does not restore a snapshot or modify cluster state;
- checksum and structural validation do not guarantee that the snapshot is fully suitable for restore into a specific cluster;
- for full confidence, periodically run a test restore in a separate environment.

## See also

- [Save a storage snapshot](./save/)
- [Restore from a snapshot](./restore/)
- [Automated snapshots](./automated-snapshots/)
