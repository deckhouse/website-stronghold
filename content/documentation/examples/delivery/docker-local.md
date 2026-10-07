---
title: "Local Stronghold for development"
linkTitle: "Local run in Docker"
description: "Running Stronghold in Docker on a developer workstation: server -dev mode and a single-node Raft configuration in Docker Compose."
weight: 80
---

For development and debugging integrations, you can run Stronghold locally in Docker. Use such an instance for development only: do not store real secrets in it.

## Building the image

The image is built from the Stronghold binary with [`stronghold bootstrap docker`](../../../install/standalone/bootstrap/#build-a-docker-image):

```bash
stronghold bootstrap docker
docker load -i stronghold-v1.19.0.tar
```

By default, the image is tagged `stronghold:<version>`, for example `stronghold:1.19.0`. <!-- TODO(verify): availability of an official Stronghold image in a public registry -->

## Development mode

The default container command is `stronghold server -dev`. In development mode, Stronghold:

- keeps data in memory — everything is lost when the container stops;
- initializes and unseals automatically;
- prints the root token to the log;
- runs without TLS.

Start the container:

```bash
docker run --rm --name stronghold-dev -p 8200:8200 stronghold:1.19.0
```

Find the root token in the container log (the `Root Token` line) and configure the client:

```bash
export STRONGHOLD_ADDR=http://127.0.0.1:8200
export STRONGHOLD_TOKEN=<root token from the log>
d8 stronghold status
```

For Vault client libraries, set the same values in `VAULT_ADDR` and `VAULT_TOKEN` (see [Application clients](../app-clients/)).

To set a predictable root token and listen address, pass development mode parameters:

```bash
docker run --rm --name stronghold-dev -p 8200:8200 stronghold:1.19.0 \
  stronghold server -dev -dev-root-token-id=root -dev-listen-address=0.0.0.0:8200
```

## Single-node Raft configuration in Docker Compose

If data must persist across restarts, run Stronghold with a configuration file and Raft storage.

1. Create `config/stronghold.hcl` (the format is described in [Configuration](../../../install/standalone/configuration/)):

   ```hcl
   ui            = true
   disable_mlock = true
   api_addr      = "http://127.0.0.1:8200"
   cluster_addr  = "http://127.0.0.1:8201"

   storage "raft" {
     path    = "/stronghold/data"
     node_id = "dev-node-1"
   }

   listener "tcp" {
     address     = "0.0.0.0:8200"
     tls_disable = true
   }
   ```

   {{< alert level="warning" >}}
   `tls_disable = true` is acceptable only on a local developer machine.
   {{< /alert >}}

1. Create `compose.yaml`:

   ```yaml
   services:
     stronghold:
       image: stronghold:1.19.0
       command: ["stronghold", "server", "-config=/stronghold/config/stronghold.hcl"]
       ports:
         - "8200:8200"
       volumes:
         - ./config:/stronghold/config:ro
         - stronghold-data:/stronghold/data
       restart: unless-stopped

   volumes:
     stronghold-data:
   ```

   The image runs as root (no user is set in the image), so no additional permissions on the `/stronghold/data` volume are needed. The image has no ENTRYPOINT: the full command is passed, as in the example.

1. Start the container:

   ```bash
   docker compose up -d
   ```

1. Initialize and unseal Stronghold. For local development, a single key is enough:

   ```bash
   export STRONGHOLD_ADDR=http://127.0.0.1:8200
   d8 stronghold operator init -key-shares=1 -key-threshold=1
   d8 stronghold operator unseal
   ```

   Save the unseal key and root token from the `operator init` output. After each container restart, run `d8 stronghold operator unseal`.

1. Log in with the root token and enable the engines you need, for example KV version 2:

   ```bash
   export STRONGHOLD_TOKEN=<root token>
   d8 stronghold secrets enable -path=secret -version=2 kv
   d8 stronghold kv put -mount=secret myapp/db username=app password=S3cr3t
   ```

The web UI is available at `http://127.0.0.1:8200/ui`.

## Running an application next to Stronghold

An application in the same `compose.yaml` reaches Stronghold by service name. The example uses the `root` root token set with `-dev-root-token-id`:

```yaml
services:
  app:
    image: registry.example.com/myapp:1.0.0
    environment:
      VAULT_ADDR: http://stronghold:8200
      VAULT_TOKEN: root
    depends_on:
      - stronghold
```

A token in an environment variable is acceptable only for local development. In real environments, use the auth methods described in [Delivering secrets to Kubernetes pods](../kubernetes-workloads/) and [Application clients](../app-clients/).
