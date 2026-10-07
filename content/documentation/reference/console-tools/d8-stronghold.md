---
title: "d8 stronghold"
description: "Reference for d8 stronghold subcommands: server operations, unsealing, authentication, tokens, policies, secrets engines, audit, namespaces, and plugins."
weight: 20
---

The `d8 stronghold` command of [Deckhouse CLI](/products/kubernetes-platform/documentation/v1/cli/d8/) is the Stronghold client. Its subcommands and flags match the HashiCorp Vault CLI (see [HashiCorp Vault compatibility](../../../about/vault-compatibility/)).

In a standalone installation, the same subcommands are available through the `stronghold` executable:

```shell
# Stronghold in DKP:
d8 stronghold <command> [flags] [arguments]
# Stronghold on Linux:
stronghold <command> [flags] [arguments]
```

To get help for any subcommand, use the `-help` flag, for example `d8 stronghold kv put -help`.

## Connecting to the server

### Environment variables

| Variable | Purpose | Example |
| --- | --- | --- |
| `STRONGHOLD_ADDR` | Stronghold server address | `export STRONGHOLD_ADDR=https://stronghold.example.com` |
| `STRONGHOLD_TOKEN` | Token used for requests. Set automatically after `login` | `export STRONGHOLD_TOKEN=<TOKEN>` |
| `STRONGHOLD_CACERT` | Path to the CA certificate used to verify the server TLS certificate | `export STRONGHOLD_CACERT=/opt/stronghold/tls/stronghold-ca.pem` |

The client also reads variables that set connection parameters. Before the client starts, every `STRONGHOLD_*` variable is copied to the matching `VAULT_*` variable, so all variables below can be set with either prefix. If both are set, `STRONGHOLD_*` takes effect.

| Variable | Purpose |
| --- | --- |
| `STRONGHOLD_NAMESPACE` | The namespace in which commands run (Stronghold EE) |
| `STRONGHOLD_CAPATH` | A directory with CA certificates |
| `STRONGHOLD_CLIENT_CERT`, `STRONGHOLD_CLIENT_KEY` | A client certificate and key for mutual TLS |
| `STRONGHOLD_TLS_SERVER_NAME` | The server name used to verify the TLS certificate (SNI) |
| `STRONGHOLD_SKIP_VERIFY` | Disable verification of the server TLS certificate. Not recommended |
| `STRONGHOLD_CLIENT_TIMEOUT` | The client request timeout |
| `STRONGHOLD_MAX_RETRIES` | The maximum number of request retries on errors |
| `STRONGHOLD_WRAP_TTL` | The time to live for [response wrapping](../../../concepts/response-wrapping/) |
| `STRONGHOLD_MFA` | MFA data for the request |
| `STRONGHOLD_FORMAT` | The default output format, for example `json` |

In DP, you can get the Stronghold server address from the Ingress object:

```shell
export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
```

### Common flags

| Flag | Purpose |
| --- | --- |
| `-address=<URL>` | Server address. Overrides `STRONGHOLD_ADDR`, for example to query a secondary cluster |
| `-namespace=<path>` | [Namespace](../../../admin/namespaces/overview/) in which the command runs (Stronghold EE) |
| `-format=<format>` | Output format, for example `json` |
| `-field=<field>` | Print only the value of the specified response field (for `read`, `write`, `kv get`, and others) |
| `-help` | Command help |

Flags for connecting to the server:

| Flag | Purpose |
| --- | --- |
| `-ca-cert=<path>` | A file with a CA certificate used to verify the server certificate. Takes precedence over `-ca-path` |
| `-ca-path=<path>` | A directory with CA certificates |
| `-tls-skip-verify` | Disable verification of the server TLS certificate. Not recommended |
| `-output-curl-string` | Do not run the request; print an equivalent `curl` command instead |

## Server and status

| Command | Purpose | Details |
| --- | --- | --- |
| `server -config=<file>` | Start the Stronghold server with the specified configuration file (standalone). In DP, the `stronghold` module manages the server | [Installation](../../../install/standalone/installation/), [configuration](../../../install/standalone/configuration/) |
| `status` | Show the server status: initialization, `Sealed`, seal type, HA mode, Raft indexes | [Access setup](../../../user/get-started/access/) |
| `version` | Show the Stronghold version and edition | — |
| `monitor -log-level=<level>` | Stream node logs in real time | [Logs](../../../admin/operations/logs/) |

