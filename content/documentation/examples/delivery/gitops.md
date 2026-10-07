---
title: "GitOps"
description: "Keeping secrets out of Git with Argo CD and Flux: External Secrets Operator, argocd-vault-plugin, SOPS, and the GitOps secrets engine."
weight: 40
---

In a GitOps workflow, Git is the source of truth for manifests, but secret values must not end up in the repository. Git stores only references to secrets, and the values are fetched from Stronghold in the cluster.

The tools on this page (External Secrets Operator, argocd-vault-plugin, SOPS) are built for HashiCorp Vault. Stronghold is API-compatible with Vault, so they can be pointed at Stronghold. <!-- TODO(verify): tested compatibility of ESO, argocd-vault-plugin, and SOPS (hc_vault_transit) with Stronghold -->

## Choosing an approach

| Approach | What is stored in Git | Where the secret value ends up | Tools |
| --- | --- | --- | --- |
| External Secrets Operator | `ExternalSecret` manifests referencing Stronghold paths | Kubernetes Secret (etcd) | Argo CD, Flux |
| argocd-vault-plugin | Manifests with `<path:...#key>` placeholders | Argo CD rendered manifests and Kubernetes Secret (etcd) | Argo CD |
| SOPS with a Transit key | Encrypted files | Kubernetes Secret (etcd) after decryption | Flux, Argo CD with a plugin |
| CSI or env-injector of the `secrets-store-integration` module | Pod manifests with annotations or CSI volumes | Only in the pod | Argo CD, Flux |

The recommended option is External Secrets Operator or the [`secrets-store-integration`](/products/kubernetes-platform/documentation/v1/modules/secrets-store-integration/) module: neither values nor encrypted data reach Git, and read access is controlled by Stronghold roles. For a comparison of ways to deliver secrets to pods, see [Delivering secrets to Kubernetes pods](../kubernetes-workloads/).

## External Secrets Operator with Argo CD or Flux

Store `SecretStore` (or `ClusterSecretStore`) and `ExternalSecret` manifests in Git. Argo CD or Flux applies them as regular resources, and the operator creates a Secret with values from Stronghold:

```yaml
apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: myapp-db
  namespace: myapp
spec:
  refreshInterval: 15m
  secretStoreRef:
    name: stronghold
    kind: SecretStore
  target:
    name: myapp-db
  dataFrom:
    - extract:
        key: myapp/db
```

Configuring `SecretStore` and the Kubernetes auth role is described in [Delivering secrets to Kubernetes pods](../kubernetes-workloads/#external-secrets-operator).

If Argo CD shows the Secret as out of sync, exclude operator-generated Secrets from Argo CD tracking, since they are not in Git.

## argocd-vault-plugin

argocd-vault-plugin substitutes values into manifests during rendering in Argo CD. Git stores placeholders:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myapp-db
  namespace: myapp
type: Opaque
stringData:
  password: <path:secret/data/myapp/db#password>
```

The plugin is configured with environment variables in the Argo CD repo-server, for example:

```yaml
env:
  - name: AVP_TYPE
    value: vault
  - name: VAULT_ADDR
    value: https://stronghold.example.com
  - name: AVP_AUTH_TYPE
    value: k8s
  - name: AVP_K8S_ROLE
    value: argocd-repo-server
```

{{< alert level="warning" >}}
Rendered manifests with plaintext values are visible to Argo CD and may be cached. Grant the plugin role access only to the required paths and restrict who can view application manifests in Argo CD.
{{< /alert >}}

## SOPS with the Transit engine

SOPS can encrypt files with a [Transit secrets engine](../../../user/secrets-engines/transit/) key through the Vault-compatible API. Encrypted files are stored in Git, and Flux decrypts them when applying.

1. Create an encryption key:

   ```bash
   d8 stronghold secrets enable transit
   d8 stronghold write -f transit/keys/sops
   ```

1. Encrypt a file:

   ```bash
   export VAULT_ADDR=https://stronghold.example.com
   export VAULT_TOKEN=<token with access to transit/encrypt/sops>
   sops --encrypt --hc-vault-transit $VAULT_ADDR/v1/transit/keys/sops secret.yaml > secret.enc.yaml
   ```

To decrypt in Flux, the kustomize-controller needs a Stronghold token with access to `transit/decrypt/sops`. The token is stored in a cluster Secret, so use a dedicated policy and a short, renewable lifetime. <!-- TODO(verify): Flux hc-vault decryption Secret format and compatibility with GOST Transit keys -->

## GitOps secrets engine

To manage the configuration of Stronghold itself (policies, auth methods, mounts), use the built-in [GitOps secrets engine](../../../user/secrets-engines/gitops/overview/). It watches a Git repository, verifies commit signatures, and applies declarative configuration through the Stronghold API. The format is described in [Configuration format](../../../user/secrets-engines/gitops/configuration-format/).

The GitOps engine does not deliver secrets to applications: use it together with one of the approaches above.
