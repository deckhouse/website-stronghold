---
title: "HashiCorp Vault compatibility"
linkTitle: "Vault compatibility"
description: "Stronghold compatibility with the HashiCorp Vault API, CLI, and ecosystem tools, and the differences between Stronghold and upstream Vault."
weight: 40
---

Stronghold is built on HashiCorp Vault: the first version of Stronghold (v1.0, February 2024) is based on Vault v1.14.x (see the [release notes](../../release-notes/)). Stronghold therefore keeps the data model, HTTP API, configuration format, and CLI of upstream Vault and extends them with its own features.

The upstream code base is Vault OSS 1.14.8 (the last version under the MPL 2.0 license). Stronghold version numbers (1.19) do not match Vault version numbers: features of newer Vault versions are not carried over automatically.

Base Stronghold corresponds to Vault CE. Some features that upstream Vault offers only in Vault Enterprise are part of Stronghold EE (see [Editions](../editions/)).

## API compatibility

The Stronghold HTTP API is compatible with the Vault API:

- Requests use the same paths with the `/v1/` prefix (`/v1/sys/...`, `/v1/auth/...`, `/v1/<mount>/...`).
- The token is passed in the `X-Vault-Token` header.
- Request and response formats of secrets engines and auth methods match Vault.
- `Transit` ciphertext uses the `vault:v<key version>:...` format, so data encrypted in Vault keeps its familiar format.
- The `KV1/KV2` API is compatible with Vault to the extent that Stronghold can [replicate KV stores](../../admin/replication/kv-replication/) from Vault CE, Vault Enterprise, or any store with a compatible Vault KV/KV2 API.

