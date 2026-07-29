---
title: "Installation"
description: "Quick single-node Stronghold setup with bootstrap service and manual HA cluster deployment."
weight: 10
---

You can install Stronghold in two ways:

- **quick installation** — a single-node server with `stronghold bootstrap service`, available from `v1.19`;
- **manual deployment** — a multi-node cluster in high availability (HA) mode with Raft storage.

The `stronghold bootstrap` command group also generates other deployment artifacts: a Helm chart archive and a Docker image tarball. For the full CLI interface, see [Bootstrap](../bootstrap/).

Prerequisites for both scenarios:

- a supported OS is installed on the server (Ubuntu, RedOS, Astra Linux);
- the Stronghold distribution is copied to the server, and the `stronghold` binary is available to run.

## Quick installation

The `stronghold bootstrap service` command generates a shell script that prepares a single-node Stronghold environment: directories, a configuration file, a systemd unit, a system user, binary installation, and optional self-signed TLS certificates.

{{< alert level="warning" >}}
Self-signed TLS certificates are suitable for testing only.
For production, provide your own certificates and use the `-no-tls` flag.
{{< /alert >}}

By default, the script is written to stdout so you can review it before applying:

```shell
stronghold bootstrap service | less
```

Apply the script in one of the following ways:

```shell
stronghold bootstrap service | sudo sh
```

or run the installation directly:

```shell
sudo stronghold bootstrap service -apply
```

The script creates the environment under `/opt/stronghold` (`-base-dir`), the systemd unit `/etc/systemd/system/stronghold.service`, and the `stronghold` system user, then enables and starts the service.

After the service starts, open `https://<node-name>:8200` in a browser and complete initialization through the web UI.
Alternatively, continue with manual initialization as described below.

Set the path to the CA certificate created by bootstrap:

```shell
export STRONGHOLD_CACERT=/opt/stronghold/tls/stronghold-ca.pem
```

Initialize Stronghold:

```shell
stronghold operator init
```

If necessary, you can specify the following parameters:

- `-key-shares` — number of key shares (default: 5);
- `-key-threshold` — minimum number of shares required to unseal the storage (default: 3).

{{< alert level="warning" >}}
After initialization, all key shares and the root token are displayed in the terminal.
Be sure to save them in a secure place.
Without the required number of key shares, access to Stronghold data will be impossible.
{{< /alert >}}

Unseal the storage. Run the command multiple times, entering the unseal keys:

```shell
stronghold operator unseal
```

If the `-key-threshold` parameter was not changed, you need to enter 3 key shares.

Check the status:

```shell
stronghold status
```

### Main bootstrap service options

Frequently used options:

| Option | Purpose |
| --- | --- |
| `-apply` | Execute the generated script instead of printing it to stdout |
| `-base-dir` | Base installation directory (default: `/opt/stronghold`) |
| `-force` | Overwrite existing files |
| `-no-tls` | Do not generate TLS certificates (use your own certificates) |
| `-tls-disable` | Disable TLS in the listener configuration (development only) |
| `-no-systemd` | Do not create or enable a systemd unit |
| `-listener-address` | TCP listener address in `host:port` form (default: `0.0.0.0:8200`) |
| `-api-addr` | API address advertised to clients |
| `-cluster-addr` | Cluster address for Raft |
| `-node-id` | Raft node identifier (defaults to hostname) |

For the full list of options and other bootstrap subcommands (`helm`, `docker`), see [Bootstrap](../bootstrap/) or run `stronghold bootstrap service -help`.

For further server configuration, see [Configuration](../configuration/).

## Deploying a cluster in HA mode

Stronghold supports multi-server mode to ensure high availability (HA). This mode is automatically enabled when using a compatible data storage backend and protects the system from failures by running multiple Stronghold servers.

To check if your storage supports high availability mode, start the server and make sure the message `HA available` is displayed next to the storage information. In this case, Stronghold will automatically use HA mode.

To provide high availability, one of the Stronghold nodes acquires a lock in the storage system and becomes active, while the other nodes switch to standby mode. If standby nodes receive requests, they either forward them or redirect clients according to the configuration and the current state of the cluster.

To run Stronghold in HA mode with the integrated Raft storage, you need at least three Stronghold servers. This requirement is necessary to achieve quorum — without it, the cluster cannot operate with the storage.

Additional prerequisites:

- individual certificates are issued for each node in the Raft cluster;
- a root certificate authority (CA) certificate is prepared.

