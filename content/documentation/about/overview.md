---
title: "Deckhouse Stronghold purpose and features"
linkTitle: "Purpose and features"
description: "Purpose of Deckhouse Stronghold and its key features for centralized secrets management."
weight: 10
---

Deckhouse Stronghold is designed for secure secrets storage, centralized access management,
and lifecycle management of secrets in enterprise infrastructure.

The product provides tools for secrets management, cryptographic operations,
and integration with enterprise systems.
Deckhouse Stronghold enables centralized management of passwords, tokens, API keys,
certificates, and other secrets.

## Centralized secrets management

Deckhouse Stronghold enables centralized storage of secrets and access management instead of storing secrets
in application code, configuration files, and other unmanaged sources. Secrets can be managed through the API and CLI.

Key features:

- Storing static secrets in a versioned key-value store.
- Providing applications and users with controlled access to secrets.
- A unified interface for performing operations with secrets.

## Authentication and access management

Deckhouse Stronghold supports various authentication methods and allows you to configure access according
to your organization's requirements, including integration with existing IAM infrastructure.

Key features:

- User authentication through external identity providers.
- Application and service authentication.
- Access policy management.
- Issuing and using access tokens.

## Dynamic secrets and lifecycle management

Deckhouse Stronghold enables issuing secrets and credentials with a limited lifetime and managing their lifecycle.

Key features:

- Issuing dynamic credentials for external systems.
- Managing secret lifetimes.
- Automatic rotation and revocation of secrets.
- Automating the delivery of secrets to applications and infrastructure components.

## Cryptographic operations

Deckhouse Stronghold enables applications to perform cryptographic operations without providing them
with access to the underlying key material.

Key features:

- Data encryption and decryption.
- Creating and verifying digital signatures.
- Issuing and managing certificates.
- Integration with external KMS and HSM systems.

## Audit and operations

Deckhouse Stronghold provides auditing and administration capabilities for monitoring operations
and managing the state of the product.

Key features:

- Auditing operations with secrets and user actions.
- Backup and recovery.
- Health diagnostics and configuration management.
- Scaling and updates in supported deployment scenarios.
