---
title: "Policy examples"
description: "A library of ready-to-use Stronghold policies: read-only KV, application access to its own secrets via templating, CI/CD, namespace admin, backup operator, auditor, PKI, Transit, and security admin."
weight: 42
---

This page collects ready-to-use policies for common scenarios. Policy syntax, the list of capabilities, and templating rules are described in [Policies](../policy/).

Before using a policy, replace the mount paths (`secret/`, `pki_int/`, `transit/`, and so on), role names, and key names with your own. To find out which capabilities a command needs, add the `-output-policy` flag to it.

Save a policy to a file and upload it to Stronghold:

```bash
d8 stronghold policy write kv-read-only kv-read-only.hcl
```

## Read-only KV access

Read access to the `secret/app/*` secrets in the KV version 2 secrets engine. In KV v2, data is located under `secret/data/`, and metadata and key lists under `secret/metadata/`.

```hcl
# Read secrets.
path "secret/data/app/*" {
  capabilities = ["read"]
}

# List keys and read metadata.
path "secret/metadata/app/*" {
  capabilities = ["read", "list"]
}
```

For KV version 1, use the path without `data/` and `metadata/`:

```hcl
path "kv/app/*" {
  capabilities = ["read", "list"]
}
```

## Application access to its own secrets

A single policy for all applications: each application can access only the secrets under a path with its own entity name. This works when the entity name matches the application name, for example with AppRole or Kubernetes login.

```hcl
# Full access to the application's own secrets.
path "secret/data/apps/{{identity.entity.name}}/*" {
  capabilities = ["create", "read", "update", "patch", "delete"]
}

path "secret/metadata/apps/{{identity.entity.name}}/*" {
  capabilities = ["read", "list", "delete"]
}

# Read-only access to shared secrets.
path "secret/data/shared/*" {
  capabilities = ["read"]
}
```

Instead of the entity name, you can use alias metadata, for example the Kubernetes namespace of the service account:

```hcl
path "secret/data/{{identity.entity.aliases.auth_kubernetes_xxxx.metadata.service_account_namespace}}/*" {
  capabilities = ["read"]
}
```

Replace `auth_kubernetes_xxxx` with the accessor of the Kubernetes auth method (`d8 stronghold auth list`).

## CI/CD policy

A policy for a deployment job: reading deployment secrets, getting dynamic database credentials, issuing tokens from a token store role, and managing its own token.

```hcl
# Deployment secrets.
path "secret/data/ci/deploy/*" {
  capabilities = ["read"]
}

# Dynamic database credentials.
path "database/creds/deploy" {
  capabilities = ["read"]
}

# Issue tokens for the deployed application from a token store role.
path "auth/token/create/app-runtime" {
  capabilities = ["update"]
}

# Manage its own token.
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/revoke-self" {
  capabilities = ["update"]
}
```

## Namespace admin

A policy for a namespace administrator in Stronghold EE. Create it inside the namespace: all paths in it are relative to that namespace.

```bash
d8 stronghold policy write -namespace=team-a namespace-admin namespace-admin.hcl
```

```hcl
# Manage child namespaces.
path "sys/namespaces/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Manage policies.
path "sys/policies/acl" {
  capabilities = ["list"]
}

path "sys/policies/acl/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Manage auth methods.
path "sys/auth" {
  capabilities = ["read"]
}

path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}

path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Manage secrets engines.
path "sys/mounts" {
  capabilities = ["read"]
}

path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Manage identity.
path "identity/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Work with secrets in the namespace.
path "secret/*" {
  capabilities = ["create", "read", "update", "patch", "delete", "list"]
}

# Check its own capabilities.
path "sys/capabilities-self" {
  capabilities = ["update"]
}
```

For more on namespaces, see [Namespaces](../../admin/namespaces/overview/).

## Backup operator

A policy for taking manual Raft snapshots and managing automated backups without access to secrets. Snapshot endpoints are described in [Stronghold backups](../../admin/backups/overview/).

```hcl
# Take a snapshot (d8 stronghold operator raft snapshot save).
path "sys/storage/raft/snapshot" {
  capabilities = ["read"]
}

# Configure automated snapshots.
path "sys/storage/raft/snapshot-auto/config" {
  capabilities = ["list"]
}

path "sys/storage/raft/snapshot-auto/config/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

# Automated snapshot status.
path "sys/storage/raft/snapshot-auto/status/*" {
  capabilities = ["read"]
}

# View the Raft cluster configuration.
path "sys/storage/raft/configuration" {
  capabilities = ["read"]
}
```

