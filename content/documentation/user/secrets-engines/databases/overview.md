---
title: "Overview"
weight: 10
---

The database secrets engine generates dynamic credentials based on configured roles.
It works with different databases through plugins.
This allows services to avoid storing credentials in plain text and instead request them from Stronghold and use the [lease mechanism](../../../concepts/lease/).

This approach simplifies auditing access to data.
Each service uses unique credentials, so suspicious activity can be associated with a specific service instance by its SQL username.

Stronghold uses an internal credential revocation mechanism.
This ensures that database users become invalid some time after the lease expires.

## How it works

The process usually consists of the following steps:

1. Enable the database secrets engine.
1. Configure the database connection.
1. Create a role with instructions for issuing credentials.
1. Request credentials by role name.

Configuration details depend on the specific database.
Examples and the list of parameters are available on the pages for individual plugins.

## Dynamic and static roles

When using dynamic roles, Stronghold generates a unique username and password pair for each credentials request.
Some plugins also support static roles.

A static role is a 1:1 mapping between a Stronghold role and a database user.
For such roles, Stronghold stores the password of the associated database user and automatically rotates it at configurable time intervals.
When a client requests credentials for a static role, Stronghold returns the current password of the database user associated with that role.
Any user with the corresponding Stronghold policies can access this database account.

{{< alert level="warning" >}}
Do not use the same database root credentials for static roles that are specified in `config/`.
Stronghold does not distinguish between regular and root credentials during password rotation.
If you assign root credentials to a static role, all dynamic and static users managed by this database configuration will stop working after the password is rotated.
If you need to rotate the root account, use the `rotate-root-credentials` API endpoint.
{{< /alert >}}

## Supported databases

All listed databases support dynamic roles, static roles, and root credential rotation.

| Database | Root user rotation | Dynamic roles | Static roles | Username customization | Credential type |
| --- | --- | --- | --- | --- | --- |
| [MySQL/MariaDB](./mysql-maria/) | Yes | Yes | Yes | Yes | password |
| [PostgreSQL](./postgresql/) | Yes | Yes | Yes | Yes | password |
| [ClickHouse](./clickhouse/) | Yes | Yes | Yes | Yes | password |

## Credential types

Database systems support different authentication methods and credential types.
The database secrets engine can manage credentials other than a username and password pair.
The `credential_type` and `credential_config` parameters for dynamic and static roles determine which credentials Stronghold generates and passes to database plugins.
Supported credential types and usage examples are described in the documentation for individual plugins.

## Password generation

Passwords are generated using a password policy.
Each database has a default password policy.
It defines a 20-character password that contains at least:

- one uppercase character;
- one lowercase character;
- one digit;
- one hyphen character.

The default password generation policy looks like this:

```hcl
length = 20
```
