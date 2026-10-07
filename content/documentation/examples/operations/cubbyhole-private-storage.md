---
title: "Private token storage (cubbyhole)"
linkTitle: "Cubbyhole"
description: "The Cubbyhole secrets engine: every token has its own private storage that is inaccessible to other tokens, including root; the data is deleted together with the token."
weight: 140
params:
  relatedLinks:
    - title: "Cubbyhole secrets engine"
      url: ../../../user/secrets-engines/cubbyhole/
    - title: "Response wrapping"
      url: ../../../concepts/response-wrapping/
---

Cubbyhole is a secrets engine that is always enabled and mounted at `cubbyhole/`. Every token has its own isolated storage in it: other tokens, including root, cannot see it. When a token expires or is revoked, its storage is deleted. [Response wrapping](../../../concepts/response-wrapping/) is built on this.

## Goal

Make sure that data in cubbyhole is available only to the token that created it and disappears together with it.

## Prerequisites

- A Stronghold token with permission to create tokens.
- The `default` policy on the created tokens: it allows working with `cubbyhole/*`.

## Step 1. Write data with your token

```bash
d8 stronghold write cubbyhole/notes text="personal note"
d8 stronghold read cubbyhole/notes
```

## Step 2. Create a second token and check isolation

1. Create a token with the `default` policy:

   ```bash
   TOKEN2=$(d8 stronghold token create -policy=default -ttl=10m -field=token)
   ```

1. The second token does not see the first token's note, although the path is the same:

   ```bash
   # reading with another token
   STRONGHOLD_TOKEN="$TOKEN2" d8 stronghold read cubbyhole/notes
   ```

   The command returns `No value found at cubbyhole/notes`.

1. The second token has its own storage:

   ```bash
   STRONGHOLD_TOKEN="$TOKEN2" d8 stronghold write cubbyhole/notes text="second token note"
   STRONGHOLD_TOKEN="$TOKEN2" d8 stronghold read cubbyhole/notes
   d8 stronghold read cubbyhole/notes
   ```

   The first read command returns the second token's note, and the second returns the first token's note.

## Step 3. Check that data is deleted together with the token

Revoke the second token. Its storage is deleted with it, and the data cannot be recovered:

```bash
d8 stronghold token revoke "$TOKEN2"
```

## Verification

The first token's note is still in place:

```bash
d8 stronghold read -field=text cubbyhole/notes
```

## Usage

Cubbyhole is suitable for temporary storage of data that must not be available to anyone except one token: intermediate script results, one-time values when handing over secrets. For long-term storage and shared access, use [KV](../kv-v2-versioning/).

## Cleanup

```bash
d8 stronghold delete cubbyhole/notes
```