Example API request:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  "${STRONGHOLD_ADDR}/v1/sys/health"
```

For a full description of the endpoints, refer to the [API reference](../../reference/api/).

## CLI

The Stronghold CLI mirrors the commands and flags of the `vault` CLI. Only the way you call it differs:

| Environment | Command | Example |
| --- | --- | --- |
| HashiCorp Vault | `vault <command>` | `vault kv get -mount=secret app` |
| Stronghold in DP (via Deckhouse CLI) | `d8 stronghold <command>` | `d8 stronghold kv get -mount=secret app` |
| Stronghold on Linux (standalone) | `stronghold <command>` | `stronghold kv get -mount=secret app` |

Subcommands, flags, and their behavior are described in [d8 stronghold](../../reference/console-tools/d8-stronghold/).

### Environment variables

Client environment variables in the Stronghold documentation use the `STRONGHOLD_` prefix:

| Stronghold | Vault equivalent | Purpose |
| --- | --- | --- |
| `STRONGHOLD_ADDR` | `VAULT_ADDR` | Server address |
| `STRONGHOLD_TOKEN` | `VAULT_TOKEN` | Client token |
| `STRONGHOLD_CACERT` | `VAULT_CACERT` | Path to the CA certificate used to verify the server TLS certificate |

The CLI also accepts variables with the `VAULT_` prefix (for example, `VAULT_ADDR`, `VAULT_TOKEN`). If both are set, `STRONGHOLD_*` takes precedence.

Environment variables that override server configuration parameters keep the `VAULT_` prefix, for example `VAULT_API_ADDR`, `VAULT_CLUSTER_ADDR`, `VAULT_RAFT_PATH`, `VAULT_RAFT_NODE_ID`, and `VAULT_ENABLE_FILE_PERMISSIONS_CHECK` (see [Configuration](../../install/standalone/configuration/)).

### Server and agent configuration

- The server configuration file uses the HCL or JSON format and the same sections as Vault: `listener`, `storage`, `seal`, `telemetry`, `ui`, and others (see [Configuration](../../install/standalone/configuration/)).
- [Stronghold Agent](../../user/agent/overview/) uses the `stronghold` section to connect to the server. To connect to HashiCorp Vault, use the `vault` section (see [Agent settings](../../user/agent/settings/)).

## Ecosystem tools

Because the API is compatible, you can point most Vault ecosystem tools at Stronghold by specifying its address. Compatibility status:

| Tool | Connection method | Status |
| --- | --- | --- |
| DP module [`secrets-store-integration`](/modules/secrets-store-integration/) (CSI driver, env-injector) | Native integration | Supported |
| [Stronghold Agent](../../user/agent/overview/) | Native integration | Supported |
| Ansible, Terraform | Vault providers and modules | Supported (see [Editions](../editions/)) <!-- TODO(verify): tested versions of the hashicorp/vault Terraform provider and the community.hashi_vault Ansible collection --> |
| [External Secrets Operator](../../examples/delivery/kubernetes-workloads/#external-secrets-operator) | `vault` provider | Not confirmed by tests <!-- TODO(verify): tested ESO version and external-secrets.io API version --> |
| cert-manager | `vault` Issuer | Not confirmed by tests <!-- TODO(verify): compatibility of the cert-manager Vault Issuer with the Stronghold PKI secrets engine --> |
| Jenkins (HashiCorp Vault Plugin), GitLab CI, GitHub Actions | Vault plugins and actions, JWT authentication | See [Integrations](../../examples/) <!-- TODO(verify): test status for each tool --> |
| Client SDKs (Go `github.com/hashicorp/vault/api`, Python `hvac`, and others) | Vault client libraries | Not confirmed by tests <!-- TODO(verify): list of tested SDKs and versions --> |

Methods of delivering secrets to Kubernetes are compared in [Delivering secrets to Kubernetes pods](../../examples/delivery/kubernetes-workloads/).

## Features not available in upstream Vault

| Feature | Edition | Description |
| --- | --- | --- |
| [KV1/KV2 replication](../../admin/replication/kv-replication/) | Stronghold EE | API-based pull replication of individual KV stores, including from Vault CE and Vault Enterprise |
| [GitOps secrets engine](../../user/secrets-engines/gitops/overview/) | Stronghold EE | Managing Stronghold configuration through a quorum of Git commit signatures |
| [Built-in `trdl` plugin](../../user/secrets-engines/trdl/) | Stronghold EE | Managing artifact builds and signing through a quorum of Git commit signatures |
| [`inner-cluster` seal](../../concepts/seal/) | Stronghold EE, Stronghold CSE | Built-in auto unseal: unsealed cluster nodes unseal other nodes without an external KMS |
| [`yandexcloudkms` seal](../../admin/kms-hsm/yandexcloudkms/) | Stronghold EE | Auto unseal and root key protection with Yandex Cloud KMS (standalone only) |
| [GOST cryptography](../../admin/cryptography/overview/) | Stronghold, Stronghold EE | `TLS 1.3` with `Magma` and `Kuznyechik` GOST encryption, `GOST 34.10` certificates in `PKI`, GOST algorithms in `Transit` |
| `CryptoPro seal wrapper` | Build with HSM support (Linux) | Seal wrapper for scenarios with Russian cryptography |
| [Managed Keys](../../user/managed-keys/overview/) with Yandex KMS | Stronghold EE | Working with key material in Yandex KMS and PKCS #11 devices from `Transit`, `PKI`, and `SSH` |
| [Namespace locks](../../admin/namespaces/overview/) in the web interface | Stronghold EE | Locking and unlocking the Namespace API |
| [`stronghold bootstrap` commands](../../install/standalone/bootstrap/) | Stronghold EE | Generating a Linux service installation script, a Helm chart, and a Docker image |
| Russian-language web interface, role and policy management in the web interface | Role management: Stronghold EE | Localized web interface with extended administration capabilities |
| Deployment as a DP module | All editions | Automatic initialization, unsealing, and Dex integration (see [Stronghold configuration](../../install/dkp/configuration/)) |

Vault Enterprise features available in Stronghold EE: [namespaces](../../admin/namespaces/overview/), [Performance and DR replication](../../admin/replication/overview/), [performance standby](../../admin/replication/performance-standby/), [automated snapshots](../../admin/backups/automated-snapshots/), [seal wrap](../../admin/kms-hsm/sealwrap/), [HSM (PKCS #11)](../../admin/kms-hsm/hsm/), Managed Keys, and [SAML authentication](../../user/auth/saml/). When you migrate from Vault Enterprise, the automated snapshot configuration is preserved.

## Differences and limitations

- Yandex Cloud KMS is the only external cloud KMS supported for `seal`. The `awskms`, `gcpckms`, `azurekeyvault`, `ocikms`, and `alicloudkms` configurations are not supported; the `transit` seal is supported.
- In the current version, HSM (`seal "pkcs11"`) and `seal "yandexcloudkms"` are supported only for the standalone installation.
- In DP, the `stronghold` module supports the `Automatic` (default) and `Manual` modes and the `Ingress` (default), `GatewayAPI`, `LoadBalancer`, `NodePort`, and `None` inlets; in the `Automatic` mode, the module performs initialization and unsealing (see [Stronghold configuration](../../install/dkp/configuration/)).
- In Stronghold, you can disable `seal wrap` for all data except the root key with the `disable_sealwrap = true` parameter.
- <!-- TODO(verify): Stronghold status of Vault Enterprise features not mentioned in the documentation: Sentinel, Control Groups, KMIP, Transform, Key Management secrets engine, upstream Login MFA. -->

## Migrating from HashiCorp Vault

For how to transfer data and configuration from HashiCorp Vault to Stronghold, refer to [Migrating from HashiCorp Vault](../../examples/operations/migration-from-vault/).

## Usage examples

Ready-made examples that use this feature:

- [Migrating from HashiCorp Vault](../../examples/operations/migration-from-vault/)

See all examples in [Usage examples](../../examples/).
