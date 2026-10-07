---
title: "Dynamic ClickHouse credentials for an application"
linkTitle: "Dynamic ClickHouse credentials"
description: "Issuing short-lived ClickHouse users with a predefined role to an application with the database secrets engine: connection, a role, a policy, renewal, revocation, and verification with clickhouse-client."
weight: 30
params:
  relatedLinks:
    - title: "Database secrets engine"
      url: ../../../user/secrets-engines/databases/overview/
    - title: "ClickHouse"
      url: ../../../user/secrets-engines/databases/clickhouse/
    - title: "Dynamic PostgreSQL credentials"
      url: ../postgresql/
    - title: "Delivering secrets to Kubernetes pods"
      url: ../../delivery/kubernetes-workloads/
    - title: "Leases, renewal, and revocation"
      url: ../../../concepts/lease/
---

An analytics service receives a dedicated ClickHouse user from Stronghold. The user's privileges come from a predefined ClickHouse role, so the SQL statements in Stronghold stay short and the privilege set is managed in one place, in ClickHouse.

## Goal

Configure issuing dynamic ClickHouse credentials to the `reports` service: the user gets the `analytics_reader` ClickHouse role with read access to the `analytics` database, lives for 1 hour with renewal up to 24 hours, and is dropped after the lease is revoked.

## Prerequisites

- Stronghold and a token with permissions to configure secrets engines and policies.
- A ClickHouse server with SQL-driven access control enabled, reachable from Stronghold over the native protocol.
- An `analytics` database.
- The `clickhouse-client` client and the `jq` utility on your workstation for verification.

## Step 1. Prepare a role and a service account in ClickHouse

Run the following in ClickHouse as an administrator:

```sql
CREATE ROLE analytics_reader;
GRANT SELECT ON analytics.* TO analytics_reader;

CREATE USER stronghold IDENTIFIED BY '<initial_password>';
GRANT CREATE USER, ALTER USER, DROP USER ON *.* TO stronghold;
GRANT analytics_reader TO stronghold WITH ADMIN OPTION;
```

`WITH ADMIN OPTION` allows the `stronghold` account to grant the `analytics_reader` role to the users it creates.

The `stronghold` account needs privileges to create, alter, and drop users (`CREATE USER`, `ALTER USER`, `DROP USER`): they are used when issuing credentials, during `rotate-root`, and when revoking a lease by default. To grant a role, the account also needs the right to assign it, for example `WITH ADMIN OPTION`, as in the example above.

In a ClickHouse cluster, add `ON CLUSTER '<cluster_name>'` to the statements, as in the example in [ClickHouse](../../../user/secrets-engines/databases/clickhouse/).

## Step 2. Configure the connection

1. Enable the secrets engine if it is not enabled yet:

   ```bash
   d8 stronghold secrets enable database
   ```

1. Configure the connection to ClickHouse:

   ```bash
   d8 stronghold write database/config/reports-clickhouse \
     plugin_name="clickhouse-database-plugin" \
     allowed_roles="reports" \
     connection_url="clickhouse://clickhouse.example.com:9440?username={{username}}&password={{password}}&secure=true" \
     username="stronghold" \
     password="<initial_password>"
   ```

   `secure=true` enables TLS. Port `9440` is the standard ClickHouse native protocol port with TLS.

   The `connection_url` format is `clickhouse://<host>:<port>?username=...&password=...`. The port is required; the `secure` and `skip_verify` parameters control TLS.

1. Rotate the service account password so that only Stronghold knows it:

   ```bash
   d8 stronghold write -force database/rotate-root/reports-clickhouse
   ```

## Step 3. Create a role

```bash
d8 stronghold write database/roles/reports \
  db_name="reports-clickhouse" \
  creation_statements="CREATE USER '{{name}}' IDENTIFIED BY '{{password}}'; \
    GRANT analytics_reader TO '{{name}}'; \
    SET DEFAULT ROLE analytics_reader TO '{{name}}';" \
  revocation_statements="DROP USER IF EXISTS '{{name}}';" \
  default_ttl="1h" \
  max_ttl="24h"
```

