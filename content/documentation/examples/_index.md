---
title: "Usage examples"
linkTitle: "Examples"
weight: 65
description: "Catalog of practical Stronghold examples: sign-in via ALD Pro, AD, and SSO, delivering secrets to Kubernetes, CI/CD, and GitOps, dynamic credentials, PKI, encryption, migration from Vault, and operations."
---

This section collects ready-made examples for typical Stronghold tasks. Every example follows the same structure: the task, what you need, step-by-step setup, verification, and links to reference pages that describe all parameters.

Commands are given for Stronghold in Deckhouse Platform (`d8 stronghold ...`). For Stronghold on Linux, use the `stronghold` command with the same arguments. Features available only in Stronghold EE are marked in the catalog and on the example pages. For a comparison of editions, see [Editions](../about/editions/).

## Access and sign-in

| Example | What it shows |
| --- | --- |
| [Integration with ALD Pro](./access/ald-pro/) | Signing in to Stronghold with ALD Pro domain accounts: the LDAP auth method with LDAPS and group-to-policy mapping, Kerberos single sign-on, service account password rotation with the LDAP secrets engine, and common errors. |
| [Integration with Active Directory](./access/active-directory/) | Signing in to Stronghold with Active Directory accounts through the LDAP method with LDAPS, nested groups, and UPN, and rotating AD service account passwords with the LDAP secrets engine and the ad schema. |
| [Single sign-on through a corporate IdP (OIDC)](./access/sso-oidc/) | How to set up single sign-on to Stronghold through a corporate identity provider over OIDC: the general flow, a complete Keycloak example with group-to-policy mapping, login through Dex in DP, and links to provider guides. |
| [MFA for administrators](./access/admin-mfa/) | Mandatory TOTP second factor for administrator logins to Stronghold: a TOTP method, a login enforcement rule for the administrators group, QR code enrollment, login in the CLI and web UI, recovery after a lost device, and audit log monitoring. |
| [Break-glass access](./access/break-glass-access/) | Preparing emergency access to Stronghold for when the identity provider is unavailable: a backup account in a sealed envelope, the generate-root procedure, usage auditing, and rotation after use. |

## Delivering secrets to applications

| Example | What it shows |
| --- | --- |
| [Delivering secrets to Kubernetes pods](./delivery/kubernetes-workloads/) | Comparison of ways to deliver Stronghold secrets to Kubernetes pods: CSI, env-injector, Stronghold Agent, External Secrets Operator, and direct API calls. |
| [Application on a virtual machine with Stronghold Agent](./delivery/legacy-app-on-vm/) | Delivering secrets to an application on a virtual machine: Stronghold Agent under systemd, AppRole authentication with a wrapped secret_id, template rendering, and application reload. |
| [CI/CD](./delivery/ci-cd/) | Getting Stronghold secrets in GitLab CI, GitHub Actions, and Jenkins without static tokens: JWT/OIDC and AppRole. |
| [GitOps](./delivery/gitops/) | Keeping secrets out of Git with Argo CD and Flux: External Secrets Operator, argocd-vault-plugin, SOPS, and the GitOps secrets engine. |
| [Terraform and Ansible](./delivery/terraform-ansible/) | Managing Stronghold configuration as code with the hashicorp/vault Terraform provider and reading secrets in Ansible with the community.hashi_vault collection. |
| [TLS certificate for a web server on a VM from Stronghold PKI](./delivery/web-server-tls/) | Automatic issuance and renewal of an Nginx TLS certificate on a virtual machine: a PKI role, a policy, a Stronghold Agent template with pki/issue, and an Nginx reload after renewal; notes for Apache. |
| [Application clients](./delivery/app-clients/) | Using Stronghold from application code: environment variables, Go, Python, Java (Spring Cloud Vault), Node.js, reading KV, dynamic credentials, and renewing tokens and leases. |
| [Local Stronghold for development](./delivery/docker-local/) | Running Stronghold in Docker on a developer workstation: server -dev mode and a single-node Raft configuration in Docker Compose. |
| [AppRole for an application and scripts](./delivery/approle-app/) | Logging an application in without a user through AppRole: a role limited by network, token lifetime, and secret_id usage count, obtaining a token, and reading a secret. |

## Dynamic credentials

| Example | What it shows |
| --- | --- |
| [Dynamic PostgreSQL credentials for an application in DP](./dynamic-credentials/postgresql/) | Issuing short-lived PostgreSQL credentials to an application in Deckhouse Platform with the database secrets engine and the Kubernetes auth method. |
| [Dynamic MySQL and MariaDB credentials for an application](./dynamic-credentials/mysql/) | Issuing short-lived MySQL or MariaDB users with limited privileges to an application with the database secrets engine: connection, a role with GRANT, a policy, renewal, revocation, and verification with the mysql client. |
| [Dynamic ClickHouse credentials for an application](./dynamic-credentials/clickhouse/) | Issuing short-lived ClickHouse users with a predefined role to an application with the database secrets engine: connection, a role, a policy, renewal, revocation, and verification with clickhouse-client. |
| [Dynamic RabbitMQ users for a service](./dynamic-credentials/rabbitmq/) | Issuing short-lived RabbitMQ users limited to their own virtual host and queues to a service with the rabbitmq secrets engine: connection, lease, role, policy, and verification with rabbitmqctl. |
| [Temporary Kubernetes access with ServiceAccount tokens](./dynamic-credentials/kubernetes-tokens/) | Issuing short-lived ServiceAccount tokens to CI jobs and engineers with the kubernetes secrets engine: roles with allowed namespaces, generated RBAC rules, getting a token, and verification with d8 k auth can-i. |
| [Rotating service account passwords](./dynamic-credentials/static-credentials-rotation/) | Automatic password rotation for existing service accounts in databases and LDAP with static roles and service account libraries. |
| [One-time TOTP codes for an application](./dynamic-credentials/totp-codes/) | The TOTP secrets engine: a key for an application user, getting and validating a code, a minimal policy, and protection against code reuse. |

