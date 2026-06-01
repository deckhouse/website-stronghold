---
title: "Kubernetes"
linkTitle: "Kubernetes"
weight: 80
description: "Authenticate to Deckhouse Stronghold using a Kubernetes ServiceAccount token."
---

The `kubernetes auth` authentication method lets you authenticate to Deckhouse Stronghold using a Kubernetes `ServiceAccount` token.
This method simplifies using Deckhouse Stronghold in Kubernetes Pods.

A Kubernetes `ServiceAccount` token can also be used to log in through `JWT auth`.
For details, see [Using JWT auth instead of Kubernetes auth](#using-jwt-auth-instead-of-kubernetes-auth).

This method is suitable for scenarios where an application runs inside Kubernetes and must obtain Deckhouse Stronghold tokens based on its Kubernetes identity.
Conceptually, `Kubernetes auth` uses the `TokenReview API`, while `JWT auth` validates the token cryptographically using a public key.

## Authentication

### Via CLI

By default, the `/kubernetes_local` path is used.
If the authentication method is enabled at a different path, specify it with the `-path` parameter.

```shell
d8 stronghold write auth/kubernetes/login role=demo jwt=...
```

### Via API

By default, the `auth/kubernetes_local/login` endpoint is used.
If the authentication method is enabled at a different path, use that path instead of `kubernetes_local`.

```shell
curl \
  --request POST \
  --data '{"jwt": "<your service account jwt>", "role": "demo"}' \
  https://stronghold.example.com/v1/auth/kubernetes/login
```

The response contains the token in the `auth.client_token` field:

```json
{
  "auth": {
    "client_token": "38fe9691-e623-7238-f618-c94d4e7bc674",
    "accessor": "78e87a38-84ed-2692-538f-ca8b9f400ab3",
    "policies": ["default"],
    "metadata": {
      "role": "demo",
      "service_account_name": "myapp",
      "service_account_namespace": "default",
      "service_account_secret_name": "myapp-token-pd21c",
      "service_account_uid": "aa9aa8ff-98d0-11e7-9bb7-0800276d99bf"
    },
    "lease_duration": 2764800,
    "renewable": true
  }
}
```

## Configuration

Configure the authentication method in advance, before users or machines can authenticate.

These actions are typically performed by an operator or a configuration management tool.

In Deckhouse Stronghold, the Kubernetes authentication method is enabled by default at the `kubernetes_local` path.
It allows applications running in the same Kubernetes cluster as Deckhouse Stronghold to authenticate.

If necessary, you can add another Kubernetes cluster to Deckhouse Stronghold.

To configure the authentication method, follow these steps:

1. Enable the Kubernetes authentication method.

   ```shell
   d8 stronghold auth enable kubernetes
   ```

1. Use the `/config` endpoint to configure Deckhouse Stronghold to interact with the new Kubernetes cluster.
   To get the Kubernetes host address and TCP port, use the `d8 k cluster-info` command.

   ```shell
   d8 stronghold write auth/kubernetes/config \
     token_reviewer_jwt="<your reviewer service account JWT>" \
     kubernetes_host=https://192.168.99.100:<your TCP port or blank for 443> \
     kubernetes_ca_cert=@ca.crt
   ```

   {% alert level="warning" %}
   The template used by Deckhouse Stronghold to authenticate Pods depends on transmitting the JWT token over the network.
   Given the Deckhouse Stronghold security model, this is acceptable because Deckhouse Stronghold is part of the trusted computing base.

   In general, Kubernetes applications should not pass this JWT to other applications, because it allows calls to the Kubernetes API on behalf of the Pods.
   This can lead to unintended access being granted to third parties.
   {% endalert %}

1. Create a named role.

   ```shell
   d8 stronghold write auth/kubernetes/role/demo \
     bound_service_account_names=myapp \
     bound_service_account_namespaces=default \
     policies=default \
     ttl=1h
   ```

   This role allows login for the `myapp` `ServiceAccount` in the `default` namespace and assigns the `default` policy.

## Kubernetes 1.21

Starting with Kubernetes `1.21`, the `BoundServiceAccountTokenVolume` feature is enabled by default.
From this version, the JWT token mounted into containers by default has a limited lifetime and is bound to the lifetime of the Pod and the `ServiceAccount`.
The value of the JWT `iss` parameter also depends on the cluster configuration.

Changes to token lifetime are important when configuring the `token_reviewer_jwt` parameter.
If a short-lived token is used, Kubernetes revokes it as soon as the Pod or `ServiceAccount` is deleted, or when the token expires.
After that, Deckhouse Stronghold can no longer use the `TokenReview API`.

For this reason, `Kubernetes auth` does not validate the `iss` issuer by default.
The Kubernetes API performs the same validation when processing tokens, so repeating issuer validation on the Deckhouse Stronghold side is not required.

### Working with short-lived Kubernetes tokens

There are several ways to configure authentication for Kubernetes Pods if the default mounted Pod tokens are short-lived.
Each option has its own advantages and limitations.

| Option | All tokens are short-lived | Tokens can be revoked early | Notes |
| --- | --- | --- | --- |
| Use a local token as the `reviewer JWT` | Yes | Yes | Requires Deckhouse Stronghold to be deployed in a Kubernetes cluster |
| Use the client JWT as the `reviewer JWT` | Yes | Yes | Has overhead |
| Use a long-lived token as the `reviewer JWT` | No | Yes | — |
| Use `JWT auth` instead of `Kubernetes auth` | Yes | No | — |

{% alert level="info" %}
By default, Kubernetes extends the lifetime of `ServiceAccount` tokens to one year to simplify migration to short-lived tokens.
If you want to disable this behavior, set the `--service-account-extend-token-expiration=false` parameter for `kube-apiserver`, or define a custom `serviceAccountToken` volume configuration.
{% endalert %}

#### Using the Deckhouse Stronghold token as the reviewer JWT

If Deckhouse Stronghold runs in a Kubernetes Pod, it uses its own `ServiceAccount` to validate application tokens for applications running in the same Kubernetes cluster.

#### Using the client JWT as the reviewer JWT

When configuring `Kubernetes auth`, you can omit `token_reviewer_jwt`.
In this case, Deckhouse Stronghold uses the client JWT as its own token when calling the Kubernetes `TokenReview` API.

Also set `disable_local_ca_jwt=true`.

With this approach, Deckhouse Stronghold does not store the `reviewer JWT`.
This makes it possible to use short-lived tokens in all scenarios.
At the same time, overhead increases: you must maintain a `ClusterRoleBinding` for the set of `ServiceAccount` objects that are allowed to authenticate to Deckhouse Stronghold.

Each Deckhouse Stronghold client requires the `system:auth-delegator` role.

Example:

```shell
d8 k create clusterrolebinding myapp-client-auth-delegator \
  --clusterrole=system:auth-delegator \
  --group=group1 \
  --serviceaccount=default:svcaccount1
```

#### Using long-lived tokens

You can manually create a long-lived token and use it as `token_reviewer_jwt`.

In this example, the `myapp` service requires the `system:auth-delegator` role.

```shell
d8 k apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: myapp-k8s-auth-secret
  annotations:
    kubernetes.io/service-account.name: myapp
type: kubernetes.io/service-account-token
EOF
```

This approach simplifies configuration, but it does not provide the improved security benefits of short-lived tokens.

#### Using JWT auth instead of Kubernetes auth

`Kubernetes auth` uses the Kubernetes `TokenReview API`.
At the same time, JWT tokens generated by Kubernetes can also be validated through `JWT auth`, using Kubernetes as an OIDC provider.

This option makes it possible to use short-lived tokens for all clients and removes the need to use a `reviewer JWT`.

{% alert level="info" %}
Client tokens cannot be revoked before their TTL expires.
For this reason, using a short token lifetime is recommended.
{% endalert %}

## TokenReview API requirements

The `Kubernetes auth` method calls the Kubernetes `TokenReview API` to verify that the provided JWT is still valid.

`ServiceAccount` objects used in this authentication scheme must have access to the `TokenReview API`.

The following `ClusterRoleBinding` example grants these permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: role-tokenreview-binding
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
  - kind: ServiceAccount
    name: myapp-auth
    namespace: default
```
