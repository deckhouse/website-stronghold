---
title: "Delivering secrets to Kubernetes pods"
linkTitle: "Kubernetes workloads"
description: "Comparison of ways to deliver Stronghold secrets to Kubernetes pods: CSI, env-injector, Stronghold Agent, External Secrets Operator, and direct API calls."
weight: 10
---

An application in Kubernetes can get secrets from Stronghold in several ways. All of them rely on the [Kubernetes auth method](../../../user/auth/kubernetes/): a pod presents its ServiceAccount token, and Stronghold issues a token with the policies bound to the role.

## Comparison

| Method | How the secret reaches the pod | Stored in etcd | Secret updates | When to use |
| --- | --- | --- | --- | --- |
| CSI driver of the `secrets-store-integration` module | File in a volume mounted into the container | No | When the volume is remounted or the pod is restarted<!-- TODO(verify): whether the module supports periodic file refresh --> | The application reads secrets from files; the cluster is managed by Deckhouse Platform |
| env-injector of the `secrets-store-integration` module | Process environment variable | No | Only on pod restart | The application reads configuration only from environment variables |
| Stronghold Agent (init container or sidecar) | File rendered from a template, or environment variables of a child process | No | The sidecar re-renders templates when the secret changes or a lease is renewed | You need templates, dynamic secrets, or lease renewal without code changes |
| External Secrets Operator (ESO) | Kubernetes Secret that the pod mounts as a volume or env vars | Yes | By `refreshInterval`: the Secret is updated, volume files follow with kubelet delay, env vars only after a restart | Applications and manifests already expect regular Secrets; GitOps |
| Direct API call from the application | The application requests the secret itself | No | Fully controlled by the application | The application uses an SDK and can renew tokens and leases |

General recommendations:

- If a secret must not be stored in etcd, choose CSI, env-injector, Stronghold Agent, or direct API calls.
- Secrets in environment variables are visible to processes that can read `/proc/<pid>/environ` and may end up in dumps and logs. Prefer files where possible.
- For [dynamic secrets](../../../user/secrets-engines/databases/overview/) and short-lived certificates, use Stronghold Agent as a sidecar or direct API calls: only these renew leases during the pod lifetime.

## Preparation: role and policy

The example below applies to all methods. It assumes that the [KV version 2 secrets engine](../../../user/secrets-engines/kv/kv-v2/) is enabled at `secret`, and the Kubernetes auth method at `kubernetes`.

1. Create a policy that allows reading the application secrets:

   ```bash
   d8 stronghold policy write myapp-read - <<'POLICY'
   path "secret/data/myapp/*" {
     capabilities = ["read"]
   }
   POLICY
   ```

1. Store the secret:

   ```bash
   d8 stronghold kv put -mount=secret myapp/db username=app password=S3cr3t
   ```

1. Create a role that binds the `myapp` ServiceAccount in the `myapp` namespace to the policy:

   ```bash
   d8 stronghold write auth/kubernetes/role/myapp \
     bound_service_account_names=myapp \
     bound_service_account_namespaces=myapp \
     policies=myapp-read \
     ttl=1h
   ```

   In DP in `Automatic` mode, the Kubernetes auth method for the current cluster is created by the module at the `kubernetes_local` path: in this case, use `auth/kubernetes_local` instead of `auth/kubernetes` here and in the examples below (`mount_path`, login URL). The `auth/kubernetes` path applies to a manually enabled method.

1. Create the ServiceAccount in the cluster:

   ```bash
   d8 k create namespace myapp
   d8 k -n myapp create serviceaccount myapp
   ```

Configuring the auth method itself (`auth/kubernetes/config`) is described in [Kubernetes auth method](../../../user/auth/kubernetes/). In Deckhouse Platform, the `stronghold` module in `Automatic` mode [connects the current cluster automatically](../../../install/dkp/configuration/) for the `secrets-store-integration` module.

## secrets-store-integration CSI driver

The Deckhouse Platform module [`secrets-store-integration`](/products/kubernetes-platform/documentation/v1/modules/secrets-store-integration/) mounts secrets into the pod as files using a CSI driver. The secret is not stored in a Secret object and does not reach etcd.

Example (resource names, API version, and fields follow the module documentation and must be checked against your DP version):

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: SecretsStoreImport
metadata:
  name: myapp-db
  namespace: myapp
spec:
  type: CSI
  role: myapp
  files:
    - name: db-password
      source:
        path: secret/data/myapp/db
        key: password
---
apiVersion: v1
kind: Pod
metadata:
  name: myapp
  namespace: myapp
