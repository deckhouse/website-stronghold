---
title: "Application clients"
description: "Using Stronghold from application code: environment variables, Go, Python, Java (Spring Cloud Vault), Node.js, reading KV, dynamic credentials, and renewing tokens and leases."
weight: 70
---

An application can call the Stronghold API directly using a client library. Stronghold is API-compatible with HashiCorp Vault, so Vault client libraries are used: `github.com/hashicorp/vault/api` for Go, `hvac` for Python, Spring Cloud Vault for Java, `node-vault` for Node.js. <!-- TODO(verify): tested compatibility of these libraries with Stronghold -->

If the application cannot be changed, or implementing token and lease renewal in it is hard, use [Stronghold Agent](../../../user/agent/overview/): it handles authentication, renewal, and delivery of secrets to files or environment variables.

## Environment variables

Vault client libraries read the standard environment variables:

| Variable | Purpose |
| --- | --- |
| `VAULT_ADDR` | Stronghold address, for example `https://stronghold.example.com` |
| `VAULT_TOKEN` | Client token. Do not set a long-lived token in production — use auth methods |
| `VAULT_CACERT` | Path to the CA certificate used to verify the Stronghold TLS certificate |
| `VAULT_NAMESPACE` | [Namespace](../../../admin/namespaces/overview/) |

The `d8 stronghold` utility uses `STRONGHOLD_ADDR`, `STRONGHOLD_TOKEN`, and `STRONGHOLD_CACERT`. In HTTP requests, the token is passed in the `X-Vault-Token` header.

## Token and lease lifecycle

Every [token](../../../concepts/tokens/) and every dynamic secret has a lifetime ([lease](../../../concepts/lease/)). The application must:

1. Authenticate at startup (for pods, with the [Kubernetes method](../../../user/auth/kubernetes/); for virtual machines, for example, with [AppRole](../../../user/auth/approle/)).
1. Renew the token before `ttl` expires, usually after two thirds of its lifetime.
1. Re-authenticate when `max_ttl` is reached and renewal is no longer possible, and on a `403` response.
1. Renew leases of dynamic secrets or request new credentials before the lease expires.
1. Revoke leases on graceful shutdown if the credentials are no longer needed.

The examples below use the `myapp` Kubernetes role, the `secret/myapp/db` KV version 2 secret, and the `my-role` database role (see [Delivering secrets to Kubernetes pods](../kubernetes-workloads/#preparation-role-and-policy) and the [PostgreSQL secrets engine](../../../user/secrets-engines/databases/postgresql/)). The policy must allow reading `database/creds/my-role`.

## Go

Use the `github.com/hashicorp/vault/api` and `github.com/hashicorp/vault/api/auth/kubernetes` packages:

```go
package main

import (
    "context"
    "log"

    vault "github.com/hashicorp/vault/api"
    auth "github.com/hashicorp/vault/api/auth/kubernetes"
)

func main() {
    ctx := context.Background()

    // DefaultConfig reads VAULT_ADDR and VAULT_CACERT.
    client, err := vault.NewClient(vault.DefaultConfig())
    if err != nil {
        log.Fatal(err)
    }

    k8sAuth, err := auth.NewKubernetesAuth("myapp")
    if err != nil {
        log.Fatal(err)
    }
    authInfo, err := client.Auth().Login(ctx, k8sAuth)
    if err != nil {
        log.Fatal(err)
    }

    // Background token renewal.
    watcher, err := client.NewLifetimeWatcher(&vault.LifetimeWatcherInput{Secret: authInfo})
    if err != nil {
        log.Fatal(err)
    }
    go watcher.Start()
    defer watcher.Stop()

    // Read KV version 2.
    kv, err := client.KVv2("secret").Get(ctx, "myapp/db")
    if err != nil {
        log.Fatal(err)
    }
    password := kv.Data["password"].(string)
    _ = password

    // Dynamic database credentials.
    creds, err := client.Logical().ReadWithContext(ctx, "database/creds/my-role")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("lease %s, ttl %ds", creds.LeaseID, creds.LeaseDuration)

    // Renew the credentials lease.
    if _, err := client.Sys().RenewWithContext(ctx, creds.LeaseID, 3600); err != nil {
        log.Print(err)
    }

    // When the token can no longer be renewed, the watcher stops: authenticate again.
    <-watcher.DoneCh()
}
```

## Python

Use the `hvac` library:

```python
import os

import hvac

client = hvac.Client(
    url=os.environ["VAULT_ADDR"],
    verify=os.environ.get("VAULT_CACERT", True),
)

with open("/var/run/secrets/kubernetes.io/serviceaccount/token") as f:
    client.auth.kubernetes.login(role="myapp", jwt=f.read())

# Read KV version 2.
secret = client.secrets.kv.v2.read_secret_version(
    path="myapp/db",
    mount_point="secret",
    raise_on_deleted_version=True,
)
password = secret["data"]["data"]["password"]

# Dynamic database credentials.
creds = client.secrets.database.generate_credentials(name="my-role")
username = creds["data"]["username"]

# Renew the lease and the token.
client.sys.renew_lease(lease_id=creds["lease_id"], increment=3600)
client.auth.token.renew_self()
```

Run token and lease renewal periodically (for example, in a separate thread) and log in again on `hvac.exceptions.Forbidden`.

## Java (Spring Cloud Vault)

Spring Cloud Vault loads secrets into the application `Environment` and renews the token and leases itself. Example `application.yml`:

```yaml
spring:
  application:
    name: myapp
  config:
    import: vault://
  cloud:
    vault:
      uri: https://stronghold.example.com
      authentication: KUBERNETES
      kubernetes:
        role: myapp
        kubernetes-path: kubernetes
      kv:
        enabled: true
        backend: secret
        default-context: myapp/db
      database:
        enabled: true
        backend: database
        role: my-role
```

Values from `secret/myapp/db` become Spring properties (for example, `${password}`), and dynamic database credentials go to `spring.datasource.username` and `spring.datasource.password` by default.

## Node.js

Use the `node-vault` library:

```javascript
const fs = require('fs');
const vault = require('node-vault')({
  apiVersion: 'v1',
  endpoint: process.env.VAULT_ADDR,
});

async function main() {
  const jwt = fs.readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token', 'utf8');
  const login = await vault.kubernetesLogin({ role: 'myapp', jwt });
  vault.token = login.auth.client_token;

  // Read KV version 2.
  const kv = await vault.read('secret/data/myapp/db');
  const password = kv.data.data.password;

  // Dynamic database credentials.
  const creds = await vault.read('database/creds/my-role');

  // Renew the token and the lease.
  await vault.tokenRenewSelf();
  await vault.write('sys/leases/renew', { lease_id: creds.lease_id, increment: 3600 });
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
```

## Recommendations

- Verify the Stronghold TLS certificate. Do not disable verification in production.
- Do not write tokens or secret values to logs.
- Cache static secrets and re-read them on a schedule rather than on every request.
- To pass a secret to another process once, use [response wrapping](../../../concepts/response-wrapping/).
