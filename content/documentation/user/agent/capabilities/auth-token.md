---
title: "Token authentication"
linkTitle: "Token"
description: "Stronghold Agent authentication using a token."
weight: 60
---

Token authentication is the simplest way to authenticate Stronghold Agent. In this mode, the Agent reads an existing token from a file and uses it to access Stronghold.

This method is suitable for simple and temporary scenarios, but AppRole or JWT/OIDC is usually recommended for production environments.

## When to use token authentication

Use this method in the following cases:

- A simple way to start Stronghold Agent is required.
- A test or temporary environment is used.
- AppRole cannot be used.
- You need to quickly test templates, Auto-Auth, or other Agent capabilities.

If you configure a production environment, use AppRole or JWT/OIDC whenever possible.

## How it works

With token authentication, Stronghold Agent performs the following actions:

1. Reads a token from a file.
1. Uses the token to call Stronghold.
1. Passes the token to internal Agent subsystems.
1. Writes the token to an external sink if needed.
1. Tries to renew the token if its parameters allow renewal.

This method is usually used together with Auto-Auth and the `token_file` method.

## Limitations

Consider the following limitations:

- The token remains a long-lived secret.
- If the token is compromised, manual rotation is required.
- This method does not separate a public identifier from a secret value.
- This option is less suitable for production environments than AppRole.

## Creating a token

First, create a policy that allows the Agent to read the required secrets:

```shell
stronghold policy write myapp-policy - <<EOF
path "secret/data/myapp/*" {
  capabilities = ["read"]
}
path "database/creds/myapp" {
  capabilities = ["read"]
}
EOF
```

Then create a token:

```shell
stronghold token create \
  -policy=myapp-policy \
  -ttl=720h \
  -renewable=true \
  -display-name="myapp-agent" \
  -format=json
```

## Delivering the token to a server

Create a directory for Stronghold Agent configuration:

```shell
ssh root@app-server.example.com << 'ENDSSH'
mkdir -p /etc/stronghold-agent
chown root:stronghold-agent /etc/stronghold-agent
chmod 750 /etc/stronghold-agent
ENDSSH
```

Save the token to a file:

```shell
echo -n "$AGENT_TOKEN" | ssh root@app-server.example.com 'cat > /etc/stronghold-agent/token'
```

Restrict access to the file:

```shell
ssh root@app-server.example.com << 'ENDSSH'
chown stronghold-agent:stronghold-agent /etc/stronghold-agent/token
chmod 0640 /etc/stronghold-agent/token
ENDSSH
```

## Stronghold Agent configuration

Create `/etc/stronghold-agent/agent.hcl` with the following configuration:

```hcl
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
```

In this example, the Agent:

- reads the token from `/etc/stronghold-agent/token`;
- uses Auto-Auth with the `token_file` method;
- writes the working token to `/var/run/stronghold-agent/token`;
- renders the application configuration file.

## Verification

Start Stronghold Agent and check the journal:

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50
```

If authentication succeeds, the Agent can read secrets and perform related tasks, such as template rendering.

## Important token parameters

Consider the following parameters when creating a token:

| Parameter | Description |
| --- | --- |
| `ttl` | Sets the initial token lifetime |
| `renewable` | Allows token renewal |
| `period` | Sets periodic renewal |
| `explicit-max-ttl` | Limits the maximum lifetime |

If the token cannot be renewed, the Agent cannot continue working after the TTL expires without a new token.

## Best practices

When using token authentication, follow these recommendations:

- Use `renewable=true` if your scenario allows it.
- Set a reasonable TTL to reduce risk if the token is compromised.
- Limit the maximum token lifetime.
- Store the token with the minimum required access rights.
- Revoke unused tokens regularly.
- Do not store the token in Git or another version control system in plain text.

{{< alert level="warning" >}}
Token authentication is primarily suitable for test, temporary, and simplified scenarios. For production environments, use AppRole or JWT/OIDC whenever possible.
{{< /alert >}}
