---
title: "Use cases"
linkTitle: "Use cases"
description: "Common Stronghold Agent use cases"
weight: 20
---

Stronghold Agent is needed where an application must receive secrets from Stronghold but cannot work with it directly. This most often applies to applications running on virtual machines, on bare metal, legacy systems, and self-hosted services that require configuration files, environment variables, keys, or certificates.
This page shows the scenarios in which Stronghold Agent is especially useful and how it is typically used in practice.

## When Stronghold Agent is the best fit

Stronghold Agent is especially convenient if you need to:

- pass secrets to an application without changing its code.
- work outside Kubernetes, for example on virtual machines or on bare metal.
- automatically renew secrets and tokens.
- deliver secrets to configuration files, keys, certificates, or environment variables.
- integrate applications that do not have an SDK for working with Stronghold.

## Deployment on virtual machines and bare metal

**The primary use case for Stronghold Agent** is deployment on virtual machines and on bare metal.

In Kubernetes, native mechanisms for working with secrets are usually used, for example CSI drivers and sidecar containers. On virtual machines and bare metal, such mechanisms are not available, so a separate component is needed to retrieve secrets from Stronghold and deliver them to the application.

This provides the following advantages:

- centralized secret management for the entire server fleet.
- simple integration through a `systemd` service.
- automatic secret renewal without downtime.
- support for legacy systems without the need to change application code.

## Delivering secrets to legacy applications

Legacy applications often expect configuration in one of the standard formats:

- configuration files, for example `config.ini`, `application.properties`, `.env`.
- environment variables.
- key and certificate files.

Stronghold Agent makes it possible to integrate Stronghold into such applications without direct integration with the Stronghold API. To do this, the Agent can:

- render templates with secrets into files.
- start applications with environment variables populated with secrets.
- automatically restart the application when secrets are updated.

### Example: Spring Boot and environment variables

Stronghold Agent can prepare environment variables with secrets:

```shell
DB_USERNAME=v-approle-myapp-abc123
DB_PASSWORD=A1b2C3d4E5f6
DB_HOST=postgres.example.com
```

The application can then use them in `application.properties`:

```java
spring.datasource.url=jdbc:postgresql://${DB_HOST}:5432/production
spring.datasource.username=${DB_USERNAME}
spring.datasource.password=${DB_PASSWORD}
```

In this scenario, the application requires no changes. Spring Boot injects the values from the environment variables automatically.

## Integration with applications without an SDK

Stronghold Agent is useful not only for legacy applications, but also for any systems that do not have a ready-made SDK for working with Stronghold.

This can include:

- a legacy application written in C or C++.
- a specialized system, for example SCADA or industrial software.
- a binary application without source code.

In such scenarios, Stronghold Agent acts as an intermediate layer and passes secrets through standard operating system mechanisms — files and environment variables.

## Automatic credential renewal

If secrets and credentials need to be renewed regularly, Stronghold Agent helps automate this process.

Depending on the scenario, the Agent can:

- update files with new values.
- send signals to the application, for example `SIGHUP`, so that it reloads its configuration.
- run commands for reload or hot reload.
- restart processes when critical changes occur.

### Example: renewing certificates for NGINX

```hcl
template {
  source      = "/etc/nginx/ssl/cert.ctmpl"
  destination = "/etc/nginx/ssl/cert.pem"
  command     = "nginx -s reload"
}
```

In this example, Stronghold Agent renews the certificate and then runs `nginx -s reload` so that the service reloads its configuration without downtime.

## Retrieving dynamic secrets

Stronghold Agent is well suited for scenarios where secrets are issued not as permanent values but dynamically and for a limited time.

For example:
- temporary database credentials.
- TLS certificates with a limited validity period.
- other short-lived credentials.

### Temporary database credentials

Stronghold supports dynamic credential generation for different DBMSs. The source materials list the following examples:

- PostgreSQL.
- MySQL.
- MongoDB.
- Oracle.

In this scenario, Stronghold Agent can retrieve temporary usernames and passwords, write them to a file or pass them through environment variables, and then renew them automatically.

### PKI certificates

Stronghold Agent can also be used to work with certificates:

- automatically retrieve TLS certificates.
- renew them before they expire.
- inject new certificates into application files.

This is especially useful for web servers and internal services that need short-lived certificates.

## Using Stronghold Agent in CI/CD pipelines

Stronghold Agent can be used on a self-hosted CI/CD runner, for example in Jenkins or GitLab Runner.

In this scenario, the Agent runs on the runner server and provides secrets locally — through files or other local mechanisms.

A typical workflow is as follows:

1. Install Stronghold Agent on the CI/CD runner server.
1. Configure authentication in Stronghold, for example through AppRole.
1. Configure secret rendering into files or local secret delivery.
1. Use these secrets in the pipeline.

### Common CI/CD scenarios

Stronghold Agent can be used for:

- SSH keys and `kubeconfig` for deployment.
- credentials for access to a container registry.
- credentials for cloud providers.

### Example: GitLab Runner and an SSH key

```yaml
deploy:
  script:
    - ssh -i /var/run/stronghold-agent/deploy_key deploy@server.example.com "cd /app && git pull && systemctl restart app"
```

The role of Stronghold Agent in this scenario is as follows:

- the Agent retrieves the SSH key from Stronghold.
- renders it into the `/var/run/stronghold-agent/deploy_key` file with `0600` permissions.
- automatically updates the file when the key is rotated.
- GitLab Runner uses this file as a regular SSH key.

## Practical recommendations

To choose an appropriate Stronghold Agent use case, use the following guidelines:

- if the application runs on virtual machines or on bare metal, Stronghold Agent is often the simplest integration method.
- if the application reads configuration from files, use template rendering.
- if the application reads settings from environment variables, use `env_template` together with `exec`.
- if secrets change frequently, think in advance about how the application will apply updates — through reload, a signal, or a restart.
- if the application has no SDK or source code, Stronghold Agent makes it possible to work around this limitation by using standard operating system mechanisms.
