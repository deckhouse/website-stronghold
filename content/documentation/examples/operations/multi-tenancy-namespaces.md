---
title: "Multi-tenancy with namespaces"
linkTitle: "Namespaces for teams"
description: "Giving each team a separate namespace with delegated administration, its own auth methods, and templated policies."
weight: 60
params:
  relatedLinks:
    - title: "Namespaces"
      url: ../../../admin/namespaces/overview/
    - title: "Policies"
      url: ../../../concepts/policy/
    - title: "Identity"
      url: ../../../concepts/identity/
    - title: "Kubernetes auth method"
      url: ../../../user/auth/kubernetes/
---

{{< alert level="info" >}}
Namespaces are available only in Stronghold EE.
{{< /alert >}}

A namespace works as a separate virtual Stronghold with its own secrets engines, auth methods, policies, and tokens. A team gets a namespace, and its administrators manage it on their own without access to the root namespace or to other teams' namespaces.

## Goal

Create a namespace for the `team-a` team, grant its administrators delegated permissions, set up separate auth methods for people and applications, and configure templated policies that do not need changes when new applications are added.

## Prerequisites

- Stronghold EE.
- A root namespace administrator token.
- OIDC client parameters in the IdP for the team (address, `client_id`, `client_secret`) if people log in through OIDC.
- The team's Kubernetes cluster if applications authenticate with the Kubernetes method.

## Step 1. Create a namespace

```bash
d8 stronghold namespace create -custom-metadata=owner="team-a@example.com" team-a
```

All subsequent commands are run in this namespace. To avoid passing `-namespace=team-a` to each command, set an environment variable:

```bash
export STRONGHOLD_NAMESPACE=team-a
```

## Step 2. Create a delegated administrator policy

Paths in the policy are relative to the namespace. The policy allows managing secrets engines, auth methods, policies, and Identity inside `team-a`, but grants no permissions outside it:

```bash
d8 stronghold policy write -namespace=team-a team-admin - <<'POLICY'
# Managing secrets engines and auth methods.
path "sys/mounts/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "sys/mounts" {
  capabilities = ["read"]
}
path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}
path "sys/auth" {
  capabilities = ["read"]
}

# Managing policies.
path "sys/policies/acl/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Configuring auth methods and Identity.
path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "identity/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Working with team secrets.
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Nested namespaces.
path "sys/namespaces/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
POLICY
```

Remove blocks the team must not control, for example `sys/namespaces/*`.

## Step 3. Set up an auth method for people

1. Enable OIDC in the team namespace:

   ```bash
   d8 stronghold auth enable -namespace=team-a oidc
   ```

1. Configure the method and a role as described in [OIDC auth method](../../../user/auth/oidc/overview/), passing `-namespace=team-a` to each command. In the role, set the claim with user groups (`groups_claim`).

1. Bind the `team-admin` policy to the team's administrator group in the IdP:

   ```bash
   OIDC_ACCESSOR=$(d8 stronghold read -namespace=team-a -field=accessor sys/auth/oidc)

   GROUP_ID=$(d8 stronghold write -namespace=team-a -field=id identity/group \
     name=team-a-admins type=external policies=team-admin)

   d8 stronghold write -namespace=team-a identity/group-alias \
     name=team-a-admins \
     mount_accessor="$OIDC_ACCESSOR" \
     canonical_id="$GROUP_ID"
   ```

From now on, team administrators log in with `d8 stronghold login -namespace=team-a -method=oidc` and continue without the root namespace administrator.

## Step 4. Set up an auth method for applications

1. Enable the Kubernetes method for the team's cluster:

   ```bash
   d8 stronghold auth enable -namespace=team-a kubernetes
   d8 stronghold write -namespace=team-a auth/kubernetes/config \
     kubernetes_host="https://<api_server_address>:6443" \
     kubernetes_ca_cert=@ca.crt
   ```

   Configuration parameters are described in [Kubernetes auth method](../../../user/auth/kubernetes/).

1. Enable the team's KV store:

   ```bash
   d8 stronghold secrets enable -namespace=team-a -path=secret -version=2 kv
   ```

## Step 5. Configure a templated policy for applications

Instead of a separate policy for each application, use one templated policy: an application gets access to secrets in the directory that matches its Kubernetes namespace.

```bash
K8S_ACCESSOR=$(d8 stronghold read -namespace=team-a -field=accessor sys/auth/kubernetes)

d8 stronghold policy write -namespace=team-a app-by-namespace - <<POLICY
path "secret/data/{{identity.entity.aliases.${K8S_ACCESSOR}.metadata.service_account_namespace}}/*" {
  capabilities = ["read"]
}
POLICY

d8 stronghold write -namespace=team-a auth/kubernetes/role/apps \
  bound_service_account_names="*" \
  bound_service_account_namespaces="team-a-*" \
  policies=app-by-namespace \
  ttl=1h
```

`bound_service_account_namespaces` also accepts patterns with a trailing `*` (for example, `team-a-*`).

An application from the `team-a-billing` Kubernetes namespace can read `secret/data/team-a-billing/*`, but not the `team-a-orders` secrets.

In the same way, you can give each employee a personal directory:

```bash
d8 stronghold policy write -namespace=team-a personal - <<'POLICY'
path "secret/data/users/{{identity.entity.id}}/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "secret/metadata/users/{{identity.entity.id}}/*" {
  capabilities = ["read", "list", "delete"]
}
POLICY
```

## Verification

1. Log in as a team administrator and check that management inside the namespace is available:

   ```bash
   d8 stronghold login -namespace=team-a -method=oidc
   d8 stronghold secrets list -namespace=team-a
   ```

1. Check that access to the root namespace is denied; the command must return a `permission denied` error:

   ```bash
   d8 stronghold secrets list
   ```

1. Compare token capabilities on paths of different teams:

   ```bash
   d8 stronghold token capabilities -namespace=team-a secret/data/team-a-billing/db
   ```

## Cleanup

Deleting a namespace deletes all its secrets engines, auth methods, policies, and tokens:

```bash
d8 stronghold namespace delete team-a
```

Before deleting a production namespace, take a [storage snapshot](../../../admin/backups/save/).
