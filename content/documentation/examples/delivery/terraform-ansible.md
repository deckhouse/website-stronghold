---
title: "Terraform and Ansible"
description: "Managing Stronghold configuration as code with the hashicorp/vault Terraform provider and reading secrets in Ansible with the community.hashi_vault collection."
weight: 50
---

The `hashicorp/vault` Terraform provider and the `community.hashi_vault` Ansible collection work with the HashiCorp Vault API. Stronghold is API-compatible with Vault, so these tools can be pointed at Stronghold by setting its address. <!-- TODO(verify): tested versions of the hashicorp/vault provider and the community.hashi_vault collection with Stronghold -->

## Terraform

The `hashicorp/vault` provider (also works in OpenTofu) lets you describe Stronghold mounts, policies, auth methods, and roles as code.

### Connecting to Stronghold

Create a dedicated AppRole role for Terraform with a policy sufficient to manage the required paths (`sys/mounts/*`, `sys/policies/acl/*`, `sys/auth/*`, `auth/*`). Example provider configuration:

```hcl
terraform {
  required_providers {
    vault = {
      source = "hashicorp/vault"
    }
  }
}

variable "role_id" {
  type = string
}

variable "secret_id" {
  type      = string
  sensitive = true
}

provider "vault" {
  address = "https://stronghold.example.com"

  auth_login {
    path = "auth/approle/login"
    parameters = {
      role_id   = var.role_id
      secret_id = var.secret_id
    }
  }
}
```

Alternative authentication options:

- a token from the `VAULT_TOKEN` environment variable (address from `VAULT_ADDR`) if no `auth_login` block is set;
- `auth_login_jwt` for runs from CI/CD with a job OIDC token (see [CI/CD](../ci-cd/)).

By default, the provider creates a short-lived child token, so the policy needs `update` on `auth/token/create`. If this is undesirable, set `skip_child_token = true` in the `provider` block.

To work in a [namespace](../../../admin/namespaces/overview/), set the `namespace` parameter in the `provider` block or in individual resources.

### Example: KV, policy, and Kubernetes role

```hcl
resource "vault_mount" "kv" {
  path        = "secret"
  type        = "kv"
  options     = { version = "2" }
  description = "KV v2 for applications"
}

resource "vault_policy" "myapp_read" {
  name   = "myapp-read"
  policy = <<-EOT
    path "secret/data/myapp/*" {
      capabilities = ["read"]
    }
  EOT
}

resource "vault_auth_backend" "kubernetes" {
  type = "kubernetes"
}

resource "vault_kubernetes_auth_backend_config" "this" {
  backend         = vault_auth_backend.kubernetes.path
  kubernetes_host = "https://kubernetes.example.com:6443"
}

resource "vault_kubernetes_auth_backend_role" "myapp" {
  backend                          = vault_auth_backend.kubernetes.path
  role_name                        = "myapp"
  bound_service_account_names      = ["myapp"]
  bound_service_account_namespaces = ["myapp"]
  token_policies                   = [vault_policy.myapp_read.name]
  token_ttl                        = 3600
}
```

Apply:

```bash
terraform init
terraform plan -out=tfplan
terraform apply tfplan
```

{{< alert level="warning" >}}
Secret values written by resources (for example, `vault_kv_secret_v2`) or read by data sources are stored in the Terraform state file in plaintext. Keep the state in a protected, encrypted backend with restricted access. Prefer writing secret values to Stronghold outside Terraform.
{{< /alert >}}

To manage Stronghold configuration from Git without Terraform, you can use the built-in [GitOps secrets engine](../../../user/secrets-engines/gitops/overview/).

## Ansible

The `community.hashi_vault` collection provides lookup plugins and modules for reading secrets.

1. Install the collection and the `hvac` library:

   ```bash
   ansible-galaxy collection install community.hashi_vault
   pip install hvac
   ```

1. Set the Stronghold address and authentication parameters via environment variables:

   ```bash
   export ANSIBLE_HASHI_VAULT_ADDR=https://stronghold.example.com
   export ANSIBLE_HASHI_VAULT_AUTH_METHOD=approle
   export ANSIBLE_HASHI_VAULT_ROLE_ID=<role_id>
   export ANSIBLE_HASHI_VAULT_SECRET_ID=<secret_id>
   ```

1. Read a KV version 2 secret in a playbook:

   ```yaml
   - name: Configure application
     hosts: app
     tasks:
       - name: Read database credentials from Stronghold
         ansible.builtin.set_fact:
           db: "{{ lookup('community.hashi_vault.vault_kv2_get', 'myapp/db', engine_mount_point='secret').secret }}"
         no_log: true

       - name: Render application config
         ansible.builtin.template:
           src: app.conf.j2
           dest: /etc/myapp/app.conf
           mode: "0600"
         no_log: true
   ```

To get dynamic credentials, use the `community.hashi_vault.vault_read` lookup, for example for the `database/creds/my-role` path ([database secrets engine](../../../user/secrets-engines/databases/overview/)).

Recommendations:

- Set `no_log: true` for tasks that handle secrets.
- Use a dedicated AppRole role for Ansible with read-only access to the required paths and a short `token_ttl`.
- Do not store `secret_id` in the inventory. Pass it via CI/CD environment variables or get it using [response wrapping](../../../concepts/response-wrapping/).
