---
title: "Disaster recovery"
linkTitle: "Disaster recovery"
weight: 30
description: "Configure Disaster Recovery replication, promote a DR secondary during a failover, and manage the DR operation token ceremony."
---

Disaster Recovery (DR) replication keeps a hot standby cluster — a full copy of
the primary's storage together with local data such as tokens and leases. A DR
secondary does not serve clients (beyond unseal and a small internal set) and
waits for a promote to take over when the primary fails.

## Before you start

- Make sure both clusters run Stronghold EE with integrated Raft storage and
  that replication is enabled (see [Overview](../overview/)).
- Make sure the primary cluster port is reachable from the secondary and that
  you have the primary CA certificate for TLS.
- Prepare a token with permissions for `sys/replication/*` on the primary.
- Arrange access to the key shares for the promote ceremony: unseal-key shares
  for manual unsealing (Shamir), or recovery-key shares for auto-unseal. You
  need a quorum of them, according to the threshold set at initialization.

In the examples below `${PRIMARY_ADDR}` and `${SECONDARY_ADDR}` are the API
addresses of the clusters, and `${VAULT_TOKEN}` is a token authorized for
`sys/replication/*`.

{{< alert level="info" >}}
If you enable replication on a cluster that already holds data (for example,
after upgrading from a build without replication), or after a period with
replication disabled, the node re-indexes its storage on the first boot by
itself. You do not need to call `sys/replication/reindex` manually.
{{< /alert >}}

## Step 1. Enable the DR primary

```shell
d8 stronghold write -force sys/replication/dr/primary/enable
```

## Step 2. Create an activation token for the DR secondary

```shell
d8 stronghold write sys/replication/dr/primary/secondary-token id=dr-1 ttl=24h
```

The `id` parameter is required; `ttl` defaults to `24h`. The command returns a
wrapping token in the `wrap_info.token` field — pass exactly this token to the
secondary.

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

For self-signed environments, `ca_cert` (the primary CA in PEM format) is
required. The DR secondary copies the entire storage, including local data, and
does not serve client requests.

## Step 4. Verify the status

```shell
d8 stronghold read -address="${SECONDARY_ADDR}" sys/replication/dr/status
```

The status shows `mode`, `state`, `connection_state`, `last_wal`, and
`last_remote_wal`. The DR secondary has caught up with the primary when its
`last_wal` reaches `last_remote_wal`.

## Promote a DR secondary

Promoting requires a **DR operation token**, issued through a ceremony that
combines key shares with a one-time password (OTP).

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

   When `complete` becomes `true`, the response contains `encoded_token`. Decode
   it with the `otp` to obtain the DR operation token.

1. Promote the secondary with the DR operation token:

   ```shell
   curl \
     --header "X-Vault-Token: ${VAULT_TOKEN}" \
     --request POST \
     --data '{"dr_operation_token":"<dr_operation_token>"}' \
     "${SECONDARY_ADDR}/v1/sys/replication/dr/secondary/promote"
   ```

After a promote, the former secondary becomes an active DR primary and serves
clients.

## Recover the former primary

After a failover, the cluster already has an active primary — the promoted
secondary. Do not start the former primary back up as a primary, or the cluster
ends up with two primaries. Instead, reconnect it to the new primary as a DR
secondary.

1. On the new primary, create an activation token:
   `d8 stronghold write sys/replication/dr/primary/secondary-token id=<id>`.
1. If the former primary still considers itself a primary, demote it:
   `d8 stronghold write -force sys/replication/dr/primary/demote`.
1. Reconnect it to the new primary:
   `d8 stronghold write sys/replication/dr/secondary/update-primary token=<activation_token>`.

Use the token method: the new primary has a different identity, and
`primary_cluster_addr` does not apply to it.

## Management operations

| Action | Endpoint |
| --- | --- |
| Status | `d8 stronghold read sys/replication/dr/status` |
| Revoke a secondary | `d8 stronghold write sys/replication/dr/primary/revoke-secondary id=dr-1` |
| Disable on the primary | `d8 stronghold write -force sys/replication/dr/primary/disable` |
| Demote the primary | `d8 stronghold write -force sys/replication/dr/primary/demote` |
| Re-point a secondary | `d8 stronghold write sys/replication/dr/secondary/update-primary token=<activation_token>` |

Notes:

- `demote` lowers a DR primary to a disabled DR secondary while keeping the
  cluster ID and local data, so it is ready to reconnect without a wipe.
- Planned switchover: `demote` the current primary, `promote` the standby, then
  `update-primary` to bring the former primary back as a secondary (see
  [Recover the former primary](#recover-the-former-primary)).
