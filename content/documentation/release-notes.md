---
title: "Version History"
weight: 90
---

## v1.18 April 2026

Version `1.18` focuses on the development of external key management systems, `PKI`, and new audit capabilities.

- Managed Keys for working with key material in external trusted systems without storing private keys inside Stronghold. Supported secret engines: `Transit`, `PKI`, and `SSH`.
- Added support for `Yandex KMS` managed keys and keys using the `PKCS#11` standard.
- User authentication via an external `SAML 2.0` Identity Provider using the `Web SSO` profile.
- Added management of `KV` mount replication parameters in the web interface.
- Added support for single-element RDN in distinguished names within `PKI` for compatibility with OpenSSL / Microsoft CA.
- Added record filtering and field exclusion mechanisms for audit devices.
- Added snapshot inspection via the `stronghold operator raft snapshot inspect` command.
- Added ability to manage the `max_ttl` parameter for ACME certificates.
- Improved compatibility with Vault Enterprise Auto-Snapshots — snapshot configuration is preserved when migrating from Vault Enterprise.
- CVEs fixed: GHSA-jqcq-xjh3-6g23, CVE-2026-33186, CVE-2026-33487, CVE-2025-15558

## v1.17 February 2026

Version `1.17` continued development of integrations, protection mechanisms, and enterprise operational scenarios.

- Added support for `WebAuthn` — passwordless authentication (FIDO2/Passkeys).
- Added support for external Stronghold plugins running in DKP.
- Added namespace locks and an interface for managing them.
- Added support for the `LDAP secrets engine` in the web interface.
- Added support for `Yandex KMS` as a seal backend.
- Expanded `Agent` operational scenarios.
- Added support for `raft` nodes in `non-voter` mode.
- Enhanced deployment scenarios on arbiter node groups and test cluster parameters.

## v1.16 August 2025

Version `1.16` was a major functional release, introducing namespaces, new web interface features, replication mechanisms, cryptographic enhancements, and a new product edition.

- Added support for `Namespaces` in `EE`.
- Implemented multi-factor authentication (`MFA`) with support for `TOTP` and `Multifactor`.
- Introduced the Deckhouse Stronghold `CE` (`Community Edition`), available for free installation.
- Added web interface management for `OIDC` roles, `AppRole`, and password policies.
- Added replication metrics.
- Added `seal wrap` — an additional encryption mechanism for the most sensitive internal data on top of the standard Stronghold cryptographic barrier.
- Added `CryptoPro seal wrapper` for scenarios using Russian cryptography.
- The web interface received a more complete Russian localization and a dark theme.
- Added `ClickHouse` support and a web interface for working with it.
- Added `TLS 1.3` support with `Magma` and `Kuznechik` GOST encryption.
- Added support for `GOST 34.10-2012 X.509` certificates.
- Version `1.16` received FSTEC of Russia certificate No. 5038 dated February 10, 2026, for Deckhouse Stronghold software.

## v1.15 February 2025

Version numbering changed, transitioning from v1.1 to v1.15. Version `1.15` focused on operational scenarios, backup, and the first significant interface improvements.

- Added scheduled `Raft snapshots` backup with storage in `S3` or file system and API management.
- Expanded `KV` replication capabilities.
- Improved the web interface.
- Added automatic unsealing via `HSM`/`PKCS#11`, including support for Rutoken ECP 3.0.

## v1.1 March 2024

- Automatic unsealing with key storage in Stronghold node memory
- Russian language interface
- Included in the Russian Software Registry, Registry entry No. 22339 dated April 24, 2024
- Added integration with the platform secret delivery module [secrets-store-integration](/modules/secrets-store-integration/)

## v1.0 February 2024 / Vault v1.14.x

- Deployment as a DKP module
- Integration with platform DEX authentication
