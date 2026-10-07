---
title: "Network ports"
description: "The network connections Stronghold uses: API, node-to-node traffic, cross-cluster replication, seal, auth methods, secrets engines, audit, and backups."
weight: 30
---

The tables below list the connections to allow in firewalls. Stronghold ports (`8200`, `8201`) are the defaults for a standalone installation; you can change them in the `listener`, `api_addr`, and `cluster_addr` parameters of the [configuration file](../../../install/standalone/configuration/). Ports of external systems depend on their settings; the tables show the standard values.

## Stronghold ports

| Source | Destination | Port/protocol | Purpose |
| --- | --- | --- | --- |
| Clients, CLI, web UI, Stronghold Agent | Stronghold nodes (listener) | `8200`/TCP, HTTPS | API and web UI. The address is set in `listener "tcp"` and advertised through `api_addr`. |
| Stronghold node | Other cluster nodes | `8201`/TCP, TLS | Cluster port (`cluster_addr`): Raft consensus and replication, request forwarding from standby nodes to the active node. TLS is always used. |
| New Stronghold node | Leader node API | `8200`/TCP, HTTPS | Joining the Raft cluster (`retry_join`, `leader_api_addr`). |
| Stronghold node | API of other nodes | `8200`/TCP, HTTPS | `seal "inner-cluster"` auto unseal: an unsealed node passes key shares to the nodes listed in the `node` blocks. |
| Secondary cluster (Stronghold EE) | Primary cluster port | `8201`/TCP, TLS | Native Performance and Disaster Recovery replication: the WAL stream from the primary to the secondary and write forwarding to the primary. |
| Secondary cluster (Stronghold EE) | Primary API | `8200`/TCP, HTTPS | Unwrapping the activation token when enabling the secondary (`primary_api_addr`). |
| KV replication consumer cluster | Source cluster API | `8200`/TCP, HTTPS | KV1/KV2 replication at the API level. No access to the cluster port is needed. |

