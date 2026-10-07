---
title: "Comparison with other solutions"
linkTitle: "Comparison"
description: "Comparison of Stronghold with HashiCorp Vault, OpenBao, and Yandex Lockbox by deployment model, API compatibility, secrets engines, replication, HSM support, and license model."
weight: 50
---

The table below helps you compare Stronghold with other secret management systems by verifiable characteristics. The comparison does not rate the products as a whole: the choice depends on your infrastructure requirements, regulatory constraints, and in-house expertise.

{{< alert level="info" >}}
Information about third-party products is based on their public documentation and may change with new releases. Before making a decision, check the vendor's current documentation.
{{< /alert >}}

<!-- TODO(verify): before publishing, check all cells in the HashiCorp Vault, OpenBao, and Yandex Lockbox columns against the vendors' current documentation and add the verification date. -->

## Comparison table

| Characteristic | Stronghold | HashiCorp Vault | OpenBao | Yandex Lockbox |
| --- | --- | --- | --- | --- |
| Delivery model | DP module; Linux executable (standalone): Stronghold EE | Executable, container, Helm chart; HCP Vault managed service | Executable, container, Helm chart | Managed service in Yandex Cloud |
| On-premise deployment, including air-gapped environments | Yes | Yes | Yes | No <!-- TODO(verify) --> |
| Vault HTTP API compatibility | Yes ([details](../vault-compatibility/)) | — | Yes <!-- TODO(verify): degree of compatibility in current OpenBao versions --> | No, own Yandex Cloud API |
| Static secrets (KV) | Yes, KV1 and KV2 with versioning | Yes | Yes | Yes |
| Dynamic secrets (databases, Kubernetes, LDAP, and others) | Yes | Yes | Yes | No <!-- TODO(verify) --> |
| Certificate issuance (PKI) | Yes, including GOST 34.10 (GOST: except Stronghold CSE) | Yes | Yes | No, certificates are issued by the separate Yandex Certificate Manager service <!-- TODO(verify) --> |
| Encryption as a service (Transit) | Yes, including GOST algorithms (GOST: except Stronghold CSE) | Yes | Yes | No, cryptographic operations are provided by the separate Yandex KMS service <!-- TODO(verify) --> |
| Namespaces | Stronghold EE | Vault Enterprise | Yes <!-- TODO(verify): OpenBao version that introduced namespaces --> | Isolation by Yandex Cloud folders and clouds <!-- TODO(verify) --> |
| High availability within a cluster | Yes, integrated Raft; performance standby: Stronghold EE | Yes; performance standby: Vault Enterprise | Yes <!-- TODO(verify): support for reads on standby nodes --> | Provided by the cloud provider |
| Cross-cluster replication | Stronghold EE: Performance, DR, KV1/KV2 | Vault Enterprise: Performance, DR | No <!-- TODO(verify) --> | Provided by the cloud provider |
| Auto unseal and root key protection in an HSM (PKCS #11) | Stronghold EE (standalone) | Vault Enterprise | <!-- TODO(verify): availability of the pkcs11 seal in OpenBao --> | Not applicable <!-- TODO(verify): support for Yandex KMS HSM keys for encrypting Lockbox secrets --> |
| Auto unseal via a cloud KMS | Yandex Cloud KMS | AWS KMS, Azure Key Vault, GCP Cloud KMS, and others <!-- TODO(verify) --> | AWS KMS, Azure Key Vault, GCP Cloud KMS, and others <!-- TODO(verify) --> | Not applicable |
| GOST cryptography | Yes ([details](../../admin/cryptography/overview/)) | No <!-- TODO(verify) --> | No <!-- TODO(verify) --> | <!-- TODO(verify) --> |
| FSTEC of Russia certificate | Stronghold CSE (certificate No. 5038) | No <!-- TODO(verify) --> | No <!-- TODO(verify) --> | <!-- TODO(verify): Yandex Cloud certificates that cover Lockbox --> |
| Entry in the Unified Register of Russian Software | Yes, No. 22339 | No <!-- TODO(verify) --> | No <!-- TODO(verify) --> | <!-- TODO(verify) --> |
| License model | Base Stronghold: free of charge; Stronghold EE and Stronghold CSE: commercial license <!-- TODO(verify): source code license of base Stronghold --> | Business Source License 1.1 (since version 1.15); Vault Enterprise: commercial license <!-- TODO(verify) --> | Mozilla Public License 2.0 <!-- TODO(verify) --> | Pay-as-you-go cloud service <!-- TODO(verify) --> |

## How to use the comparison

- If you need deployment in an air-gapped environment and compatibility with existing Vault tools, compare Stronghold, HashiCorp Vault, and OpenBao against the features on your requirements list.
- If your infrastructure is already hosted in Yandex Cloud and you only need static secrets, evaluate the managed service. Dynamic secrets, PKI, and Transit in your own infrastructure require a solution with the Vault API.
- If requirements for certified information security tools or Russian software apply, see [Security compliance](../compliance/).

The differences between Stronghold and upstream Vault are described in detail in [HashiCorp Vault compatibility](../vault-compatibility/), and edition features are described in [Editions](../editions/).