```shell
d8 stronghold status
```

## operator

Cluster maintenance commands. They require a privileged token or unseal keys.

| Command | Purpose | Key flags | Details |
| --- | --- | --- | --- |
| `operator init` | Initialize the storage and get the key shares and the root token. In DP, the module does this automatically | `-key-shares`, `-key-threshold` | [Installation](../../../install/standalone/installation/) |
| `operator unseal` | Enter an unseal key share. Repeat as many times as set by `-key-threshold` | — | [Seal and unseal](../../../concepts/seal/) |
| `operator seal` | Seal the node | — | [Seal and unseal](../../../concepts/seal/) |
| `operator rekey` | Change the unseal or recovery keys | `-init`, `-nonce`, `-verify` | [Seal and unseal](../../../concepts/seal/) |
| `operator generate-root` | Generate a new root token using a key quorum | `-init`, `-nonce`, `-decode`, `-otp` | [Tokens](../../../concepts/tokens/) |
| `operator step-down` | Force the active HA node to step down | — | [Switching from EE to CSE](../../../install/standalone/switching-editions/ee-to-cse/) |
| `operator raft list-peers` | Show Raft nodes and their roles | — | [Raft quorum recovery](../../../install/standalone/raft-lost-quorum-recovery/) |
| `operator raft join` | Join a node to the Raft cluster | `-non-voter` | [Configuration](../../../install/standalone/configuration/) |
| `operator raft autopilot state` | Show the Raft autopilot state | — | [Monitoring](../../../admin/operations/monitoring/) |
| `operator raft snapshot save <file>` | Save a snapshot of the Raft storage | — | [Saving a snapshot](../../../admin/backups/save/) |
| `operator raft snapshot restore <file>` | Restore the storage from a snapshot | `-force` | [Restoring](../../../admin/backups/restore/) |
| `operator raft snapshot inspect <file>` | Check the contents and consistency of a snapshot | `-validate`, `-depth`, `-filter`, `-format` | [Inspecting a snapshot](../../../admin/backups/inspect/) |

The `operator seal` command seals the server, and `operator rotate` rotates the storage encryption key (see [Key management](../../../admin/operations/key-management/)).

```shell
d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
```

## Authentication

| Command | Purpose | Key flags | Details |
| --- | --- | --- | --- |
| `login` | Log in to Stronghold and save the token | `-method`, `-path`, `-no-print` | [Access setup](../../../user/get-started/access/) |
| `auth enable <type>` | Enable an auth method | `-path` | [Auth methods](../../../user/auth/overview/) |
| `auth list` | List enabled auth methods | `-format` | [Auth methods](../../../user/auth/overview/) |
| `auth tune <path>` | Change auth method parameters | for example, `-user-lockout-disable` | [userpass](../../../user/auth/userpass/) |

```shell
d8 stronghold login -method=oidc -path=oidc_deckhouse -no-print
```

## token

| Command | Purpose | Key flags | Details |
| --- | --- | --- | --- |
| `token create` | Create a token | `-policy`, `-period`, `-orphan`, `-no-default-policy`, `-namespace` | [Tokens](../../../concepts/tokens/) |
| `token renew [<token>]` | Renew the token lifetime | — | [Tokens](../../../concepts/tokens/) |
| `token revoke <token>` | Revoke the token and its child tokens | `-accessor` | [Tokens](../../../concepts/tokens/) |

```shell
d8 stronghold token create -policy=myapp-read -period=24h
```

## policy

| Command | Purpose | Details |
| --- | --- | --- |
| `policy write <name> <file>` | Create or update an ACL policy. Pass `-` instead of a file to read the policy from stdin | [Policies](../../../concepts/policy/) |

```shell
d8 stronghold policy write myapp-read - <<'POLICY'
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
POLICY
```

## secrets

| Command | Purpose | Key flags | Details |
| --- | --- | --- | --- |
| `secrets enable <type>` | Enable a secrets engine | `-path`, `-version` (for `kv`) | [Secrets engines](../../../user/secrets-engines/overview/) |
| `secrets list` | List enabled secrets engines | `-format` | [Secrets engines](../../../user/secrets-engines/overview/) |
| `secrets tune <path>` | Change mount parameters | `-max-lease-ttl` | [PKI](../../../user/secrets-engines/pki/) |
| `secrets disable <path>` | Disable a secrets engine and delete its data | — | [Secrets engines](../../../user/secrets-engines/overview/) |

