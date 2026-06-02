---
title: "Environment variables and Process Supervisor"
linkTitle: "Process Supervisor"
description: "Using Stronghold Agent to pass secrets through environment variables."
weight: 30
---

Stronghold Agent can pass secrets to an application through environment variables and manage the lifecycle of a child process. This is done with the `env_template` and `exec` blocks.

This mode is suitable for applications that read configuration from environment variables and must not receive secrets through files on disk.

## When to use this mode

Use `env_template` together with `exec` in the following cases:

- The application reads configuration from environment variables.
- Secrets must not be written to disk in plain text.
- Restarting the application when a secret changes is acceptable.
- Dynamic credentials are used with regular rotation.
- The application must be started by Stronghold Agent as a child process.

If the application needs configuration files, use the `template` mode. For details, see [Templates and file rendering](../templating/).

## How Process Supervisor works

In Process Supervisor mode, Stronghold Agent performs the following actions:

1. Authenticates to Stronghold.
1. Requests secrets specified in `env_template` blocks.
1. Creates environment variables.
1. Starts the application as a child process.
1. Tracks secret changes.
1. Restarts the child process when a secret is updated.

This mode lets you avoid writing secrets to files and pass them to the application only at the process level.

## env_template block

The `env_template` block sets the value of one environment variable. The variable name is specified in the block header.

Example:

```hcl
env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.api_key }}{{ end }}"
}
```

One `env_template` block creates only one environment variable. If the application needs several variables, create a separate block for each variable.

## exec block

The `exec` block defines the command that Stronghold Agent starts as a child process.

Example:

```hcl
exec {
  command = ["/usr/bin/java", "-jar", "/opt/myapp/demo-application.jar"]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}
```

The `exec` block is usually used together with one or more `env_template` blocks.

## Spring Boot example

In this example, Stronghold Agent passes database credentials and an API key to the application through environment variables.

### Step 1. Prepare the application

```text
server.port=8080
spring.datasource.url=${DB_URL}
spring.datasource.username=${DB_USERNAME}
spring.datasource.password=${DB_PASSWORD}
api.key=${API_KEY}
```

### Step 2. Prepare secrets in Stronghold

```shell
stronghold write database/config/postgresql \
  plugin_name=postgresql-database-plugin \
  allowed_roles="myapp-role" \
  connection_url="postgresql://{{username}}:{{password}}@postgres.prod:5432/myapp?sslmode=require" \
  username="vault_admin" \
  password="admin_password"

stronghold write database/roles/myapp-role \
  db_name=postgresql \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"

stronghold kv put secret/myapp/config \
  api_key=sk_live_1234567890abcdef
```

### Step 3. Configure Stronghold Agent

Create `/etc/stronghold-agent/agent.hcl`:

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"
}

auto_auth {
  method {
    type = "approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
      remove_secret_id_file_after_reading = false
    }
  }
}

exec {
  command = ["/usr/bin/java", "-jar", "/opt/myapp/demo-application.jar"]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}

env_template "DB_URL" {
  contents = "jdbc:postgresql://postgres.prod:5432/myapp"
}

env_template "DB_USERNAME" {
  contents = "{{ with secret \"database/creds/myapp-role\" }}{{ .Data.username }}{{ end }}"
}

env_template "DB_PASSWORD" {
  contents = "{{ with secret \"database/creds/myapp-role\" }}{{ .Data.password }}{{ end }}"
}

env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.api_key }}{{ end }}"
}
```

### Step 4. Start the Agent

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl
```

After startup, the Agent authenticates, obtains secrets, creates environment variables, starts the application, and restarts it when a secret changes.

## Process lifecycle management

When a secret changes, Stronghold Agent can restart the child process. The usual sequence is as follows:

1. The Agent obtains the new secret value.
1. The Agent creates the environment variables again.
1. The Agent sends a stop signal to the child process.
1. The Agent starts the process again with updated values.

By default, `SIGTERM` is usually used for stopping the process.

## Limitations

Consider the following specifics:

- The `exec` block must be used with at least one `env_template` block.
- Each `env_template` block defines only one environment variable.
- The `env_template` block does not create an `.env` file.
- The `destination`, `perms`, `command`, and `wait` parameters are not used for `env_template`.
- `env_template` cannot be combined with `template` and `api_proxy` in one configuration file.
- If the application starts through Docker Engine, environment variables must be passed explicitly with `--env`.

{{< alert level="info" >}}
The `env_template` block always contains the environment variable name in the block header, for example `env_template "MY_VAR" { ... }`.
{{< /alert >}}