<!-- TODO(verify): the port and connection direction for native replication in Stronghold EE (in upstream Vault the secondary connects to the primary's cluster port, 8201 by default; whether the primary needs access to the secondary). -->

Monitoring and telemetry are configured in the `telemetry` section of the configuration file: for example, with `statsite_address = "127.0.0.1:8125"`, Stronghold sends metrics to that address. Metrics are also available through the API on port `8200`.

Prometheus metrics are served by the `/v1/sys/metrics?format=prometheus` endpoint on the API port, and `statsite_address` uses TCP.

## Ports in DP

| Source | Destination | Port/protocol | Purpose |
| --- | --- | --- | --- |
| Users, CLI, applications outside the cluster | DP Ingress controller | `443`/TCP, HTTPS | Web UI and API at `stronghold.<domain>` from `publicDomainTemplate`. |
| User's browser | Dex (`user-authn` module) | `443`/TCP, HTTPS | Web UI login through OIDC. |
| Ingress controller or Application Load Balancer (`Ingress` and `GatewayAPI` inlets only) | Stronghold pods | `8500`/TCP, HTTPS | HTTPS with a mandatory client certificate (`https-mtls`). By default, Ingress and ALB send traffic to the `https-mtls` port. |
| Clients, applications (through the `stronghold` service, inlets other than Ingress) | Stronghold pods | `8200`/TCP, HTTPS | External API through the Service. The certificate is taken from the `ingress-tls` secret. |
| Stronghold pod | Other Stronghold pods | `8201`/TCP, TLS | Cluster port. It is published externally only when `publishCluster.enabled` is set (Stronghold EE). |
| Application pods, Stronghold pods, the module's unsealer | Stronghold pods | `8300`/TCP, HTTPS | Internal API: `retry_join`, `seal "inner-cluster"`, probes, and unsealing. |
| Stronghold pod | Other Stronghold pods | `8301`/TCP, TLS | Internal cluster address: Raft and request forwarding between nodes. |
| Prometheus | Stronghold pods (`kube-rbac-proxy`) | `9889`/TCP, HTTPS | Metrics. |

For how the module is placed in the cluster, see [Deployment in DP](../deployment-dkp/).

## Seal and external key stores

| Source | Destination | Port/protocol | Purpose |
| --- | --- | --- | --- |
| Stronghold node | Yandex Cloud KMS | `443`/TCP, HTTPS | Encrypting and decrypting the root key for `seal "yandexcloudkms"`. With double encryption, KMS must be available at all times. The endpoint can be overridden with the `endpoint` parameter. |
| Stronghold node | Yandex Cloud VM metadata service | `80`/TCP, HTTP | Getting the VM service account token when `oauth_token` and `service_account_key_file` are not set. |
| Stronghold node | HSM | Local or vendor port | `seal "pkcs11"` works through a local PKCS#11 library. USB tokens (Rutoken ECP 3.0) and TPM2 need no network ports. For a network HSM, open the port listed in the vendor documentation. |

<!-- TODO(verify): the metadata service address and port (in Yandex Cloud, 169.254.169.254:80). -->

## Auth methods

| Source | Destination | Port/protocol | Purpose |
| --- | --- | --- | --- |
| Stronghold node | LDAP / Active Directory server | `389`/TCP (LDAP, StartTLS), `636`/TCP (LDAPS) | The [LDAP](../../../user/auth/ldap/) auth method and the [LDAP](../../../user/secrets-engines/ldap/) secrets engine. |
| Stronghold node | OIDC / JWT provider | `443`/TCP, HTTPS | The [OIDC](../../../user/auth/oidc/) and [JWT](../../../user/auth/jwt/) methods: fetching OIDC Discovery and JWKS, exchanging the code for a token. |
| Stronghold node | Kubernetes API | `6443`/TCP or `443`/TCP, HTTPS | The [Kubernetes](../../../user/auth/kubernetes/) method: validating service account tokens (TokenReview). The port is set in `kubernetes_host`. |
| User's browser | OIDC provider, SAML IdP | `443`/TCP, HTTPS | Redirecting the user to the provider login page. |

## Secrets engines

| Source | Destination | Port/protocol | Purpose |
| --- | --- | --- | --- |
| Stronghold node | PostgreSQL | `5432`/TCP | [PostgreSQL dynamic credentials](../../../user/secrets-engines/databases/postgresql/). |
| Stronghold node | MySQL / MariaDB | `3306`/TCP | [MySQL/MariaDB dynamic credentials](../../../user/secrets-engines/databases/mysql-maria/). |
| Stronghold node | ClickHouse | `9000`/TCP (native protocol) | [ClickHouse dynamic credentials](../../../user/secrets-engines/databases/clickhouse/). The address is set in `connection_url`. |
| Stronghold node | RabbitMQ management HTTP API | `15672`/TCP, HTTP(S) | The RabbitMQ secrets engine (`connection_uri`). |
| Stronghold node | Kubernetes API | `6443`/TCP or `443`/TCP, HTTPS | The [Kubernetes secrets engine](../../../user/secrets-engines/kubernetes/): creating service accounts and tokens. |

The ClickHouse plugin connects over the native protocol (TCP, port `9000`); HTTP port `8123` is not used.

The KV, PKI, Transit, SSH, and TOTP engines make no outbound connections: they work with data in Stronghold storage.

## Audit and backups

| Source | Destination | Port/protocol | Purpose |
| --- | --- | --- | --- |
| Stronghold node | Local syslog service | Local socket | The `syslog` audit device. Configure forwarding to a remote server in the syslog service itself. |
| Stronghold node | Log receiver | Port from the `address` parameter, TCP or UDP | The `socket` audit device (for example, `address=127.0.0.1:9090 socket_type=tcp`). A UNIX socket needs no network port. |
| Active Stronghold node | S3-compatible storage | `443`/TCP, HTTPS, or the port from `aws_s3_endpoint` | [Automated snapshots](../../backups/automated-snapshots/) (Stronghold EE) with the `aws-s3` storage type. |

The `file` audit device and snapshots with the `local` type write to the node's local disk. How audit devices are written to and what happens when they are unavailable is described in [Audit](../../audit/overview/#audit-device-failures).
