---
title: "Basic settings"
description: "Structure of the Stronghold Agent HCL configuration file and the main sections: stronghold, auto_auth, template, exec, listener, and logging."
weight: 40
---

## Configuration file structure

The Stronghold Agent configuration is written in HCL format:

```hcl
# Connection to the Stronghold server.
stronghold {
  address = "https://stronghold.example.com:8200"
  ca_cert = "/etc/stronghold-agent/ca.pem"
  
  # Retry settings.
  retry {
    num_retries = 5
  }
}

# Automatic authentication.
auto_auth {
  method "approle" {
    # ... method configuration.
  }
  
  sink "file" {
    # ... sink configuration.
  }
}

# API Proxy (optional).
api_proxy {
  use_auto_auth_token = true
}

# Caching (optional).
cache {
  use_auto_auth_token = true
}

# Listener for API Proxy.
listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
}

# Templates.
template {
  source      = "/path/to/template.ctmpl"
  destination = "/path/to/output"
  # ... additional settings.
}

# PID file.
pid_file = "/var/run/stronghold-agent.pid"

# Logging.
log_level = "info"
log_file = "/var/log/stronghold-agent.log"
```

## The vault/stronghold section

To connect to a Stronghold server, use the `stronghold` section name. If you need integration with HashiCorp Vault, use `vault` as the section name.

```hcl
stronghold {
  # Server URL (required).
  address = "https://stronghold.example.com:8200"
  
  # TLS settings.
  ca_cert = "/etc/stronghold-agent/ca.pem"              # CA certificate.
  ca_path = "/etc/stronghold-agent/ca-bundle/"          # Directory with CA certificates.
  client_cert = "/etc/stronghold-agent/client.pem"      # Client certificate.
  client_key = "/etc/stronghold-agent/client-key.pem"   # Client key.
  tls_skip_verify = false                         # Do not disable in production!
  tls_server_name = "stronghold.example.com"      # SNI name.
  
  # Retry policy.
  retry {
    num_retries = 5  # Number of retries.
  }
}
```

## The auto_auth section

Configuring automatic authentication:

```hcl
auto_auth {
  # Authentication method.
  method "approle" {
    mount_path = "auth/approle"  # Mount path of the auth method
    namespace  = "myns"           # Namespace (optional)
    
    config = {
      # Method-specific parameters.
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
    }
  }
  
  # Sink: where to store the token (optional, there can be several).
  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
  
  # Sink with encryption.
  sink "file" {
    wrap_ttl = "5m"                    # Wrap the token with a TTL
    aad_env_var = "VAULT_AAD"          # Additional authenticated data for encryption
    dh_type = "curve25519"             # Diffie-Hellman type
    dh_path = "/etc/stronghold-agent/dh-pub" # Public key
    
    config = {
      path = "/var/run/stronghold-agent/encrypted-token"
    }
  }
}
```

## The template section

Configuring template rendering:

```hcl
template {
  # Path to the template file (required).
  source = "/etc/myapp/config.ctmpl"
  
  # Path to the output file (required).
  destination = "/etc/myapp/config.conf"
  
  # File permissions.
  perms = "0600"
  
  # User and group.
  user = "myapp"
  group = "myapp"
  
  # Back up the file before replacing it.
  backup = true
  
  # Command to run after rendering.
  command = "systemctl reload myapp"
  command_timeout = "30s"
  
  # Wait before rendering.
  wait {
    min = "5s"
    max = "10s"
  }
  
  # Error handling.
  error_on_missing_key = true
  
  # Create missing parent directories of destination.
  create_dest_dirs = true
}
```

## The template_config section

Global settings for all templates:

```hcl
template_config {
  # Exit Agent if template rendering fails after all retries are exhausted.
  exit_on_retry_failure = false
  
  # Interval for periodically rendering "static" secrets (for example, KV).
  # Can be set as a duration string (for example, "5m") or a number of seconds.
  static_secret_render_interval = "5m"
}
```

### The exec section (Process Supervisor)

Running a child process with injected secrets:

```hcl
exec {
  # Command to run (required).
  command = ["/usr/bin/myapp", "--config", "/etc/myapp/config.yaml"]
  
  # Restart policy when secrets change.
  restart_on_secret_changes = "always"  # always, never
  
  # Signal used to stop the process.
  restart_stop_signal = "SIGTERM"
}

# Template for environment variables.
env_template "DATABASE_URL" {
  contents = "{{ with secret \"secret/data/myapp\" }}postgresql://{{ .Data.data.username }}:{{ .Data.data.password }}@db:5432{{ end }}"
  error_on_missing_key = true
}

env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp\" }}{{ .Data.data.api_key }}{{ end }}"
  error_on_missing_key = true
}
```

### The listener section

Configuring the HTTP(S) listener for API Proxy:

```hcl
listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
  
  # TLS settings.
  tls_cert_file = "/etc/stronghold-agent/agent-cert.pem"
  tls_key_file = "/etc/stronghold-agent/agent-key.pem"
  
  # Require a special request header.
  require_request_header = true
  
  # API settings.
  agent_api {
    enable_quit = true  # Enable the /agent/v1/quit endpoint
  }
}

# Unix socket listener
listener "unix" {
  address = "/var/run/stronghold-agent.sock"
  tls_disable = true
  socket_mode = "0660"
  socket_user = "myapp"
  socket_group = "myapp"
}
```

## Logging and debugging

```hcl
# Log level: trace, debug, info, warn, error.
log_level = "info"

# Log file.
log_file = "/var/log/stronghold-agent.log"

# Log format: standard, json.
log_format = "json"

# Log rotation.
log_rotate_duration = "24h"
log_rotate_bytes = 104857600  # 100MB
log_rotate_max_files = 10
```
