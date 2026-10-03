---
title: Delivery as Linux OS services
description: Architecture of Deckhouse Stronghold components delivered as Linux OS services
weight: 30
---

## Architecture

The architecture of the Deckhouse Stronghold components when installed as a Linux OS service, at level 2 of the C4 model, is shown in the following diagram:

![Architecture of the Deckhouse Stronghold components when installed as a Linux OS service](../../../images/architecture/c4-l2-stronghold-linux.svg)

## Components

Deckhouse Stronghold consists of the following components:

1. **Stronghold**: The main component, implementing the following functionality:

    - Manages the lifecycle of secrets.
    - Supports the operation of the Stronghold cluster.
    - Performs Performance and Disaster Recovery (DR) replication of the Stronghold cluster state.
    - Creates and updates credentials for PostgreSQL, MySQL, MSSQL, ClickHouse, and other database management systems.
    - Authorizes with external systems (Lightweight Directory Access Protocol (LDAP), OpenID Connect (OIDC)).
    - Handles API requests from both the user and the [Deckhouse CLI](../../cli/d8/) utility.

## Interactions

Deckhouse Stronghold interacts with the following external components:

1. **External authorization services**: Authorizes requests in external LDAP or OIDC services.

1. **External Stronghold instance**: Performs Performance and DR replication.

1. **Database management systems**: Manages credentials for connecting to DBMS.

1. **Additional secret engines**: Retrieves and updates secrets.
