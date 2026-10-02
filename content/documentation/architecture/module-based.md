---
title: Delivery as a Deckhouse Platform module
description: Architecture of Deckhouse Stronghold components delivered as a Deckhouse Platform module
weight: 20
---

## Architecture

{{< alert level="info" >}}
The following assumptions are made to simplify the diagram:

- The diagram shows containers of different pods interacting with each other directly. In fact, they interact through the corresponding Kubernetes services (internal load balancers). Service names are omitted where they are obvious from the context. Otherwise, the service name is shown above the arrow.
- Pods may run in multiple replicas, but the diagram shows all pods as a single replica.
{{< /alert >}}

The architecture of the Deckhouse Stronghold components when installed as a Deckhouse Platform (DP) module, at level 2 of the C4 model, is shown in the following diagram:

![Architecture of the Deckhouse Stronghold components when installed as a DP module](../../../images/architecture/c4-l2-stronghold.svg)

## Components

Deckhouse Stronghold consists of the following components:

1. **Stronghold** (StatefulSet): The main component, implementing the following functionality:

    - Manages the lifecycle of secrets.
    - Supports the operation of the Stronghold cluster.
    - Performs Performance and Disaster Recovery (DR) replication of the Stronghold cluster state.
    - Creates and updates credentials for PostgreSQL, MySQL, MSSQL, ClickHouse, and other database management systems.
    - Authorizes with external systems (Lightweight Directory Access Protocol (LDAP), OpenID Connect (OIDC)).
    - Handles API requests from both the user and the [Deckhouse CLI](../../cli/d8/) utility.

    It consists of the following containers:

    - **plugin-fetcher**: Init container that downloads the plugins for Stronghold specified in the [`settings.plugins`](/modules/stronghold/configuration.html#parameters-plugins) module parameter.
    - **stronghold**: Main container.
    - **kube-rbac-proxy**: Sidecar container with an authorizing proxy based on Kubernetes RBAC, used to provide secure access to the main container. It is an [open source project](https://github.com/brancz/kube-rbac-proxy).

1. **Stronghold-automatic** (Deployment): Optional component consisting of a single **stronghold-automatic** container that performs the initial initialization of the secret storage. This component also performs [automatic unsealing](../concepts/seal/#auto-unseal-with-inner-cluster) of the secret storage.

    The component is created by the Deckhouse controller if the [`settings.management.mode`](/modules/stronghold/configuration.html#parameters-management-mode) module parameter is set to `Automatic`.

## Interactions

Deckhouse Stronghold interacts with the following external components:

1. **Kube-apiserver**:

    - Authorizes requests.
    - Manages Pod and Secret resources.

1. **Dex**: Authorizes requests in the cluster-wide OIDC.

1. **External authorization services**: Authorizes requests in external LDAP or OIDC services.

1. **External Stronghold instance**: Performs Performance and DR replication.

1. **Database management systems**: Manages credentials for connecting to DBMS.

1. **Additional secret engines**: Retrieves and updates secrets.

The following external components interact with Deckhouse Stronghold:

1. **Gateway/Ingress controller**: Forwards user requests to the Deckhouse Stronghold web interface. Depends on the method chosen for publishing resources in DP: using the Ingress controller of the [`ingress-nginx`](/modules/ingress-nginx/) module, or using the Gateway controller of the [`alb`](/modules/alb/) module.

1. **Prometheus-main**: Collects metrics from the stronghold component.
