---
title: "Deckhouse Stronghold user guide"
linkTitle: "Overview"
weight: 10
---

This section is intended for Deckhouse Stronghold users.

The user guide currently includes the following sections:

- [Project access](./get-started/access/) - how to get access to a project.
- [Usage examples](../examples/) - ready-made examples for typical tasks: sign-in via ALD Pro and AD, delivering secrets to Kubernetes and CI/CD, dynamic credentials, PKI, migration from Vault.
- [Core concepts](../concepts/auth/) - the main Stronghold concepts such as authentication, tokens, policies, identity, leases, and response wrapping.
- [Authentication methods](./auth/index/) - ways to authenticate users and services in Stronghold.
- [Managed Keys](./managed-keys/overview/) - how to use external cryptographic keys through `pkcs11` and `yandexcloudkms` with `ssh`, `pki`, and `transit`.
- [Secrets engines](./secrets-engines/index/) - how to work with KV, PKI, Transit, LDAP, Kubernetes, databases, and other secrets engines.
- [Stronghold Agent](./agent/overview/) - automatic authentication, token handling, and secret delivery without changing application code.
- [Web interface](./web-ui/) - working with Stronghold through the web interface.
