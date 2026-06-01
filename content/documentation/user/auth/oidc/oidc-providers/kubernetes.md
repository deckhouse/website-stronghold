---
title: "Kubernetes as an OIDC provider"
linkTitle: "Kubernetes"
weight: 40
description: "Using Kubernetes as an OIDC provider for JWT/OIDC authentication in Deckhouse Stronghold."
---

Kubernetes can act as an OIDC provider for Deckhouse Stronghold. In this scenario, Stronghold validates ServiceAccount JWT tokens using the `JWT/OIDC auth` authentication method.

This option is suitable if applications already use ServiceAccount tokens and need to authenticate to Stronghold without the `Kubernetes auth` method. Validation is performed either through the OIDC Discovery URL or through public keys used to verify JWT signatures.

{% alert level="warning" %}
The `JWT` authentication method does not use the Kubernetes API `TokenReview`. Instead, Stronghold validates JWTs using public key cryptography. Because of this, changes to the token state or the related ServiceAccount may not be taken into account until the token expires. To reduce this risk, use short-lived ServiceAccount tokens or configure a separate `Kubernetes auth` method that uses the Kubernetes API `TokenReview`.
{% endalert %}

## When to use

Use this scenario in the following cases:

- You need to authenticate applications in Stronghold that use ServiceAccount tokens.
- The Kubernetes cluster can act as an OIDC provider.
- You need to use `JWT auth` instead of `Kubernetes auth`.

## Configuration options

Two main configuration options are available when using Kubernetes as an OIDC provider:

- Via the OIDC Discovery URL.
- Via public keys for JWT validation.

## Configure via the OIDC Discovery URL

This is the simplest option if the `ServiceAccountIssuerDiscovery` option is enabled in the cluster and OIDC discovery metadata is available.

### Requirements

Before you begin, make sure that:

- the `ServiceAccountIssuerDiscovery` option is enabled;
- the `kube-apiserver` `--service-account-issuer` parameter contains an address that is reachable from Stronghold;
- short-lived ServiceAccount tokens are used;
- if the OIDC Discovery URL uses a custom certificate, Stronghold is configured to trust the corresponding CA certificate.

The `ServiceAccountIssuerDiscovery` option is available starting with Kubernetes 1.18 and is enabled by default starting with Kubernetes 1.20.

Short-lived ServiceAccount tokens mounted into pods are used by default starting with Kubernetes 1.21.

### Step 1: Open access to the OIDC discovery metadata

Make sure that the OIDC Discovery URL is accessible without authentication. For more information, refer to the Kubernetes documentation about ServiceAccount issuer discovery:

```shell
d8 k create clusterrolebinding oidc-reviewer \
  --clusterrole=system:service-account-issuer-discovery \
  --group=system:unauthenticated
```

### Step 2: Get the issuer value

Get the `issuer` value from the cluster OIDC configuration:

```shell
ISSUER="$(d8 k get --raw /.well-known/openid-configuration | jq -r '.issuer')"
```

### Step 3: Enable and configure JWT auth

Enable the `jwt` authentication method and specify the obtained `issuer` value in the `oidc_discovery_url` parameter:

```shell
d8 stronghold auth enable jwt
d8 stronghold write auth/jwt/config oidc_discovery_url="${ISSUER}"
```

Stronghold uses this URL to retrieve OIDC discovery metadata.

### Step 4: Create a role

After configuring the `JWT auth` method,
create a role as described in the "Create a role and authenticate" section.

## Configure via public keys

Use this option in the following cases:

- the Kubernetes API is unavailable from Stronghold;
- one `JWT auth` endpoint must serve multiple clusters;
- it is more convenient to use a set of public keys directly.

### Requirements

Before you begin, make sure that:

- the `ServiceAccountIssuerDiscovery` option is enabled, or you have access to the `/etc/kubernetes/pki/sa.pub` file;
- short-lived ServiceAccount tokens are used.

The `ServiceAccountIssuerDiscovery` option is available starting with Kubernetes 1.18
and is enabled by default starting with Kubernetes 1.20.

If you have access to the `/etc/kubernetes/pki/sa.pub` file on a master node of the cluster, you can skip the steps for obtaining the key and converting it to the `PEM` format, because the key is already stored in the required format.

Short-lived ServiceAccount tokens mounted into pods are used by default starting with Kubernetes 1.21.

1. Get the JWKS with public keys

