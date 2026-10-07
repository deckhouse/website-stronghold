---
title: "cert-manager"
description: "Issuing TLS certificates in Kubernetes with cert-manager using a vault-type Issuer, the Stronghold PKI secrets engine, and Kubernetes auth."
weight: 20
---

cert-manager issues certificates through an Issuer or ClusterIssuer of type `vault`, which calls the [PKI secrets engine](../../../user/secrets-engines/pki/) at `pki/sign/<role>`. Stronghold is API-compatible with Vault, so such an Issuer can be pointed at Stronghold. <!-- TODO(verify): tested compatibility of the cert-manager vault Issuer with Stronghold, including GOST PKI keys -->

In Deckhouse Platform, cert-manager is installed by the [`cert-manager`](/products/kubernetes-platform/documentation/v1/modules/cert-manager/) module. <!-- TODO(verify): cert-manager module documentation path, controller namespace, and ServiceAccount name in DKP -->

## Configuring Stronghold

It is assumed that the PKI engine is enabled at `pki`, a root or intermediate certificate is created, and the `example-dot-ru` role exists, as described in [PKI secrets engine](../../../user/secrets-engines/pki/).

1. Create a policy that allows signing certificate requests for the role:

   ```bash
   d8 stronghold policy write cert-manager-pki - <<'POLICY'
   path "pki/sign/example-dot-ru" {
     capabilities = ["create", "update"]
   }
   POLICY
   ```

1. Create a [Kubernetes auth](../../../user/auth/kubernetes/) role for the `stronghold-issuer` ServiceAccount in the `myapp` namespace. The `audience` parameter must match the token audience that cert-manager requests: `vault://<namespace>/<Issuer name>` for an Issuer, `vault://<ClusterIssuer name>` for a ClusterIssuer:

   ```bash
   d8 stronghold write auth/kubernetes/role/cert-manager-myapp \
     bound_service_account_names=stronghold-issuer \
     bound_service_account_namespaces=myapp \
     audience=vault://myapp/stronghold \
     policies=cert-manager-pki \
     ttl=20m
   ```

   In DP in `Automatic` mode, the Kubernetes auth method for the current cluster is created by the module at the `kubernetes_local` path: in this case, use `auth/kubernetes_local` instead of `auth/kubernetes` in the commands and in the Issuer `mountPath`. The `auth/kubernetes` path applies to a manually enabled method.

## Configuring cert-manager

1. Create a ServiceAccount and allow cert-manager to request tokens for it:

   ```yaml
   apiVersion: v1
   kind: ServiceAccount
   metadata:
     name: stronghold-issuer
     namespace: myapp
   ---
   apiVersion: rbac.authorization.k8s.io/v1
   kind: Role
   metadata:
     name: stronghold-issuer-token
     namespace: myapp
   rules:
     - apiGroups: [""]
       resources: ["serviceaccounts/token"]
       resourceNames: ["stronghold-issuer"]
       verbs: ["create"]
   ---
   apiVersion: rbac.authorization.k8s.io/v1
   kind: RoleBinding
   metadata:
     name: stronghold-issuer-token
     namespace: myapp
   roleRef:
     apiGroup: rbac.authorization.k8s.io
     kind: Role
     name: stronghold-issuer-token
   subjects:
     - kind: ServiceAccount
       name: cert-manager
       namespace: d8-cert-manager
   ```

1. Create the Issuer. In `caBundle`, specify the base64-encoded CA certificate that signed the Stronghold TLS certificate:

   ```yaml
   apiVersion: cert-manager.io/v1
   kind: Issuer
   metadata:
     name: stronghold
     namespace: myapp
   spec:
     vault:
       server: https://stronghold.example.com
       path: pki/sign/example-dot-ru
       caBundle: <base64-encoded CA>
       auth:
         kubernetes:
           role: cert-manager-myapp
           mountPath: /v1/auth/kubernetes
           serviceAccountRef:
             name: stronghold-issuer
   ```

1. Check that the Issuer is ready:

   ```bash
   d8 k -n myapp get issuer stronghold
   ```

   The `READY` column must show `True`.

1. Request a certificate:

   ```yaml
   apiVersion: cert-manager.io/v1
   kind: Certificate
   metadata:
     name: www
     namespace: myapp
   spec:
     secretName: www-tls
     commonName: www.my-website.ru
     dnsNames:
       - www.my-website.ru
     duration: 72h
     renewBefore: 24h
     issuerRef:
       name: stronghold
       kind: Issuer
   ```

cert-manager stores the certificate and private key in the `www-tls` Secret and renews the certificate before it expires. The `duration` value must not exceed the PKI role `max_ttl`.

## Other authentication methods

A `vault` Issuer also supports AppRole authentication (`auth.appRole` with `roleId` and a reference to a Secret with `secret_id`) and token authentication (`auth.tokenSecretRef`). Kubernetes auth is preferable because it does not require storing long-lived credentials in the cluster.

Using Stronghold PKI as an external CA for Istio is described in [Service mesh](../service-mesh/).
