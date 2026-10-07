---
title: "Temporary Kubernetes access with ServiceAccount tokens"
linkTitle: "Temporary Kubernetes tokens"
description: "Issuing short-lived ServiceAccount tokens to CI jobs and engineers with the kubernetes secrets engine: roles with allowed namespaces, generated RBAC rules, getting a token, and verification with d8 k auth can-i."
weight: 50
params:
  relatedLinks:
    - title: "Kubernetes secrets engine"
      url: ../../../user/secrets-engines/kubernetes/
    - title: "CI/CD integration"
      url: ../../delivery/ci-cd/
    - title: "Leases, renewal, and revocation"
      url: ../../../concepts/lease/
    - title: "Secrets engines API"
      url: ../../../reference/api/secrets/
---

Instead of a permanent kubeconfig with broad permissions, a CI job or an engineer requests a ServiceAccount token from Stronghold for the duration of the work. Stronghold creates a service account, a role, and a role binding with the required permissions in the cluster, issues a token with a limited lifetime, and deletes the created objects when the lease expires.

![Kubernetes token issuance flow](../../../images/ex-kubernetes-tokens.en.png)

## Goal

Configure two roles of the Kubernetes secrets engine:

- `ci-deploy` for CI jobs: permissions to manage Deployments, Services, and ConfigMaps in the `myapp` namespace; the token lives for 15 minutes.
- `engineer-view` for engineers: permissions of the built-in `view` ClusterRole in the `myapp` and `myapp-stage` namespaces; the token lives for 1 hour, at most 8 hours.

## Prerequisites

- Stronghold deployed in DP and a token with permissions to configure secrets engines and policies.
- Cluster access with `d8 k` and permissions to create ClusterRoles and ClusterRoleBindings.
- The `myapp` and `myapp-stage` namespaces.
- A configured auth method for CI (for example, JWT as in [CI/CD integration](../../delivery/ci-cd/)) and for engineers (for example, OIDC).

## Step 1. Grant Stronghold permissions in the cluster

Stronghold creates RBAC objects on behalf of its service account. Because Kubernetes prevents privilege escalation, this account needs the `bind` and `escalate` verbs on roles.

1. Find the name of the Stronghold service account:

   ```bash
   d8 k -n d8-stronghold get pod stronghold-0 -o jsonpath='{.spec.serviceAccountName}'
   ```

   In DP, this is the `stronghold` service account in the `d8-stronghold` namespace. By default, the module grants it only `system:auth-delegator` and access to its own namespace (Secrets and Pods), so the permissions below must be added.

1. Create a ClusterRole and bind it to the Stronghold service account. Replace `stronghold` in `subjects` with the name from the previous step:

   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: stronghold-k8s-secrets-engine
   rules:
     - apiGroups: [""]
       resources: ["serviceaccounts", "serviceaccounts/token"]
       verbs: ["create", "update", "delete"]
     - apiGroups: ["rbac.authorization.k8s.io"]
       resources: ["rolebindings", "clusterrolebindings"]
       verbs: ["create", "update", "delete"]
     - apiGroups: ["rbac.authorization.k8s.io"]
       resources: ["roles", "clusterroles"]
       verbs: ["bind", "escalate", "create", "update", "delete"]
   ---
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRoleBinding
   metadata:
     name: stronghold-k8s-secrets-engine
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: ClusterRole
     name: stronghold-k8s-secrets-engine
   subjects:
     - kind: ServiceAccount
       name: stronghold
       namespace: d8-stronghold
   ```

   ```bash
   d8 k apply -f stronghold-k8s-secrets-engine.yaml
   ```

{{< alert level="warning" >}}
A Stronghold service account with these permissions is effectively a cluster administrator. Restrict access to the Kubernetes secrets engine configuration with Stronghold policies.
{{< /alert >}}

## Step 2. Enable the secrets engine

```bash
d8 stronghold secrets enable kubernetes
d8 stronghold write -f kubernetes/config
```

An empty configuration means that Stronghold connects to the API of the current cluster using the CA certificate and JWT of its pod. For another cluster, set `kubernetes_host`, `kubernetes_ca_cert`, and `service_account_jwt`.

## Step 3. Create a role for CI

A role with `generated_role_rules` creates the whole chain of objects on each request: a ServiceAccount, a Role, and a RoleBinding.

```bash
d8 stronghold write kubernetes/roles/ci-deploy \
  allowed_kubernetes_namespaces="myapp" \
  token_default_ttl="15m" \
  token_max_ttl="1h" \
  generated_role_rules='{"rules":[
    {"apiGroups":["apps"],"resources":["deployments"],"verbs":["get","list","watch","create","update","patch"]},
    {"apiGroups":[""],"resources":["services","configmaps"],"verbs":["get","list","create","update","patch"]}
  ]}'
```

- `allowed_kubernetes_namespaces`: namespaces where a token can be requested. Instead of a list, you can set the `allowed_kubernetes_namespace_selector` label selector; Stronghold then needs `get` on `namespaces`.
- `generated_role_rules`: Role rules in JSON or YAML.
- `token_default_ttl` and `token_max_ttl`: the default and maximum token lifetime.

## Step 4. Create a role for engineers

A role with `kubernetes_role_name` binds the created service account to an existing role, here the built-in `view` ClusterRole:

```bash
d8 stronghold write kubernetes/roles/engineer-view \
  allowed_kubernetes_namespaces="myapp,myapp-stage" \
  kubernetes_role_type="ClusterRole" \
  kubernetes_role_name="view" \
  token_default_ttl="1h" \
  token_max_ttl="8h"
