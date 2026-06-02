---
title: "MySQL"
weight: 30
---

MySQL is one of the supported plugins for the database secrets engine in Stronghold.
The plugin dynamically generates database credentials based on configured MySQL roles.
Static roles are also supported.

Several variants of this plugin are available in Stronghold.
Each variant is intended for different MySQL drivers.
The main difference between them is the allowed username length,
because different MySQL versions support different username lengths.

The following plugins are available:

- `mysql-database-plugin`
- `mysql-aurora-database-plugin`
- `mysql-rds-database-plugin`
- `mysql-legacy-database-plugin`

## Features

The following features are supported:

| Plugin name | Root credential rotation | Dynamic roles | Static roles | Username customization |
| --- | --- | --- | --- | --- |
| `mysql-database-plugin` | Yes | Yes | Yes | Yes |
| `mysql-aurora-database-plugin` | Yes | Yes | Yes | Yes |
| `mysql-rds-database-plugin` | Yes | Yes | Yes | Yes |
| `mysql-legacy-database-plugin` | Yes | Yes | Yes | Yes |

## Connection setup

To configure the MySQL plugin, complete the following steps:

1. Enable the database secrets engine if it is not already enabled.

   ```shell
   d8 stronghold secrets enable database
   ```

1. Configure Stronghold by specifying the required plugin and connection parameters.

   ```shell
   d8 stronghold write database/config/my-mysql-database \
     plugin_name=mysql-database-plugin \
     connection_url="{{username}}:{{password}}@tcp(127.0.0.1:3306)/" \
     allowed_roles="my-role" \
     username="strongholduser" \
     password="strongholdpass"
   ```

1. Create a Stronghold role that maps the role name to the SQL statement for creating an account in MySQL.

   ```shell
   d8 stronghold write database/roles/my-role \
     db_name=my-mysql-database \
     creation_statements="CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%';" \
     default_ttl="1h" \
     max_ttl="24h"
   ```

## Getting credentials

To create a new account, use the `database/creds/<role-name>` endpoint:

```shell
d8 stronghold read database/creds/my-role
```
