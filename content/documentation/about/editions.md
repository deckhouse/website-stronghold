---
title: "Editions"
description: "Feature comparison of the Stronghold, Stronghold EE, and Stronghold CSE editions and the conditions for using them in Deckhouse Kubernetes Platform."
weight: 20
---

Deckhouse Stronghold is available in three editions:

- **Stronghold** (base Stronghold, Community Edition): available for use in any edition of Deckhouse Platform (DP).
- **Stronghold EE** (Enterprise Edition): licensed separately and available for use in any **commercial edition** of DP.
- **Stronghold CSE** (Certified Security Edition): certified by FSTEC of Russia for environments with increased information security requirements, licensed separately, and available for use only in DKP CSE.

To switch between editions, refer to [Switching Stronghold to Stronghold EE](../../install/dkp/platform-management/switching-editions/ce-to-ee/) and [Switching Stronghold from EE to CSE](../../install/dkp/platform-management/switching-editions/ee-to-cse/).

## Feature comparison

| Feature | Stronghold | Stronghold EE | Stronghold CSE |
| --- | --- | --- | --- |
| Secure management of the secret lifecycle (storage, creation, delivery, revocation, and rotation) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Support of IaC automation tools (Ansible, Terraform) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Support of authentication methods | JWT, OIDC, Kubernetes, LDAP, Token, **WebAuthn** | JWT, OIDC, Kubernetes, LDAP, Token, **WebAuthn**, **[SAML](../../user/auth/saml/)** | JWT, OIDC, Kubernetes, LDAP, Token |
| Support of KV, Kubernetes, Database, SSH, and PKI secrets engines | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [GitOps secrets engine](../../user/secrets-engines/gitops/overview/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | To be confirmed <!-- TODO(verify): availability in Stronghold CSE --> |
| [Built-in `trdl` plugin](../../user/secrets-engines/trdl/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | To be confirmed <!-- TODO(verify): availability in Stronghold CSE --> |
| Support of Russian operating systems ([full list of supported OS](/products/kubernetes-platform/documentation/v1/supported_versions.html)) | RED OS, ALT Linux, Astra Linux Special Edition, **ROSA Server** | RED OS, ALT Linux, Astra Linux Special Edition, **ROSA Server** | RED OS, ALT Linux, Astra Linux Special Edition |
| Deploying to an air-gapped environment | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Web interface | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Role and access policy management through the web interface | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Namespaces](../../admin/namespaces/overview/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Built-in automatic unsealing (auto unseal) without external services or KMS](../../concepts/seal/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Automatic unsealing and root key protection with HSM (PKCS #11)](../../admin/kms-hsm/hsm/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | To be confirmed <!-- TODO(verify): availability in Stronghold CSE --> |
| [Double encryption (seal wrap)](../../admin/kms-hsm/sealwrap/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | To be confirmed <!-- TODO(verify): availability in Stronghold CSE --> |
| High availability (HA) configurations | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Performance standby](../../admin/replication/performance-standby/): HA standby nodes serve reads | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | To be confirmed <!-- TODO(verify): availability in Stronghold CSE --> |
| [Cross-cluster data replication](../../admin/replication/overview/) | {{< icon-edition type="not_supported" >}} | KV1/KV2, Performance, DR | KV1/KV2, Performance, DR |
| [Scheduled automatic backups](../../admin/backups/automated-snapshots/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Audit logging](../../admin/audit/overview/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| [Managed Keys](../../user/managed-keys/overview/) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| [GOST algorithm support for PKI/Transit](../../admin/cryptography/overview/) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Delivered as a [standalone](../../install/standalone/) executable file | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Certificate of conformity with FSTEC of Russia Order No. 76, trust level 4 | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} |
| Can be launched in DP CE | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} |

{{< alert level="info" >}}
In the current version, HSM and `seal "yandexcloudkms"` are supported only for the standalone installation of Stronghold.
{{< /alert >}}