```

With `kubernetes_role_type="ClusterRole"`, Stronghold creates a RoleBinding in the requested namespace by default. To grant cluster-wide permissions, pass `cluster_role_binding=true` when requesting credentials, and a ClusterRoleBinding is created instead. Allow this only for dedicated roles and policies.

To issue tokens for an existing service account, use `service_account_name`. It cannot be combined with `kubernetes_role_name` or `generated_role_rules`.

## Step 5. Create policies

The `creds` endpoint is called with `POST`, so the policy needs the `update` capability:

```bash
d8 stronghold policy write k8s-ci-deploy - <<'POLICY'
path "kubernetes/creds/ci-deploy" {
  capabilities = ["update"]
}
POLICY

d8 stronghold policy write k8s-engineer-view - <<'POLICY'
path "kubernetes/creds/engineer-view" {
  capabilities = ["update"]
}
POLICY
```

Assign `k8s-ci-deploy` to the CI auth method role, and `k8s-engineer-view` to the engineers' group in the OIDC auth method or in Identity.

## Step 6. Get a token

1. Request a token for CI:

   ```bash
   d8 stronghold write kubernetes/creds/ci-deploy kubernetes_namespace=myapp
   ```

   Example output:

   ```text
   Key                          Value
   ---                          -----
   lease_id                     kubernetes/creds/ci-deploy/cujRLYjKZUMQk6dkHBGGWm67
   lease_duration               15m
   lease_renewable              false
   service_account_name         v-token-ci-deplo-1653001548-5z6hrgsxnmzncxejztml4arz
   service_account_namespace    myapp
   service_account_token        eyJHbGci0iJSUzI1Ni...
   ```

   The lease is not renewable: when it expires, request a new token.

1. An engineer requests a token for the required time within `token_max_ttl`:

   ```bash
   d8 stronghold write kubernetes/creds/engineer-view \
     kubernetes_namespace=myapp-stage \
     ttl=2h
   ```

1. Create a temporary kubeconfig with the token:

   ```bash
   TOKEN="$(d8 stronghold write -field=service_account_token kubernetes/creds/engineer-view kubernetes_namespace=myapp-stage)"
   SERVER="$(d8 k config view --minify -o jsonpath='{.clusters[].cluster.server}')"
   d8 k config view --minify --raw -o jsonpath='{.clusters[].cluster.certificate-authority-data}' | base64 -d > ca.crt

   export KUBECONFIG="$PWD/tmp-kubeconfig"
   d8 k config set-cluster dkp --server="$SERVER" --certificate-authority=ca.crt --embed-certs=true
   d8 k config set-credentials stronghold-token --token="$TOKEN"
   d8 k config set-context tmp --cluster=dkp --user=stronghold-token --namespace=myapp-stage
   d8 k config use-context tmp
   ```

In a CI job, get a Stronghold token with the JWT auth method as described in [CI/CD integration](../../delivery/ci-cd/), then request a Kubernetes token and pass it to `d8 k` with `--token`:

```yaml
deploy:
  script:
    - export STRONGHOLD_TOKEN="$(d8 stronghold write -field=token auth/gitlab/login role=myproject-deploy jwt=$STRONGHOLD_ID_TOKEN)"
    - export K8S_TOKEN="$(d8 stronghold write -field=service_account_token kubernetes/creds/ci-deploy kubernetes_namespace=myapp)"
    - d8 k --server="$K8S_SERVER" --certificate-authority="$K8S_CA_FILE" --token="$K8S_TOKEN" -n myapp apply -f deploy/
```

## Verification

1. Request a CI token and check the allowed actions:

   ```bash
   TOKEN="$(d8 stronghold write -field=service_account_token kubernetes/creds/ci-deploy kubernetes_namespace=myapp)"
   d8 k --token="$TOKEN" auth can-i create deployments -n myapp
   d8 k --token="$TOKEN" auth can-i patch configmaps -n myapp
   ```

   Both commands must return `yes`.

1. Check that there are no extra permissions:

   ```bash
   d8 k --token="$TOKEN" auth can-i get secrets -n myapp
   d8 k --token="$TOKEN" auth can-i create deployments -n kube-system
   ```

   Both commands must return `no`.

   If the current kubeconfig uses another authentication method (for example, a client certificate), run the check with the temporary kubeconfig from step 6.

1. Make sure a token request for a namespace that is not allowed is rejected:

   ```bash
   d8 stronghold write kubernetes/creds/ci-deploy kubernetes_namespace=default
   ```

1. List the objects Stronghold created in the namespace:

   ```bash
   d8 k -n myapp get serviceaccounts,roles,rolebindings | grep v-token
   ```

1. Revoke the lease early and make sure the token no longer works:

   ```bash
   d8 stronghold lease revoke -prefix kubernetes/creds/ci-deploy/
   d8 k --token="$TOKEN" auth can-i create deployments -n myapp
   ```

   Once the service account is deleted, requests with its token fail with `Unauthorized`.

For roles with `service_account_name`, Stronghold does not create a service account, so revoking the lease does not invalidate the token before it expires. Use a short `token_default_ttl` for such roles.

## Cleanup

1. Revoke all issued tokens:

   ```bash
   d8 stronghold lease revoke -prefix kubernetes/creds/
   ```

1. Delete the roles, policies, and configuration:

   ```bash
   d8 stronghold delete kubernetes/roles/ci-deploy
   d8 stronghold delete kubernetes/roles/engineer-view
   d8 stronghold policy delete k8s-ci-deploy
   d8 stronghold policy delete k8s-engineer-view
   d8 stronghold secrets disable kubernetes
   ```

1. Remove the Stronghold permissions in the cluster and the temporary files:

   ```bash
   d8 k delete -f stronghold-k8s-secrets-engine.yaml
   rm -f tmp-kubeconfig ca.crt
   ```
