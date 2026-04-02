---
title: "Deckhouse Stronghold user guide"
linkTitle: "Introduction"
weight: 10
---

This section is intended for Deckhouse Stronghold users.

The user guide currently includes the following sections:

- [Project access](./access/) - how to get access to a project.
- [Core concepts](./concepts/auth/) - the main Stronghold concepts such as authentication, tokens, policies, identity, leases, and response wrapping.
- [Authentication methods](./auth/index/) - ways to authenticate users and services in Stronghold.
- [Managed Keys](./managed-keys/overview/) - how to use external cryptographic keys through `pkcs11` and `yandexcloudkms` with `ssh`, `pki`, and `transit`.
- [Secrets engines](./secrets-engines/index/) - how to work with KV, PKI, Transit, LDAP, Kubernetes, databases, and other secrets engines.
- [Stronghold Agent](./agent/overview/) - automatic authentication, token handling, and secret delivery without changing application code.
