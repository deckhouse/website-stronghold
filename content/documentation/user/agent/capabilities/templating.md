---
title: "Templates and file rendering"
linkTitle: "Templates"
description: "Using Stronghold Agent to render secrets to files."
weight: 20
---

Stronghold Agent can render secrets from Stronghold into configuration files. This mode is suitable for applications that read settings from files rather than from environment variables.

This page describes how to use the `template` mode, which capabilities it supports, and when to choose it.

## When to use template mode

Use the `template` mode in the following cases:

- The application reads configuration from files.
- Secrets must be passed through `.conf`, `.ini`, `.yaml`, or `.properties` files.
- Certificates, keys, or other sensitive data must be stored in files.
- The application cannot work with the Stronghold API directly.
- Dynamic credentials must be renewed without manual intervention.

If the application reads secrets from environment variables, use `env_template` together with `exec`. For details, see [Environment variables and Process Supervisor](../process-supervisior/).

## How template rendering works

Stronghold Agent usually performs the following actions:

1. Reads a template file.
1. Requests secrets from Stronghold.
1. Substitutes values into the template.
1. Writes the result to the target file.
1. Runs a command to reload the service if needed.

If a secret value changes, the Agent renders the file again. If needed, it can also run the command specified in the `command` parameter again.

## Template syntax

Stronghold Agent uses Consul Template-style templates. A template usually requests a secret by path and extracts the required fields.

Basic template structure:

```go
{{ with secret "path/to/secret" }}
{{ .Data.field_name }}
{{ end }}
```

KV v2 example:

```go
{{ with secret "secret/data/myapp" }}
username={{ .Data.data.username }}
password={{ .Data.data.password }}
{{ end }}
```

Dynamic database secret example:

```go
{{ with secret "database/creds/myapp" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
{{ end }}
```

## Template functions

Templates can use the following functions:

| Function | Purpose | Example |
| --- | --- | --- |
| `secret` | Gets a secret | `{{ with secret "secret/data/myapp" }}{{ .Data.data.password }}{{ end }}` |
| `base64Encode` | Encodes a string in Base64 | `{{ "password" \| base64Encode }}` |
| `base64Decode` | Decodes a string from Base64 | `{{ .Data.cert \| base64Decode }}` |
| `toJSON` | Converts a value to JSON | `{{ .Data \| toJSON }}` |
| `toYAML` | Converts a value to YAML | `{{ .Data \| toYAML }}` |
| `toLower` | Converts a string to lowercase | `{{ .Data.name \| toLower }}` |
| `toUpper` | Converts a string to uppercase | `{{ .Data.name \| toUpper }}` |
| `trim` | Removes whitespace from both ends of a string | `{{ .Data.value \| trim }}` |
| `range` | Iterates over list items | `{{ range .Items }}{{ .Name }}{{ end }}` |
| `env` | Reads an environment variable | `{{ env "HOME" }}` |
| `timestamp` | Returns the current time | `{{ timestamp "2006-01-02 15:04:05" }}` |

## Rendering a configuration file

In this example, Stronghold Agent creates `application.properties` for a Java application.

### Step 1. Store secrets in Stronghold

```shell
stronghold kv put secret/myapp/config \
  db_host=postgres.prod.example.com \
  db_port=5432 \
  db_name=production \
  db_user=app_user \
  db_password=SecureP@ssw0rd
```

### Step 2. Create a template file

Create `/etc/myapp/templates/application.properties.ctmpl`:

```text
{{ with secret "secret/data/myapp/config" }}
spring.datasource.url=jdbc:postgresql://{{ .Data.data.db_host }}:{{ .Data.data.db_port }}/{{ .Data.data.db_name }}
spring.datasource.username={{ .Data.data.db_user }}
spring.datasource.password={{ .Data.data.db_password }}
{{ end }}
spring.datasource.hikari.maximum-pool-size=10
spring.datasource.hikari.minimum-idle=5
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

  sink {
    type = "file"
    config = {
      path = "/var/run/stronghold-agent/token"
    }
  }
}

template {
  source = "/etc/myapp/templates/application.properties.ctmpl"
  destination = "/etc/myapp/application.properties"
  perms = "0600"
  user = "myapp"
  group = "myapp"
  command = "systemctl reload myapp"
  command_timeout = "30s"

  wait {
    min = "2s"
    max = "10s"
  }

  error_on_missing_key = true
}
```

### Step 4. Check the configuration

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

During execution, the Agent reads `agent.hcl`, connects to Stronghold, authenticates through AppRole, obtains a token, writes it to the sink, requests secrets, renders the template, creates `/etc/myapp/application.properties`, and exits with code `0`.

### Check the result

Run the following commands:

```shell
ls -la /var/run/stronghold-agent/token
ls -la /etc/myapp/application.properties
sudo cat /etc/myapp/application.properties
```

After successful verification, you can start the Agent as a systemd service:

```shell
systemctl start stronghold-agent
systemctl status stronghold-agent
journalctl -u stronghold-agent -f
```

## Advanced scenarios

### Dynamic database credentials

A template can obtain temporary database credentials and write them to a file:

```go
{{ with secret "database/creds/myapp-role" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
DB_LEASE_ID={{ .LeaseID }}
DB_LEASE_DURATION={{ .LeaseDuration }}
{{ end }}
```

In this case, the Agent renders the file again when the secret is renewed. If needed, it can also restart or reload the service.

### PKI certificate rendering

Certificate template:

```go
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.certificate }}
{{ .Data.ca_chain }}
{{ end }}
```

Private key template:

```go
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.private_key }}
{{ end }}
```

Example configuration:

```hcl
template {
  source = "/etc/nginx/ssl/cert.pem.ctmpl"
  destination = "/etc/nginx/ssl/cert.pem"
  perms = "0644"
}

template {
  source = "/etc/nginx/ssl/key.pem.ctmpl"
  destination = "/etc/nginx/ssl/key.pem"
  perms = "0600"
  command = "systemctl reload nginx"
}
```

## template block parameters

The `template` block has the following main parameters:

| Parameter | Description | Example |
| --- | --- | --- |
| `source` | Path to the template file | `/etc/app/template.ctmpl` |
| `destination` | Path to the rendered file | `/etc/app/config.conf` |
| `perms` | File permissions | `"0600"` |
| `user` | File owner | `"myapp"` |
| `group` | File group | `"myapp"` |
| `command` | Command after rendering | `"systemctl reload app"` |
| `command_timeout` | Command timeout | `"30s"` |
| `error_on_missing_key` | Fails if a key is missing | `true` |
| `wait.min` | Minimum delay before update | `"2s"` |
| `wait.max` | Maximum delay before update | `"10s"` |
| `backup` | Creates a backup file | `true` |

## Limitations

Consider the following specifics of `template` mode:

- Secrets are written to disk as the rendered file.
- The application must have access to the destination file.
- If the Agent runs under systemd with `ProtectSystem=strict`, add the directory from `destination` to `ReadWritePaths`.
- If the template uses a missing key and `error_on_missing_key` is enabled, rendering fails.
