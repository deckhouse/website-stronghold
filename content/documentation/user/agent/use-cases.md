---
title: "Use cases"
description: "Typical Stronghold Agent scenarios: VM and bare-metal deployments, legacy applications, dynamic secrets, and CI/CD pipelines."
weight: 20
---

## Deployment on VMs and bare metal

**The main use case for Stronghold Agent** is deployment on virtual machines (VMs) and physical (bare-metal) servers.
Unlike Kubernetes, which has native mechanisms for working with secrets via CSI drivers and sidecar containers, VMs and bare metal require a separate solution.

**Benefits:**

- Centralized secret management for the entire server fleet.
- Simple integration via a systemd service.
- Automatic secret updates without downtime.
- Support for legacy systems without changing application code.

## Delivering secrets to legacy applications

Legacy applications often expect configuration in the form of:

- **Configuration files** (config.ini, application.properties, .env).
- **Environment variables**.
- **Key and certificate files**.

Stronghold Agent can:

- Render templates with secrets into files.
- Run applications with injected environment variables.
- Automatically restart the application when secrets are updated.

**Example:** a Spring Boot application receives credentials via environment variables.

Agent automatically creates environment variables with secrets:

```bash
DB_USERNAME=v-approle-myapp-abc123
DB_PASSWORD=A1b2C3d4E5f6
DB_HOST=postgres.example.com
```

In `application.properties`, simply use these variables:

```java
spring.datasource.url=jdbc:postgresql://${DB_HOST}:5432/production
spring.datasource.username=${DB_USERNAME}
spring.datasource.password=${DB_PASSWORD}
```

The application requires no changes: Spring Boot substitutes the values from the environment automatically.

## Integrating applications without SDK support

Many applications are written in languages or frameworks that have no ready-made SDK for working with Stronghold:

- Legacy C/C++ applications.
- Specialized systems (SCADA, industrial software).
- Binary applications without source code.

Stronghold Agent lets such applications use secrets through standard OS mechanisms.

## Automatic credential updates without restarts

Stronghold Agent can track secret changes and:

- **Update files** with new values.
- **Send signals** to the application (SIGHUP to reload the configuration).
- **Run commands** (reload scripts, hot reload).
- **Restart processes** on critical changes.

**Example:** the Nginx web server receives updated TLS certificates:

```hcl
template {
  source      = "/etc/nginx/ssl/cert.ctmpl"
  destination = "/etc/nginx/ssl/cert.pem"
  command     = "nginx -s reload"  # Reload without downtime
}
```

## Obtaining dynamic secrets

Stronghold supports dynamic generation of temporary credentials for various systems:

**Database credentials:**

- PostgreSQL, MySQL/MariaDB, ClickHouse.
- Temporary users with a limited TTL.
- Automatic rotation.

**PKI certificates:**

- Automatic issuance of TLS certificates.
- Renewal before expiration.
- Support for various CAs.

## Using in CI/CD pipelines

For self-hosted CI/CD runners (Jenkins, GitLab Runner), Agent runs on the runner server and provides secrets to pipelines.

**How it works:**

1. Agent is installed on the CI/CD runner server.
1. It authenticates in Stronghold via AppRole.
1. It renders secrets into files or provides them via API Proxy.
1. The pipeline reads secrets from local files.

**Typical scenarios:**

- **Deployment credentials**: SSH keys, kubeconfig for deployment.
- **Registry access**: Docker registry credentials for pulling and pushing images.
- **Cloud providers**: cloud credentials for infrastructure.

**Example for GitLab Runner:**

```yaml
# .gitlab-ci.yml
deploy:
  script:
    - ssh -i /var/run/stronghold-agent/deploy_key deploy@server.example.com "cd /app && git pull && systemctl restart app"
```

**The role of Stronghold Agent:**

- Agent gets the SSH key from Stronghold (for example, from `secret/data/ci/deploy_key`).
- It renders the key to the `/var/run/stronghold-agent/deploy_key` file with `0600` permissions.
- When the key is rotated in Stronghold, Agent automatically updates the file.
- GitLab Runner simply uses this key, and it is always up to date.

## Usage examples

Ready-made examples that use this feature:

- [Application on a virtual machine with Stronghold Agent](../../../examples/delivery/legacy-app-on-vm/)
- [TLS certificate for a web server on a VM from Stronghold PKI](../../../examples/delivery/web-server-tls/)

See all examples in [Usage examples](../../../examples/).
