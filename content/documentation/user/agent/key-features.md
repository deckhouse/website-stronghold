---
title: "Key features"
description: "Stronghold Agent features: template rendering, Process Supervisor mode, Auto-Auth, API Proxy, and caching, with step-by-step examples."
weight: 30
---

## Templating

Templating lets you create configuration files filled with secrets from Stronghold, using the [Consul Template](https://github.com/hashicorp/consul-template) template language to render files.

There are two template modes:

1. `template` (render to a file): Agent generates or updates a file on disk (for example, `application.properties`, `nginx.conf`, `*.pem`) and, if needed, runs a command (`command`) to reload the service.
1. `env_template` + `exec` (render to environment variables and start a process): Agent builds environment variable values and starts the application as a child process (`exec`). When secrets change, the process can be restarted.

How Agent works:

1. Reads a template file with placeholders.
1. Requests secrets from Stronghold.
1. Renders the final file, substituting the real values.
1. Saves the file with the specified permissions.
1. (Optional) Runs a command to reload the application.

When to use:

- Legacy applications that read configuration files.
- Applications without Stronghold/Vault API support.
- When secrets need to be delivered in standard formats (.properties, .conf, .ini, .yaml).
- Dynamic database credentials.
- PKI certificates.

### Template syntax

Basic structure:

```go
{{ with secret "path/to/secret" }}
  {{ .Data.field_name }}
{{ end }}
```

For KV v2 (`secret/data/...`):

```go
{{ with secret "secret/data/myapp" }}
username = {{ .Data.data.username }}
password = {{ .Data.data.password }}
{{ end }}
```

For dynamic secrets (database, PKI):

```go
{{ with secret "database/creds/myapp" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
{{ end }}
```

Main functions:

| Function | Description                | Example |
|---------|----------------------------|--------|
| `secret` | Get a secret               | `{{ with secret "secret/data/myapp" }}{{ .Data.data.password }}{{ end }}` |
| `base64Encode` | Encode to base64           | `{{ "password" \| base64Encode }}` |
| `base64Decode` | Decode from base64         | `{{ .Data.cert \| base64Decode }}` |
| `toJSON` | Convert to JSON            | `{{ .Data \| toJSON }}` |
| `toYAML` | Convert to YAML            | `{{ .Data \| toYAML }}` |
| `toLower` / `toUpper` | Change case                | `{{ .Data.name \| toUpper }}` |
| `trim` | Remove whitespace          | `{{ .Data.value \| trim }}` |
| `range` | Iterate over an array      | `{{ range .Items }}{{ .Name }}{{ end }}` |
| `env` | Get an environment variable | `{{ env "HOME" }}` |
| `timestamp` | Get the current time       | `{{ timestamp "2006-01-02 15:04:05" }}` |

### Configuring templating: a step-by-step example

**Scenario:** a legacy Java application reads database credentials from `application.properties`.

Step 1: Store the secrets in Stronghold.

```bash
# Create a static secret.
d8 stronghold kv put secret/myapp/config \
  db_host=postgres.prod.example.com \
  db_port=5432 \
  db_name=production \
  db_user=app_user \
  db_password=SecureP@ssw0rd
```

Step 2: Create a template file.

Create `/etc/myapp/templates/application.properties.ctmpl`:

```text
# Database Configuration.
{{ with secret "secret/data/myapp/config" }}
spring.datasource.url=jdbc:postgresql://{{ .Data.data.db_host }}:{{ .Data.data.db_port }}/{{ .Data.data.db_name }}
spring.datasource.username={{ .Data.data.db_user }}
spring.datasource.password={{ .Data.data.db_password }}
{{ end }}

# Connection pool.
spring.datasource.hikari.maximum-pool-size=10
spring.datasource.hikari.minimum-idle=5
```

Step 3: Configure Agent.

Create `/etc/stronghold-agent/agent.hcl`:

```hcl
# Connection to Stronghold.
stronghold {
  address = "https://stronghold.example.com:8200"
}

# Auto-Auth with AppRole.
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

# Templating block.
template {
  # Path to the template.
  source      = "/etc/myapp/templates/application.properties.ctmpl"
  
  # Path to the final file.
  destination = "/etc/myapp/application.properties"
  
  # Permissions (0600 or 0400 is required for secrets).
  perms       = "0600"
  
  # File owner (optional).
  user        = "myapp"
  group       = "myapp"
  
  # Command to reload the application after a change.
  command     = "systemctl reload myapp"
  
  # Command timeout.
  command_timeout = "30s"
  
  # Run the command only when the content changes
  # (not on every lease renewal).
  wait {
    min = "2s"
    max = "10s"
  }
  
  # Fail if a key is missing.
  error_on_missing_key = true
}
```

Step 4: Start Agent.

```bash
# Validate the configuration with a trial run:
# Agent reads the config, authenticates, creates the file, and exits immediately.
stronghold agent -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

What happens when you run the command:

1. Agent reads and parses the configuration (`agent.hcl`).
1. Connects to the Stronghold server.
1. Authenticates (AppRole: reads role-id and secret-id).
1. Gets a token and saves it to the sink (`/var/run/stronghold-agent/token`).
1. Requests secrets from Stronghold.
1. Renders the template and creates the file (`/etc/myapp/application.properties`).
1. **Exits with code 0** (success) thanks to the `-exit-after-auth` flag.

Checking the result:

```bash

# 1. Check that the token has been created.
ls -la /var/run/stronghold-agent/token
# There should be a file with a recent date.

# 2. Check that the target file has been created.
ls -la /etc/myapp/application.properties
# There should be a file with 0600 permissions.

# 3. Check the content (careful: it contains passwords!).
sudo cat /etc/myapp/application.properties
# It should contain the real secret values, not {{ ... }}.
```

After a successful check, run Agent as a systemd service:

```bash
systemctl start stronghold-agent
systemctl status stronghold-agent

# Check the logs.
journalctl -u stronghold-agent -f
```

### Advanced scenarios

#### Example 1: Dynamic database credentials

```hcl
# Template: /etc/myapp/db-config.conf.ctmpl
{{ with secret "database/creds/myapp-role" }}
# Auto-generated credentials (TTL: 1h)
# Rotation: automatic
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
DB_LEASE_ID={{ .LeaseID }}
DB_LEASE_DURATION={{ .LeaseDuration }}
{{ end }}
```

Agent automatically:

- Requests temporary credentials.
- Updates the file on rotation (before the TTL expires).
- Runs the application reload command.

#### Example 2: PKI certificates

```hcl
# Template: /etc/nginx/ssl/cert.pem.ctmpl
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.certificate }}
{{ .Data.ca_chain }}
{{ end }}