If OIDC metadata is available, get the JWKS via `jwks_uri`:

```shell
d8 k get --raw "$(d8 k get --raw /.well-known/openid-configuration | jq -r '.jwks_uri' | sed -r 's/.*\.[^/]+(.*)/\1/')"
```

1. Convert keys to PEM

Convert the keys from the `JWK` format to the `PEM` format.

1. Configure JWT auth

Configure the `JWT auth` method to use the obtained keys:

```shell
d8 stronghold write auth/jwt/config \
  jwt_validation_pubkeys="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9...
-----END PUBLIC KEY-----","-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9...
-----END PUBLIC KEY-----"
```

1. Create a role

After configuring the `JWT auth` method, create a role as described in the "Create a role and authenticate" section.

## Create a role and authenticate

After configuring the `JWT auth` authentication method, create a role and perform a test authentication.

By default, the examples below use a ServiceAccount token mounted into a pod. If you need to control the `aud` value or TTL, use a separate `serviceAccountToken` mount as shown in the "Configure TTL and API audience for the ServiceAccount token" section.

### How to get the API audience

Select the value from the `aud` field that is used in the ServiceAccount token. Two ways to get the `aud` value are shown below.

The example that retrieves `aud` by creating a token requires `kubectl` version 1.24.0 or later:

```shell
d8 k create token default | cut -f2 -d. | base64 --decode
```

Example of retrieving the token from a pod:

```shell
d8 k exec my-pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/token | cut -f2 -d. | base64 --decode
```

For example, the `aud` value may look like `https://kubernetes.default.svc.cluster.local`. Always verify the actual value in your cluster.

### Create a role

The following example creates a role that can be used by the `default` ServiceAccount in the `default` namespace:

```shell
d8 stronghold write auth/jwt/role/my-role \
  role_type="jwt" \
  bound_audiences="<AUDIENCE-FROM-PREVIOUS-STEP>" \
  user_claim="sub" \
  bound_subject="system:serviceaccount:default:default" \
  policies="default" \
  ttl="1h"
```

### Authenticate

After that, applications in pods and other clients that have access to the ServiceAccount JWT token can authenticate:

```shell
d8 stronghold write auth/jwt/login \
  role=my-role \
  jwt=@/var/run/secrets/kubernetes.io/serviceaccount/token
```

Equivalent API request:

```shell
curl \
  --fail \
  --request POST \
  --data '{"jwt":"<JWT-TOKEN-HERE>","role":"my-role"}' \
  "${STRONGHOLD_ADDR}/v1/auth/jwt/login"
```

## Configure TTL and API audience for the ServiceAccount token

If you need to set custom TTL values or `aud` values for ServiceAccount tokens, use a separate `serviceAccountToken` mount in the pod.

This option is especially useful if you cannot disable the `kube-apiserver` `--service-account-extend-token-expiration` parameter and still need to use a short token TTL.

For the token from the example below, specify the following value in the role:

```text
bound_audiences=stronghold
```

Example:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  # The automountServiceAccountToken parameter is redundant in this example
  # because the mountPath used matches the default path.
  # This override prevents the default token from being created.
  # With a different mount path, this parameter helps ensure
  # that only one token is mounted into the pod.
  automountServiceAccountToken: false
  containers:
    - name: nginx
      image: nginx
      volumeMounts:
        - name: custom-token
          mountPath: /var/run/secrets/kubernetes.io/serviceaccount
  volumes:
    - name: custom-token
      projected:
        defaultMode: 420
        sources:
          - serviceAccountToken:
              path: token
              expirationSeconds: 600
              audience: stronghold
          - configMap:
              name: kube-root-ca.crt
              items:
                - key: ca.crt
                  path: ca.crt
          - downwardAPI:
              items:
                - fieldRef:
                    apiVersion: v1
                    fieldPath: metadata.namespace
                  path: namespace
```

## What's next

After configuring Kubernetes as an OIDC provider, perform the following steps:

1. Verify that the `JWT auth` method is configured correctly.
1. Create a role with the required `bound_audiences` and `bound_subject` values.
1. Perform a test authentication from a pod.
1. Make sure that the issued Stronghold token contains the expected policies.

## Additional information

For more information, use the following materials:

- [ServiceAccount issuer discovery in the Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#service-account-issuer-discovery)
- [kube-apiserver parameters in the Kubernetes documentation](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options)
