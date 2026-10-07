---
title: "Standalone deployment"
description: "The architecture of a standalone Stronghold installation: a single-node server and an HA cluster of three or more servers with integrated Raft storage."
weight: 25
---

In a standalone installation, Stronghold runs as a Linux service (systemd) on dedicated servers. The administrator manages the configuration, TLS certificates, unsealing, and networking. The step-by-step installation is described in [Installation](../../../install/standalone/installation/), and the configuration file parameters in [Configuration](../../../install/standalone/configuration/).

## Deployment options

| Option | Composition | Storage | Purpose |
| --- | --- | --- | --- |
| Single-node server | One server prepared with `stronghold bootstrap service` | Raft on a single node or `filesystem` | Testing, development, small non-critical installations |
| HA cluster | Three or more servers: one active node, the rest standby | Integrated Raft storage | Production installations with fault tolerance |

The `filesystem` backend does not support HA. Use Raft for HA.

By default, `stronghold bootstrap service` creates a configuration with Raft storage.

The `stronghold bootstrap` command also generates a Helm chart and a Docker image for deployment in a third-party Kubernetes cluster. See [Bootstrap](../../../install/standalone/bootstrap/).

## HA cluster layout

![layout_plan_standalone.en.png](../../../images/layout_plan_standalone.en.png)

- One node acquires a lock in storage and becomes active; the others switch to standby.
- Requests that reach a standby node are forwarded to the active node, or the client is redirected to the active node's `api_addr`, depending on the configuration and cluster state. In Stronghold EE, performance standby nodes serve reads locally; see [Performance standby](../../replication/performance-standby/).
- Raft replicates data between all nodes: each node keeps a full copy of the data encrypted by the barrier.
- Quorum requires at least three servers. The quorum is `(n+1)/2` voting nodes. Nodes joined with `retry_join_as_non_voter = true` (the `-non-voter` flag) do not vote.

## Cluster node

On each server:

- binary and configuration: `/opt/stronghold/stronghold`, `/opt/stronghold/config.hcl`;
- Raft data: `/opt/stronghold/data` with `0700` permissions, owned by the `stronghold` system user;
- TLS: the node certificate and key and the root CA in `/opt/stronghold/tls`. Each node gets its own certificate;
- the systemd unit `/etc/systemd/system/stronghold.service`; the service runs as the `stronghold` user with the `CAP_IPC_LOCK` capability.

Key node configuration parameters:

```hcl
cluster_addr  = "https://10.20.30.10:8201"   # address for Raft and request forwarding
api_addr      = "https://10.20.30.10:8200"   # address for clients and redirects
disable_mlock = true

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
  tls_key_file  = "/opt/stronghold/tls/node-1-key.pem"
}

storage "raft" {
  path    = "/opt/stronghold/data"
  node_id = "raft-node-1"

  retry_join {
    leader_api_addr     = "https://10.20.30.11:8200"
    leader_ca_cert_file = "/opt/stronghold/tls/stronghold-ca.pem"
  }
}
```

- `api_addr` is advertised to other nodes for client redirects.
- `cluster_addr` is advertised for node-to-node communication. Traffic between cluster members always uses TLS; the scheme in the address is ignored.
- The `retry_join` blocks list the nodes a new node tries to join at startup. Joining is done through the leader node's API (port `8200`).
- With Raft, `disable_mlock = true` and disabled swap are recommended. A separate `ha_storage` is not declared with Raft.

## Initialization and unsealing

1. The cluster is initialized once, on the first node. By default, the unseal key is split into 5 shares, and 3 are required to unseal (`-key-shares`, `-key-threshold`).
1. With a Shamir seal, each node is unsealed separately after every restart.
1. To avoid unsealing nodes manually, configure auto unseal:
   - [HSM via PKCS#11](../../kms-hsm/hsm/) (`seal "pkcs11"`, Stronghold EE);
   - [Yandex Cloud KMS](../../kms-hsm/yandexcloudkms/) (`seal "yandexcloudkms"`);
   - [`seal "inner-cluster"`](../../../concepts/seal/#auto-unseal-with-inner-cluster): Stronghold EE; unsealed nodes unseal the others, without an external KMS.

All these seal mechanisms are available only in a standalone installation.

## Backups

- Raft snapshots are taken manually; see [Backups](../../backups/overview/).
- Stronghold EE offers scheduled [automated snapshots](../../backups/automated-snapshots/) to a local disk or S3-compatible storage. Snapshots are taken by the active node, so for production keep them in external object storage.

What to do when quorum is lost is described in [Recover from lost quorum](../../../install/standalone/raft-lost-quorum-recovery/).

## Cross-cluster replication

Several standalone Stronghold EE clusters can be connected with Performance or Disaster Recovery replication. The secondary cluster needs access to the primary's cluster port. The architecture and topologies are described in ["Architecture: Stronghold and Stronghold EE"](../../replication/architecture/), and the network requirements in [Network ports](../ports/).