# Template: /etc/nginx/ssl/key.pem.ctmpl
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.private_key }}
{{ end }}
```

Agent configuration:

```hcl
template {
  source      = "/etc/nginx/ssl/cert.pem.ctmpl"
  destination = "/etc/nginx/ssl/cert.pem"
  perms       = "0644"
}

template {
  source      = "/etc/nginx/ssl/key.pem.ctmpl"
  destination = "/etc/nginx/ssl/key.pem"
  perms       = "0600"
  command     = "systemctl reload nginx"
}
```

#### Example 3: Conditional logic and loops

```go
# Template with conditions
{{ with secret "secret/data/myapp/config" }}
{{ if eq .Data.data.environment "production" }}
LOG_LEVEL=ERROR
DEBUG_MODE=false
{{ else }}
LOG_LEVEL=DEBUG
DEBUG_MODE=true
{{ end }}

API_KEY={{ .Data.data.api_key }}
{{ end }}

# Loop over a list
{{ with secret "secret/data/myapp/allowed-ips" }}
{{ range $index, $ip := .Data.data.ips }}
allow {{ $ip }};
{{ end }}
{{ end }}
```

#### Example 4: Multiple secrets in one file

```go
# Database credentials
{{ with secret "database/creds/app" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
{{ end }}

# API Keys
{{ with secret "secret/data/myapp/api-keys" }}
STRIPE_KEY={{ .Data.data.stripe_key }}
SENDGRID_KEY={{ .Data.data.sendgrid_key }}
{{ end }}

# Redis credentials
{{ with secret "secret/data/myapp/redis" }}
REDIS_HOST={{ .Data.data.host }}
REDIS_PASSWORD={{ .Data.data.password }}
{{ end }}
```

### Important template block parameters

| Parameter | Description | Example |
|----------|----------|--------|
| `source` | Path to the template file | `/etc/app/template.ctmpl` |
| `destination` | Path to the resulting file | `/etc/app/config.conf` |
| `perms` | Permissions (octal) | `"0600"`, `"0644"` |
| `user` | File owner | `"myapp"` |
| `group` | File group | `"myapp"` |
| `command` | Command to run after rendering | `"systemctl reload app"` |
| `command_timeout` | Command timeout | `"30s"` |
| `error_on_missing_key` | Fail if a key is missing | `true` / `false` |
| `wait.min` | Minimum time between updates | `"2s"` |
| `wait.max` | Maximum time between updates | `"10s"` |
| `backup` | Create a backup before overwriting | `true` / `false` |

### template vs env_template: the difference and how to choose

`template` renders secrets **to a file on disk**.

- Use `template` if the application or service reads its configuration from files: `.conf/.ini/.yaml/.properties`, TLS `*.pem`, keys, certificates, and so on.
- `template` offers the full set of file options: `destination`, `perms`, `user/group`, `backup`, `wait`, and `command` to reload the service after a change (for example, `systemctl reload nginx`).

`env_template` + `exec` starts the application as a child process, and secrets are passed **in the process environment variables**.

- Use `env_template` if the application reads its configuration from environment variables (12-factor style) and a restart on secret rotation is acceptable.
- In Stronghold Agent, each `env_template` sets the value of **exactly one** environment variable and is always written as `env_template "VAR_NAME" { ... }`.
- Important: `env_template` does **not** create a `.env` file. The `destination/perms/command/wait/...` fields are not supported for `env_template` (they are available only in `template`).

Practical notes:

- If you use `template` and run Agent under systemd with hardening (`ProtectSystem=strict`), make sure `ReadWritePaths` includes the `template.destination` directory (otherwise rendering fails because writing is forbidden).
- If you use `env_template` to start a Docker container, pass the environment variables to `docker run` explicitly via `--env VAR_NAME` (or in another way); otherwise, they remain only in the environment of Agent itself.

## Process Supervisor mode

Process Supervisor mode lets Agent start the application as a child process and inject secrets directly into environment variables.

In this mode, Agent:

1. Starts as the parent process.
1. Requests secrets from Stronghold.
1. Builds environment variables from the template.
1. Starts the application as a child process with these variables.
1. Tracks secret changes.
1. When secrets change, restarts the application with the new values.

Mode limitations:

- `exec` must be used together with at least one `env_template`.
- `env_template` **cannot** be combined with `template` and `api_proxy` in one config (these are different operating modes).
- `env_template` is always defined as `env_template "VAR_NAME" { ... }` and sets the value of exactly one environment variable.

Benefits:

- Secrets are **never written to disk**.
- Automatic restart when secrets are updated.
- Secret isolation at the process level.
- Suitable for 12-factor applications.
- Easy migration of legacy applications to environment variables.

When to use:

- Applications that read configuration from environment variables.
- High security requirements (secrets must not touch the disk).
- Containerized applications on VMs.
- Dynamic credentials with frequent rotation.
- Development and testing.

### Configuring Process Supervisor: a step-by-step example

Scenario: a Java Spring Boot application reads secrets from environment variables.

Step 1: Prepare the application.

The Spring Boot application must read its configuration from environment variables:

The `application.properties` configuration:

```text
# application.properties - uses environment variables.
server.port=8080

# Database - values are taken from the environment.
spring.datasource.url=${DB_URL}
spring.datasource.username=${DB_USERNAME}
spring.datasource.password=${DB_PASSWORD}
spring.datasource.driver-class-name=org.postgresql.Driver

# JPA.
spring.jpa.hibernate.ddl-auto=validate
spring.jpa.properties.hibernate.dialect=org.hibernate.dialect.PostgreSQLDialect

# API Key.
api.key=${API_KEY}
```

Step 2: Store the secrets in Stronghold.

```bash
# Configure the database secrets engine for dynamic credentials.
d8 stronghold write database/config/postgresql \
  plugin_name=postgresql-database-plugin \
  allowed_roles="myapp-role" \
  connection_url="postgresql://{{username}}:{{password}}@postgres.prod:5432/myapp?sslmode=require" \
  username="vault_admin" \
  password="admin_password"

# Create a role for the application.
d8 stronghold write database/roles/myapp-role \
  db_name=postgresql \
  creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
  default_ttl="1h" \
  max_ttl="24h"

# Static secrets (API keys).
d8 stronghold kv put secret/myapp/config \
  api_key=sk_live_1234567890abcdef
```

Step 3: Configure Agent in Supervisor mode.

Create `/etc/stronghold-agent/agent.hcl`:

```hcl
# Connection to Stronghold.
stronghold {
  address = "https://stronghold.example.com:8200"
}

# Auto-Auth with AppRole
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

# Process Supervisor: start Spring Boot.
exec {
  # Command to start the application.
  command = ["/usr/bin/java", "-jar", "/opt/myapp/demo-application.jar"]
  
  # Restart when secrets change (database credential rotation).
  restart_on_secret_changes = "always"
  
  # Signal used to stop the process (SIGTERM by default).
  restart_stop_signal = "SIGTERM"
}

# Environment variable templates.
# Important: `env_template` is defined as `env_template "VARIABLE_NAME" { ... }`.
# Each block sets the value of one environment variable that is passed to the child process.
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

env_template "JAVA_OPTS" {
  contents = "-Xmx2g -Xms512m -XX:+UseG1GC"
}

env_template "SPRING_PROFILES_ACTIVE" {
  contents = "production"
}
```

Step 4: Start Agent.

```bash
# Agent command; it starts the Java application with secrets in the environment.
stronghold agent -config=/etc/stronghold-agent/agent.hcl
```

Agent automatically:

- Authenticates in Stronghold.
- Gets database credentials and API keys.
- Starts the Java application with secrets in environment variables.
- On credential rotation, restarts the application with the new values.

### Examples for different programming languages

#### Go application

```hcl
exec {
  command = ["/opt/myapp/myapp-server"]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}

env_template "DB_HOST" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_host }}{{ end }}"
}
env_template "DB_PORT" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_port }}{{ end }}"
}
env_template "DB_NAME" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_name }}{{ end }}"
}
env_template "DB_USER" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_user }}{{ end }}"
}
env_template "DB_PASSWORD" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.db_password }}{{ end }}"
}
env_template "API_KEY" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.api_key }}{{ end }}"
}
env_template "LOG_LEVEL" {
  contents = "info"
}
```

#### Docker container on a VM

```hcl
exec {
  command = [
    "/usr/bin/docker", "run", "--rm",
    "--name", "myapp",
    "-p", "8080:8080",
    # Pass environment variables from the Stronghold Agent environment into the container:
    "--env", "DOCKER_ENV_API_KEY",
    "--env", "DOCKER_ENV_DATABASE_URL",
    "myapp:latest"
  ]
  restart_on_secret_changes = "always"
  restart_stop_signal = "SIGTERM"
}

