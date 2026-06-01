---
title: "Deckhouse Stronghold User Guide"
linkTitle: "Introduction"
description: "Introduction to the Deckhouse Stronghold User Guide"
weight: 10
---

## About this guide

This section is intended for Deckhouse Stronghold users — developers, engineers, and other specialists who work with secrets, tokens, authentication methods, and cryptographic capabilities of the product through the command-line interface (CLI), API, or web interface.

The user guide helps you solve everyday tasks:

- access Stronghold.
- authenticate.
- work with tokens and policies.
- use secrets engines.
- apply Stronghold in application scenarios.
- integrate applications with Stronghold Agent and other product capabilities.

{{< alert level="info" >}}
If you need to install, upgrade, or administer Stronghold, use the [administrator guide](../admin/overview) and the [installation guide](../install/overview).
{{< /alert >}}

## Who this guide is for

This guide is intended for:

- developers who retrieve secrets from Stronghold in applications.
- users who work with Stronghold through the web interface.
- engineers who use the CLI and API.
- application operations teams that configure access to secrets and use authentication methods.

## Guide structure

### Getting started

This section helps you take your first steps in Stronghold:

- [Access to the project](../get-started/access) — getting access to Stronghold and preparing your working environment.
- [First login](../get-started/first-login) — signing in to Stronghold by using supported mechanisms.
- [First secret](../get-started/first-secret) — a basic scenario for writing and reading a secret.

### Authentication methods

The [Authentication methods](./auth/overview) section describes how users, applications, and services sign in to Stronghold, from OIDC and LDAP to AppRole, Kubernetes, SAML, and MFA.

### Managed keys

The [Managed keys](./managed-keys/overview) section describes how to use external key management systems and perform basic operations with managed keys.

### Secrets engines

The [Secrets engines](./secrets-engines/overview) section describes how to work with the main Stronghold secrets engines: `KV`, `PKI`, `Transit`, `SSH`, `TOTP`, `Identity`, `Kubernetes`, `LDAP`, and databases.

### Stronghold Agent

The [Stronghold Agent](./stronghold-agent/overview) section describes automatic authentication, secret delivery, and agent usage scenarios in applications and infrastructure.

## Important considerations

When using this guide, keep the following in mind:

- some capabilities depend on the Deckhouse Stronghold edition.
- some scenarios depend on how Stronghold is configured by the administrator.
- the availability of authentication methods, secrets engines, and policies depends on your installation configuration.
- in some scenarios, Stronghold is used as a standalone product, and in others, as part of Deckhouse Kubernetes Platform (DKP), while the core user concepts remain the same.

## Support

If you have questions about using Deckhouse Stronghold, you can ask for help in the [Deckhouse Telegram channel](https://t.me/deckhouse_ru).

{{< alert level="info" >}}
If you use the Enterprise edition, you can also contact technical support at [support@deckhouse.ru](mailto:support@deckhouse.ru).
{{< /alert >}}
