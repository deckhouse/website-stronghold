---
title: "Main capabilities"
linkTitle: "Capabilities"
description: "Main capabilities of Stronghold agent"
weight: 30
---

Stronghold Agent helps applications receive secrets from Stronghold without direct integration with the API. It supports several operating modes: rendering templates into files, passing secrets through environment variables, automatic authentication, token caching, secret renewal, and operation through a local API Proxy.

This page describes the key features of Stronghold Agent and typical usage scenarios.

## Templating

Templating allows you to create configuration files populated with secrets from Stronghold. To render files, Stronghold Agent uses the Consul Template template language.

There are two modes for working with templates:

1. `template` — rendering to a file. Agent creates or updates a file on disk, for example `application.properties`, `nginx.conf`, or `*.pem`, and, if needed, runs a command to reload a service.
1. `env_template` together with `exec` — rendering to environment variables and starting a process. Agent forms the values of environment variables and launches the application as a child process. If secrets change, the process can be restarted.

### How it works

Typically, Stronghold Agent performs the following actions:

1. Reads a template file with placeholders.
1. Requests secrets from Stronghold.
1. Substitutes actual values into the template.
1. Saves the resulting file with the required permissions.
1. If needed, runs a command to reload the application.

### When to use

Templating is suitable for the following scenarios:

- legacy applications that read configuration from files;
- applications without support for the Stronghold API;
- delivery of secrets into `.properties`, `.conf`, `.ini`, `.yaml`;
- working with dynamic credentials for databases;
- delivery of PKI certificates.

### Template syntax

Basic structure:

```go
{{ with secret "path/to/secret" }}
  {{ .Data.field_name }}
{{ end }}
```

For KV v2:

```go
{{ with secret "secret/data/myapp" }}
username = {{ .Data.data.username }}
password = {{ .Data.data.password }}
{{ end }}
```

For dynamic secrets, for example for a database or PKI:

```go
{{ with secret "database/creds/myapp" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
{{ end }}
```

### Main template functions

| Function | Description | Example |
| --- | --- | --- |
| `secret` | Retrieve a secret | `{{ with secret "secret/data/myapp" }}{{ .Data.data.password }}{{ end }}` |
| `base64Encode` | Encode to Base64 | `{{ "password" \| base64Encode }}` |
| `base64Decode` | Decode from Base64 | `{{ .Data.cert \| base64Decode }}` |
| `toJSON` | Convert to JSON | `{{ .Data \| toJSON }}` |
| `toYAML` | Convert to YAML | `{{ .Data \| toYAML }}` |
| `toLower` / `toUpper` | Change case | `{{ .Data.name \| toUpper }}` |
| `trim` | Remove spaces | `{{ .Data.value \| trim }}` |
| `range` | Iterate over an array | `{{ range .Items }}{{ .Name }}{{ end }}` |
| `env` | Get an environment variable | `{{ env "HOME" }}` |
| `timestamp` | Get the current time | `{{ timestamp "2006-01-02 15:04:05" }}` |

### Step-by-step example: rendering a configuration file

Scenario: a legacy Java application reads database credentials from `application.properties`.

#### Step 1. Save secrets in Stronghold

```shell
stronghold kv put secret/myapp/config \
  db_host=postgres.prod.example.com \
  db_port=5432 \
  db_name=production \
  db_user=app_user \
  db_password=SecureP@ssw0rd
```

#### Step 2. Create a template file

Create the `/etc/myapp/templates/application.properties.ctmpl` file:

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

#### Step 3. Configure Stronghold Agent

