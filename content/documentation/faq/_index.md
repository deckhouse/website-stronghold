---
title: "FAQ"
linkTitle: "FAQ"
description: "Frequently asked questions about Deckhouse Stronghold"
weight: 80
---

This page provides brief answers to frequently asked questions about Deckhouse Stronghold and its core features.

## What is Deckhouse Stronghold?

Deckhouse Stronghold is a secret storage system that manages access to secrets, issues temporary credentials, and performs cryptographic operations through various secret engines such as KV, Transit, PKI, SSH, LDAP, Kubernetes, and databases.

## How is KV v1 different from KV v2?

`KV v1` stores only the current value of a secret. Writing a new value overwrites the old one.

`KV v2` supports versioning, metadata, soft deletion, restoration of deleted versions, permanent destruction of versions, and partial updates via `patch`.

If you only need to store the current value and don’t require a change history, `KV v1` will suffice. If you need versions, metadata, and more flexible secret lifecycle management, use `KV v2`.

## Does Transit store my data?

No. The `Transit` secret engine performs cryptographic operations on data but does not store the submitted plaintext or the returned ciphertext. Data storage remains the responsibility of the application or the calling system.

## Can Stronghold be used to issue temporary credentials?

Yes. Deckhouse Stronghold supports issuing temporary credentials for various scenarios, for example:

- Kubernetes tokens for `ServiceAccount`;
- dynamic LDAP credentials;
- temporary credentials for PostgreSQL, MySQL, and ClickHouse;
- short-lived certificates via PKI.

## Can Stronghold be used as an OIDC identity provider?

Yes. Deckhouse Stronghold can act as an OIDC identity provider. To do so, you configure an authentication method, a client application, and use the OIDC discovery configuration, including `issuer`, `client_id`, and `client_secret`.

## Can Stronghold issue OIDC tokens with identity data?

Yes. Stronghold can issue signed JWTs that conform to the OIDC ID token structure. Such tokens are created based on a role associated with a signing key and can include additional claims via a template.

## Can secrets and tokens be automatically renewed in an application?

Yes, but the method depends on the scenario. For example:

- `Transit` supports key rotation and `rewrap` without exposing plaintext;
- the `Kubernetes` secret engine revokes tokens after their lease expires;
- `PKI` allows issuing certificates with a limited TTL;
- `LDAP` supports password rotation for static roles and account sets;
- `KV v2` supports versioning and managing deletion/restoration of versions.