---
title: "ClickHouse"
description: "Information about ClickHouse in Deckhouse Stronghold."
hidden: true
weight: 40
---

ClickHouse is one of the supported plugins for the database secrets engine.
The plugin dynamically generates database credentials based on configured roles.
It also supports static roles.

## Features

The following features are supported:

| Plugin name | Root credential rotation | Dynamic roles | Static roles | Username customization |
| --- | --- | --- | --- | --- |
| `clickhouse-database-plugin` | Yes | Yes | Yes | Yes |

## Connection setup

To configure the ClickHouse plugin, complete the following steps:

1. Enable the database secrets engine if it is not already enabled.

   ```shell
   d8 stronghold secrets enable database
   ```

1. Configure the ClickHouse connection.

   ```shell
   d8 stronghold write database/config/my-clickhouse-database \
     plugin_name="clickhouse-database-plugin" \
     allowed_roles="my-role" \
     connection_url="clickhouse://clickhouse-server.my:9000??username={{username}}&password={{password}}&secure=true&skip_verify=true" \
     username="strongholduser" \
     password="strongholdpass"
   ```

1. Create a Stronghold role.
   In this example, the `readonly` role is assumed to have already been created in the `my_cluster` database cluster.

   ```shell
   d8 stronghold write database/roles/my-role \
     db_name="my-clickhouse-database" \
     creation_statements="CREATE USER '{{name}}' IDENTIFIED BY '{{password}}' ON CLUSTER 'my_cluster'; \
       GRANT readonly TO '{{name}}' ON CLUSTER 'my_cluster'; \
       SET DEFAULT ROLE readonly TO '{{name}}';" \
     default_ttl="1h" \
     max_ttl="24h"
   ```

## Getting credentials

To generate a new account, run the following command:

```shell
d8 stronghold read database/creds/my-role
```

Example output:

```console
Key                Value
---                -----
lease_id           database/creds/my-role/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
```
