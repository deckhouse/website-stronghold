---
title: "Secret engines"
linkTitle: "Overview"
weight: 5
---

Secret engines are components that store, generate, or encrypt data.
It is most convenient to describe them in terms of the functions they perform.
A secret engine receives a set of data, performs specific actions on it, and returns the result.

Some secret engines only store and read data, for example K/V.
Others connect to external services and generate dynamic credentials on request.
Others provide encryption as a service, TOTP code generation, certificates, and other capabilities.

Secret engines are mounted at a specific path in Stronghold.
When Stronghold receives a request, the router automatically forwards everything that matches the route prefix to the corresponding secret engine.
As a result, each secret engine defines its own paths and properties.
For users, secret engines work similarly to a virtual file system and support read, write, and delete operations.

## Secret engine lifecycle

Most secret engines can be enabled, disabled, tuned, and moved through the CLI or API.

The following operations are available:

- `enable` — Enables a secret engine at the specified path. With some exceptions, a secret engine can be enabled at multiple paths. Each instance is isolated within its own path. By default, a secret engine is mounted at a path that matches its type. For example, `kv` is mounted at `kv/`.
- `disable` — Disables an existing secret engine. When a secret engine is disabled, all of its secrets are revoked if this is supported. All data stored for this secret engine in physical storage is deleted.
- `move` — Moves the path of an existing secret engine. During the move, all secrets are revoked because secret leases are tied to the path where they were created. The secret engine configuration is preserved.
- `tune` — Tunes the global configuration of a secret engine, for example TTL values.

{{< alert level="warning" >}}
The path where secret engines are mounted is case-sensitive.
For example, a K/V secret engine mounted at `kv/` and `KV/` is treated as two separate instances.
{{< /alert >}}

After a secret engine is mounted, you can interact with it directly at its path according to its own API.
To see which paths are supported, use the `d8 stronghold path-help` command.

Mount points in Stronghold cannot conflict with each other.
This imposes two restrictions:

- you cannot create a mount whose prefix matches an existing mount;
- you cannot create a mount point whose name is a prefix of an existing mount.

For example, the `foo/bar` and `foo/baz` mounts can coexist, but `foo` and `foo/baz` cannot.

## Barrier view

Secret engines receive a _barrier view_ of the configured Stronghold physical storage.
This is similar to the `chroot` mechanism.

When a secret engine is mounted, a random UUID is generated.
This UUID becomes the root directory for that engine's data.
Whenever the engine writes to physical storage, the path is prefixed with a directory that contains this UUID.
Because the Stronghold storage layer does not support relative access such as `../`, a mounted secret engine cannot access the data of other engines.

This is an important Stronghold security feature.
Even a malicious engine cannot access the data of another engine.