The following scenario describes a manual deployment of a Stronghold cluster with three nodes: one active and two standby.

### Infrastructure preparation

#### Launch via systemd unit

{{< alert level="warning" >}}
All examples assume that a system user `stronghold` has been created and the service runs under this user.
If you need to use another user, replace `stronghold` with the appropriate name.
{{< /alert >}}

You can prepare the base environment on each node (directories, user, systemd unit, binary) with `stronghold bootstrap service` and the `-no-tls` flag, then place your own certificates and HA configuration.

Alternatively, create the systemd unit manually:

1. Create the file `/etc/systemd/system/stronghold.service` with the following content:

   ```ini
   [Unit]
   Description=Stronghold service
   Documentation=https://deckhouse.io/products/stronghold/
   After=network.target

   [Service]
   Type=simple
   ExecStart=/opt/stronghold/stronghold server -config=/opt/stronghold/config.hcl
   ExecReload=/bin/kill -HUP $MAINPID
   KillMode=process
   Restart=on-failure
   RestartSec=5
   User=stronghold
   Group=stronghold
   LimitNOFILE=65536
   CapabilityBoundingSet=CAP_IPC_LOCK
   AmbientCapabilities=CAP_IPC_LOCK
   SecureBits=noroot

   [Install]
   WantedBy=multi-user.target
   ```

1. Apply the systemd configuration changes:

   ```shell
   systemctl daemon-reload
   ```

1. Enable the service to start automatically:

   ```shell
   systemctl enable stronghold.service
   ```

1. Create the `/opt/stronghold/data` directory and set the appropriate permissions:

   ```shell
   mkdir -p /opt/stronghold/data
   chown stronghold:stronghold /opt/stronghold/data
   chmod 0700 /opt/stronghold/data
   ```

#### Preparing the required certificates

To configure TLS, you need a set of certificates and keys in the `/opt/stronghold/tls` directory:

- root certificate authority (CA) `stronghold-ca.pem` — the certificate used to sign Stronghold TLS certificates;
- Raft node certificates (three nodes in this scenario):
  - `node-1-cert.pem`;
  - `node-2-cert.pem`;
  - `node-3-cert.pem`;
- private keys of the node certificates:
  - `node-1-key.pem`;
  - `node-2-key.pem`;
  - `node-3-key.pem`.

In this example, a root certificate and a set of self-signed certificates for each node are created.

{{< alert level="warning" >}}
Self-signed certificates are suitable only for testing and experimentation.
For production use, it is strongly recommended to use certificates issued and signed by a trusted certificate authority (CA).
{{< /alert >}}

1. On the first node, create a directory for storing certificates (if it does not already exist) and switch to it:

   ```shell
   mkdir -p /opt/stronghold/tls
   cd /opt/stronghold/tls/
   ```

1. Generate a key for the root certificate:

   ```shell
   openssl genrsa 2048 > stronghold-ca-key.pem
   ```

1. Issue the root certificate:

   ```console
   openssl req -new -x509 -nodes -days 3650 -key stronghold-ca-key.pem -out stronghold-ca.pem

   Country Name (2 letter code) [XX]:RU
   Locality Name (eg, city) [Default City]:Moscow
   Organization Name (eg, company) [Default Company Ltd]:MyOrg
   Common Name (eg, your name or your server hostname) []:demo.tld
   ```

   > The certificate attributes are provided as an example.

1. To issue node certificates, create configuration files that contain the `subjectAltName` (SAN) parameter.
   For example, for the `raft-node-1` node:

   ```shell
   cat << EOF > node-1.cnf
   [v3_ca]
   subjectAltName = @alt_names
   [alt_names]
   DNS.1 = raft-node-1.demo.tld
   IP.1 = 10.20.30.10
   IP.2 = 127.0.0.1
   EOF
   ```

   Each node must have valid FQDN and IP addresses.
   The `subjectAltName` field in the certificate must contain values relevant to the specific node.
   Similarly, create a separate configuration file for each node.

1. Generate certificate signing requests (CSRs) and keys for the nodes:

   ```shell
   openssl req -newkey rsa:2048 -nodes -keyout node-1-key.pem -out node-1-csr.pem -subj "/CN=raft-node-1.demo.tld"
   openssl req -newkey rsa:2048 -nodes -keyout node-2-key.pem -out node-2-csr.pem -subj "/CN=raft-node-2.demo.tld"
   openssl req -newkey rsa:2048 -nodes -keyout node-3-key.pem -out node-3-csr.pem -subj "/CN=raft-node-3.demo.tld"
   ```

