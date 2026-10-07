---
title: "Application on a virtual machine with Stronghold Agent"
linkTitle: "Application on a VM"
description: "Delivering secrets to an application on a virtual machine: Stronghold Agent under systemd, AppRole authentication with a wrapped secret_id, template rendering, and application reload."
weight: 20
params:
  relatedLinks:
    - title: "Stronghold Agent"
      url: ../../../user/agent/overview/
    - title: "Agent settings"
      url: ../../../user/agent/settings/
    - title: "Running and managing Agent"
      url: ../../../user/agent/launch-and-control/
    - title: "AppRole auth method"
      url: ../../../user/auth/approle/
    - title: "Response wrapping"
      url: ../../../concepts/response-wrapping/
---

An application that reads passwords from a configuration file and cannot call the Stronghold API can be connected without code changes. Stronghold Agent authenticates to Stronghold, renders the configuration file from a template, and reloads the application when secrets change.

## Goal

Set up Stronghold Agent under systemd on a virtual machine. Agent logs in with AppRole using a one-time wrapped `secret_id`, builds the `application.properties` file, and runs `systemctl reload myapp` when secrets change.

## Prerequisites

- A Linux virtual machine with systemd and the `myapp` application running as the `myapp.service` service.
- The Stronghold binary on the VM (`/usr/local/bin/stronghold` in the examples) and network access to Stronghold.
- A Stronghold token with permissions to configure AppRole, policies, and KV.
- KV version 2 enabled at the `secret` path.

## Step 1. Store the secret and create a policy

```bash
d8 stronghold kv put -mount=secret myapp/config \
  db_user=app_user \
  db_password='S3cure-P@ss'

d8 stronghold policy write myapp-vm - <<'POLICY'
path "secret/data/myapp/config" {
  capabilities = ["read"]
}
POLICY
```

## Step 2. Create an AppRole role

```bash
d8 stronghold auth enable approle

d8 stronghold write auth/approle/role/myapp-vm \
  token_policies=myapp-vm \
  token_ttl=1h \
  token_max_ttl=24h \
  secret_id_ttl=720h \
  secret_id_bound_cidrs="10.0.10.15/32"
```

The `secret_id_bound_cidrs` parameter allows login only from the VM address. `secret_id_ttl` limits how long Agent can log in again with one `secret_id`.

## Step 3. Deliver role_id and a wrapped secret_id to the VM

1. Get the `role_id`. It is not a secret and can be distributed by a configuration management system:

   ```bash
   d8 stronghold read -field=role_id auth/approle/role/myapp-vm/role-id > role-id
   ```

1. Issue a wrapped `secret_id`. Instead of the `secret_id` itself, a one-time token with a 10-minute TTL is returned:

   ```bash
   d8 stronghold write -wrap-ttl=10m -field=wrapping_token -f \
     auth/approle/role/myapp-vm/secret-id > secret-id
   ```

1. Copy both files to the VM into `/etc/stronghold-agent/` and restrict access:

   ```bash
   sudo install -o stronghold-agent -g stronghold-agent -m 0640 role-id /etc/stronghold-agent/role-id
   sudo install -o stronghold-agent -g stronghold-agent -m 0600 secret-id /etc/stronghold-agent/secret-id
   ```

If the wrapping token is intercepted and unwrapped by someone else, Agent cannot unwrap it, which indicates an incident. The verification procedure is described in [Response wrapping](../../../concepts/response-wrapping/#verifying-a-wrapping-token).

## Step 4. Create a template

Create the `/etc/myapp/templates/application.properties.ctmpl` file:

```text
{{ with secret "secret/data/myapp/config" }}
spring.datasource.username={{ .Data.data.db_user }}
spring.datasource.password={{ .Data.data.db_password }}
{{ end }}
```

## Step 5. Configure Agent

Create the `/etc/stronghold-agent/agent.hcl` file:

```hcl
stronghold {
  address = "https://stronghold.example.com"
}

auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path                   = "/etc/stronghold-agent/role-id"
      secret_id_file_path                 = "/etc/stronghold-agent/secret-id"
      secret_id_response_wrapping_path    = "auth/approle/role/myapp-vm/secret-id"
      remove_secret_id_file_after_reading = true
    }
  }

  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0600
    }
  }
}

template_config {
  static_secret_render_interval = "5m"
}

template {
  source          = "/etc/myapp/templates/application.properties.ctmpl"
  destination     = "/etc/myapp/application.properties"
  perms           = "0640"
  command         = "systemctl reload myapp"
  command_timeout = "30s"
  error_on_missing_key = true
}
```

- `secret_id_response_wrapping_path`: Agent recognizes the wrapped `secret_id`, checks its creation path, and unwraps it.
- `remove_secret_id_file_after_reading = true`: the `secret_id` file is deleted after reading.
- `static_secret_render_interval`: how often Agent re-reads static KV secrets.

The unwrapped `secret_id` is kept only in Agent memory. After Agent or the VM restarts, or when `secret_id_ttl` expires, Agent needs a new wrapped `secret_id`. Build issuing and delivering `secret_id` into your deployment process, for example in CI/CD.

## Step 6. Run Agent under systemd

1. Create the `/etc/systemd/system/stronghold-agent.service` unit file following the example in [Running and managing Agent](../../../user/agent/launch-and-control/). In `ReadWritePaths`, list the `/var/run/stronghold-agent` and `/etc/myapp` directories.

1. Allow the `stronghold-agent` user to run only the application reload, for example through sudoers:

   ```text
   stronghold-agent ALL=(root) NOPASSWD: /usr/bin/systemctl reload myapp
   ```

   In this case, set `command = "sudo /usr/bin/systemctl reload myapp"` in the template.

1. Start the service:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now stronghold-agent
   ```

## Verification

1. Check the Agent log: it must contain messages about successful authentication and rendering:

   ```bash
   sudo journalctl -u stronghold-agent -n 50
   ```

1. Make sure the secret file is deleted and the configuration is built:

   ```bash
   ls -la /etc/stronghold-agent/secret-id /etc/myapp/application.properties
   ```

1. Change the secret and wait for the file to be re-rendered and the application to reload (no longer than `static_secret_render_interval`):

   ```bash
   d8 stronghold kv patch -mount=secret myapp/config db_password='N3w-P@ss'
   sudo journalctl -u myapp -f
   ```

## Cleanup

1. Stop and disable Agent: `sudo systemctl disable --now stronghold-agent`.
1. Delete the `/etc/stronghold-agent/role-id` and `/etc/myapp/application.properties` files and the token file in `/var/run/stronghold-agent/`.
1. Delete the role and the policy:

   ```bash
   d8 stronghold delete auth/approle/role/myapp-vm
   d8 stronghold policy delete myapp-vm
   ```
