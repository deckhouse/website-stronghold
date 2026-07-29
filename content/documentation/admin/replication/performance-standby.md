---
title: "Performance standby"
linkTitle: "Performance standby"
weight: 40
params:
  edition: ee
description: "Serve read requests from non-active HA nodes within a single Stronghold cluster."
---

Performance standby lets the non-active HA nodes **inside a single cluster**
serve read requests locally instead of forwarding every request to the active
node. This spreads read load across all cluster nodes and reduces latency, while
writes and other operations that change state are still forwarded to the active
node.

Unlike [Performance replication](../performance/) and
[Disaster recovery](../disaster-recovery/), which operate across separate
clusters, performance standby works within one cluster and requires no
cross-cluster setup.

## How it works

- In an HA cluster, one node is active and the others are standby nodes.
- In Stronghold EE, each standby node becomes a performance standby: it handles
  read requests against its local state.
- Requests that modify state (writes, deletes) are forwarded to the active node
  and applied there.

Performance standby is enabled by default in EE builds. No additional
configuration is required beyond running an HA cluster on integrated Raft
storage.

## Verify performance standby

Send a read request directly to a standby node with the
`X-Vault-No-Request-Forwarding` header. A performance standby still returns
`200` for a read because it serves it locally; a request that requires the
active node returns a non-2xx status when forwarding is suppressed.

```shell
curl \
  --header "X-Vault-Token: ${VAULT_TOKEN}" \
  --header "X-Vault-No-Request-Forwarding: true" \
  "${STANDBY_ADDR}/v1/secret/data/hello"
```

You can also check the node role in the health and status output — a performance
standby reports itself as a standby that serves reads.

## Disable performance standby

To turn off performance standby without affecting replication, set the following
parameter in the server configuration:

```hcl
disable_performance_standby = true
```

With this parameter, standby nodes forward all requests to the active node, as
they would without the Enterprise standby capability.

{{< alert level="info" >}}
The `disable_performance_standby` parameter applies to the server configuration
of a Standalone Stronghold installation. Turning off the whole replication
pipeline with `disable_wal_replication = true` also disables performance
standby.
{{< /alert >}}
