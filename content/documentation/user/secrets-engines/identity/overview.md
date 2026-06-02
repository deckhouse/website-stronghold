---
title: "Overview"
description: "Information about Overview in Deckhouse Stronghold."
weight: 10
---
## Identity secrets engine

The `Identity` secrets engine is used to manage identity in Deckhouse Stronghold.
It links authenticated clients to entities and allows you to use those entities
for access management.

In Deckhouse Stronghold, clients are represented as entities.
Each entity can have multiple aliases [1].

For example, a single user can sign in through GitHub and LDAP.
In this case, both accounts can be linked to one entity
that has two aliases: GitHub and LDAP.

After a successful authentication through a supported backend,
except for the Token backend, Deckhouse Stronghold creates a new entity
and adds an alias to it if no suitable entity already exists.

The entity ID is then associated with the authenticated token.
When such tokens are used, entity IDs are written to the audit log.
This makes it possible to track user actions.

The identity store allows you to manage entities in Deckhouse Stronghold.
You can create entities and their aliases and link them through the ACL API.
You can also assign policies to entities to extend the capabilities of tokens
associated with entity IDs.

These capabilities complement the existing token capabilities rather than replace them.
The capabilities inherited by a token from an entity are determined dynamically
at request time.
This allows you to manage access flexibly for already issued tokens.

{{< alert level="warning" >}}
The `Identity` secrets engine is installed by default.
It cannot be disabled or moved.
For a detailed description of the concept, see [Identity](../../../../concepts/identity/).
{{< /alert >}}

## What the engine supports

The `Identity` secrets engine supports the following features:

- [OIDC identity tokens](../oidc-token/);
- [OIDC identity provider](../oidc-provider/).
