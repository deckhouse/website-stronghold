---
title: "Basic settings"
linkTitle: "Configuration"
description: "Basic settings for Stronghold Agent"
weight: 40
---

Stronghold Agent configuration is defined in HCL format. Using the configuration file, you configure the connection to the Stronghold server, automatic authentication, template rendering, child process execution, API Proxy, and logging parameters.
This page helps you understand the structure of the configuration file and the purpose of its main sections.

## Configuration file structure

The following example shows the general structure of a Stronghold Agent configuration:

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"
  ca_cert = "/etc/stronghold-agent/ca.pem"
  retry {
    num_retries = 5
  }
}
auto_auth {
  method "approle" {
    # ... method configuration.
  }
  sink "file" {
    # ... sink configuration.
  }
}
api_proxy {
  use_auto_auth_token = true
}
cache {
  use_auto_auth_token = true
}
listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
}
template {
  source      = "/path/to/template.ctmpl"
  destination = "/path/to/output"
}
pid_file = "/var/run/stronghold-agent.pid"
log_level = "info"
log_file = "/var/log/stronghold-agent.log"
```

The following sections are typically used in the configuration:
- `stronghold` — connection to the Stronghold server;
- `auto_auth` — automatic authentication and token sink;
- `api_proxy` — local proxy for the Stronghold API;
- `cache` — caching;
- `listener` — local listener for API Proxy;
- `template` — template rendering;
- `template_config` — global template parameters;
- `exec` — child process execution;
- `env_template` — passing secrets through environment variables;
- `pid_file`, `log_level`, `log_file`, and other parameters — process and logging management.

## stronghold section

Use the `stronghold` section to connect to the Stronghold server. If you need integration with HashiCorp Vault, use `vault` instead of `stronghold`.

Configuration example:

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"
  ca_cert = "/etc/stronghold-agent/ca.pem"
  ca_path = "/etc/stronghold-agent/ca-bundle/"
  client_cert = "/etc/stronghold-agent/client.pem"
  client_key = "/etc/stronghold-agent/client-key.pem"
  tls_skip_verify = false
  tls_server_name = "stronghold.example.com"
  retry {
    num_retries = 5
  }
}
```

You can configure the following parameters in this section:
- `address` — Stronghold server address;
- `ca_cert` — path to the CA certificate;
- `ca_path` — directory with a set of CA certificates;
- `client_cert` and `client_key` — client certificate and key;
- `tls_skip_verify` — disables TLS verification;
- `tls_server_name` — server name for SNI;
- `retry` — retry policy.

{{< alert level="warning" >}}
Do not disable `tls_skip_verify` in a production environment.
{{< /alert >}}

## auto_auth section

The `auto_auth` section is responsible for Stronghold Agent automatic authentication and for storing the received token.

Configuration example:

```hcl
auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    namespace  = "myns"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
    }
  }
  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
  sink "file" {
    wrap_ttl = "5m"
    aad_env_var = "VAULT_AAD"
    dh_type = "curve25519"
    dh_path = "/etc/stronghold-agent/dh-pub"
    config = {
      path = "/var/run/stronghold-agent/encrypted-token"
    }
  }
}
```

Configure the following in this section:
- `method` — authentication method;
- `mount_path` — auth method mount path;
- `namespace` — namespace, if used;
- `config` — parameters of the specific method;
- `sink` — location where the Agent stores the token.

A sink can be plain or use additional encryption.

## template section

The `template` section is used to render templates into files.

Configuration example:

```hcl
template {
  source = "/etc/myapp/config.ctmpl"
  destination = "/etc/myapp/config.conf"
  perms = "0600"
  user = "myapp"
  group = "myapp"
  backup = true
  command = "systemctl reload myapp"
  command_timeout = "30s"
  wait {
    min = "5s"
    max = "10s"
  }
  error_on_missing_key = true
  create_dest_dirs = true
}
```