env_template "DOCKER_ENV_API_KEY" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.api_key }}{{ end }}"
}

env_template "DOCKER_ENV_DATABASE_URL" {
  contents = "{{ with secret \"secret/data/myapp/config\" }}{{ .Data.data.database_url }}{{ end }}"
}
```

### Important exec block parameters

| Parameter | Description | Default value |
|----------|----------|---------------------|
| `command` | Command that starts the application (array) | (required) |
| `restart_on_secret_changes` | Restart when secrets change: `never`, `always` | `always` |
| `restart_stop_signal` | Signal used to stop the process | `SIGTERM` |

### Important env_template block parameters

| Parameter | Description | Example |
|----------|----------|--------|
| `contents` | Inline environment variable template | `<<-EOT ... EOT` |
| `source` | Path to a template file (alternative to `contents`) | `"/etc/app/env.ctmpl"` |
| `error_on_missing_key` | Fail if a key is missing | `true` / `false` |

> Note: in Stronghold Agent, the `env_template` block always has an environment variable name: `env_template "MY_VAR" { ... }`.
> The `destination/perms/command/wait/...` fields are **not supported** in `env_template` (they are available only in a regular `template` block).

### Process lifecycle management

When secrets change (for example, when database credentials are rotated), Agent:

1. Gets the new secrets.
1. Builds new environment variables.
1. Sends `SIGTERM` to the child process (the application must handle `SIGTERM` correctly).
1. Restarts the process with the updated variables.

## Token caching and rotation

Token caching:

- The token is cached after authentication.
- The cached token is used for all requests.
- The load on the Stronghold server is reduced.

Token renewal:

Agent receives a token with a limited lifetime (for example, 1 hour) and renews it in advance. If renewal is not possible, Agent authenticates again.
This lets the application run continuously without manual intervention.

Lease renewal:

Dynamic secrets (for example, database credentials) also have a lifetime.
Agent automatically renews them before they expire, then updates configuration files and can reload the application.
This way, credentials are always up to date, and the application does not fail because of expired passwords.

## API Proxy

Agent can act as a proxy for the Stronghold API:

Capabilities:

- A local HTTP(S) endpoint for applications.
- Automatic addition of the authentication token.
- Response caching (optional).
- Reduced network load.

Configuration:

```hcl
api_proxy {
  use_auto_auth_token = true
}

listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
}
```

How the application uses it:

```bash
# The application calls the local Agent.
curl http://127.0.0.1:8200/v1/secret/data/myapp

# Agent automatically adds the token and proxies the request to the Stronghold server.
```

## Auto-Auth

Auto-Auth is a key Stronghold Agent feature that fully automates obtaining and renewing the authentication token.

How it works:

1. Agent starts with a configured authentication method.
1. It automatically authenticates in Stronghold.
1. It gets a token and uses it for its own operations (templating, API proxy).
1. If a sink is configured, it writes the token to a file for use by other processes. A sink is a file where Agent writes the obtained token.
1. It automatically renews the token before the TTL expires.
1. It re-authenticates when necessary.

Configuring a sink is optional:

- If a sink is configured, the token is written to a file (for example, `/var/run/stronghold-agent/token`) that other processes can read.
- If no sink is configured, the token is used by Agent only for internal operations (templating, caching).

Supported authentication methods:

- **AppRole**: recommended for VMs and bare metal.
- **Token**: for simple scenarios.
- **JWT/OIDC**: for integration with identity providers.
- **Cloud providers**.

## AppRole (recommended for VMs and bare metal)

AppRole is an authentication method designed for machines and applications.

Concept:

- **Role ID**: the role identifier (similar to a username).
- **Secret ID**: the secret identifier (similar to a password).
- Both IDs are required for authentication.

Benefits:

- Separation of duties (Role ID and Secret ID are delivered via different channels).
- Flexible policy configuration.
- Support for CIDR restrictions.
- A Secret ID can be single-use.

Configuration on the Stronghold server:

```bash
# Enable AppRole.
d8 stronghold auth enable approle

