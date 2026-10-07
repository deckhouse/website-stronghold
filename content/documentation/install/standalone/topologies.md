---
title: "Reference topologies"
description: "Ready-to-use Stronghold configuration templates: a single node for development and testing, three- and five-node Raft HA clusters with TLS and retry_join, and auto unseal via HSM, Yandex Cloud KMS, and inner-cluster."
weight: 25
---

This page contains `/opt/stronghold/config.hcl` templates for typical standalone topologies. The parameters are described in [Configuration](../configuration/), and the deployment procedure is described in [Installation](../installation/).

| Topology | Purpose | Fault tolerance |
| --- | --- | --- |
| [Single node](#single-node) | Development, testing, sandboxes | None |
| [Three Raft nodes](#three-node-ha-cluster) | Production | Failure of one node |
| [Five Raft nodes](#five-node-ha-cluster) | Production with increased availability requirements | Failure of two nodes |
| [DR pair](#dr-pair) (Stronghold EE) | Hot standby at another site | Failure of an entire cluster |

All examples use the `/opt/stronghold/data` (Raft data) and `/opt/stronghold/tls` (certificates) directories, the systemd unit from [Installation](../installation/), and ports `8200` (API) and `8201` (cluster traffic). Replace the IP addresses, node names, and paths with the values of your infrastructure.

## Single node

Suitable for development and testing. The fastest way to get this installation is the [`stronghold bootstrap service`](../bootstrap/) command, which creates the directories, configuration, systemd unit, and, if needed, self-signed certificates.

If you prepare the configuration manually, use this template:

```hcl
ui            = true
cluster_addr  = "https://127.0.0.1:8201"
api_addr      = "https://stronghold.demo.tld:8200"
disable_mlock = true

listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/opt/stronghold/tls/stronghold-cert.pem"
  tls_key_file  = "/opt/stronghold/tls/stronghold-key.pem"
}

storage "raft" {
  path    = "/opt/stronghold/data"
  node_id = "raft-node-1"
}
```

{{< alert level="warning" >}}
A single node provides no fault tolerance. Do not use this topology in production.
{{< /alert >}}

## Three-node HA cluster

The minimum production configuration: the Raft quorum is two out of three nodes, and the cluster keeps running if one node fails.

Before deployment:

- Issue a certificate for each node with the node FQDN and IP address in the `subjectAltName` field (see [Installation](../installation/)).
- Place the node certificate and key and the `stronghold-ca.pem` CA certificate on each node.
- Open TCP ports `8200` and `8201` between the nodes.

Configuration of the first node (`raft-node-1`, `10.20.30.10`):

```hcl
ui            = true
cluster_addr  = "https://10.20.30.10:8201"
api_addr      = "https://10.20.30.10:8200"
disable_mlock = true

listener "tcp" {
  address         = "0.0.0.0:8200"
  tls_cert_file   = "/opt/stronghold/tls/node-1-cert.pem"
  tls_key_file    = "/opt/stronghold/tls/node-1-key.pem"
  tls_min_version = "tls12"
}

storage "raft" {
  path    = "/opt/stronghold/data"
  node_id = "raft-node-1"

  retry_join {
    leader_tls_servername   = "raft-node-1.demo.tld"
    leader_api_addr         = "https://10.20.30.10:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-2.demo.tld"
    leader_api_addr         = "https://10.20.30.11:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-3.demo.tld"
    leader_api_addr         = "https://10.20.30.12:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }
}
```

The `retry_join` blocks list all cluster nodes, so the same structure works for every node. On the other nodes, change only the values that refer to the node itself:

| Parameter | `raft-node-2` | `raft-node-3` |
| --- | --- | --- |
| `cluster_addr` | `https://10.20.30.11:8201` | `https://10.20.30.12:8201` |
| `api_addr` | `https://10.20.30.11:8200` | `https://10.20.30.12:8200` |
| `tls_cert_file`, `leader_client_cert_file` | `/opt/stronghold/tls/node-2-cert.pem` | `/opt/stronghold/tls/node-3-cert.pem` |
| `tls_key_file`, `leader_client_key_file` | `/opt/stronghold/tls/node-2-key.pem` | `/opt/stronghold/tls/node-3-key.pem` |
| `node_id` | `raft-node-2` | `raft-node-3` |

Start the service on all nodes, initialize the cluster on one node (`stronghold operator init`), and unseal each node (`stronghold operator unseal`). Check the cluster membership:

```shell
stronghold operator raft list-peers
```

## Five-node HA cluster

The Raft quorum is three out of five nodes, and the cluster keeps running if two nodes fail. Use this topology if the nodes are spread across several availability zones or if you need to service nodes without reducing fault tolerance.

Configuration of the first node (`raft-node-1`, `10.20.30.10`):

```hcl
ui            = true
cluster_addr  = "https://10.20.30.10:8201"
api_addr      = "https://10.20.30.10:8200"
disable_mlock = true

listener "tcp" {
  address         = "0.0.0.0:8200"
  tls_cert_file   = "/opt/stronghold/tls/node-1-cert.pem"
  tls_key_file    = "/opt/stronghold/tls/node-1-key.pem"
  tls_min_version = "tls12"
}

storage "raft" {
  path    = "/opt/stronghold/data"
  node_id = "raft-node-1"

  retry_join {
    leader_tls_servername   = "raft-node-1.demo.tld"
    leader_api_addr         = "https://10.20.30.10:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-2.demo.tld"
    leader_api_addr         = "https://10.20.30.11:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-3.demo.tld"
    leader_api_addr         = "https://10.20.30.12:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-4.demo.tld"
    leader_api_addr         = "https://10.20.30.13:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }

  retry_join {
    leader_tls_servername   = "raft-node-5.demo.tld"
    leader_api_addr         = "https://10.20.30.14:8200"
    leader_ca_cert_file     = "/opt/stronghold/tls/stronghold-ca.pem"
    leader_client_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
    leader_client_key_file  = "/opt/stronghold/tls/node-1-key.pem"
  }
}
```

On nodes `raft-node-2` to `raft-node-5` (addresses `10.20.30.11` to `10.20.30.14`), change `cluster_addr`, `api_addr`, `node_id`, and the paths to the node certificate and key in the same way as for the three-node cluster.

{{< alert level="info" >}}
Use an odd number of voting nodes. Join nodes that should only receive the replication stream and not participate in the quorum with the `retry_join_as_non_voter = true` parameter in the `storage "raft"` section.
{{< /alert >}}

## Auto unseal

Add one of the `seal` sections below to the configuration of every node. When you initialize a cluster with an HSM or KMS, Stronghold returns recovery keys instead of unseal keys (see [Seal and unseal](../../../concepts/seal/)). For initialization, set the `-recovery-shares` and `-recovery-threshold` parameters.

For HSM and Yandex Cloud KMS, recovery keys are set during initialization instead of unseal keys:

```shell
stronghold operator init -recovery-shares=5 -recovery-threshold=3
```

### HSM (PKCS #11)

Available in Stronghold EE. Create the keys in the HSM beforehand as described in [HSM support](../../../admin/kms-hsm/hsm/).

```hcl
seal "pkcs11" {
  lib         = "/usr/lib/librtpkcs11ecp.so"
  token_label = "my_token"
  pin         = "<PIN>"
  key_label   = "vault-rsa-key"
}
```

### Yandex Cloud KMS

The parameters and authentication methods are described in [Yandex Cloud KMS](../../../admin/kms-hsm/yandexcloudkms/).

```hcl
seal "yandexcloudkms" {
  kms_key_id               = "<KMS_KEY_ID>"
  service_account_key_file = "/etc/stronghold/yc-sa-key.json"
}
```

### inner-cluster

Available in Stronghold EE and Stronghold CSE. Unsealed cluster nodes unseal the other nodes, and no external KMS is required (see [Seal and unseal](../../../concepts/seal/)).

```hcl
seal "inner-cluster" {
  node {
    name        = "raft-node-1"
    address     = "https://10.20.30.10:8200"
    tls_ca_cert = "/opt/stronghold/tls/stronghold-ca.pem"
  }
  node {
    name        = "raft-node-2"
    address     = "https://10.20.30.11:8200"
    tls_ca_cert = "/opt/stronghold/tls/stronghold-ca.pem"
  }
  node {
    name        = "raft-node-3"
    address     = "https://10.20.30.12:8200"
    tls_ca_cert = "/opt/stronghold/tls/stronghold-ca.pem"
  }
}
```

{{< alert level="warning" >}}
Do not store the HSM PIN and service account keys in a configuration file that other system users can read. Restrict the permissions on `/opt/stronghold/config.hcl` to the `stronghold` user.
{{< /alert >}}

The PIN can be passed through the `VAULT_HSM_PIN` or `PKCS11_WRAPPER_PIN` environment variable instead of the `pin` parameter in the configuration, so the PIN does not end up in the file.

## DR pair

Available in Stronghold EE. A DR pair consists of two independent HA clusters, for example three nodes each at different sites. Deploy each cluster using the [three-node HA cluster](#three-node-ha-cluster) template with its own certificates and its own seal. Seal-wrapped data is not included in the replication stream, so each cluster seals data with its own seal.

Specifics:

- Both clusters run Stronghold EE with integrated Raft.
- The primary cluster port must be reachable from the secondary.
- Replication is enabled through the `sys/replication/dr/*` API after both clusters are initialized.

For how to enable replication and perform a failover, refer to [Disaster recovery](../../../admin/replication/disaster-recovery/).
