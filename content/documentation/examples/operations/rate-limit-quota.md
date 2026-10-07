---
title: "Request rate limiting"
linkTitle: "Rate limit"
description: "Protecting Stronghold from overload by a single client: a rate limit quota on a path, checking the 429 response, and viewing and deleting the quota."
weight: 130
params:
  relatedLinks:
    - title: "Quotas"
      url: ../../../admin/operations/quotas/
    - title: "Quotas API"
      url: ../../../reference/api/system/
---

A single faulty or overly active client can consume all cluster resources. A rate limit quota restricts the number of requests per unit of time, and excess requests receive the `429 Too Many Requests` response.

## Goal

Limit the request rate to the `secret/` secrets engine and make sure that exceeding the limit results in a `429` response.

## Prerequisites

- A Stronghold token with permissions on `sys/quotas/rate-limit/*`.
- The `curl` utility.

## Step 1. Create a quota

```bash
d8 stronghold write sys/quotas/rate-limit/secret-limit \
  path="secret/" \
  rate=5 \
  interval=60
```

Quota parameters:

| Parameter | Purpose |
| --- | --- |
| `path` | The path the quota applies to. For a mount path, the quota covers all requests to it. If the parameter is not set, the quota applies globally |
| `rate` | The allowed number of requests per `interval` |
| `interval` | The interval length in seconds |
| `block_interval` | How many seconds a client is blocked after exceeding the limit. By default, there is no blocking |
| `role` | For authentication method paths: the role the quota applies to |

The quota is counted separately for each client IP address.

## Step 2. Check that the limit works

Send ten requests in a row:

```bash
for i in $(seq 1 10); do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -H "X-Vault-Token: $STRONGHOLD_TOKEN" \
    "$STRONGHOLD_ADDR/v1/secret/data/probe"
done | sort | uniq -c
```

The first five requests succeed (the response is `404` because the `probe` secret does not exist, or `200`), and the rest receive `429`:

```text
      5 404
      5 429
```

Requests to other paths, such as `sys/health`, are not affected by the quota.

## Step 3. View and change the quota

```bash
d8 stronghold list sys/quotas/rate-limit
d8 stronghold read sys/quotas/rate-limit/secret-limit
d8 stronghold write sys/quotas/rate-limit/secret-limit path="secret/" rate=100 interval=1
```

After a change, the quota takes effect immediately.

## Verification

```bash
d8 stronghold read -field=rate sys/quotas/rate-limit/secret-limit
```

The command returns `100`.

## Cleanup

```bash
d8 stronghold delete sys/quotas/rate-limit/secret-limit
```
