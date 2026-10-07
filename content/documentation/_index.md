---
title: "Deckhouse Stronghold documentation"
linkTitle: "Overview"
weight: 10
outputs:
  - HTML
  - markdown
  - search
  - llms
  - corpus
  - print
params:
  no_list: true
cascade:
  params:
    simple_list: true
---

{{< downloads >}}

Welcome to the home page of the Deckhouse Stronghold documentation.

Deckhouse Stronghold provides secure storage and lifecycle management for sensitive data (secrets).
The secret storage uses a key-value format and is compatible with the HashiCorp Vault API.

To quickly find the information you need:

- Use search if you are looking for a specific Stronghold feature, parameter, or other object.
- Use the side menu to navigate through documentation sections.

## Where to start

Ready-made solutions for typical tasks are collected in [Usage examples](./examples/): sign-in via ALD Pro and Active Directory, secrets in Kubernetes, CI/CD, and GitOps, dynamic credentials, PKI, encryption, migration from Vault.

- **Developers** — [first secret](./user/get-started/first-secret/), [client libraries](./examples/delivery/app-clients/), [secrets in Kubernetes pods](./examples/delivery/kubernetes-workloads/).
- **DevOps engineers** — [CI/CD](./examples/delivery/ci-cd/), [GitOps](./examples/delivery/gitops/), [Terraform and Ansible](./examples/delivery/terraform-ansible/), [all examples](./examples/).
- **Administrators** — [installation](./install/), [architecture](./admin/architecture/overview/), [operations](./admin/operations/), [backups](./admin/backups/overview/).
- **Security specialists** — [threat model](./admin/architecture/threat-model/), [security compliance](./about/compliance/), [audit](./admin/audit/overview/), [cryptographic algorithms](./admin/cryptography/overview/).
- **Moving from HashiCorp Vault** — [compatibility](./about/vault-compatibility/) and [migration guide](./examples/operations/migration-from-vault/).

If you need help:

- Ask a question in the ["Deckhouse | EN-community"](https://t.me/deckhouse) Telegram channel.
- If you use the Enterprise Edition, contact support at [`support@deckhouse.io`](mailto:support@deckhouse.io).
