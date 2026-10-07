---
title: "Quotas"
description: "Limiting request rate and lease count in Stronghold with sys/quotas quotas: configuration, examples, and verification."
weight: 70
---

Quotas protect Stronghold from being overloaded by individual clients: a misconfigured application that authenticates on every request or creates leases uncontrollably can degrade the availability of the entire cluster.

Stronghold supports two quota types:

- **rate limit quotas** — limit the number of requests per unit of time;
- **lease count quotas** — limit the number of leases that exist at the same time.

Managing quotas requires a token with permissions on `sys/quotas/*`.

## Rate limit quotas

A quota applies globally, to a namespace, to a mount, or to an auth method role. Quota parameters are described in the [`/sys/quotas/rate-limit/{name}`](../../../reference/api/system/#post-sysquotasrate-limitname) API reference:

| Parameter | Description |
| --- | --- |
| `rate` | Maximum number of requests per interval. A positive number |
| `interval` | Interval over which `rate` is counted. Default is `1s` |
| `block_interval` | If set, a client that exceeds the limit is blocked for this duration |
| `path` | Mount or namespace path. An empty value means a global quota |
| `role` | Auth method role the quota applies to |

When a quota is exceeded, Stronghold returns HTTP code `429`.

### Examples

A global quota for the whole cluster:

```shell
d8 stronghold write sys/quotas/rate-limit/global rate=1000
```

A quota on the AppRole auth method to protect against login storms:

```shell
d8 stronghold write sys/quotas/rate-limit/approle-login \
  path=auth/approle \
  rate=50 \
  interval=1s
```

A quota on a specific role that blocks the offender for a minute:

```shell
d8 stronghold write sys/quotas/rate-limit/ci-role \
  path=auth/approle \
  role=ci \
  rate=10 \
  interval=1s \
  block_interval=60s
```

A quota on a secrets engine mount:

```shell
d8 stronghold write sys/quotas/rate-limit/kv-apps \
  path=secret/ \
  rate=200
```

A quota on a namespace (Stronghold EE, see [Namespaces](../../namespaces/overview/)):

```shell
d8 stronghold write sys/quotas/rate-limit/team-a \
  path=team-a/ \
  rate=300
```

If several quotas match a request, the most specific one applies: role, then mount, then namespace, then the global quota.

### Viewing and deleting

```shell
d8 stronghold list sys/quotas/rate-limit
d8 stronghold read sys/quotas/rate-limit/approle-login
d8 stronghold delete sys/quotas/rate-limit/approle-login
```

## General quota settings

The [`/sys/quotas/config`](../../../reference/api/system/#post-sysquotasconfig) endpoint sets general parameters:

| Parameter | Description |
| --- | --- |
| `rate_limit_exempt_paths` | Paths exempt from rate limit quotas |
| `enable_rate_limit_audit_logging` | Log requests rejected by a quota to the audit log |
| `enable_rate_limit_response_headers` | Add HTTP headers with rate limit information to responses |

Example: exempt the health check from quotas and enable response headers:

```shell
d8 stronghold write sys/quotas/config \
  rate_limit_exempt_paths="sys/health" \
  enable_rate_limit_response_headers=true \
  enable_rate_limit_audit_logging=true
```

{{< alert level="info" >}}
The `rate_limit_exempt_paths` parameter replaces the whole list. Before changing it, read the current values with `d8 stronghold read sys/quotas/config`.
{{< /alert >}}

## Lease count quotas

Lease count quotas limit the number of leases that exist at the same time on a mount, in a namespace, or for a role. When the limit is reached, new requests that create leases are rejected until existing leases expire or are revoked.

{{< alert level="warning" >}}
Lease count quotas are part of Stronghold EE and are enabled by a license feature. Without it, the [`/sys/quotas/lease-count/{name}`](../../../reference/api/system/#post-sysquotaslease-countname) endpoints are unavailable (the API reference marks them as `enterprise-stub`). Quota parameters: `path`, `role`, and `max_leases`.
{{< /alert >}}

An example following upstream Vault:

```shell
d8 stronghold write sys/quotas/lease-count/db-creds \
  path=database/ \
  max_leases=5000
```

For accurate per-role lease counting, do not enable the `imprecise_lease_role_tracking` server parameter (see [Configuration](../../../install/standalone/configuration/)).

## Choosing values

1. Collect the actual load from the `stronghold_core_handle_request` and `stronghold_expire_num_leases` metrics (see [Monitoring](../monitoring/)) over several weeks.
1. Set limits with headroom above peak values.
1. Enable `enable_rate_limit_audit_logging` and analyze rejected requests.
1. Gradually lower the limits for clients with abnormal load.