# Create a role.
d8 stronghold write auth/approle/role/myapp \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="myapp-policy"

# Get the Role ID.
d8 stronghold read auth/approle/role/myapp/role-id

# Create a Secret ID.
d8 stronghold write -f auth/approle/role/myapp/secret-id
```

Agent configuration:

```hcl
auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
      remove_secret_id_file_after_reading = true
    }
  }
}
```

Delivering and storing credentials:

Role ID:

- **What it is:** a public role identifier; it is not a secret.
- **How it is delivered:** via configuration management (Ansible, Puppet), in the VM image, or manually.
- **Where it is stored:** the path is set in the Agent configuration with the `role_id_file_path` parameter.
  - Typical path: `/etc/stronghold-agent/role-id`.
  - Permissions: 0640, owner: stronghold-agent.
  - You can use any path of your choice.
- **Deletion:** it is NOT deleted after use; it is reused on re-authentication.
- **Security:** it can be stored in a git repository; compromising it is not critical (it is useless without a Secret ID).

Secret ID:

- **What it is:** a secret identifier, similar to a password.
- **How it is delivered:**
  - Manually by an administrator on the first server start.
  - Via a secure SSH connection.
  - Via an internal self-service portal (if available).
  - Via an encrypted variable in the CI/CD system.
  - It must NOT be in git or configuration management.
- **Where it is stored:** the path is set in the Agent configuration with the `secret_id_file_path` parameter.
  - Typical path: `/etc/stronghold-agent/secret-id`.
  - Permissions: 0640, owner: stronghold-agent.
  - You can use any path of your choice.
- **Deletion:** it can be deleted after use (the `remove_secret_id_file_after_reading = true` parameter in the Agent config).
- **Security:** a **critical** secret that must be protected.

Secret ID types:

The `secret_id_num_uses` and `secret_id_ttl` parameters are set on the Stronghold server when creating the role or generating a Secret ID. Agent simply uses an already created Secret ID.

1. **Single-use (num_uses=1):**

   ```bash
   # On the Stronghold server when creating the role:
   d8 stronghold write auth/approle/role/myapp \
     secret_id_num_uses=1 \
     policies="myapp-policy"
   
   # Or when generating a specific Secret ID:
   d8 stronghold write -f auth/approle/role/myapp/secret-id num_uses=1
   ```

   Features:

   - Used for authentication only once.
   - Becomes invalid after use.
   - The most secure option for production.
   - Agent must have `remove_secret_id_file_after_reading = true` in the configuration file.

1. **Multi-use (num_uses=0):**

   ```bash
   # On the Stronghold server:
   d8 stronghold write auth/approle/role/myapp \
     secret_id_num_uses=0 \
     policies="myapp-policy"
   ```

   Features:

   - Can be used many times.
   - Convenient for testing and development.
   - Less secure (if compromised, manual rotation is required).

1. **With a limited TTL:**

   ```bash
   # On the Stronghold server:
   d8 stronghold write auth/approle/role/myapp \
     secret_id_ttl=24h \
     policies="myapp-policy"
   ```

   Features:

   - Expires after the specified time (24 hours in the example).
   - A balance between security and convenience.
   - After the TTL expires, a new Secret ID is required.

A complete example of configuring and delivering credentials:

```bash
# STEP 1: Configuration on the Stronghold server (performed by an administrator).