1. Issue certificates based on the created CSRs:

   ```shell
   openssl x509 -req -set_serial 01 -days 3650 -in node-1-csr.pem -out node-1-cert.pem -CA stronghold-ca.pem -CAkey stronghold-ca-key.pem -extensions v3_ca -extfile ./node-1.cnf
   openssl x509 -req -set_serial 02 -days 3650 -in node-2-csr.pem -out node-2-cert.pem -CA stronghold-ca.pem -CAkey stronghold-ca-key.pem -extensions v3_ca -extfile ./node-2.cnf
   openssl x509 -req -set_serial 03 -days 3650 -in node-3-csr.pem -out node-3-cert.pem -CA stronghold-ca.pem -CAkey stronghold-ca-key.pem -extensions v3_ca -extfile ./node-3.cnf
   ```

   > It is recommended to use unique `-set_serial` values for each certificate.

1. Copy the required files to each node:

   - node certificate;
   - node private key;
   - root certificate.

   For example, for the `raft-node-2` and `raft-node-3` nodes:

   ```shell
   scp ./node-2-key.pem ./node-2-cert.pem ./stronghold-ca.pem raft-node-2.demo.tld:/opt/stronghold/tls
   scp ./node-3-key.pem ./node-3-cert.pem ./stronghold-ca.pem raft-node-3.demo.tld:/opt/stronghold/tls
   ```

   > If the `/opt/stronghold/tls` directory does not exist on the target nodes, create it.

### Deploying a Raft cluster

1. Connect to the first server where the Stronghold cluster initialization will be performed.

1. Allow network connections for TCP ports `8200` and `8201`. Example for `firewalld`:

   ```shell
   firewall-cmd --add-port=8200/tcp --permanent
   firewall-cmd --add-port=8201/tcp --permanent
   firewall-cmd --reload
   ```

   > If necessary, you can use other ports by specifying them in the `/opt/stronghold/config.hcl` configuration file.

1. Create the `/opt/stronghold/config.hcl` configuration file for Raft:

   ```hcl
   ui = true
   cluster_addr  = "https://10.20.30.10:8201"
   api_addr      = "https://10.20.30.10:8200"
   disable_mlock = true

   listener "tcp" {
     address            = "0.0.0.0:8200"
     tls_cert_file      = "/opt/stronghold/tls/node-1-cert.pem"
     tls_key_file       = "/opt/stronghold/tls/node-1-key.pem"
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

1. Start the Stronghold service:

   ```shell
   systemctl start stronghold
   ```

1. Set the path to the CA certificate:

   ```shell
   export STRONGHOLD_CACERT=/opt/stronghold/tls/stronghold-ca.pem
   ```

1. Initialize the cluster:

   ```shell
   stronghold operator init
   ```

   If necessary, you can specify the following parameters:

   - `-key-shares` — number of key shares (default: 5);
   - `-key-threshold` — minimum number of shares required to unseal the storage (default: 3).

   {{< alert level="warning" >}}
   After initialization, all key shares and the root token are displayed in the terminal.
   Be sure to save them in a secure place.
   Without the required number of key shares, access to Stronghold data will be impossible.
   {{< /alert >}}

1. Unseal the cluster. Run the command multiple times, entering the unseal keys:

   ```shell
   stronghold operator unseal
   ```

   > If the `-key-threshold` parameter was not changed, you need to enter 3 key shares.

1. Configure the remaining nodes:

   - set the appropriate `cluster_addr` and `api_addr` values in `/opt/stronghold/config.hcl`, as well as the certificate paths for that node;
   - skip the initialization step;
   - immediately proceed to unsealing the cluster (`operator unseal`).

1. Verify the cluster status:

   ```console
   stronghold status
   Key                     Value
   ---                     -----
   Seal Type               shamir
   Initialized             true
   Sealed                  false
   Total Shares            5
   Threshold               3
   Version                 1.15.2
   Build Date              2025-03-07T16:10:46Z
   Storage Type            raft
   Cluster Name            stronghold-cluster-a3fcc270
   Cluster ID              f682968d-5e6c-9ad4-8303-5aecb259ca0b
   HA Enabled              true
   HA Cluster              https://10.20.30.10:8201
   HA Mode                 active
   Active Node Address     https://10.20.30.10:8200
   Raft Committed Index    40
   Raft Applied Index      40
   ```
