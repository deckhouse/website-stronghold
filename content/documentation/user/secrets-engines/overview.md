---
title: "Secrets engines"
linkTitle: "Overview"
weight: 5
---

Secrets engines are components which store, generate, or encrypt data. Secrets
engines are incredibly flexible, so it is easiest to think about them in terms
of their function. Secrets engines are provided some set of data, they take some
action on that data, and they return a result.

Some secrets engines simply store and read data - like the K/V plugin. Other
secrets engines connect to other services and generate dynamic credentials on
demand. Other secrets engines provide encryption as a service, totp
generation, certificates, and much more.

Secrets engines are enabled at a **path** in Stronghold. When a request comes to
Stronghold, the router automatically routes anything with the route prefix to the
secrets engine. In this way, each secrets engine defines its own paths and
properties. To the user, secrets engines behave similar to a virtual filesystem,
supporting operations like read, write, and delete.

## Available secrets engines

| Secrets engine | Type (`secrets enable`) | Purpose | Editions |
|----------------|-------------------------|---------|----------|
| [Cubbyhole](../cubbyhole/) | `cubbyhole` | Per-token private secret storage. Enabled by default. | All editions |
| [Databases](../databases/overview/): [PostgreSQL](../databases/postgresql/), [MySQL/MariaDB](../databases/mysql-maria/), [ClickHouse](../databases/clickhouse/) | `database` | Dynamic and static database credentials. | Stronghold, Stronghold EE, Stronghold CSE |
| [GitOps](../gitops/overview/) | `gitops` | Applying Stronghold configuration from a Git repository gated by a quorum of commit signatures. | Stronghold EE |
| [Identity](../identity/overview/) | `identity` | Entities, groups, the OIDC provider, and identity tokens. Enabled by default. | All editions |
| [Kubernetes](../kubernetes/) | `kubernetes` | Dynamic Kubernetes service account tokens and accounts. | Stronghold, Stronghold EE, Stronghold CSE |
| [KV version 1](../kv/kv-v1/) and [KV version 2](../kv/kv-v2/) | `kv` | Storing arbitrary static secrets; KV v2 supports versioning. | Stronghold, Stronghold EE, Stronghold CSE |
| [LDAP](../ldap/) | `ldap` | Managing LDAP/Active Directory accounts: static and dynamic credentials, service account check-out. | Not specified |
| [PKI](../pki/) | `pki` | Issuing and revoking X.509 certificates, including GOST. | Stronghold, Stronghold EE, Stronghold CSE (GOST: Stronghold and Stronghold EE) |
| [RabbitMQ](../rabbitmq/) | `rabbitmq` | Dynamic RabbitMQ user credentials. | Not specified |
| [SSH](../signed-ssh-certificates/) | `ssh` | SSH certificate signing (CA) and [one-time passwords (OTP)](../ssh-otp/). | Stronghold, Stronghold EE, Stronghold CSE |
| [TOTP](../totp/) | `totp` | Generating and validating TOTP codes. | Not specified |
| [Transit](../transit/) | `transit` | Encryption as a service: encryption, signing, and HMAC without storing data. | Not specified (GOST: Stronghold and Stronghold EE) |
| [trdl](../trdl/) | `trdl` | Building, signing, and publishing releases from Git gated by a quorum of signatures. | Stronghold EE |

For an edition comparison, see [Editions](../../../about/editions/). KV replication is available in Stronghold EE and Stronghold CSE; see [KV1/KV2 replication](../kv/kv-replication/).

<!-- TODO(verify): edition availability of ldap, rabbitmq, totp, transit; the editions page does not list them. The gitops and trdl edition is taken from the release notes. -->

## Secrets engines lifecycle

Most secrets engines can be enabled, disabled, tuned, and moved via the CLI or
API.

- `Enable` - This enables a secrets engine at
  a given path. With a few exceptions, secrets engines can be enabled at multiple
  paths. Each secrets engine is isolated to its path. By default, they are
  enabled at their "type" (e.g. "kv" enables at `kv/`).

{{< alert level="critical" >}}

 **Case-sensitive:** The path where you enable secrets engines is case-sensitive. For
  example, the KV secrets engine enabled at `kv/` and `KV/` are treated as two
  distinct instances of KV secrets engine.

{{< /alert >}}
- `Disable` - This disables an existing
  secrets engine. When a secrets engine is disabled, all of its secrets are
  revoked (if they support it), and all the data stored for that engine in
  the physical storage layer is deleted.

- `Move` - This moves the path for an existing
  secrets engine. This process revokes all secrets, since secret leases are tied
  to the path where they were created. The configuration data stored for the engine
  persists through the move.

- `Tune` - This tunes global configuration for
  the secrets engine such as the TTLs.

Once a secrets engine is enabled, you can interact with it directly at its path
according to its own API. Use `d8 stronghold path-help` to determine the paths it
responds to.

Note that mount points cannot conflict with each other in Stronghold. There are
two broad implications of this fact. The first is that you cannot have
a mount which is prefixed with an existing mount. The second is that you
cannot create a mount point that is named as a prefix of an existing mount.
As an example, the mounts `foo/bar` and `foo/baz` can peacefully coexist
with each other whereas `foo` and `foo/baz` cannot

## Barrier view

Secrets engines receive a _barrier view_ to the configured Stronghold physical
storage. This is a lot like a [chroot](https://en.wikipedia.org/wiki/Chroot).

When a secrets engine is enabled, a random UUID is generated. This becomes the
data root for that engine. Whenever that engine writes to the physical storage
layer, it is prefixed with that UUID folder. Since the Stronghold storage layer
doesn't support relative access (such as `../`), this makes it impossible for an
enabled secrets engine to access other data.

This is an important security feature in Stronghold - even a malicious engine
cannot access the data from any other engine.