Create the `/etc/stronghold-agent/agent.hcl` file:

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
  source      = "/etc/myapp/templates/application.properties.ctmpl"
  destination = "/etc/myapp/application.properties"
  perms       = "0600"
  user        = "myapp"
  group       = "myapp"
  command     = "systemctl reload myapp"
  command_timeout = "30s"

  wait {
    min = "2s"
    max = "10s"
  }

  error_on_missing_key = true
}
```

#### Step 4. Check the configuration

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

During execution, Agent:

1. reads and parses `agent.hcl`;
1. connects to the Stronghold server;
1. authenticates through AppRole;
1. receives a token and saves it in a sink;
1. requests secrets;
1. renders the template and creates `/etc/myapp/application.properties`;
1. exits with code `0`.

#### Verify the result

```shell
ls -la /var/run/stronghold-agent/token
ls -la /etc/myapp/application.properties
sudo cat /etc/myapp/application.properties
```

After successful verification, Agent can be started as a `systemd` service:

```shell
systemctl start stronghold-agent
systemctl status stronghold-agent
journalctl -u stronghold-agent -f
```

### Advanced templating scenarios

#### Dynamic credentials for databases

```hcl
{{ with secret "database/creds/myapp-role" }}
# Auto-generated credentials (TTL: 1h)
# Rotation: automatic
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
DB_LEASE_ID={{ .LeaseID }}
DB_LEASE_DURATION={{ .LeaseDuration }}
{{ end }}
```

In this scenario, Agent:

- requests temporary credentials;
- updates the file on rotation;
- runs a command to reload the application.

#### PKI certificates

```hcl
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.certificate }}
{{ .Data.ca_chain }}
{{ end }}
```

```hcl
{{ with secret "pki/issue/web-server" "common_name=app.example.com" "ttl=720h" }}
{{ .Data.private_key }}
{{ end }}
```

Configuration example:

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

#### Conditional logic and loops

```go
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
```

```go
{{ with secret "secret/data/myapp/allowed-ips" }}
{{ range $index, $ip := .Data.data.ips }}
allow {{ $ip }};
{{ end }}
{{ end }}
```

#### Multiple secrets in one file

```go
{{ with secret "database/creds/app" }}
DB_USER={{ .Data.username }}
DB_PASS={{ .Data.password }}
{{ end }}
{{ with secret "secret/data/myapp/api-keys" }}
STRIPE_KEY={{ .Data.data.stripe_key }}
SENDGRID_KEY={{ .Data.data.sendgrid_key }}
{{ end }}
{{ with secret "secret/data/myapp/redis" }}
REDIS_HOST={{ .Data.data.host }}
REDIS_PASSWORD={{ .Data.data.password }}
{{ end }}
```

### Important parameters of the `template` block

| Parameter | Description | Example |
| --- | --- | --- |
| `source` | Path to the template file | `/etc/app/template.ctmpl` |
| `destination` | Path to the resulting file | `/etc/app/config.conf` |
| `perms` | Access permissions | `"0600"`, `"0644"` |
| `user` | File owner | `"myapp"` |
| `group` | File group | `"myapp"` |
| `command` | Command after rendering | `"systemctl reload app"` |
| `command_timeout` | Command timeout | `"30s"` |
| `error_on_missing_key` | Error if a key is missing | `true` / `false` |
| `wait.min` | Minimum time between updates | `"2s"` |
| `wait.max` | Maximum time between updates | `"10s"` |
| `backup` | Create a backup copy | `true` / `false` |

## Template and env_template

`template` renders secrets into a file on disk.

Use this mode if the application reads configuration from files:

- `.conf`, `.ini`, `.yaml`, `.properties`;
- TLS files `*.pem`;
- keys and certificates.

For `template`, file-related parameters are available: `destination`, `perms`, `user`, `group`, `backup`, `wait`, and also `command` for service reload.

`env_template` together with `exec` passes secrets into environment variables of a child process.

Use this mode if:

- the application reads configuration from environment variables;
- restarting the process on secret rotation is acceptable;
- secrets must not be written to disk.

Keep the following specifics in mind:

- each `env_template` sets the value of exactly one environment variable;
- the block is always written as `env_template "VAR_NAME" { ... }`;
- `env_template` does not create a `.env` file;
- the `destination`, `perms`, `command`, `wait`, and similar fields are not supported in `env_template`.

Practical nuances:

- if Agent runs under `systemd` with hardening and `ProtectSystem=strict`, you must add the directory from `template.destination` to `ReadWritePaths`, otherwise writing will be denied;
- if you run a Docker container through `env_template`, environment variables must be passed explicitly to `docker run` through `--env`.

## Process Supervisor mode

Process Supervisor allows Stronghold Agent to start an application as a child process and pass secrets directly into environment variables.

### How it works

In this mode, Agent:

1. starts as the parent process;
1. requests secrets from Stronghold;
1. forms environment variables;
1. starts the application as a child process;
1. tracks secret changes;
1. restarts the application with new values when secrets change.

### Limitations

This mode has several limitations:

- `exec` must be used with at least one `env_template`;
- `env_template` cannot be combined with `template` and `api_proxy` in the same configuration file;
- each `env_template` forms only one environment variable.

### Advantages

This mode is useful because:

- secrets are not written to disk;
- the application is restarted automatically when secrets are renewed;
- secrets are isolated at the process level;
- the mode is well suited for 12-factor applications;
- it simplifies migration of legacy applications to environment variables.

### When to use

Process Supervisor is suitable if:

- the application reads configuration from environment variables;
- there are increased security requirements;
- the application is containerized but runs on a VM;
- dynamic credentials with frequent rotation are used;
- a convenient mode is needed for development and testing.

### Step-by-step example

Scenario: a Java Spring Boot application reads secrets from environment variables.

#### Step 1. Prepare the application

```text
server.port=8080
spring.datasource.url=${DB_URL}
spring.datasource.username=${DB_USERNAME}
spring.datasource.password=${DB_PASSWORD}
spring.datasource.driver-class-name=org.postgresql.Driver
spring.jpa.hibernate.ddl-auto=validate
spring.jpa.properties.hibernate.dialect=org.hibernate.dialect.PostgreSQLDialect
api.key=${API_KEY}
```

#### Step 2. Prepare secrets in Stronghold

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

#### Step 3. Configure Agent

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

env_template "JAVA_OPTS" {
  contents = "-Xmx2g -Xms512m -XX:+UseG1GC"
}

env_template "SPRING_PROFILES_ACTIVE" {
  contents = "production"
}
```