# Enable the AppRole authentication method.
d8 stronghold auth enable approle

# Create an access policy for the application.
d8 stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF

# Create an AppRole role with settings.
d8 stronghold write auth/approle/role/myapp \
  token_ttl=1h \                    # Token lifetime (renewed automatically).
  token_max_ttl=4h \                # Maximum token lifetime (re-authentication is required afterwards).
  policies="myapp-policy" \         # Access policy (which secrets it can read).
  secret_id_num_uses=1 \            # Number of Secret ID uses (1 = single-use).
  secret_id_ttl=24h                 # Secret ID lifetime (expires in 24 hours).

# Get the Role ID.
d8 stronghold read auth/approle/role/myapp/role-id
# Output: role_id    abc123-def456-ghi789.

# Generate a single-use Secret ID.
d8 stronghold write -f auth/approle/role/myapp/secret-id
# Output: secret_id    xyz789-abc123-def456.

# STEP 2: Delivering credentials to the target server.

# Create a directory on the target server.
ssh root@app-server.example.com << 'ENDSSH'
  mkdir -p /etc/stronghold-agent
  chown root:stronghold-agent /etc/stronghold-agent
  chmod 750 /etc/stronghold-agent
ENDSSH

# Deliver the Role ID (via automation or manually).
ssh root@app-server.example.com << 'ENDSSH'
  echo -n "abc123-def456-ghi789" > /etc/stronghold-agent/role-id
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/role-id
  chmod 0640 /etc/stronghold-agent/role-id
ENDSSH

# Deliver the Secret ID via a secure channel (single-use).
ssh root@app-server.example.com << 'ENDSSH'
  echo -n "xyz789-abc123-def456" > /etc/stronghold-agent/secret-id
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/secret-id
  chmod 0640 /etc/stronghold-agent/secret-id
ENDSSH

# STEP 3: Configuring Agent on the target server.

# Create the configuration file.
cat > /etc/stronghold-agent/agent.hcl <<EOF
stronghold {
  address = "https://stronghold.example.com:8200"
}

auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
      remove_secret_id_file_after_reading = true
    }
  }
  
  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
}
EOF

# Set permissions on the configuration.
chown root:stronghold-agent /etc/stronghold-agent/agent.hcl
chmod 0640 /etc/stronghold-agent/agent.hcl

# STEP 4: Starting Agent.

# Start Agent.
systemctl start stronghold-agent

# Check that authentication succeeded.
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"

# Check that the token exists.
ls -la /var/run/stronghold-agent/token

# Check that the Secret ID has been deleted (if remove_secret_id_file_after_reading = true).
ls -la /etc/stronghold-agent/secret-id
# Expected: No such file or directory.

# STEP 5: Further operation.

