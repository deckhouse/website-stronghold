---
title: "Terraform и Ansible"
description: "Управление конфигурацией Stronghold как кодом через Terraform-провайдер hashicorp/vault и получение секретов в Ansible через коллекцию community.hashi_vault."
weight: 50
---

Terraform-провайдер `hashicorp/vault` и коллекция Ansible `community.hashi_vault` работают с API HashiCorp Vault. Stronghold совместим с API Vault, поэтому эти инструменты можно направить на Stronghold, указав его адрес. <!-- TODO(verify): протестированные версии провайдера hashicorp/vault и коллекции community.hashi_vault со Stronghold -->

## Terraform

Провайдер `hashicorp/vault` (также работает в OpenTofu) позволяет описывать точки монтирования, политики, методы аутентификации и роли Stronghold как код.

### Подключение к Stronghold

Создайте для Terraform отдельную роль AppRole с политикой, достаточной для управления нужными путями (`sys/mounts/*`, `sys/policies/acl/*`, `sys/auth/*`, `auth/*`). Пример конфигурации провайдера:

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

Альтернативные способы аутентификации:

- токен из переменной окружения `VAULT_TOKEN` (адрес — из `VAULT_ADDR`), если блок `auth_login` не указан;
- `auth_login_jwt` — для запуска из CI/CD с OIDC-токеном задания (см. [«CI/CD»](../ci-cd/)).

По умолчанию провайдер создаёт дочерний токен с коротким сроком жизни, поэтому политике нужно право `update` на `auth/token/create`. Если это нежелательно, укажите в блоке `provider` параметр `skip_child_token = true`.

Для работы в [пространстве имён](../../../admin/namespaces/overview/) укажите параметр `namespace` в блоке `provider` или в отдельных ресурсах.

### Пример: KV, политика и роль Kubernetes

```hcl
resource "vault_mount" "kv" {
  path        = "secret"
  type        = "kv"
  options     = { version = "2" }
  description = "KV v2 для приложений"
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

Применение:

```bash
terraform init
terraform plan -out=tfplan
terraform apply tfplan
```

{{< alert level="warning" >}}
Значения секретов, записанные ресурсами (например, `vault_kv_secret_v2`) или прочитанные источниками данных, сохраняются в state-файле Terraform в открытом виде. Храните state в защищённом бэкенде с шифрованием и ограниченным доступом. Значения секретов предпочтительно записывать в Stronghold вне Terraform.
{{< /alert >}}

Для управления конфигурацией Stronghold из Git без Terraform можно использовать встроенный [механизм секретов GitOps](../../../user/secrets-engines/gitops/overview/).

## Ansible

Коллекция `community.hashi_vault` содержит lookup-плагины и модули для чтения секретов.

1. Установите коллекцию и библиотеку `hvac`:

   ```bash
   ansible-galaxy collection install community.hashi_vault
   pip install hvac
   ```

1. Задайте адрес Stronghold и параметры аутентификации через переменные окружения:

   ```bash
   export ANSIBLE_HASHI_VAULT_ADDR=https://stronghold.example.com
   export ANSIBLE_HASHI_VAULT_AUTH_METHOD=approle
   export ANSIBLE_HASHI_VAULT_ROLE_ID=<role_id>
   export ANSIBLE_HASHI_VAULT_SECRET_ID=<secret_id>
   ```

1. Прочитайте секрет KV версии 2 в плейбуке:

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

Для получения динамических учётных данных используйте lookup `community.hashi_vault.vault_read`, например для пути `database/creds/my-role` ([механизм секретов баз данных](../../../user/secrets-engines/databases/overview/)).

Рекомендации:

- Указывайте `no_log: true` для задач, работающих с секретами.
- Используйте для Ansible отдельную роль AppRole с правами только на чтение нужных путей и коротким `token_ttl`.
- Не храните `secret_id` в инвентаре. Передавайте его через переменные окружения CI/CD или получайте с помощью [обертывания ответа](../../../concepts/response-wrapping/).
