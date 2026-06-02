---
title: "Cubbyhole secrets engine"
weight: 30
---

The `cubbyhole` secrets engine is used to store arbitrary secrets in storage bound to a token.
Paths in `cubbyhole` are scoped to tokens.
A token cannot access another token’s `cubbyhole`.
When a token expires, its `cubbyhole` is destroyed.

Unlike the `kv` secrets engine, the lifetime of `cubbyhole` is tied to the lifetime of the authentication token.
Because of this, values stored in `cubbyhole` do not use TTL or a renewal interval.

Writing to a key in the `cubbyhole` secrets engine completely replaces the previous value.

## Configuration

Most secrets engines must be configured before they can perform their functions.
These actions are usually performed by an operator or a configuration management tool.

The `cubbyhole` secrets engine is enabled by default.
It cannot be disabled, moved, or enabled multiple times.

## Usage

After the secrets engine is configured and you have a token with the required permissions, you can write keys with arbitrary values.

Perform the following steps:

1. Write arbitrary data.

   ```console
   $ d8 stronghold write cubbyhole/my-secret my-value=s3cr3t

   Success! Data written to: cubbyhole/my-secret
   ```

1. Read arbitrary data.

   ```console
   $ d8 stronghold read cubbyhole/my-secret

   Key         Value
   ---         -----
   my-value    s3cr3t
   ```
