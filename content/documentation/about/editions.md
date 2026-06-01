---
title: "Editions"
weight: 10
---

Deckhouse Stronghold is available as Community Edition (CE) and Enterprise Edition (EE).

Deckhouse Stronghold CE is available for use in any of the Deckhouse Kubernetes Platform editions.

Deckhouse Stronghold EE is licensed separately and available for use in any **commercial edition** of DKP.

The table below provides a brief comparison of the Deckhouse Stronghold editions, listing their main features and details:

| Features | Stronghold CE | Stronghold EE | Stronghold CSE |
| --- | --- | --- | --- |
| **Secret Engines** | | | |
| KV | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| PKI | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Database | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| SSH | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Transit | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Authentication** | | | |
| IP-based, client-based, etc. authentication restrictions | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| External authentication systems support (Identity Plugins) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| UserPass | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| AppRole | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Kubernetes | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| LDAP | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Dex | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| OIDC | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| JWT | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| MFA (external integration) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| WebAuth (FIDO2/Passkeys) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| SAML | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Secret Management** | | | |
| Secret lifecycle management (storage, creation, delivery, revocation, rotation) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Dynamic secrets support | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Create own secret storage inaccessible to other root tokens | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Secure secret delivery to applications as environment variables | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Secure secret delivery to applications as files | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Secure delivery of binary files as secrets (keytab, GPG keys, software licenses, etc.) to containers | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Storage Backend** | | | |
| Raft integrated storage | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Support for external storage (upon vendor agreement) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Auto Unseal Mechanisms** | | | |
| Built-in auto unseal via Raft | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Auto unseal via external hardware security module (TPM, HSM) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Auto unseal via external key management service (KMS) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Auto unseal via Transit Secrets Engine | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **High Availability** | | | |
| HA configuration out of the box (3-node cluster) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Replication filters support for secret transfer between HA nodes | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Scheduled backups without external schedulers and scripts | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Cross-cluster data replication (KV1/KV2) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Disaster recovery replication | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Performance** | | | |
| Performance replication | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Data Protection** | | | |
| Encryption-as-a-Service (on-the-fly encryption without storage) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Transparent encryption key rotation | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Signature verification during data decryption | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Create own secret storage inaccessible to other root tokens | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Key splitting for storage unsealing (Shamir's Secret Sharing) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Managed Keys (cryptographic keys located in external trusted system) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Root key encryption via external hardware security module (HSM) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Double data encryption via external hardware security module (HSM) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Russian cryptographic algorithms (GOST) support for Seal Wrap, TLS, PKI, Transit secret engine | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Access Control** | | | |
| Namespaces | In progress | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Configurable access control policies | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Identification via issued access tokens | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Access management via local access groups | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Local user account lifecycle management | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Administration** | | | |
| Web interface | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Command-line interface (CLI) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Centralized cluster management via common API (Cluster Management) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Centralized event collection | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Application role management (AppRole, OIDC/JWT Role) via web UI | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Monitoring** | | | |
| Centralized event collection | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Built-in system audit logging | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Distribution of audit events to different logs based on filters | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Audit event viewing via web UI and API | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **External Systems and Components Support** | | | |
| IaC automation tools (Ansible, Terraform) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Database management systems (PostgreSQL, MySQL, MongoDB, PostgresPro Enterprise, etc.) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Russian operating systems (Astra Linux, Red OS, ALT Linux, ROSA Server) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Russian identity providers (Identity Blitz) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Russian directory services (ALD Pro) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Russian backup systems (RuBackup) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Russian two-factor authentication systems (MULTIFACTOR, MFA SAS) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Russian hardware security modules (CryptoPro HSM, Rutoken ECP 3.0) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Key management systems (Yandex KMS) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| **Execution Environments** | | | |
| Deployment as a module in Deckhouse Kubernetes Platform Community Edition | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} |
| Deployment as a module in commercial editions of Deckhouse Kubernetes Platform | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Standalone binary execution on Linux OS (outside Deckhouse Kubernetes Platform) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| **Certification** | | | |
| FSTEC of Russia Order No. 118 dated July 4, 2022 compliance certificate (functional module within containerization platform) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| FSTEC of Russia Order No. 76 dated June 2, 2020 compliance certificate (Technical Specifications) | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} |