#### Step 4. Start Agent

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl
```

After start, Agent:

- authenticates;
- receives database credentials and the API key;
- starts the Java application with secrets in environment variables;
- restarts the application when credentials are rotated.

### Examples for different applications

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

### Important parameters of the `exec` block

| Parameter | Description | Default value |
| --- | --- | --- |
| `command` | Application start command | required |
| `restart_on_secret_changes` | Restart on secret changes: `never`, `always` | `always` |
| `restart_stop_signal` | Signal to stop the process | `SIGTERM` |

### Important parameters of the `env_template` block

| Parameter | Description | Example |
| --- | --- | --- |
| `contents` | Inline template of the environment variable | `<<-EOT ... EOT` |
| `source` | Path to the template file | `"/etc/app/env.ctmpl"` |
| `error_on_missing_key` | Error if a key is missing | `true` / `false` |

{% alert level="info" %}
The `env_template` block always has an environment variable name: `env_template "MY_VAR" { ... }`. The `destination`, `perms`, `command`, `wait`, and similar fields are not supported in `env_template` [3].
{% endalert %}

### Process lifecycle management

When secrets change, for example during database credential rotation, Agent:

1. receives new secrets;
1. forms new environment variables;
1. sends `SIGTERM` to the child process;
1. restarts the process with updated values.

## Token caching and rotation

Stronghold Agent supports token caching and automatic renewal.

### Token caching

Caching helps:

- preserve the token after authentication;
- use the same token for all requests;
- reduce load on the Stronghold server.

### Token renewal

If Agent receives a token with a limited TTL, it renews it in advance. If renewal is not possible, Agent authenticates again.

This is needed to keep the application running continuously without manual intervention.

### Lease renewal

Dynamic secrets also have an expiration time. Agent renews them in advance, then updates configuration files and can reload the application.

This way, secrets remain up to date, and the application does not lose access because of expired credentials.

## API Proxy

Stronghold Agent can work as a proxy for the Stronghold API.

### What this gives you

- a local HTTP(S) endpoint for applications;
- automatic addition of the authentication token;
- response caching;
- reduced network load.

### Configuration example

```hcl
api_proxy {
  use_auto_auth_token = true
}

listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
}
```

### How this looks for the application

```shell
curl http://127.0.0.1:8200/v1/secret/data/myapp
```

In this scenario, the application connects to the local Agent, and Agent itself adds the token and proxies the request to the Stronghold server.

## Auto-Auth

Auto-Auth is one of the key features of Stronghold Agent. It automates receiving and renewing the authentication token.

### How it works

1. Agent starts with a configured authentication method.
1. It authenticates to Stronghold on its own.
1. It receives a token and uses it for templating, API Proxy, and other tasks.
1. If a sink is configured, it writes the token to a file.
1. It renews the token before TTL expires.
1. If necessary, it authenticates again.

### Sink

Sink is the place where Agent writes the received token.

If a sink is configured, the token is written to a file, for example `/var/run/stronghold-agent/token`, and other processes can use it. If no sink is configured, the token is used only by Agent itself.

### Supported authentication methods

Stronghold Agent supports the following authentication methods:

- AppRole — the recommended option for VM and bare metal;
- Token — for simple scenarios;
- JWT/OIDC — for integration with an identity provider;
- cloud providers.

## AppRole

AppRole is the recommended authentication method for machines and applications on VM and bare metal.

### How AppRole works

- `Role ID` — the role identifier, similar to a username;
- `Secret ID` — the secret identifier, similar to a password;
- both values are required for authentication.

### Advantages

AppRole has the following advantages:

- separation of duties;
- flexible policy configuration;
- support for CIDR restrictions;
- the ability to use one-time `Secret ID`.

### Configuration on the Stronghold side

```shell
stronghold auth enable approle
stronghold write auth/approle/role/myapp \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="myapp-policy"
stronghold read auth/approle/role/myapp/role-id
stronghold write -f auth/approle/role/myapp/secret-id
```

### Agent configuration

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

### Storage and delivery of credentials

#### Role ID

- this is the public role identifier;
- it can be delivered through Ansible, Puppet, a VM image, or manually;
- it is usually stored in `/etc/stronghold-agent/role-id`;
- it is not deleted after use;
- by itself, it is not considered a critical secret.

#### Secret ID

- this is a sensitive secret;
- it is better to deliver it through a secure channel;
- it is usually stored in `/etc/stronghold-agent/secret-id`;
- if needed, it can be deleted after reading;
- it should not be stored in Git or in a configuration management system without additional protection.

### Secret ID types

#### One-time

```shell
stronghold write auth/approle/role/myapp \
  secret_id_num_uses=1 \
  policies="myapp-policy"
```

or:

```shell
stronghold write -f auth/approle/role/myapp/secret-id num_uses=1
```

Features:

- used only once;
- becomes invalid after use;
- this is the safest option for the production environment.

#### Reusable

```shell
stronghold write auth/approle/role/myapp \
  secret_id_num_uses=0 \
  policies="myapp-policy"
```

Features:

- can be used multiple times;
- suitable for development and testing;
- requires manual rotation if compromised.

#### With a limited TTL

```shell
stronghold write auth/approle/role/myapp \
  secret_id_ttl=24h \
  policies="myapp-policy"
```

Features:

- expires after the specified time;
- provides a balance between security and convenience;
- requires a new `Secret ID` after expiration.

### Full example of AppRole setup and delivery

```shell
stronghold auth enable approle
stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF

stronghold write auth/approle/role/myapp \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="myapp-policy" \
  secret_id_num_uses=1 \
  secret_id_ttl=24h

stronghold read auth/approle/role/myapp/role-id
stronghold write -f auth/approle/role/myapp/secret-id
```

Delivery to the target server:

```shell
ssh root@app-server.example.com << 'ENDSSH'
  mkdir -p /etc/stronghold-agent
  chown root:stronghold-agent /etc/stronghold-agent
  chmod 750 /etc/stronghold-agent
ENDSSH
```

```shell
ssh root@app-server.example.com << 'ENDSSH'
  echo -n "abc123-def456-ghi789" > /etc/stronghold-agent/role-id
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/role-id
  chmod 0640 /etc/stronghold-agent/role-id
ENDSSH
```

```shell
ssh root@app-server.example.com << 'ENDSSH'
  echo -n "xyz789-abc123-def456" > /etc/stronghold-agent/secret-id
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/secret-id
  chmod 0640 /etc/stronghold-agent/secret-id
ENDSSH
```

Agent configuration:

```shell
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
```

Configuration permissions:

```shell
chown root:stronghold-agent /etc/stronghold-agent/agent.hcl
chmod 0640 /etc/stronghold-agent/agent.hcl
```

Start and verification:

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"
ls -la /var/run/stronghold-agent/token
ls -la /etc/stronghold-agent/secret-id
```

### Practical recommendations for AppRole

- use one-time `Secret ID` for the production environment;
- deliver `Role ID` and `Secret ID` through different channels;
- restrict access by CIDR;
- log `Secret ID` usage for audit.

## Token

Direct token use is the simplest authentication method. Agent reads a ready-to-use token from a file and uses it.

### When to use

This option is suitable for:

- test environments;
- temporary installations;
- scenarios where AppRole cannot be used;
- simple cases without increased security requirements.

### Limitations

This method has the following limitations:

