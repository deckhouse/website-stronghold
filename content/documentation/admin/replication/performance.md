---
title: "Performance replication"
linkTitle: "Performance replication"
weight: 20
description: "Configure Performance replication between Stronghold clusters, verify it, and manage path filters."
---

Performance replication copies the primary cluster's storage (except node-local
paths) to one or more performance secondaries. A secondary serves reads locally
and forwards writes to the primary. This lets you scale read traffic and place
it closer to consumers.

## Before you start

- Make sure both clusters run Stronghold EE with integrated Raft storage and
  that replication is enabled (see [Overview](../overview/)).
- Make sure the primary cluster port is reachable from the secondary and that
  you have the primary CA certificate for TLS.
- Prepare a token with permissions for `sys/replication/*` on both clusters.

In the examples below `${PRIMARY_ADDR}` and `${SECONDARY_ADDR}` are the API
addresses of the clusters, and `${VAULT_TOKEN}` is a token authorized for
`sys/replication/*`.

{{< alert level="info" >}}
If you enable replication on a cluster that already holds data (for example,
after upgrading from a build without replication), or after a period with
replication disabled, the node re-indexes its storage on the first boot by
itself. You do not need to call `sys/replication/reindex` manually.
{{< /alert >}}

## Step 1. Enable the primary

```shell
d8 stronghold write -force sys/replication/performance/primary/enable
```

{{< alert level="warning" >}}
Enabling a primary restarts the core, and the node is unavailable for a few
seconds. Before continuing, wait until it responds again and
`sys/replication/performance/status` shows `mode: primary`.
{{< /alert >}}

## Step 2. Create an activation token for the secondary

Generate a wrapping activation token for a specific secondary identified by
`id`:

```shell
d8 stronghold write sys/replication/performance/primary/secondary-token \
  id=sec-1 ttl=24h
```

The `id` parameter is required; `ttl` defaults to `24h`. The command returns a
short-lived wrapping token in the `wrap_info.token` field — pass exactly this
token to the secondary.

## Step 3. Enable the secondary

Provide the wrapping token from the previous step. For self-signed environments,
`ca_cert` (the primary CA in PEM format) is required, otherwise the TLS
connection to the primary cannot be verified.

```shell
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request POST \
  --data @secondary-enable.json \
  "${SECONDARY_ADDR}/v1/sys/replication/performance/secondary/enable"
```

Example `secondary-enable.json`:

```json
{
  "token": "<wrapping_token_from_step_2>",
  "primary_api_addr": "<primary_api_address>",
  "ca_cert": "<primary_ca_certificate_in_pem>"
}
```

Fields of `secondary/enable`:

- `token` — required; the wrapping token from step 2.
- `primary_api_addr` — the primary address used to unwrap the token; required if
  the token does not contain an `addr` claim.
- `ca_cert` — inline PEM CA of the primary for the TLS unwrap.
- `ca_file` / `ca_path` — path to a PEM file or a directory of PEM files, read
  on the server.
- `client_cert_pem` / `client_key_pem` — optional client certificate and key for
  the unwrap TLS connection.

{{< alert level="warning" >}}
After bootstrap the secondary wipes its own token store, so the secondary's
original root token stops working. Log in through a **replicated** auth method
(for example, `userpass`) that was configured on the primary **before**
replication was enabled.
{{< /alert >}}

## Step 4. Verify the status

```shell
d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/performance/status
```

When the secondary is connected and pulling the WAL, `state` is `stream-wals`
and `connection_state` is `ready`. The secondary has caught up with the primary
when its `last_wal` reaches `last_remote_wal`.

## Step 5. Verify data replication

Run the commands against the secondary with a token from the replicated auth
method you logged in with in step 3 — the primary's root token is not valid on
the secondary.

```shell
# Write on the primary.
d8 stronghold kv put -address="${PRIMARY_ADDR}" secret/hello value=world

# Read on the secondary once it has converged.
d8 stronghold kv get -address="${SECONDARY_ADDR}" secret/hello

# A write on the secondary is forwarded to the primary and returns through the WAL.
d8 stronghold kv put -address="${SECONDARY_ADDR}" secret/fromsec value=1
```

## Path filters

A primary can selectively withhold specific mounts or namespaces from a specific
secondary. Filters are configured per secondary by `id`:

- `deny` — replicate everything **except** the listed paths.
- `allow` — replicate **only** the listed paths.

The default mode is `deny`.

```shell
# Deny the mount blocked/ for sec-1.
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request POST \
  --data '{"mode":"deny","paths":["blocked/"]}' \
  "${PRIMARY_ADDR}/v1/sys/replication/performance/primary/paths-filter/sec-1"

# Read the filter.
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  "${PRIMARY_ADDR}/v1/sys/replication/performance/primary/paths-filter/sec-1"

# Read the expanded filter (the actual storage prefixes that will apply).
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  "${PRIMARY_ADDR}/v1/sys/replication/performance/primary/paths-filter/sec-1/dynamic"

# Remove the filter.
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request DELETE \
  "${PRIMARY_ADDR}/v1/sys/replication/performance/primary/paths-filter/sec-1"
```

A filter change is applied live: when a path is added to `deny` (or removed from
`allow`), the primary streams an invalidation and the secondary drops the
previously allowed data; when the change is reversed, the data is replicated
again.

{{< alert level="warning" >}}
A performance secondary cannot be promoted while path filters are configured on
it. Remove the filters first.
{{< /alert >}}

When a filtered mount is unmounted, remounted, or disabled, its storage prefix
is automatically removed from all filters, so a stale entry does not govern a
new mount at the same path.

## Management operations

| Action | Endpoint |
| --- | --- |
| Status | `d8 stronghold read sys/replication/performance/status` |
| Revoke a secondary | `d8 stronghold write sys/replication/performance/primary/revoke-secondary id=sec-1` |
| Disable on the primary | `d8 stronghold write -force sys/replication/performance/primary/disable` |
| Disable on the secondary | `d8 stronghold write -force sys/replication/performance/secondary/disable` |
| Demote the primary | `d8 stronghold write -force sys/replication/performance/primary/demote` |
| Promote a secondary | `d8 stronghold write -force sys/replication/performance/secondary/promote` |
| Re-point a secondary | `d8 stronghold write sys/replication/performance/secondary/update-primary token=<activation_token>` |

Notes:

- `revoke-secondary` also removes the per-secondary filter, so it leaves no
  orphaned entries.
- For `update-primary`, when the new primary has a **new** identity (for
  example, a promoted former secondary), use the token method with a fresh
  activation token from the new primary. The address method
  (`primary_cluster_addr`) is only valid if the primary kept its identity.
