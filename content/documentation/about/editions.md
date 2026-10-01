---
title: "Editions"
weight: 20
---

Deckhouse Stronghold is available as Enterprise Edition (EE) and Certified Security Edition (CSE), which is certified by the FSTEC of Russia for environments with increased information security requirements.

Deckhouse Stronghold EE and Deckhouse Stronghold CSE are licensed separately. Deckhouse Stronghold EE is available for use in any **commercial edition** of DP. Deckhouse Stronghold CSE is available only in the DKP CSE edition.

The table below provides a brief comparison of the Deckhouse Stronghold editions and HashiCorp Vault, listing their main features and details:

| Feature | Vault | EE | CSE |
| --- | --- | --- | --- |
| Secure management of the secret lifecycle (storage, creation, delivery, revocation, and rotation) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Support of IaC automation tools (Ansible, Terraform) | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Support of authentication methods | JWT, OIDC, Kubernetes, LDAP, Token | JWT, OIDC, Kubernetes, LDAP, Token, **WebAuthn**, **SAML** | JWT, OIDC, Kubernetes, LDAP, Token |
| Support of KV, Kubernetes, Database, SSH, and PKI secret engines | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Support of Russian operating systems ([full list of supported OSes](/products/kubernetes-platform/documentation/v1/supported_versions.html)) | - | RED OS, ALT Linux, Astra Linux Special Edition, **ROSA Server** | RED OS, ALT Linux, Astra Linux Special Edition |
| Deploying to an air-gapped environment | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Web interface | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Role and access policy management through a web interface | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Support for namespaces | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Built-in automatic vault unsealing (auto unseal) without requiring any external services or KMS | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| HA configurations | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Cross-cluster data replication | {{< icon-edition type="not_supported" >}} | KV1/KV2, Performance, DR | KV1/KV2, Performance, DR |
| Automatic backup creation on a schedule | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Audit logging support | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Managed Keys | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| GOST algorithm support for PKI/Transit | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="not_supported" >}} |
| Delivered as a standalone executable file | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} | {{< icon-edition type="supported" >}} |
| Certificate of compliance with FSTEC of Russia Order No. 76, trust level 4 | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="not_supported" >}} | {{< icon-edition type="supported" >}} |
