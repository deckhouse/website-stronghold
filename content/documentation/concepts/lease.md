---
title: "Lease"
description: "Leases of dynamic secrets and tokens: renewal, revocation, prefix-based revocation, lease lookup, and preventing lease explosion."
weight: 10
---

## Lease, renewal, and revoke

For every dynamic secret and authentication service, Stronghold creates a _lease_,
which includes metadata containing information such as duration, renewability, and more.
Stronghold guarantees that the data will be valid for the given duration (Time to Live, TTL).
Once the lease is expired, Stronghold can automatically revoke the data,
and the consumer of the secret can no longer be certain that it is valid.

Consumers of secrets have to check in with Stronghold routinely to either renew the lease (if allowed)
or request a replacement secret.
This improves the value of Stronghold audit logs and significantly simplifies the key replacement process.

All dynamic secrets in Stronghold are required to have a lease.
Even if the data is meant to be valid "forever", a lease is required to force the consumer to check in routinely.

In addition to renewals, a lease can be _revoked_.
When a lease is revoked, it invalidates that secret immediately and prevents any further renewals.

The revocation can be done manually via the API, via the `d8 stronghold lease revoke` CLI command,
via the user interface under the "Access" tab, or automatically by Stronghold.
When a lease is expired, Stronghold automatically revokes it.
When a token is revoked, Stronghold revokes all leases that were created using it.

{{< alert level="info" >}}
The Key/Value backend, which stores arbitrary secrets, doesn't issue leases but sometimes returns a lease duration.
For details, refer to the [`kv` secrets engine documentation](../../user/secrets-engines/kv/overview/).
{{< /alert >}}

## Lease IDs

When reading a dynamic secret (for example, using the `d8 stronghold read`command), Stronghold always returns a `lease_id`.
This ID can be used in commands such as `d8 stronghold lease renew` and `d8 stronghold lease revoke` to manage the lease of a secret.

## Lease duration and renewal

_Lease duration_ is returned along with the lease ID as a Time To Live (TTL) value, time in seconds for which the lease is valid.
A consumer of this secret must renew the lease within that timeframe.

When renewing the lease, the user can request a specific amount of time they want remaining on the lease,
which is called the `increment`.
This increment to the lease duration won't be added at the end of the current TTL but rather at the request time.
For example, the command `d8 stronghold lease renew -increment=3600 my-lease-id` would request
that the TTL of the lease be adjusted to 1 hour (3600 seconds).
Having the increment be rooted at the current time instead of the end of the lease lets users
increase or reduce the length of leases if they don't need a secret for the full possible lease period.

The requested `increment` is completely advisory.
The backend in charge of the secret can choose to completely ignore it.
For most secrets, the backend does its best to respect the `increment`, but often limits it to ensure renewals every so often.

The return value of renewals should be carefully inspected to determine what the new lease TTL is.

## Prefix-based revocation

In addition to revoking a single secret, users with proper access control can revoke multiple secrets based on their lease ID prefix.

The lease ID prefixes always contain the path where the secret was requested from.
This lets you revoke groups of secrets.
For example, to revoke all Userpass logins, it would be enough to run `d8 stronghold lease revoke -prefix auth/userpass/`.

This can be useful if there is an intrusion within a system.
All secrets of a specific backend or a certain configuration can be revoked quickly.

## Managing leases in practice

Use the `d8 stronghold lease` commands and the `sys/leases/*` endpoints to work with leases. Most operations on other users' leases require a policy with access to the corresponding `sys/leases/` paths; listing and prefix-based revocation also require the `sudo` capability.

### Looking up leases

- Get lease metadata: issue time, expiry time, last renewal time, TTL, and whether the lease is renewable:

  ```bash
  d8 stronghold lease lookup database/creds/my-role/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
  ```

  The equivalent API request is `POST sys/leases/lookup` with the `lease_id` parameter.

- List leases under a path prefix (requires `sudo`):

  ```bash
  d8 stronghold list sys/leases/lookup/database/creds/my-role/
  ```

  A call with an empty prefix (`sys/leases/lookup/`) returns the top level of paths that have leases.

- Get the total number of leases and their distribution across mounts:

  ```bash
  d8 stronghold read sys/leases/count
  ```

### Renewing a lease

```bash
d8 stronghold lease renew -increment=1h database/creds/my-role/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
```

Check `lease_duration` in the response: the backend can cap the requested increment at `max_ttl`.

### Revoking leases

- Revoke a single lease:

  ```bash
  d8 stronghold lease revoke database/creds/my-role/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
  ```

- Revoke all leases under a prefix (requires `sudo`):

  ```bash
  d8 stronghold lease revoke -prefix database/creds/my-role/
  ```

  The equivalent API request is `POST sys/leases/revoke-prefix/<prefix>`. The `sync` parameter (defaults to `true`) controls whether revocation is synchronous.

### Force revocation

If the external system is unavailable or has been removed, revocation through the backend fails and the lease stays in Stronghold. In this case, use force revocation (requires `sudo`):

```bash
d8 stronghold lease revoke -force -prefix database/creds/my-role/
```

The equivalent API request is `POST sys/leases/revoke-force/<prefix>`.

{{< alert level="danger" >}}
With force revocation, Stronghold deletes the leases on its side and ignores backend errors. The credentials in the external system may remain valid, so delete them manually.
{{< /alert >}}

The `sys/leases/tidy` endpoint cleans up lease bookkeeping data, which may be needed after errors:

```bash
d8 stronghold write -force sys/leases/tidy
```

## Lease explosion

A lease explosion happens when clients create leases faster than they expire or are revoked. Typical causes:

- an application gets a new token or new dynamic credentials on every request instead of reusing them;
- dynamic secrets and tokens have large TTL values, so leases stay in storage for a long time;
- automation (CI/CD, scripts) logs in again on every run without revoking the tokens it received.

A large number of leases increases storage size, slows down startup and active node failover (leases are reloaded), and a mass simultaneous expiry puts load on Stronghold and external systems.

### Prevention

- Set short `default_ttl` and `max_ttl` for dynamic secret roles and auth methods. Configure mount defaults with `d8 stronghold secrets tune -default-lease-ttl=1h -max-lease-ttl=24h <path>` or `d8 stronghold auth tune ...`.
- Reuse tokens and credentials during their TTL and renew them instead of requesting new ones. For applications, use [Stronghold Agent](../../user/agent/overview/) with caching and automatic renewal.
- Revoke tokens and leases when they are no longer needed, for example at the end of a CI job (`d8 stronghold token revoke -self`).
- For large numbers of short-lived operations, use batch tokens: they are not persisted in Stronghold. See [Tokens](../tokens/).
- Limit request rates with `rate-limit` quotas (`sys/quotas/rate-limit/<name>`), for example for an auth method path:

  ```bash
  d8 stronghold write sys/quotas/rate-limit/login-limit \
    path="auth/approle/" \
    rate=50 \
    interval=1s
  ```

- Monitor the number of leases with `sys/leases/count` and monitoring metrics.

Lease count quotas (`sys/quotas/lease-count`) are part of Stronghold EE and are enabled by a license feature, see [Quotas](../../admin/operations/quotas/).
