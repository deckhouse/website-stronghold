---
title: "Overview"
linkTitle: "Overview"
description: "Overview of Stronghold Agent in Deckhouse Stronghold"
weight: 10
---
Stronghold Agent is a client daemon that simplifies application integration with Stronghold. It provides automatic authentication, token management, and secret delivery without requiring changes to application code.

Stronghold Agent is especially useful for applications that do not have built-in support for secret management systems. This often applies to legacy applications, services running on virtual machines and on bare metal, as well as binary applications without an SDK for working with Stronghold.

Stronghold Agent works as an intermediary between an application and the Stronghold server:

- automatically authenticates.
- obtains and renews tokens.
- requests secrets.
- writes secrets to files or passes them through environment variables.
- updates secrets without changing application code.

![How Stronghold Agent works](/images/stronghold/agent.png)

The diagram shows a typical scenario: Stronghold Agent retrieves secrets from the Stronghold server over HTTPS, reads local configuration, and then passes secrets to applications through configuration files, template rendering, or environment variables.

## When to use Stronghold Agent

Use Stronghold Agent if you need to:

- connect an application without native integration to Stronghold.
- deliver secrets to configuration files, certificates, or environment variables.
- automatically renew secrets and tokens.
- run applications on virtual machines or on bare metal.
- simplify the integration of legacy applications with a secret management system.

**The primary use case for Stronghold Agent** is deployment on virtual machines and bare-metal servers. In such environments, there are usually no native mechanisms like those available in Kubernetes, so a separate component is needed to deliver secrets to the application.

## Stronghold Agent capabilities

Stronghold Agent addresses several practical tasks.

### Automatic authentication

The Agent can authenticate to Stronghold on its own and obtain a token for subsequent operations. This frees the application from having to handle authentication directly.

### Secret delivery

The Agent can deliver secrets in several ways:

- render secrets into files.
- pass secrets through environment variables.
- prepare configuration files for applications.
- update keys and certificates.

### Automatic updates

The Agent can track secret changes and update them without manual intervention. If the application can reload its configuration or correctly handle restarts, secrets can be updated without changing application code.

### Simplified integration

If an application cannot work with Stronghold directly, Stronghold Agent takes over this task. This is especially useful for legacy applications, specialized software, and services without an SDK.

### Additional capabilities

Depending on the scenario, Stronghold Agent can be used for:

- template rendering via `template`.
- passing secrets through `env_template` and `exec`.
- automatic authentication via Auto-Auth.
- operating through a local API Proxy.

These capabilities are described in detail on the child pages of this section.

## Common use cases

Stronghold Agent is suitable for the following scenarios:

- delivering secrets to legacy applications.
- working with configuration files, keys, and certificates.
- obtaining dynamic credentials for databases.
- renewing TLS certificates.
- running a self-hosted CI/CD runner.
- integrating applications that do not have an SDK for working with Stronghold.

For example, Stronghold Agent can be used to:

- write secrets to `application.properties`, `config.ini`, or other configuration files.
- inject environment variables before starting an application.
- reload a service after a configuration update.
- automatically renew temporary database credentials.
- renew TLS certificates before they expire.

## Getting started

If you are just getting started with Stronghold Agent, use the following order:

1. Open [Use cases](./use-cases) to understand whether Stronghold Agent fits your scenario.
1. Go to [Core capabilities](./capabilities) to choose an appropriate secret delivery method.
1. Then review [Basic settings](./settings) and prepare the configuration file.
1. After that, use [Launch and management](./launch-and-control) to validate the configuration and start the Agent.

## Practical recommendations

To get started with Stronghold Agent faster and without unnecessary application changes:

- first determine how the application receives configuration — from files or through environment variables.
- for virtual machines and bare metal, consider Stronghold Agent first as the primary integration method.
- if the application does not support an SDK, use templating or `env_template`.
- think in advance about how the application will react to secret updates: reload files, receive a signal, or restart.
- store the Agent configuration and sensitive files with the minimum required access permissions.