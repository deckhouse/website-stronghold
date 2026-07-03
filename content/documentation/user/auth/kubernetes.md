---
title: "Kubernetes auth method"
linkTitle: "Kubernetes"
weight: 80
---

The Kubernetes auth method can be used to authenticate with Stronghold using a
Kubernetes Service Account Token. This method of authentication makes it easy to
introduce a Stronghold token into a Kubernetes Pod.

You can also use a Kubernetes Service Account Token to [log in via JWT auth][k8s-jwt-auth].
See the section ["How to work with short-lived Kubernetes tokens"](#how-to-work-with-short-lived-kubernetes-tokens)
for a summary of why you might want to use JWT auth instead and how it compares to
Kubernetes auth.

{{< alert level="info" >}}
If you are upgrading to Kubernetes v1.21+, ensure the configuration option
`disable_iss_validation` is set to true. Assuming the default mount path, you
can check with `d8 stronghold read -field disable_iss_validation auth/kubernetes/config`.
See ["Changes in JWT token behavior in Kubernetes 1.21+"](#changes-in-jwt-token-behavior-in-kubernetes-121) for more details.
{{< /alert >}}

## Authentication

### Via the CLI

The auth method name depends on how it was created. When Stronghold is deployed as part of Deckhouse Kubernetes Platform (DKP), the `kubernetes_local` auth method is created automatically and configured for the Kubernetes cluster where Stronghold is running. If the Kubernetes auth method is created manually, the default name is `kubernetes` unless a different path is specified.

If the auth method was created under a different name, specify it using the `-path` parameter in the CLI. For example:

```shell-session
d8 stronghold write -path=your-path auth/kubernetes/login role=demo jwt=...
```

### Via the API

Use the endpoint that matches the auth method name. When Stronghold is deployed as part of DKP, the automatically created auth method uses the `auth/kubernetes_local/login` endpoint. If the auth method was created under a different name, use the corresponding endpoint. The example below uses an auth method named `kubernetes`.

```shell-session
$ curl \
    --request POST \
    --data '{"jwt": "<your service account jwt>", "role": "demo"}' \
curl \
  --request POST \
  --data '{"jwt": "<your service account jwt>", "role": "demo"}' \
  http://127.0.0.1:8200/v1/auth/kubernetes/login
```

The response will contain a token at `auth.client_token`:

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

Before users or applications can authenticate, the Kubernetes auth method must be configured. This is typically done by an operator or configuration management tool.

To configure authentication for another Kubernetes cluster, enable an additional instance of the Kubernetes auth method:

1. Enable the Kubernetes auth method:

   ```bash
   d8 stronghold auth enable kubernetes
   ```

1. Use the `/config` endpoint to configure Stronghold to talk to Kubernetes. Use
  `d8 k cluster-info` to validate the Kubernetes host address and TCP port.

   ```bash
   d8 stronghold write auth/kubernetes/config \
   token_reviewer_jwt="<your reviewer service account JWT>" \
   kubernetes_host=https://192.168.99.100:<your TCP port or blank for 443> \
   kubernetes_ca_cert=@ca.crt
   ```

   {{< alert level="warning" >}}
   Stronghold uses the service account JWT token to authenticate with the Kubernetes API. Do not share this token with other applications or services, as they will be able to make requests to the Kubernetes API with the permissions of the corresponding service account.
   {{< /alert >}}

1. Create a named role:

   ```text
   d8 stronghold write auth/kubernetes/role/demo \
     bound_service_account_names=myapp \
     bound_service_account_namespaces=default \
     policies=default \
     ttl=1h
   ```

   This role authorizes the `myapp` service account in the `default`
   namespace and assigns it the default policy.

## Changes in JWT token behavior in Kubernetes 1.21+

Starting in version [Kubernetes 1.21](https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.21.md#api-change-2), the Kubernetes
feature gate `BoundServiceAccountTokenVolume` is enabled by default. As a result, the JWT tokens automatically mounted into containers have the following features:

* It has an expiry time and is bound to the lifetime of the pod and service account.
* They have an expiry time.
* They are bound to the lifetime of a pod and service account.
* The value of the JWT's `iss` claim depends on the cluster's configuration.

The changes to token lifetime are important when configuring the
`token_reviewer_jwt` option.
If a short-lived token is used,
Kubernetes will revoke it as soon as the pod or service account are deleted, or
if the expiry time passes, and Stronghold will no longer be able to use the
`TokenReview` API. For details on handling this change, see ["How to work with short-lived Kubernetes tokens"](#how-to-work-with-short-lived-kubernetes-tokens).

Kubernetes auth does not validate the `iss` field by default because Kubernetes already validates it when processing `TokenReview` requests.

### How to work with short-lived Kubernetes tokens

There are a few different ways to configure auth for Kubernetes pods when
default mounted pod tokens are short-lived, each with their own tradeoffs. This
table summarizes the options, each of which is explained in more detail below.

| Option                               | All tokens are short-lived | Can revoke tokens early | Other considerations                                      |
|--------------------------------------|----------------------------|-------------------------|-----------------------------------------------------------|
| Use local token as reviewer JWT      | Yes                        | Yes                     | Requires Stronghold to be deployed on the Kubernetes cluster |
| Use client JWT as reviewer JWT       | Yes                        | Yes                     | Operational overhead                                      |
| Use long-lived token as reviewer JWT | No                         | Yes                     | Simpler to configure                                                          |
| Use JWT auth instead                 | Yes                        | No                      | Client tokens cannot be revoked until they expire                                                          |

{{< alert level="info" >}}
By default, Kubernetes currently extends the lifetime of admission
injected service account tokens to a year to help smooth the transition to short-lived tokens. If you would like to disable this, set [`--service-account-extend-token-expiration=false`](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options) for `kube-apiserver` or specify your own `serviceAccountToken` volume mount. See the [Specifying TTL and API audience section](../oidc/kubernetes/#specifying-ttl-and-api-audience) for an example.
{{< /alert >}}

#### Using the Stronghold token as the reviewer JWT

When running Stronghold in a Kubernetes pod, Stronghold uses its local service account token as the reviewer JWT. Stronghold periodically re-reads the token file, which allows it to work with short-lived tokens.

To use the local token and CA certificate, omit
`token_reviewer_jwt` and `kubernetes_ca_cert` when configuring the auth method.
Stronghold will attempt to load them from `token` and `ca.crt` respectively inside
the default mount folder `/var/run/secrets/kubernetes.io/serviceaccount/`.

```bash
d8 stronghold write auth/kubernetes/config \
  kubernetes_host=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT
```

#### Using the Stronghold client's JWT as the reviewer JWT

When configuring Kubernetes auth, you can omit the `token_reviewer_jwt`, and Stronghold
will use the Stronghold client's JWT as its own auth token when communicating with
the Kubernetes `TokenReview` API. If Stronghold is running in Kubernetes, you also need
to set `disable_local_ca_jwt=true`.

This means Stronghold does not store any JWTs and allows you to use short-lived tokens
everywhere but adds some operational overhead to maintain the cluster role
bindings on the set of service accounts you want to be able to authenticate with
Stronghold. Each client of Stronghold would need the `system:auth-delegator` ClusterRole:

```bash
d8 k create clusterrolebinding myapp-client-auth-delegator \
  --clusterrole=system:auth-delegator \
  --group=group1 \
  --serviceaccount=default:svcaccount1 \
  ...
```

#### Using long-lived tokens

You can create a long-lived token using the [Kubernetes][k8s-create-secret] instructions
and use that as the `token_reviewer_jwt`. In this example, the `myapp` service
account would need the `system:auth-delegator` ClusterRole:

```bash
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

Using this maintains previous workflows but does not benefit from the improved
security posture of short-lived tokens.

[k8s-create-secret]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#manually-create-a-service-account-api-token

#### Using JWT auth

Kubernetes auth uses `TokenReview` API. However, the
JWT tokens Kubernetes generates can also be verified using Kubernetes as an OIDC
provider. The JWT auth method documentation has [instructions][k8s-jwt-auth] for
setting up JWT auth with Kubernetes as the OIDC provider.

[k8s-jwt-auth]: ../oidc/kubernetes/

This solution allows you to use short-lived tokens for all clients and removes
the need for a `token_reviewer_jwt` configuration. However, the client tokens cannot be revoked before
their TTL expires, so it is recommended to keep the TTL short with that
limitation in mind.

## Configuring Kubernetes

This auth method uses the `Kubernetes TokenReview API` to validate the provided JWT.

The service account used by this auth method must have permission to access the `TokenReview` API. The ClusterRoleBinding example below grants the required permissions:

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