{{< alert level="warning" >}}
A snapshot contains all Stronghold data in encrypted form. Store snapshot files in a secure location. Restoring from a snapshot (`POST sys/storage/raft/snapshot`) is not included in this policy: grant `update` on that path to administrators only.
{{< /alert >}}

The `sys/storage/raft/snapshot-auto/config/*` path is protected by root rights, so configuring automatic snapshots requires the `sudo` capability. Creating a snapshot (`sys/storage/raft/snapshot`) and reading the status do not require `sudo`.

## Auditor

A policy for reviewing the security configuration without access to secrets: audit devices, policies, mounts, auth methods, and identity. The `sys/audit` path is root-protected, so it requires the `sudo` capability.

```hcl
# List audit devices.
path "sys/audit" {
  capabilities = ["read", "sudo"]
}

# Read policies.
path "sys/policies/acl" {
  capabilities = ["list"]
}

path "sys/policies/acl/*" {
  capabilities = ["read"]
}

# Read mounts and auth methods.
path "sys/mounts" {
  capabilities = ["read"]
}

path "sys/auth" {
  capabilities = ["read"]
}

# Read entities and groups.
path "identity/entity/id" {
  capabilities = ["list"]
}

path "identity/entity/id/*" {
  capabilities = ["read"]
}

path "identity/group/id" {
  capabilities = ["list"]
}

path "identity/group/id/*" {
  capabilities = ["read"]
}
```

This policy grants no access to secrets: in Stronghold, anything not explicitly allowed is denied.

## PKI certificate issuer

A policy for a service that issues certificates from the `web-server` role of an intermediate CA.

```hcl
# Issue a certificate with the key generated in Stronghold.
path "pki_int/issue/web-server" {
  capabilities = ["update"]
}

# Sign a CSR generated on the client side.
path "pki_int/sign/web-server" {
  capabilities = ["update"]
}

# Read the CA certificate and chain.
path "pki_int/cert/ca" {
  capabilities = ["read"]
}

path "pki_int/ca_chain" {
  capabilities = ["read"]
}
```

See [PKI secrets engine](../../user/secrets-engines/pki/).

## Transit encrypt-only

The application can encrypt data with the `app-key` key but cannot decrypt it.

```hcl
path "transit/encrypt/app-key" {
  capabilities = ["update"]
}

# Explicitly deny decryption.
path "transit/decrypt/app-key" {
  capabilities = ["deny"]
}
```

For a decrypt-only service, swap the capabilities. See [Transit secrets engine](../../user/secrets-engines/transit/).

## Token self-management

A minimal policy similar to the built-in `default` policy: the token can look up, renew, and revoke itself, check its capabilities, and use cubbyhole and response wrapping.

```hcl
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/revoke-self" {
  capabilities = ["update"]
}

path "sys/capabilities-self" {
  capabilities = ["update"]
}

path "sys/leases/renew" {
  capabilities = ["update"]
}

path "sys/leases/lookup" {
  capabilities = ["update"]
}

path "cubbyhole/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "sys/wrapping/wrap" {
  capabilities = ["update"]
}

path "sys/wrapping/lookup" {
  capabilities = ["update"]
}

path "sys/wrapping/unwrap" {
  capabilities = ["update"]
}
```

To view the current contents of the built-in `default` policy, run `d8 stronghold read sys/policy/default`.

## Security admin

A security admin manages policies, auth methods, audit, and identity but has no access to secret contents. The `deny` capability has the highest priority, so the denial applies even if another token policy allows access.

```hcl
# Manage policies.
path "sys/policies/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Manage auth methods.
path "sys/auth" {
  capabilities = ["read"]
}

path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}

path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Manage audit devices.
path "sys/audit" {
  capabilities = ["read", "sudo"]
}

path "sys/audit/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}

# Manage identity.
path "identity/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Read mounts.
path "sys/mounts" {
  capabilities = ["read"]
}

# Deny access to secret contents.
path "secret/data/*" {
  capabilities = ["deny"]
}

path "kv/*" {
  capabilities = ["deny"]
}

path "database/creds/*" {
  capabilities = ["deny"]
}
```

List every secrets mount used in your installation in the `deny` blocks.

{{< alert level="warning" >}}
An administrator who can change policies and auth methods is technically able to grant themselves access to secrets. Monitor such changes with the [audit log](../../admin/audit/overview/).
{{< /alert >}}