## Certificates

| Example | What it shows |
| --- | --- |
| [Internal PKI with Stronghold](./certificates/internal-pki/) | Building a two-tier PKI: a root CA outside Stronghold or in a separate mount, an intermediate CA in Stronghold, roles, certificate issuance, CRL, OCSP, cert-manager, and ACME. |
| [cert-manager](./certificates/cert-manager/) | Issuing TLS certificates in Kubernetes with cert-manager using a vault-type Issuer, the Stronghold PKI secrets engine, and Kubernetes auth. |
| [Automatic certificates with ACME](./certificates/acme/) | Automatic issuance and renewal of TLS certificates for internal servers through the ACME server of the Stronghold PKI secrets engine: config/cluster, config/acme, EAB, certbot, lego, and Caddy. |
| [SSH access with signed certificates](./certificates/ssh-certificates-access/) | Setting up SSH access to servers with short-lived certificates that users obtain from Stronghold after an OIDC login. |
| [Service mesh](./certificates/service-mesh/) | Using Stronghold PKI as an external certificate authority for Istio via cert-manager and istio-csr. |

## Encryption and keys

| Example | What it shows |
| --- | --- |
| [Encrypting database fields with Transit](./encryption/transit-encryption/) | Encrypting individual fields in an application database with the Transit engine, envelope encryption with datakey, key rotation, rewrap, and min_decryption_version. |
| [Auto-unseal with Yandex Cloud KMS](./encryption/yandex-kms-unseal/) | A standalone Stronghold server with auto-unseal through Yandex Cloud KMS: key and service account in yc CLI, the seal yandexcloudkms stanza, initialization with recovery keys, migration from Shamir, and verification after a restart. |
| [Auto-unseal and seal wrap with an HSM](./encryption/hsm-unseal/) (Stronghold EE) | Protecting the Stronghold EE root key with a hardware module over PKCS #11: the seal pkcs11 stanza, a SoftHSM2 lab, initialization with recovery keys, double encryption, and recommendations for a production HSM. |

## Operations

| Example | What it shows |
| --- | --- |
| [Migrating from HashiCorp Vault](./operations/migration-from-vault/) | Moving secrets, policies, auth methods, PKI, and dynamic secrets from HashiCorp Vault to Stronghold, with a client cutover and rollback plan. |
| [Multi-tenancy with namespaces](./operations/multi-tenancy-namespaces/) | Giving each team a separate namespace with delegated administration, its own auth methods, and templated policies. |
| [Shipping audit logs to SIEM](./operations/siem/) | Sending Stronghold audit logs to SIEM systems: audit devices, collection with log-shipper, Vector, and Fluent Bit, and field mapping for ELK, OpenSearch, Splunk, MaxPatrol SIEM, and KUMA. |
| [One-time secret handover with response wrapping](./operations/secret-sharing-wrapping/) | Securely handing a password or key to another person once with response wrapping: -wrap-ttl, token lookup, and unwrap. |
| [Secret versions in KV v2](./operations/kv-v2-versioning/) | The lifecycle of a secret in KV version 2: new versions, patch, rollback, history limits, check-and-set, deletion, recovery, and destruction of versions. |
| [Request rate limiting](./operations/rate-limit-quota/) | A rate limit quota on a path: protection from overload by a single client, checking the 429 response, and viewing, changing, and deleting the quota. |
| [Private token storage (cubbyhole)](./operations/cubbyhole-private-storage/) | Data isolation in cubbyhole: every token has its own storage that other tokens cannot access and that is deleted together with the token. |
| [Regular disaster recovery drills](./operations/dr-drill/) | Running regular drills: restoring a snapshot into an isolated cluster, a test promotion of a DR secondary, a checklist, and frequency. |

## Stronghold as an IdP

| Example | What it shows |
| --- | --- |
| [Stronghold as an OIDC provider for internal applications](./identity-provider/oidc-provider-for-apps/) | Setting up single sign-on for internal applications with the Stronghold OIDC provider: signing key, assignments, templated scopes, client, and provider. |

## How to add an example

If you did not find the scenario you need, ask in the ["Deckhouse | EN-community"](https://t.me/deckhouse) Telegram channel. To contribute an example, create a page in the matching subsection and follow the structure:

1. Goal — which task the example solves.
1. What you need — edition, environment (DP or Standalone), access, and external systems.
1. Steps — commands and configuration to run in order.
1. Verification — how to confirm that everything works.
1. Troubleshooting and cleanup of test resources.
1. Links to reference pages in `params.relatedLinks`.

Add a row for the new example to the table above and a link to the example in `params.relatedLinks` of the reference pages it uses.