# On restart, Agent uses only the Role ID.
# The token is renewed automatically every ~59 minutes (1 minute before the TTL expires).
# The Secret ID is no longer required.
```

**Recommendations:**

- Use single-use Secret IDs in production.
- Separate delivery: Role ID via automation, Secret ID via a secure channel.
- Restrict access by CIDR (`secret_id_bound_cidrs`).
- Log all Secret ID uses in Stronghold for auditing.

## Token (for simple scenarios)

Using a token directly is the simplest authentication method. Agent reads a ready-made token from a file and uses it.

Concept:

- An administrator creates the token in advance on the Stronghold server.
- The token is delivered to the target server in any convenient way.
- Agent simply reads the token from the file and uses it.

When to use:

- Test environments and development.
- Temporary installations.
- Scenarios where AppRole cannot be used.
- Simple cases without strict security requirements.

Drawbacks:

- Less secure (a token is a long-lived credential).
- No separation of duties (unlike AppRole).
- If compromised, manual rotation is required.
- Not recommended for production environments.

A complete configuration example:

```bash
# STEP 1: Creating a token on the Stronghold server

# Create an access policy.
d8 stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF

# Create a token with parameters.
d8 stronghold token create \
  -policy=myapp-policy \
  -ttl=720h \                  # Lifetime: 30 days.
  -renewable=true \            # Can be renewed.
  -display-name="myapp-agent" \
  -format=json

# Output:
# {
#   "auth": {
#     "client_token": "hvs.CAES...xyz123",
#     "policies": ["default", "myapp-policy"],
#     "renewable": true,
#     "lease_duration": 2592000
#   }
# }

# Save the token.
export AGENT_TOKEN="hvs.CAES...xyz123"

# STEP 2: Delivering the token to the target server

# Create a directory.
ssh root@app-server.example.com << 'ENDSSH'
  mkdir -p /etc/stronghold-agent
  chown root:stronghold-agent /etc/stronghold-agent
  chmod 750 /etc/stronghold-agent
ENDSSH

# Deliver the token via a secure channel.
echo -n "$AGENT_TOKEN" | ssh root@app-server.example.com 'cat > /etc/stronghold-agent/token'
ssh root@app-server.example.com << 'ENDSSH'
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/token
  chmod 0640 /etc/stronghold-agent/token
ENDSSH

# STEP 3: Configuring Agent

cat > /etc/stronghold-agent/agent.hcl <<EOF
stronghold {
  address = "https://stronghold.example.com:8200"
}

auto_auth {
  method "token_file" {
    config = {
      token_file_path = "/etc/stronghold-agent/token"
    }
  }
  
  # A sink is optional for the token method.
  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
}

# Template example.
template {
  source = "/etc/stronghold-agent/templates/database.conf.ctmpl"
  destination = "/etc/myapp/database.conf"
  perms = "0600"
}
EOF

chown root:stronghold-agent /etc/stronghold-agent/agent.hcl
chmod 0640 /etc/stronghold-agent/agent.hcl

# STEP 4: Starting Agent

systemctl start stronghold-agent

# Check that it works.
journalctl -u stronghold-agent -n 50

# The token is renewed automatically until max_ttl expires.
```

Important token parameters:

- **ttl**: the initial token lifetime.
- **renewable**: whether the token can be renewed (must be true for Agent).
- **period**: if set, the token is renewed for this period (for example, with `period=24h`, the token is renewed every 24 hours).
- **explicit-max-ttl**: the absolute maximum lifetime (after it, the token can no longer be renewed).

Security recommendations:

1. Use tokens with `renewable=true` for automatic renewal.
1. Set a reasonable TTL (for example, 30 days).
1. Configure explicit-max-ttl to limit the total lifetime.
1. Regularly review and revoke unused tokens.
1. Store the token with minimal permissions (0640).
1. For production, consider using AppRole instead of a token.

## JWT/OIDC (for integration with identity providers)

JWT/OIDC authentication lets you use your existing identity management infrastructure to authenticate in Stronghold.

Concept:

- The application gets a JWT from an identity provider (Keycloak, Azure AD, Google, etc.).
- The JWT contains claims about the user or service.
- Stronghold verifies the JWT signature and extracts the claims.
- Based on the claims, a Stronghold token with the corresponding policies is issued.

Use cases:

- Integration with corporate SSO (Single Sign-On).
- Using service accounts from the identity provider.
- Federated authentication between organizations.
- CI/CD integration via OIDC (GitHub Actions, GitLab CI).

Benefits:

- Centralized identity management.
- No need to create separate credentials for each application.
- Automatic JWT rotation by the identity provider.
- Support for MFA and other IdP features.

A complete configuration example (with Keycloak):

```bash
# STEP 1: Configuring the JWT auth method on the Stronghold server.

# Enable the JWT auth method
d8 stronghold auth enable jwt

