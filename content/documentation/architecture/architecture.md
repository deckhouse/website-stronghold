---
title: Deckhouse Stronghold Architecture
description: Architecture of Deckhouse Stronghold components
weight: 20
---

Deckhouse Stronghold is designed for secure secrets storage, centralized access management,
and lifecycle management of secrets in enterprise infrastructure.

The product provides tools for secrets management, cryptographic operations,
and integration with enterprise systems.
Deckhouse Stronghold enables centralized management of passwords, tokens, API keys,
certificates, and other secrets.

Key capabilities:

- Centralized secret management.
- Authentication and access management.
- Dynamic secrets and lifecycle management.
- Cryptographic operations.
- Ensuring Stronghold cluster operation.
- Performing [Performance and Disaster Recovery (DR)](../admin/replication/overview/) replication of Stronghold cluster state.
- Processing user API requests coming through the web interface and via [Deckhouse CLI](/products/kubernetes-platform/documentation/v1/cli/d8/).

For more details about key capabilities, refer to the [corresponding documentation section](../about/overview/).

Delivery options

- Linux package.
- [`stronghold`](/modules/stronghold/) module for Deckhouse Platform (DP).

## Architecture

{{< alert level="info" >}}
The following assumptions are made to simplify the diagram:

- The diagram shows containers of different pods interacting with each other directly. In fact, they interact through the corresponding Kubernetes services (internal load balancers). Service names are omitted where they are obvious from the context. Otherwise, the service name is shown above the arrow.
- Pods may run in multiple replicas, but the diagram shows all pods as a single replica.
{{< /alert >}}

The Deckhouse Stronghold architecture (for all delivery options) at C4 model level 1 is shown in the following diagram:
![C4 L1 diagram of Deckhouse Stronghold](../../../images/architecture/c4-l1-stronghold.svg)

The Deckhouse Stronghold architecture for the DP module delivery option at C4 model level 2 is shown in the following diagram:
![C4 L2 diagram of Deckhouse Stronghold components when installed as a DP module](../../../images/architecture/c4-l2-stronghold.svg)

## Components

### Delivery as a Linux OS package

Deckhouse Stronghold, delivered as a Linux package, is installed as a service consisting of a single component (one background process), implementing the full functionality of the Stronghold server both as a single instance in Standalone configuration and as a Stronghold cluster node in high-availability (HA) configuration.

### Delivery as a Deckhouse Platform module

Deckhouse Stronghold consists of the following components:

1. **Stronghold** (StatefulSet): Component implementing the core system functionality.  

    It consists of the following containers:

    - **plugin-fetcher**: Init container that downloads Stronghold plugins specified in the [`settings.plugins`](/modules/stronghold/configuration.html#parameters-plugins) module setting.
    - **stronghold**: Main container.
    - **kube-rbac-proxy**: Sidecar container with an authorization proxy based on Kubernetes RBAC for secure access to component metrics. It is an [Open Source project](https://github.com/brancz/kube-rbac-proxy).

1. **Stronghold-automatic** (Deployment): Optional component consisting of one **stronghold-automatic** container and providing initial secret store initialization. Stronghold-automatic also performs [automatic unsealing](../../concepts/seal//#auto-unseal-with-inner-cluster) of the secret store.  

    The component is created by the Deckhouse controller if the [`settings.management.mode`](/modules/stronghold/configuration.html#parameters-management-mode) module setting is set to `Automatic`.

## Interactions

Deckhouse Stronghold interacts with the following external components:

1. **External authentication providers**: Requests authentication.

1. **Secondary Stronghold cluster**: Performs Performance and DR replication.

1. **External systems using secrets**: Retrieves and updates secrets.

### Delivery as a Deckhouse Platform module

Deckhouse Stronghold delivered as a DP module also interacts with the following external components:

1. **Kube-apiserver**:

    - Authenticates for the Kubernetes auth method.
    - Manages Pod and Secret resources.

1. **Dex**: Requests authentication in cluster OIDC.

The following external components interact with Deckhouse Stronghold delivered as a DP module:

1. **Gateway/Ingress controller**: Forwards user request to the Deckhouse Stronghold web interface. Depends on the chosen method of publishing resources: using the Ingress controller of the [`ingress-nginx`](/modules/ingress-nginx/) module or the Gateway controller of the [`alb`](/modules/alb/) module.

1. **Prometheus-main**: Collects metrics from the stronghold component.