- a token is a long-lived credential;
- there is no separation of duties as with AppRole;
- manual rotation is required if compromised;
- this option is not recommended for the production environment.

### Full example

Creating a token:

```shell
stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF

stronghold token create \
  -policy=myapp-policy \
  -ttl=720h \
  -renewable=true \
  -display-name="myapp-agent" \
  -format=json
```

Token delivery:

```shell
ssh root@app-server.example.com << 'ENDSSH'
  mkdir -p /etc/stronghold-agent
  chown root:stronghold-agent /etc/stronghold-agent
  chmod 750 /etc/stronghold-agent
ENDSSH
```

```shell
echo -n "$AGENT_TOKEN" | ssh root@app-server.example.com 'cat > /etc/stronghold-agent/token'
```

```shell
ssh root@app-server.example.com << 'ENDSSH'
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/token
  chmod 0640 /etc/stronghold-agent/token
ENDSSH
```

Agent configuration:

```shell
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
```

Verification:

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50
```

### Important token parameters

- `ttl` — initial token lifetime;
- `renewable` — whether it can be renewed;
- `period` — renewal period;
- `explicit-max-ttl` — absolute maximum lifetime.

### Practical recommendations for token authentication

- use `renewable=true`;
- set a reasonable TTL;
- limit the total lifetime through `explicit-max-ttl`;
- regularly revoke unused tokens;
- store the token with minimal access permissions;
- if possible, switch to AppRole for production.

## JWT/OIDC

JWT/OIDC authentication allows you to use an existing identity management system to log in to Stronghold.

### How it works

1. The application receives a JWT from an identity provider.
1. The JWT contains claims about a user or service.
1. Stronghold verifies the signature and extracts the claims.
1. Based on the claims, it issues its own Stronghold token.

### When to use

This option is suitable for:

- integration with corporate SSO;
- using a service account from an identity provider;
- federated authentication;
- CI/CD through OIDC, for example GitHub Actions or GitLab CI.

### Advantages

JWT/OIDC provides the following advantages:

- centralized identity management;
- no need to create separate credentials for each application;
- JWT is rotated on the identity provider side;
- you can use MFA and other identity provider capabilities.

### Full example with Keycloak

Method configuration on the Stronghold side:

```shell
stronghold auth enable jwt
stronghold write auth/jwt/config \
  oidc_discovery_url="https://keycloak.example.com/realms/myrealm" \
  oidc_client_id="stronghold" \
  oidc_client_secret="client-secret-from-keycloak" \
  default_role="default"
```

Creating a policy:

```shell
stronghold policy write myapp-jwt-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF
```

Creating a role:

```shell
stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_subject="service-account-myapp" \
  token_ttl=1h \
  token_max_ttl=4h \
  token_policies="myapp-jwt-policy"
```

Variant with additional claims:

```shell
stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_claims='{"environment":"production","app":"myapp"}' \
  claim_mappings='{"department":"dept"}' \
  token_policies="myapp-jwt-policy"
```

Obtaining a JWT from the identity provider:

```shell
curl -X POST "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=myapp-service" \
  -d "client_secret=service-secret" \
  -d "grant_type=client_credentials" \
  | jq -r '.access_token' > /tmp/jwt-token.txt
```

Delivering the JWT to the server:

```shell
scp /tmp/jwt-token.txt root@app-server.example.com:/etc/stronghold-agent/jwt-token
```

```shell
ssh root@app-server.example.com << 'ENDSSH'
  chown stronghold-agent:stronghold-agent /etc/stronghold-agent/jwt-token
  chmod 0640 /etc/stronghold-agent/jwt-token
ENDSSH
```

Agent configuration:

```shell
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
```

Verification:

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"
cat /var/run/stronghold-agent/token
```

### JWT method specifics

- the JWT token and the Stronghold token are different tokens;
- JWT usually lives for 5–60 minutes;
- after login, Agent receives a Stronghold token;
- the Stronghold token is renewed automatically;
- Agent does not automatically renew the JWT itself.

If the JWT expires, you need to obtain a new token from the identity provider. The source materials suggest a separate periodic renewal for this.

### JWT verification

```shell
cat /etc/stronghold-agent/jwt-token | cut -d. -f2 | base64 -d | jq
```

### Practical recommendations for JWT/OIDC

- use a short TTL for JWT;
- configure `bound_audiences`;
- use `bound_subject` or `bound_claims` for stricter validation;
- use OIDC discovery for production;
- log authentication for audit.
