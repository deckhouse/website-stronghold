---
title: "Purpose and Key Features of Deckhouse Stronghold"
linkTitle: "Purpose and Features"
description: "The purpose of Deckhouse Stronghold and its key features for centralized secret management"
weight: 10
---

**Deckhouse Stronghold** is designed for secure secret storage, centralized access management, and lifecycle control of confidential data in enterprise infrastructure.

The product combines tools for secret management, cryptographic operations, and integration with enterprise systems. It helps organizations reduce risks associated with using passwords, tokens, API keys, certificates, and other sensitive data.

## Key Features

### Centralized Secret Management

Deckhouse Stronghold eliminates the need to store secrets in application code, configuration files, or external unmanaged sources. All sensitive data is stored centrally and accessible via a unified API and CLI.

The product supports:

- storing static secrets in a key-value store with versioning;
- securely providing secrets to applications and users;
- a unified interface for all operations with confidential data.

### Flexible Authentication and Access Control

The product supports various authentication methods and allows adapting the access model to organizational requirements. This simplifies integration with existing IAM infrastructure.

The following scenarios are available:

- user authentication via external identity providers;
- application and service authentication;
- role-based access policy management;
- issuing and using access tokens.

### Dynamic Secrets and Lifecycle Automation

Issuing secrets and credentials for a limited time reduces the risk of compromise. This helps implement the principle of least privilege.

Deckhouse Stronghold supports:

- issuing dynamic credentials for external systems;
- secret leasing with expiration control;
- automatic rotation, revocation, and deletion of secrets;
- automating secret management in application and infrastructure scenarios.

### Built-in Cryptographic Mechanisms

The product provides tools for performing cryptographic operations without exposing raw encryption keys to applications.

Deckhouse Stronghold can be used for:

- data encryption and decryption;
- digital signing and verification;
- issuing and managing certificates;
- integration with external KMS and HSM.

### Control, Audit, and Managed Operation

Auditing and administration tools allow monitoring user and service actions. This helps maintain security requirements and simplifies product maintenance.

Capabilities include:

- auditing all secret operations and user actions;
- backup and restore;
- state diagnostics and configuration management;
- scaling and updating in supported deployment scenarios.

## Use Cases

Deckhouse Stronghold is suitable for the following environments and scenarios:

- **production environments** — where controlled and secure secret management is required;
- **Kubernetes environments** — for delivering secrets to applications and services;
- **enterprise information systems** — for centralizing access to confidential data;
- **DevOps and CI/CD** — to avoid storing secrets in plain text;
- **integration with external systems** — KMS, HSM, enterprise identity providers.

The product can be used both in standalone deployments and as part of the **Deckhouse Kubernetes Platform**.

## Target Audience

Deckhouse Stronghold is aimed at the following roles:

- **administrators** — responsible for secure infrastructure operation;
- **DevOps engineers and platform teams** — implementing secure secret management practices;
- **developers** — needing a secure way to obtain secrets in applications;
- **information security professionals** — controlling access, audit, and manageability of confidential data operations.

## Compatibility with the HashiCorp Vault Ecosystem

The Deckhouse Stronghold API is compatible with the **HashiCorp Vault API**. This simplifies migrating existing scenarios, reusing client libraries, and integrating with tools already working within the HashiCorp Vault ecosystem.
