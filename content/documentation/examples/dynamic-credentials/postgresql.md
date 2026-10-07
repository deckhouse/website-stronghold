---
title: "Dynamic PostgreSQL credentials for an application in DKP"
linkTitle: "Dynamic PostgreSQL credentials"
description: "Issuing short-lived PostgreSQL credentials to an application in Deckhouse Kubernetes Platform with the database secrets engine and the Kubernetes auth method."
weight: 10
params:
  relatedLinks:
    - title: "Database secrets engine"
      url: ../../../user/secrets-engines/databases/overview/
    - title: "PostgreSQL"
      url: ../../../user/secrets-engines/databases/postgresql/
    - title: "Kubernetes auth method"
      url: ../../../user/auth/kubernetes/
    - title: "Delivering secrets to Kubernetes pods"
      url: ../../delivery/kubernetes-workloads/
    - title: "Leases, renewal, and revocation"
      url: ../../../concepts/lease/
---

The application receives a unique PostgreSQL username and password with a limited lifetime from Stronghold. Stronghold creates the database user on request and removes it when the lease expires. The password is not stored in manifests or Secret objects.

## Goal

Configure issuing dynamic PostgreSQL credentials to the `myapp` application running in the `myapp` namespace of a Deckhouse Platform cluster, and deliver them to the pod with automatic renewal.

## Prerequisites

- Stronghold deployed in DP and a token with permissions to configure secrets engines, policies, and auth methods.
- A PostgreSQL server reachable from Stronghold pods and an account that can create roles (`CREATEROLE`), for example `stronghold`.
- Cluster access with `d8 k`.

In DP, the Kubernetes auth method for the current cluster is created automatically at the `kubernetes_local` path. If you use a different path, replace `kubernetes_local` in the commands below.

## Step 1. Configure the database secrets engine

1. Enable the secrets engine:

   ```bash
   d8 stronghold secrets enable database
   ```

1. Configure the PostgreSQL connection:

   ```bash
   d8 stronghold write database/config/myapp-postgres \
     plugin_name="postgresql-database-plugin" \
     allowed_roles="myapp" \
     connection_url="postgresql://{{username}}:{{password}}@postgres.db.svc:5432/myapp?sslmode=require" \
     username="stronghold" \
     password="<initial_password>" \
     password_authentication="scram-sha-256"
   ```

1. Rotate the service account password so that only Stronghold knows it:

   ```bash
   d8 stronghold write -force database/rotate-root/myapp-postgres
   ```

## Step 2. Create a database role

The role defines the SQL statements for creating a user and the credential lifetime:

```bash
d8 stronghold write database/roles/myapp \
  db_name="myapp-postgres" \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  revocation_statements="REVOKE ALL ON ALL TABLES IN SCHEMA public FROM \"{{name}}\"; DROP ROLE IF EXISTS \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"
```

- `default_ttl`: lease duration on issue and on each renewal.
- `max_ttl`: maximum credential lifetime; after it, the application needs new credentials.

## Step 3. Create a policy

```bash
d8 stronghold policy write myapp-db - <<'POLICY'
path "database/creds/myapp" {
  capabilities = ["read"]
}
POLICY
```

Renewing and revoking your own leases and token is allowed by the built-in `default` policy.

## Step 4. Create a Kubernetes auth role

1. Create a namespace and a service account:

   ```bash
   d8 k create namespace myapp
   d8 k -n myapp create serviceaccount myapp
   ```

1. Bind the service account to the policy:

   ```bash
   d8 stronghold write auth/kubernetes_local/role/myapp \
     bound_service_account_names=myapp \
     bound_service_account_namespaces=myapp \
     policies=myapp-db \
     ttl=1h
   ```

## Step 5. Deliver credentials to the pod

For dynamic credentials, use [Stronghold Agent](../../../user/agent/overview/) as a sidecar: it renews the lease and re-renders the file when credentials change. Methods that inject values only at pod start (env-injector, init container) fit only if the pod restarts more often than `max_ttl` expires.

Agent configuration for this scenario:

```hcl
stronghold {
  address = "https://stronghold.example.com"
}

auto_auth {
  method "kubernetes" {
    mount_path = "auth/kubernetes_local"
    config = {
      role = "myapp"
    }
  }
}

template {
  destination = "/secrets/db.env"
  contents    = <<-EOT
  {{ with secret "database/creds/myapp" }}
  DB_USER={{ .Data.username }}
  DB_PASSWORD={{ .Data.password }}
  {{ end }}
  EOT
}
```

A pod manifest with an Agent sidecar, an in-memory `emptyDir` volume, and a ConfigMap with the configuration is given in [Delivering secrets to Kubernetes pods](../../delivery/kubernetes-workloads/#stronghold-agent).

## Step 6. Handle credential changes in the application

Agent renews the lease every `default_ttl` until `max_ttl` is reached. After that, Agent requests new credentials and overwrites the `/secrets/db.env` file. The old user is removed when its lease is revoked.

The application must either:

- re-read the file on change and reconnect to the database;
- or exit on an authentication error so that Kubernetes restarts the container with new credentials.

Choose `max_ttl` so that credentials change no more often than the application can reconnect correctly.

## Verification

1. Request credentials manually with an administrator token:

   ```bash
   d8 stronghold read database/creds/myapp
   ```

   The response must contain the `lease_id`, `lease_duration`, `username`, and `password` fields.

1. Check that the user is created in PostgreSQL:

   ```sql
   SELECT rolname, rolvaliduntil FROM pg_roles WHERE rolname LIKE 'v-%';
   ```

1. Make sure the credentials file exists in the pod:

   ```bash
   d8 k -n myapp exec deploy/myapp -c app -- cat /secrets/db.env
   ```

1. List active leases for the role:

   ```bash
   d8 stronghold list sys/leases/lookup/database/creds/myapp/
   ```

## Cleanup

1. Delete test resources in the cluster: `d8 k delete namespace myapp`.
1. Revoke all credentials issued for the role:

   ```bash
   d8 stronghold lease revoke -prefix database/creds/myapp/
   ```

1. Delete the role, policy, and configuration:

   ```bash
   d8 stronghold delete auth/kubernetes_local/role/myapp
   d8 stronghold policy delete myapp-db
   d8 stronghold delete database/roles/myapp
   d8 stronghold delete database/config/myapp-postgres
   ```
