---
title: "Overview"
weight: 10
---

The `kv` secrets engine is a general-purpose key-value store for arbitrary secrets in the configured Stronghold physical storage.
It can operate in two modes:

- version 1 of `kv` — for storing a single current value per key;
- version 2 of `kv` — with versioning and the ability to store a configurable number of versions for each key.

## KV version 1

In version 1 of `kv`, which does not support versioning, only the most recently updated value is stored for a key.
The main advantage of this mode is lower storage consumption for each key,
because no additional metadata or change history is stored.
In addition, requests to this secrets engine are usually processed faster,
because each request requires fewer accesses to the data store and no locking occurs when the key value changes.

## KV version 2

In version 2 of `kv`, you can store a configurable number of versions for each key.
By default, 10 versions are retained.
From each stored version, you can retrieve both data and metadata.
To protect against accidental data overwrites, you can use Check-and-Set operations.
When a version is deleted, the underlying data is not removed immediately and is instead marked as deleted.
Such deletion can be undone.
To permanently remove version data, use the `destroy` command or send a request to the corresponding API endpoint.
To delete all versions and metadata of a key, use the metadata `delete` command or the corresponding API endpoint.
You can configure separate ACLs for these operations to manage permissions for soft deletion,
restoration, and permanent data removal independently.

## Usage

To work with the `kv` secrets engine, use the `d8 stronghold kv <subcommand> [options] [args]` command.
The available subcommands are listed in the table below:

| Subcommand | `kv` v1 | `kv` v2 | Description |
| --- | --- | --- | --- |
| `delete` | x | x | Deletes secret versions from `kv` |
| `destroy` | — | x | Permanently deletes one or more versions of a secret |
| `enable-versioning` | — | x | Enables versioning for an existing `kv` v1 store |
| `get` | x | x | Retrieves data |
| `list` | x | x | Lists data or secrets |
| `metadata` | — | x | Performs operations on `kv` store metadata |
| `patch` | — | x | Updates secrets without overwriting existing values |
| `put` | x | x | Creates or updates secrets by replacing existing values |
| `rollback` | — | x | Rolls back a secret to a previous version |
| `undelete` | — | x | Restores a deleted version of a secret |
