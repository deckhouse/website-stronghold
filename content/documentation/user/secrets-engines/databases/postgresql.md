---
title: "PostgreSQL"
description: "Information about PostgreSQL in Deckhouse Stronghold."
weight: 20
---

PostgreSQL is one of the supported plugins for the database secrets engine.
The plugin dynamically generates database credentials based on configured PostgreSQL roles.
Static roles are also supported.

## Features

The following features are supported:

| Plugin name | Root credential rotation | Dynamic roles | Static roles | Username customization |
| --- | --- | --- | --- | --- |
| `postgresql-database-plugin` | Yes | Yes | Yes | Yes |

## Connection setup

To configure the PostgreSQL plugin, complete the following steps:

1. Enable the database secrets engine if it is not already enabled.

   ```shell
   d8 stronghold secrets enable database
   ```

1. Configure the PostgreSQL connection.

   ```shell
   d8 stronghold write database/config/my-postgresql-database \
     plugin_name="postgresql-database-plugin" \
     allowed_roles="my-role" \
     connection_url="postgresql://{{username}}:{{password}}@localhost:5432/database-name" \
     username="strongholduser" \
     password="strongholdpass" \
     password_authentication="scram-sha-256"
   ```

1. Create a Stronghold role that maps the role name to SQL statements for creating an account in PostgreSQL.

   ```shell
   d8 stronghold write database/roles/my-role \
     db_name="my-postgresql-database" \
     creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
       GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
     default_ttl="1h" \
     max_ttl="24h"
   ```

## Getting credentials

To generate a new account, run the following command:

```shell
d8 stronghold read database/creds/my-role
```