```shell
d8 stronghold secrets enable -path=secret -version=2 kv
```

## kv

| Command | Purpose | Key flags | Details |
| --- | --- | --- | --- |
| `kv put <path> <key>=<value>` | Write a secret | `-mount` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv get <path>` | Read a secret | `-mount`, `-version`, `-field` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv patch <path> <key>=<value>` | Partially update a secret | `-mount` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv list <path>` | List keys at a path | `-mount`, `-recursive` | [KV](../../../user/secrets-engines/kv/overview/) |
| `kv delete <path>` | Delete the latest version of a secret (recoverable in KV2) | `-mount`, `-versions` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv undelete <path>` | Restore deleted versions | `-mount`, `-versions` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv destroy <path>` | Permanently delete versions | `-mount`, `-versions` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv metadata get\|put\|patch\|delete <path>` | Manage KV2 secret metadata | `-mount`, `-custom-metadata` | [KV2](../../../user/secrets-engines/kv/kv-v2/) |
| `kv enable-versioning <path>` | Upgrade a KV1 store to KV2 | — | [KV2](../../../user/secrets-engines/kv/kv-v2/) |

```shell
d8 stronghold kv put -mount=secret myapp/db password=s3cr3t
d8 stronghold kv get -mount=secret -field=password myapp/db
```

## Generic commands

These commands work with any API path: secrets engines, auth methods, and the `sys/` system backend.

| Command | Purpose | Key flags |
| --- | --- | --- |
| `read <path>` | Read data | `-field`, `-format`, `-address` |
| `write <path> [<key>=<value>...]` | Write data or perform an operation | `-f`/`-force` (request without data), `-field` |
| `list <path>` | List keys at a path | `-format` |
| `delete <path>` | Delete data at a path | — |
| `path-help <path>` | Show help for the paths of an engine | — |

```shell
d8 stronghold write -f transit/keys/orders/rotate
d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/dr/status
d8 stronghold path-help pki
```

## lease

| Command | Purpose | Key flags | Details |
| --- | --- | --- | --- |
| `lease renew <lease_id>` | Renew a lease | `-increment` | [Leases](../../../concepts/lease/) |
| `lease revoke <lease_id>` | Revoke a lease | `-prefix` | [Leases](../../../concepts/lease/) |

```shell
d8 stronghold lease revoke -prefix database/creds/myapp/
```

## audit

Audit is available in Stronghold EE.

| Command | Purpose | Details |
| --- | --- | --- |
| `audit enable <type> [parameters]` | Enable a `file`, `syslog`, or `socket` audit device | [Audit](../../../admin/audit/overview/) |
| `audit list` | List enabled audit devices | [Audit](../../../admin/audit/overview/) |
| `audit disable <path>` | Disable an audit device | [Audit](../../../admin/audit/overview/) |

```shell
d8 stronghold audit enable file file_path=/var/log/stronghold_audit.log
```

## namespace

Namespaces are available in Stronghold EE.

| Command | Purpose | Key flags | Details |
| --- | --- | --- | --- |
| `namespace create <name>` | Create a namespace | `-namespace` (parent) | [Namespaces](../../../admin/namespaces/overview/) |
| `namespace list` | List child namespaces | `-namespace` | [Namespaces](../../../admin/namespaces/overview/) |
| `namespace lookup <name>` | Show namespace details | `-namespace` | [Namespaces](../../../admin/namespaces/overview/) |
| `namespace delete <name>` | Delete a namespace | `-namespace` | [Namespaces](../../../admin/namespaces/overview/) |
| `namespace lock [<path>]` | Lock the Namespace API | — | [Namespaces](../../../admin/namespaces/overview/) |
| `namespace unlock [<path>]` | Unlock the Namespace API | `-unlock-key` | [Namespaces](../../../admin/namespaces/overview/) |

```shell
d8 stronghold namespace create team-a
```

## plugin

| Command | Purpose | Key flags | Details |
| --- | --- | --- | --- |
| `plugin register <type> <name>` | Register an external plugin in the catalog | `-sha256`, `-version` | [Plugins in standalone](../../../admin/plugins/standalone/), [plugins in DP](../../../admin/plugins/dkp/) |
| `plugin deregister <type> <name>` | Remove a plugin from the catalog | — | [Plugins](../../../admin/plugins/overview/) |

```shell
d8 stronghold plugin register -sha256="${PLUGIN_SHA}" secret my-custom-plugin
```
