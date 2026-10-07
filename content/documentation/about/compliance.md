---
title: "Security compliance"
description: "How Stronghold mechanisms map to typical information security requirements: authentication, access control, audit, cryptographic protection, backup, and availability."
weight: 60
---

This page maps typical information security requirements to the Stronghold mechanisms that help you meet them. Use it as a starting point when you prepare a threat model, a technical specification, or a document package for certification of your information system.

{{< alert level="warning" >}}
Compliance of an information system with regulatory requirements is determined by the system as a whole: its architecture, organizational measures, and the configuration of all components. The presence of a mechanism in Stronghold does not mean that a requirement is met automatically.
{{< /alert >}}

## Certification and registries

| Document | Details | Edition |
| --- | --- | --- |
| FSTEC of Russia certificate | Certificate No. 5038 dated February 10, 2026, for Deckhouse Stronghold software (version `1.16`). Conformity with FSTEC of Russia Order No. 76, trust level 4 | Stronghold CSE |
| Unified Register of Russian Software | Registry entry No. 22339 dated April 24, 2024 | <!-- TODO(verify): which editions the registry entry covers --> |

<!-- TODO(verify): validity period of FSTEC certificate No. 5038, the list of certified Stronghold CSE versions, and the documents (logbook/formular, operating manuals) included in the delivery set. -->

The certificate and registry details are also listed in the [release notes](../../release-notes/) and on the [Editions](../editions/) page.

## Mapping requirements to mechanisms

