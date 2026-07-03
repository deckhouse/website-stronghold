---
title: "Kubernetes as an OIDC provider"
linkTitle: "Kubernetes"
weight: 40
---

Kubernetes can act as an OIDC provider so that Stronghold can validate service account tokens using JWT auth or OIDC auth.

{{< alert >}}
The JWT auth mechanism does **not** use the Kubernetes API `TokenReview` for token validation.
Instead, it uses public-key cryptography to verify the JWT contents.
This means that tokens revoked by Kubernetes remain valid until they expire.

To reduce this risk, use short TTLs for service account tokens or use [Kubernetes auth](../../kubernetes/), which uses the `TokenReview` API.
{{< /alert >}}

## Using the discovery URL

When using discovery, you only need to specify the OIDC discovery URL.
If the OIDC URL uses a custom certificate, you also need a trusted CA.
This is the simplest configuration mode if the Kubernetes cluster meets the requirements.

The Kubernetes cluster must meet the following requirements:

- The [`ServiceAccountIssuerDiscovery`][k8s-sa-issuer-discovery] feature gate must be enabled.
  - It is available starting from version 1.18 and enabled by default starting from version 1.20.
- The URL specified in the `--service-account-issuer` parameter of the `kube-apiserver` component must contain an address accessible from Stronghold.
  For most managed Kubernetes services, this address is public.
- Short-lived Kubernetes service account tokens must be used.
  - This behavior is enabled by default for tokens mounted into pods starting from Kubernetes 1.21.

To enable automatic configuration, follow these steps:

1. Make sure the OIDC discovery URL does not require authentication, as described in the [Kubernetes documentation][k8s-sa-issuer-discovery].

   ```bash
   d8 k create clusterrolebinding oidc-reviewer  \
     --clusterrole=system:service-account-issuer-discovery \
     --group=system:unauthenticated
   ```

1. Determine the issuer URL for your cluster.

   ```bash
   ISSUER="$(d8 k get --raw /.well-known/openid-configuration | jq -r '.issuer')"
   ```

1. Enable and configure JWT auth in Stronghold.

   ```bash
   d8 stronghold auth enable jwt
   d8 stronghold write auth/jwt/config oidc_discovery_url="${ISSUER}"
   ```

1. Configure the [required roles](#creating-roles-and-authenticating).

[k8s-sa-issuer-discovery]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#service-account-issuer-discovery

## Using public keys to validate JWTs

This method can be useful if the Kubernetes API is not accessible from Stronghold or if you want a single JWT auth endpoint to serve multiple Kubernetes clusters using a public key chain.

The Kubernetes cluster must meet the following requirements:

- The [`ServiceAccountIssuerDiscovery`][k8s-sa-issuer-discovery] feature gate must be enabled.
  - It is available starting from version 1.18 and enabled by default starting from version 1.20.
  - This requirement is optional if you have access to the `/etc/kubernetes/pki/sa.pub` file on the cluster master node.
    In this case, you can skip the steps for retrieving the key and converting it to PEM format because the key is already stored in the required format.
- Short-lived Kubernetes service account tokens must be used.
  - This behavior is enabled by default for tokens mounted into pods starting from Kubernetes 1.21.

To configure JWT auth using Kubernetes public keys, follow these steps:

1. Retrieve the public key used to sign service account tokens from your cluster's JWKS URI.

   ```bash
   # The jwks_uri is available in /.well-known/openid-configuration.
   d8 k get --raw "$(d8 k get --raw /.well-known/openid-configuration | jq -r '.jwks_uri' | sed -r 's/.*\.[^/]+(.*)/\1/')"
   ```

1. Convert the keys from JWK format to PEM format.
   You can use a CLI utility or an [online JWK-to-PEM converter][jwk-to-pem].

1. Configure the JWT auth endpoint to use the retrieved keys.

   ```bash
   d8 stronghold write auth/jwt/config \
      jwt_validation_pubkeys="-----BEGIN PUBLIC KEY-----
   MIIBIjANBgkqhkiG9...
   -----END PUBLIC KEY-----","-----BEGIN PUBLIC KEY-----
   MIIBIjANBgkqhkiG9...
   -----END PUBLIC KEY-----"
   ```

1. Configure the [required roles](#creating-roles-and-authenticating).

[jwk-to-pem]: https://8gwifi.org/jwkconvertfunctions.jsp

## Creating roles and authenticating

After the JWT auth endpoint is configured, you can configure a role and authenticate.
The examples below assume that you use the service account token available by default in all pods.
If you need to control the API audience or TTL, see [Specifying TTL and API audience](#specifying-ttl-and-api-audience).

Choose any value from the default API audiences.
In these examples, the `aud` array contains only one API audience, `https://kubernetes.default.svc.cluster.local`.

To find the default API audience, create a new token.
This requires the [Deckhouse CLI `d8` tool](/products/kubernetes-platform/documentation/v1/cli/d8/) v1.24.0 or later:

```shell-session
d8 k create token default | cut -f2 -d. | base64 --decode
{"aud":["https://kubernetes.default.svc.cluster.local"], ... "sub":"system:serviceaccount:default:default"}
```

You can also read the token from a running pod:

```shell-session
d8 k exec my-pod -- cat /var/run/secrets/kubernetes.io/serviceaccount/token | cut -f2 -d. | base64 --decode
{"aud":["https://kubernetes.default.svc.cluster.local"], ... "sub":"system:serviceaccount:default:default"}
```

Create a role for JWT auth that the `default` service account in the `default` namespace can use.

```bash
d8 stronghold write auth/jwt/role/my-role \
  role_type="jwt" \
  bound_audiences="<AUDIENCE-FROM-PREVIOUS-STEP>" \
  user_claim="sub" \
  bound_subject="system:serviceaccount:default:default" \
  policies="default" \
  ttl="1h"
```

Pods or clients that have access to the service account JWT can now authenticate with this token.

Authentication example using the Deckhouse CLI:


## Specifying TTL and API audience

If you need to specify a custom TTL or API audience for service account tokens, the following Pod manifest shows a volume mount that overrides the automatically mounted default token.
This is especially relevant if you cannot disable the [`--service-account-extend-token-expiration`][k8s-extended-tokens] flag for `kube-apiserver` and want to use short TTLs.

When using the resulting token, set `bound_audiences=stronghold` when creating roles in JWT auth.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  # The automountServiceAccountToken parameter is redundant in this example because the
  # selected mountPath matches the default path. This override prevents creation of the
  # default token. However, you can use this parameter to ensure that only one token is
  # mounted if you choose a different mount path.
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
              expirationSeconds: 600 # Minimum TTL is 10 minutes.
              audience: stronghold   # Must match your role's `bound_audiences` parameter.
          # The remaining parameters are added to mimic the standard token creation behavior.
          # They create the same objects as when automountServiceAccountToken is enabled.
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

[k8s-extended-tokens]: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options