spec:
  serviceAccountName: myapp
  containers:
    - name: app
      image: registry.example.com/myapp:1.0.0
      volumeMounts:
        - name: secrets
          mountPath: /mnt/secrets
          readOnly: true
  volumes:
    - name: secrets
      csi:
        driver: secrets-store.csi.deckhouse.io
        volumeAttributes:
          secretsStoreImport: myapp-db
```

The application reads the password from `/mnt/secrets/db-password`.

## secrets-store-integration env-injector

The env-injector from the same module substitutes secret values into environment variables when the container starts. The values exist only in process memory and are not written to the pod spec.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: myapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
      annotations:
        secrets-store.deckhouse.io/role: myapp
    spec:
      serviceAccountName: myapp
      containers:
        - name: app
          image: registry.example.com/myapp:1.0.0
          env:
            - name: DB_USER
              value: secrets-store:secret/data/myapp/db#username
            - name: DB_PASSWORD
              value: secrets-store:secret/data/myapp/db#password
```

To make the application pick up a new secret value, restart the pod, for example with `d8 k -n myapp rollout restart deployment myapp`.

## Stronghold Agent

[Stronghold Agent](../../../user/agent/overview/) authenticates using the Kubernetes method, fetches secrets, and renders them to files using templates. In a pod, Agent runs:

- as an init container with `exit_after_auth = true` — secrets are rendered once at pod start;
- as a sidecar container — Agent renews the token and leases and re-renders templates when secrets change.

Files are shared with the application through an `emptyDir` volume with `medium: Memory`, so secrets are not written to the node disk.

Example Agent configuration in a ConfigMap (the section format is described in [Settings](../../../user/agent/settings/)):

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-agent
  namespace: myapp
data:
  agent.hcl: |
    stronghold {
      address = "https://stronghold.example.com"
    }

    auto_auth {
      method "kubernetes" {
        mount_path = "auth/kubernetes"
        config = {
          role = "myapp"
        }
      }
    }

    template {
      destination = "/secrets/db.env"
      contents    = <<-EOT
      {{ with secret "secret/data/myapp/db" }}
      DB_USER={{ .Data.data.username }}
      DB_PASSWORD={{ .Data.data.password }}
      {{ end }}
      EOT
    }
```

The Stronghold image uses `stronghold` as its entrypoint, so `args` contains `agent` and `-config=<path to agent.hcl>`; use the image name from your registry.

Example pod with Agent as a sidecar:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
  namespace: myapp
spec:
  serviceAccountName: myapp
  containers:
    - name: stronghold-agent
      image: registry.example.com/stronghold:<version>
      args: ["agent", "-config=/etc/stronghold-agent/agent.hcl"]
      volumeMounts:
        - name: agent-config
          mountPath: /etc/stronghold-agent
        - name: secrets
          mountPath: /secrets
    - name: app
      image: registry.example.com/myapp:1.0.0
      volumeMounts:
        - name: secrets
          mountPath: /secrets
          readOnly: true
  volumes:
    - name: agent-config
      configMap:
        name: myapp-agent
    - name: secrets
      emptyDir:
        medium: Memory
```

For the init container variant, move the `stronghold-agent` container to `initContainers` and add `exit_after_auth = true` to `agent.hcl`.

## External Secrets Operator

[External Secrets Operator](https://external-secrets.io/) is a third-party operator with a `vault` provider. Stronghold is API-compatible with Vault, so the provider can be pointed at Stronghold. <!-- TODO(verify): tested ESO compatibility with Stronghold and the supported external-secrets.io API version -->

ESO creates a regular Secret object, so the secret value is stored in etcd. Enable encryption of Secret objects in etcd using cluster tooling and restrict access to Secrets with RBAC.

```yaml
apiVersion: external-secrets.io/v1
kind: SecretStore
metadata:
  name: stronghold
  namespace: myapp
spec:
  provider:
    vault:
      server: https://stronghold.example.com
      path: secret
      version: v2
      auth:
        kubernetes:
          mountPath: kubernetes
          role: myapp
          serviceAccountRef:
            name: myapp
---
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
  data:
    - secretKey: password
      remoteRef:
        key: myapp/db
        property: password
```

## Direct API call

The application can authenticate with its ServiceAccount token and read the secret on its own. Example with `curl`:

```bash
SA_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

STRONGHOLD_TOKEN=$(curl -s --request POST \
  --data "{\"role\": \"myapp\", \"jwt\": \"${SA_TOKEN}\"}" \
  https://stronghold.example.com/v1/auth/kubernetes/login | jq -r '.auth.client_token')

curl -s --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  https://stronghold.example.com/v1/secret/data/myapp/db | jq '.data.data'
```

SDK examples and guidance on renewing tokens and leases are in [Application clients](../app-clients/).
