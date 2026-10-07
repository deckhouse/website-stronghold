---
title: "Secret versions in KV v2"
linkTitle: "KV v2 secret versions"
description: "The lifecycle of a secret in KV version 2: new versions, partial updates, reading and rolling back a version, limiting the number of versions, check-and-set, deletion, recovery, and destruction."
weight: 120
params:
  relatedLinks:
    - title: "KV version 2 secrets engine"
      url: ../../../user/secrets-engines/kv/kv-v2/
    - title: "Policies"
      url: ../../../concepts/policy/
---

KV version 2 stores several versions of every secret. This lets you roll back an erroneous change, protect against concurrent overwrites, and control how much history to keep.

## Goal

Go through the full lifecycle of a secret: create several versions, read and roll back a version, limit history retention, enable check-and-set, delete and recover a version, and destroy it permanently.

## Prerequisites

- A Stronghold token with permissions on the `secret/*` path.
- The KV version 2 secrets engine mounted at `secret/` (it is enabled by default in development mode).

## Step 1. Create secret versions

1. Every write creates a new version:

   ```bash
   d8 stronghold kv put -mount=secret app/config username=app password=v1-pass
   d8 stronghold kv put -mount=secret app/config username=app password=v2-pass
   ```

1. The `kv patch` command changes only the specified fields and keeps the others:

   ```bash
   d8 stronghold kv patch -mount=secret app/config password=v3-pass
   ```

   The `username` field stays the same, and `password` becomes `v3-pass`. The result is version 3.

## Step 2. Read a specific version

```bash
d8 stronghold kv get -mount=secret app/config
d8 stronghold kv get -mount=secret -version=1 app/config
d8 stronghold kv metadata get -mount=secret app/config
```

The first command returns the current version, and the second returns version 1. The third shows the metadata: the current version, the creation time of each version, and deletion flags.

## Step 3. Roll back a version

```bash
d8 stronghold kv rollback -mount=secret -version=1 app/config
```

A rollback does not rewrite history: the content of version 1 is written as a new version 4.

## Step 4. Limit history

```bash
d8 stronghold kv metadata put -mount=secret \
  -max-versions=5 \
  -delete-version-after=720h \
  app/config
```

- `max-versions` is how many versions to keep. The oldest version beyond the limit is deleted automatically.
- `delete-version-after` is how long after creation a version is marked as deleted.

## Step 5. Enable check-and-set

Check-and-set protects against overwrites when two clients change a secret at the same time: a write succeeds only if the client specifies the current version.

1. Require a version on every write. The `kv metadata patch` command changes only the specified fields, while `kv metadata put` overwrites all metadata with default values, so `patch` is needed here:

   ```bash
   d8 stronghold kv metadata patch -mount=secret -cas-required=true app/config
   ```

1. A write without a version is now rejected:

   ```bash
   # write without cas
   d8 stronghold kv put -mount=secret app/config password=without-cas
   ```

1. A write with the current version succeeds:

   ```bash
   d8 stronghold kv put -mount=secret -cas=4 app/config username=app password=v5-pass
   ```

   If another client has changed the secret in the meantime, the version does not match and the write is rejected.

## Step 6. Delete, recover, and destroy a version

1. A soft delete marks the latest version as deleted, and the data remains in storage:

   ```bash
   d8 stronghold kv delete -mount=secret app/config
   ```

1. Recover the version:

   ```bash
   d8 stronghold kv undelete -mount=secret -versions=5 app/config
   ```

1. Destroying removes the version data permanently:

   ```bash
   d8 stronghold kv destroy -mount=secret -versions=1,2 app/config
   ```

## Verification

```bash
d8 stronghold kv metadata get -mount=secret app/config
```

In the output, versions 1 and 2 are marked as destroyed (`destroyed: true`), version 5 is recovered and current, `max_versions` is `5`, and `cas_required` is `true`.

## Cleanup

Delete the secret together with all versions and metadata:

```bash
d8 stronghold kv metadata delete -mount=secret app/config
```