# Configure the JWT method with Keycloak parameters.
d8 stronghold write auth/jwt/config \
  oidc_discovery_url="https://keycloak.example.com/realms/myrealm" \
  oidc_client_id="stronghold" \
  oidc_client_secret="client-secret-from-keycloak" \
  default_role="default"

# Create an access policy.
d8 stronghold policy write myapp-jwt-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF

# Create a role for JWT authentication.
d8 stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_subject="service-account-myapp" \
  token_ttl=1h \
  token_max_ttl=4h \
  token_policies="myapp-jwt-policy"

# Role parameters:
# - bound_audiences: which audiences must be present in the JWT.
# - user_claim: which claim to use as the username.
# - bound_subject: a specific subject value (optional).
# - bound_claims: additional requirements for claims.

# Example with more complex conditions:
d8 stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_claims='{"environment":"production","app":"myapp"}' \
  claim_mappings='{"department":"dept"}' \
  token_policies="myapp-jwt-policy"

# STEP 2: Getting a JWT from the identity provider.

# Example 1: Keycloak service account.
curl -X POST "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=myapp-service" \
  -d "client_secret=service-secret" \
  -d "grant_type=client_credentials" \
  | jq -r '.access_token' > /tmp/jwt-token.txt

# Example 2: GitHub Actions OIDC token.
# In a GitHub Actions workflow:
# - uses: actions/checkout@v3
# - name: Get OIDC token
#   run: |
#     curl -H "Authorization: bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
#          "$ACTIONS_ID_TOKEN_REQUEST_URL" | jq -r '.value' > jwt-token.txt

# STEP 3: Delivering the JWT to the target server.

# Create a directory.
ssh root@app-server.example.com << 'ENDSSH'
  mkdir -p /etc/stronghold-agent
  chown root:stronghold-agent /etc/stronghold-agent
  chmod 750 /etc/stronghold-agent
ENDSSH

# Deliver the JWT.
scp /tmp/jwt-token.txt root@app-server.example.com:/etc/stronghold-agent/jwt-token
ssh root@app-server.example.com << 'ENDSSH'
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/jwt-token
  chmod 0640 /etc/stronghold-agent/jwt-token
ENDSSH

# STEP 4: Configuring Agent

cat > /etc/stronghold-agent/agent.hcl <<EOF
stronghold {
  address = "https://stronghold.example.com:8200"
}

auto_auth {
  method "jwt" {
    mount_path = "auth/jwt"
    config = {
      path = "/etc/stronghold-agent/jwt-token"
      role = "myapp-role"
    }
  }
  
  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
}

template {
  source = "/etc/stronghold-agent/templates/database.conf.ctmpl"
  destination = "/etc/myapp/database.conf"
  perms = "0600"
}
EOF

chown root:stronghold-agent /etc/stronghold-agent/agent.hcl
chmod 0640 /etc/stronghold-agent/agent.hcl

# STEP 5: Starting Agent.

systemctl start stronghold-agent

# Check that authentication succeeded.
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"

# Check the Stronghold token.
cat /var/run/stronghold-agent/token
```

JWT method specifics:

1. **JWT vs Stronghold token:**
   - The JWT is a credential from the identity provider (short-lived, usually 5–60 minutes).
   - After authentication, Agent gets a Stronghold token.
   - The Stronghold token is renewed automatically (as with AppRole).
   - Agent does NOT renew the JWT automatically.

1. **Periodic JWT refresh:**
   - When the JWT expires, you need to get a new one from the IdP.
   - You can configure a cron job for periodic refresh:

      ```bash
      # /etc/cron.d/refresh-jwt
      */30 * * * * stronghold-agent /usr/local/bin/refresh-jwt-token.sh
      ```

Checking the JWT:

```bash
# Decode the JWT to check its claims.
cat /etc/stronghold-agent/jwt-token | cut -d. -f2 | base64 -d | jq

# The output shows the claims:
# {
#   "sub": "service-account-myapp",
#   "aud": "stronghold",
#   "iss": "https://keycloak.example.com/realms/myrealm",
#   "exp": 1234567890,
#   "iat": 1234567800
# }
```

Recommendations:

1. Use a short TTL for JWTs (5–15 minutes).
1. Configure bound_audiences to protect against token reuse.
1. Use bound_subject or bound_claims for strict validation.
1. In production, use OIDC discovery (automatic key refresh).
1. Log all authentications for auditing.
