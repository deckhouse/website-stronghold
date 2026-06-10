---
title: "Deckhouse Stronghold documentation"
linkTitle: "Overview"
description: "Deckhouse Stronghold product documentation"
weight: 10
outputs:
  - HTML
  - search
params:
  no_list: true
cascade:
  params:
    simple_list: true
---

Welcome to the home page of the Deckhouse Stronghold documentation.

Deckhouse Stronghold provides secure storage and lifecycle management for confidential data and secrets.
The secret storage is implemented as a key-value store and is compatible with the HashiCorp Vault API.

The documentation includes the following sections:

- [Getting started](/products/stronghold/gs/) — step-by-step instructions for installing a standard Deckhouse Stronghold configuration.
- [About the product](./about/overview/) — information about product purpose, editions, and technical requirements.
- [Installation](./install/dkp/install/steps/prepare/) — Deckhouse Stronghold installation procedure.
- [Configuration](./install/dkp/platform-management/node-management/node-group/) — access configuration, key backup, certificate setup, and authentication setup.
- [Working with policies](./concepts/policy/) — managing access to secrets and operations with them.
- [Working with access tokens](./user/auth/token/) — user authentication methods and access token management.
- [Working with secrets](./user/secrets-engines/kv/overview/) — secret engines and ways to deliver secrets to applications.
- [Reference](/products/kubernetes-platform/documentation/v1/cli/d8/) — reference information about resources, modules, and their configuration.

If you have any questions, feel free to contact us via [our Telegram channel](https://t.me/deckhouse).
We will do our best to help you.

If you are an Enterprise Edition user, reach out via <a href="mailto:support@deckhouse.io">support@deckhouse.io</a>.
We will be glad to provide assistance.