You can configure the following parameters in this section:
- `source` — path to the template file;
- `destination` — path to the output file;
- `perms` — access permissions;
- `user` and `group` — owner and group;
- `backup` — creates a backup copy;
- `command` — command to run after rendering;
- `command_timeout` — command timeout;
- `wait` — delay before rendering;
- `error_on_missing_key` — fails rendering if a key is missing;
- `create_dest_dirs` — creates missing destination directories.

## template_config section

The `template_config` section defines global parameters for all templates.

Configuration example:

```hcl
template_config {
  exit_on_retry_failure = false
  static_secret_render_interval = "5m"
}
```

You can configure the following parameters in this section:
- `exit_on_retry_failure` — exits after failed retries;
- `static_secret_render_interval` — interval for periodic rendering of static secrets, for example from KV.

## exec section

The `exec` section is used in Process Supervisor mode, where Stronghold Agent starts a child process and passes secrets to it through environment variables.

Configuration example:

```hcl
exec {
  command = ["/usr/bin/myapp", "--config", "/etc/myapp/config.yaml"]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}
```

You can configure the following parameters in this section:
- `command` — application start command;
- `restart_on_secret_changes` — restarts the process when secrets change;
- `restart_stop_signal` — signal to stop the process.

### env_template section

`env_template` is usually used together with `exec`:

```hcl
env_template "DATABASE_URL" {
  contents = "{{ with secret \"secret/data/myapp\" }}postgresql://{{ .Data.data.username }}:{{ .Data.data.password }}@db:5432{{ end }}"
  error_on_missing_key = true
}
env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp\" }}{{ .Data.data.api_key }}{{ end }}"
  error_on_missing_key = true
}
```

Each `env_template` block forms the value of one environment variable.

Consider the following limitations:
- the block must always be written as `env_template "VAR_NAME" { ... }`;
- `env_template` does not create a `.env` file;
- fields such as `destination`, `perms`, `command`, `wait`, and similar parameters are not supported in `env_template`.

## listener section

The `listener` section configures an HTTP(S) listener for API Proxy.

TCP listener example:

```hcl
listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
  tls_cert_file = "/etc/stronghold-agent/agent-cert.pem"
  tls_key_file = "/etc/stronghold-agent/agent-key.pem"
  require_request_header = true
  agent_api {
    enable_quit = true
  }
}
```

Unix socket listener example:

```hcl
listener "unix" {
  address = "/var/run/stronghold-agent.sock"
  tls_disable = true
  socket_mode = "0660"
  socket_user = "myapp"
  socket_group = "myapp"
}
```

You can configure the following parameters in this section:
- `address` — TCP listener address or path to the Unix socket;
- `tls_disable` — disables TLS;
- `tls_cert_file` and `tls_key_file` — TLS certificate and key;
- `require_request_header` — requires a special header;
- `agent_api.enable_quit` — enables the `POST /agent/v1/quit` endpoint;
- `socket_mode`, `socket_user`, `socket_group` — Unix socket parameters.

## Logging and debugging

Stronghold Agent supports configuration of the logging level, log format, and log file rotation.

Configuration example:

```hcl
log_level = "info"
log_file = "/var/log/stronghold-agent.log"
log_format = "json"
log_rotate_duration = "24h"
log_rotate_bytes = 104857600
log_rotate_max_files = 10
```

You can configure the following parameters:
- `log_level` — logging level: `trace`, `debug`, `info`, `warn`, `error`;
- `log_file` — path to the log file;
- `log_format` — log format: `standard` or `json`;
- `log_rotate_duration` — rotation period;
- `log_rotate_bytes` — maximum file size;
- `log_rotate_max_files` — number of retained files.

## Practical recommendations

To simplify Stronghold Agent configuration, use the following recommendations:
- first determine which mode you need: `template` or `exec` with `env_template`;
- use TLS for a production environment and do not disable certificate verification;
- store the configuration file, tokens, `role-id`, and `secret-id` with the minimum required access permissions;
- first verify the basic connection and `auto_auth`, and then move on to templates and Process Supervisor;
- if you use `template`, verify write permissions for the destination directory in advance;
- if you use `listener`, restrict access to the local listener or Unix socket in advance.