| Requirement group | Stronghold mechanism | Edition | Where to configure |
| --- | --- | --- | --- |
| **Identification and authentication** of users and services | LDAP, OIDC, JWT, Kubernetes, AppRole, userpass, token, and WebAuthn auth methods | All | [Auth methods](../../user/auth/overview/) |
| | Authentication via an external SAML 2.0 IdP | Stronghold EE | [SAML](../../user/auth/saml/) |
| | Multi-factor authentication (TOTP, Multifactor) | <!-- TODO(verify): editions --> | [MFA](../../user/auth/mfa/) |
| | Account lockout after failed login attempts (`user_lockout`) | All | [user_lockout section](../../install/standalone/configuration/#user_lockout) |
| | A single user entity across multiple login methods | All | [Identity](../../concepts/identity/) |
| **Access control** | Deny-by-default ACL policies, explicit `deny` with the highest priority | All | [Policies](../../concepts/policy/) |
| | Tokens with a limited lifetime and revocation | All | [Tokens](../../concepts/tokens/) |
| | Isolation of configuration and secrets between teams, Namespace API lock | Stronghold EE | [Namespaces](../../admin/namespaces/overview/) |
| | Granting administrator rights by IdP groups in DP | All | [Access management](../../install/dkp/configuration/#access-management) |
| **Security event logging** | Audit log of all API requests and responses (`file`, `syslog`, `socket` devices) | Stronghold EE | [Audit](../../admin/audit/overview/) |
| | Hashing of sensitive data in the log (HMAC-SHA256) | Stronghold EE | [Audit log record schema](../../admin/audit/log-format/) |
| | Record filtering and field exclusion | Stronghold EE | [Filtering](../../admin/audit/filtering/), [field exclusion](../../admin/audit/exclusion/) |
| **Cryptographic protection** of data at rest | `AES-256-GCM` cryptographic barrier | All | [Cryptographic algorithms](../../admin/cryptography/overview/) |
| | Root key protection in an HSM (PKCS #11) or Yandex Cloud KMS | HSM: Stronghold EE | [HSM](../../admin/kms-hsm/hsm/), [Yandex Cloud KMS](../../admin/kms-hsm/yandexcloudkms/) |
| | Additional encryption of critical data (seal wrap) | Stronghold EE | [Double encryption](../../admin/kms-hsm/sealwrap/) |
| | Unseal key splitting (Shamir), key rotation (rekey) | All | [Seal and unseal](../../concepts/seal/) |
| **Cryptographic protection** of data in transit | TLS `1.2`/`1.3`, client certificate verification (mTLS) | All | [Cryptographic algorithms](../../admin/cryptography/overview/), [listener section](../../install/standalone/configuration/#listener) |
| | `TLS 1.3` with GOST encryption, GOST algorithms in `PKI` and `Transit` | Stronghold, Stronghold EE | [GOST algorithms](../../admin/cryptography/overview/) |
| **Key and certificate management** | Certificate issuance, encryption and signing as a service, managed keys in external systems | `PKI`, `Transit`: all; Managed Keys: Stronghold EE | [PKI](../../user/secrets-engines/pki/), [Transit](../../user/secrets-engines/transit/), [Managed Keys](../../user/managed-keys/overview/) |
| **Secret lifecycle management** | Dynamic credentials with a limited lifetime, rotation, and revocation | All | [Leases](../../concepts/lease/), [database secrets engines](../../user/secrets-engines/databases/overview/) |
| | Complexity policies for generated passwords | All | [Password policies](../../concepts/password-policy/) |
| **Integrity** | Data integrity control by the `AES-GCM` barrier | All | [Cryptographic algorithms](../../admin/cryptography/overview/) |
| | SHA256 verification of external plugin binaries | All | [Plugins](../../admin/plugins/overview/) |
| | Configuration changes only through a quorum of Git commit signatures | Stronghold EE | [GitOps secrets engine](../../user/secrets-engines/gitops/overview/) |
| | Snapshot consistency check | All | [Inspecting a snapshot](../../admin/backups/inspect/) |
| **Backup and recovery** | Manual Raft storage snapshots | All | [Saving a snapshot](../../admin/backups/save/), [restoring](../../admin/backups/restore/) |
| | Scheduled snapshots to local or S3-compatible storage | Stronghold EE | [Automated snapshots](../../admin/backups/automated-snapshots/) |
| **Availability and fault tolerance** | HA cluster on integrated Raft | All | [Installation](../../install/standalone/installation/) |
| | Performance standby, cross-cluster Performance and DR replication | Stronghold EE | [Replication](../../admin/replication/overview/) |
| | Recovery after losing Raft quorum | All | [Standalone](../../install/standalone/raft-lost-quorum-recovery/), [DP](../../install/dkp/raft-lost-quorum-recovery/) |

## Russian regulatory requirements

The table lists regulations that are usually taken into account when deploying secret management systems in Russia. Applicability and specific measures are determined separately for your system.

| Regulation | How Stronghold helps meet the requirements |
| --- | --- |
| FSTEC of Russia Order No. 76 (information security requirements establishing trust levels) | Stronghold CSE is certified for trust level 4 (certificate No. 5038) |
| Federal Law No. 152-FZ "On Personal Data" and personal data protection requirements | <!-- TODO(verify): personal data protection levels for which Stronghold CSE is applicable and the corresponding measures from FSTEC of Russia Order No. 21 --> |
| Federal Law No. 187-FZ "On the Security of Critical Information Infrastructure" | <!-- TODO(verify): CII object significance categories and measures from FSTEC of Russia Order No. 239 for which Stronghold CSE is applicable --> |
| Requirements for state information systems (FSTEC of Russia Order No. 17) | <!-- TODO(verify): state information system protection classes for which Stronghold CSE is applicable --> |
| Use of Russian software | Entry No. 22339 in the Unified Register of Russian Software |
| GOST cryptographic protection | Stronghold and Stronghold EE support GOST algorithms in TLS, `PKI`, and `Transit`. Stronghold CSE does not support GOST algorithms for `PKI`/`Transit` (see [Editions](../editions/)). <!-- TODO(verify): whether Stronghold holds an FSB of Russia certificate as a cryptographic protection tool (KS1/KS2/KS3 classes); which crypto tools (for example, CryptoPro seal wrapper) are used in certified scenarios --> |

<!-- TODO(verify): list of protection measures (IAF, UPD, RSB, ZIS, OTsL, ODT, and others) covered by Stronghold CSE according to certification tests. -->

## Configuration recommendations

- Enable at least two audit devices and forward logs to a SIEM.
- Set the minimum TLS version to `tls12`.
- Store unseal and recovery keys separately with several responsible persons.
- Use short token lifetimes and dynamic secrets instead of static ones.
- Configure automated snapshots stored outside the protected cluster and test recovery regularly.
- Do not use the root token for day-to-day work: perform administration under accounts with assigned policies.