- `creation_statements`: create the user, grant the `analytics_reader` role, and make it the default role so that the privileges apply right after login.
- `revocation_statements`: drop the user when the lease is revoked.
- `default_ttl` and `max_ttl`: the lease duration on issue and renewal, and the maximum credential lifetime.

If `revocation_statements` are set in the role, the plugin runs them when a lease is revoked. If they are not set, the plugin runs `DROP USER IF EXISTS`.

## Step 4. Create a policy

```bash
d8 stronghold policy write reports-clickhouse - <<'POLICY'
path "database/creds/reports" {
  capabilities = ["read"]
}
POLICY
```

Assign the policy to the auth method role the service logs in with, for example a Kubernetes auth role as in [Dynamic PostgreSQL credentials](../postgresql/#step-4-create-a-kubernetes-auth-role).

## Step 5. Get credentials

1. Request credentials:

   ```bash
   d8 stronghold read database/creds/reports
   ```

   Example output:

   ```text
   Key                Value
   ---                -----
   lease_id           database/creds/reports/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   lease_duration     1h
   lease_renewable    true
   password           SsnoaA-8Tv4t34f41baD
   username           v-token-reports-x
   ```

1. Renew the lease before `lease_duration` expires:

   ```bash
   d8 stronghold lease renew database/creds/reports/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   ```

1. When the credentials are no longer needed, revoke the lease:

   ```bash
   d8 stronghold lease revoke database/creds/reports/2f6a614c-4aa2-7b19-24b9-ad944a8d4de6
   ```

To deliver the credentials to the service and renew them automatically, use Stronghold Agent with a `{{ with secret "database/creds/reports" }}` template. The Agent configuration is in [Delivering secrets to Kubernetes pods](../../delivery/kubernetes-workloads/#stronghold-agent).

## Verification

1. Request credentials and store them in variables:

   ```bash
   creds="$(d8 stronghold read -format=json database/creds/reports)"
   CH_USER="$(echo "$creds" | jq -r '.data.username')"
   CH_PASSWORD="$(echo "$creds" | jq -r '.data.password')"
   LEASE_ID="$(echo "$creds" | jq -r '.lease_id')"
   ```

1. Connect with `clickhouse-client` and check the current user and its grants:

   ```bash
   clickhouse-client --host clickhouse.example.com --port 9440 --secure \
     --user "$CH_USER" --password "$CH_PASSWORD" \
     --query "SELECT currentUser(); SHOW GRANTS;" --multiquery
   ```

   The output must contain the `analytics_reader` role.

1. Make sure writes are denied:

   ```bash
   clickhouse-client --host clickhouse.example.com --port 9440 --secure \
     --user "$CH_USER" --password "$CH_PASSWORD" \
     --query "CREATE TABLE analytics.t (x UInt8) ENGINE = Memory"
   ```

   The command must fail with `ACCESS_DENIED`.

1. Revoke the lease and check that the user is dropped:

   ```bash
   d8 stronghold lease revoke "$LEASE_ID"
   clickhouse-client --host clickhouse.example.com --port 9440 --secure \
     --user "$CH_USER" --password "$CH_PASSWORD" --query "SELECT 1"
   ```

   The connection must fail with an authentication error.

## Cleanup

1. Revoke all credentials issued for the role:

   ```bash
   d8 stronghold lease revoke -prefix database/creds/reports/
   ```

1. Delete the role, the policy, and the connection:

   ```bash
   d8 stronghold policy delete reports-clickhouse
   d8 stronghold delete database/roles/reports
   d8 stronghold delete database/config/reports-clickhouse
   ```

1. If needed, drop the service account and the role in ClickHouse: `DROP USER stronghold; DROP ROLE analytics_reader;`.
