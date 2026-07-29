---
title: "Disaster recovery"
linkTitle: "Disaster recovery"
weight: 30
params:
  edition: ee
description: "Configure Disaster Recovery replication, promote a DR secondary during a failover, and manage the DR operation token ceremony."
---

Disaster Recovery (DR) replication keeps a hot standby cluster that mirrors the **entire** non-ignored keyspace of the primary, including local data such as tokens and leases. A DR secondary does not serve client requests (beyond unseal and a small internal set) — it waits for a promote and takes over when the primary fails.

## Before you start

- Make sure both clusters run Stronghold EE with integrated Raft storage and that replication is enabled (see [Overview](../overview/)).
- Make sure the primary cluster port is reachable from the secondary and that you have the primary CA certificate for TLS.
- Prepare a token with permissions for `sys/replication/*` on the primary.
- Keep the holders of unseal or recovery key shares available — they are required for the promote ceremony.

In the examples below `${PRIMARY_ADDR}` and `${SECONDARY_ADDR}` are the API addresses of the clusters, and `${VAULT_TOKEN}` is a token authorized for `sys/replication/*`.

## Step 1. Enable the DR primary

```shell
d8 stronghold write -force sys/replication/dr/primary/enable
```

## Step 2. Create an activation token for the DR secondary

```shell
d8 stronghold write sys/replication/dr/primary/secondary-token id=dr-1
```

The command returns a wrapping token in the `wrap_info.token` field — pass exactly this token to the secondary.

## Step 3. Enable the DR secondary

```shell
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --request POST \
  --data @dr-secondary-enable.json \
  "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/enable"
```

Example `dr-secondary-enable.json`:

```json
{
  "token": "<wrapping_token_from_step_2>",
  "primary_api_addr": "<primary_api_address>",
  "ca_cert": "<primary_ca_certificate_in_pem>"
}
```

For self-signed environments, `ca_cert` (the primary CA in PEM format) is required. The DR secondary replicates the entire keyspace, including local data, and does not serve client requests.

## Step 4. Verify the status

```shell
d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/dr/status
```

## Promote a DR secondary

Promoting a DR secondary requires a **DR operation token**, which is issued through a multi-step ceremony similar to generate-root: it combines shares of the unseal or recovery keys with a one-time password (OTP).

1. Start the ceremony. The response returns a `nonce` and an `otp`:

   ```shell
   curl \
     --request PUT \
     --data '{}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/generate-operation-token/attempt"
   ```

1. Submit a key share. Repeat for each required share:

   ```shell
   curl \
     --request PUT \
     --data '{"key":"<unseal_or_recovery_key_share>","nonce":"<nonce>"}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/generate-operation-token/update"
   ```

   When `complete` becomes `true`, the response contains `encoded_token`. Decode it with the `otp` to obtain the DR operation token.

1. Promote the secondary with the DR operation token:

   ```shell
   curl \
     --header "X-Vault-Token: ${VAULT_TOKEN}" \
     --request POST \
     --data '{"dr_operation_token":"<dr_operation_token>"}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/promote"
   ```

After a successful promote, the former secondary becomes an active DR primary and serves client requests.

## Management operations

| Action | Endpoint |
| --- | --- |
| Status | `d8 stronghold read sys/replication/dr/status` |
| Revoke a secondary | `d8 stronghold write sys/replication/dr/primary/revoke-secondary id=dr-1` |
| Disable on the primary | `d8 stronghold write -force sys/replication/dr/primary/disable` |
| Demote the primary | `d8 stronghold write -force sys/replication/dr/primary/demote` |
| Re-point a secondary | `d8 stronghold write sys/replication/dr/secondary/update-primary token=<activation_token>` |

Notes:

- `demote` lowers a DR primary to a disabled DR secondary while **keeping** the cluster ID and local data, so it is ready to reconnect without a wipe.
- Use `demote` on the old primary and `promote` on the standby to perform a controlled failover; use `update-primary` to reconnect the old primary to the newly promoted one.
