---
title: "JWT/OIDC"
linkTitle: "JWT/OIDC"
description: "Stronghold Agent authentication using JWT/OIDC."
weight: 70
---

JWT/OIDC authentication lets Stronghold Agent use an external identity provider to obtain a Stronghold token. This method is suitable for scenarios that already use a centralized identity management system.

This page describes how JWT/OIDC authentication works, when to use it, and how to configure Stronghold Agent to work with JWT.

## When to use JWT/OIDC

Use JWT/OIDC in the following cases:

- Stronghold must be integrated with a corporate single sign-on system.
- The application receives a JWT from an external identity provider.
- Federated authentication is required.
- Authentication is performed in CI/CD through OIDC, for example in GitHub Actions or GitLab CI.

For virtual machines and bare metal without an external identity provider, AppRole is usually simpler.

## How JWT/OIDC authentication works

The process usually looks as follows:

1. The application or an external process obtains a JWT from the identity provider.
1. Stronghold Agent reads the JWT from a file.
1. The Agent sends the JWT to the Stronghold authentication method.
1. Stronghold verifies the token signature and claims.
1. After successful verification, Stronghold issues its own token.
1. The Agent uses this token to work with secrets.

JWT and the Stronghold token are different tokens. JWT is used for sign-in, and the Stronghold token is used for further Agent work.

## Benefits

JWT/OIDC authentication provides the following benefits:

- Lets you use an existing identity provider.
- Reduces the need for separate credentials for each application.
- Lets you use claims to restrict access.
- Simplifies integration with corporate SSO and CI/CD.

## Limitations

Consider the following specifics:

- The Agent does not issue JWTs by itself.
- JWTs usually have a short lifetime.
- After JWT expiration, a new token must be obtained from the identity provider.
- Stronghold Agent renews the Stronghold token automatically, but it cannot always renew the original JWT automatically.

For long-running workloads, prepare a mechanism for renewing the JWT in an external process.

## Configuring the JWT method in Stronghold

The following example configures the JWT method with Keycloak.

### Step 1. Enable the authentication method

```shell
stronghold auth enable jwt
```

### Step 2. Configure JWT/OIDC

```shell
stronghold write auth/jwt/config \
  oidc_discovery_url="https://keycloak.example.com/realms/myrealm" \
  oidc_client_id="stronghold" \
  oidc_client_secret="client-secret-from-keycloak" \
  default_role="default"
```

### Step 3. Create a policy

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

### Step 4. Create a role

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

If needed, additionally restrict the role by claims:

```shell
stronghold write auth/jwt/role/myapp-role \
  role_type="jwt" \
  bound_audiences="stronghold" \
  user_claim="sub" \
  bound_claims='{"environment":"production","app":"myapp"}' \
  claim_mappings='{"department":"dept"}' \
  token_policies="myapp-jwt-policy"
```

## Obtaining a JWT from an identity provider

The following example obtains a JWT from Keycloak:

```shell
curl -X POST "https://keycloak.example.com/realms/myrealm/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=myapp-service" \
  -d "client_secret=service-secret" \
  -d "grant_type=client_credentials" \
  | jq -r '.access_token' > /tmp/jwt-token.txt
```

Then pass the JWT to the server where Stronghold Agent runs.

## Delivering the JWT to a server

Copy the JWT file to the target server:

```shell
scp /tmp/jwt-token.txt root@app-server.example.com:/etc/stronghold-agent/jwt-token
```

Restrict access to the file:

```shell
ssh root@app-server.example.com << 'ENDSSH'
chown stronghold-agent:stronghold-agent /etc/stronghold-agent/jwt-token
chmod 0640 /etc/stronghold-agent/jwt-token
ENDSSH
```

## Stronghold Agent configuration

Create `/etc/stronghold-agent/agent.hcl`:

```hcl
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
```

In this example, the Agent:

- reads the JWT from `/etc/stronghold-agent/jwt-token`;
- authenticates through the `auth/jwt` method;
- obtains a Stronghold token;
- writes the token to a file sink;
- uses it to render the template.

## Verification

Start Stronghold Agent and check the journal:

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50 | grep -i "authentication successful"
cat /var/run/stronghold-agent/token
```

## Checking JWT contents

If needed, check token claims:

```shell
cat /etc/stronghold-agent/jwt-token | cut -d. -f2 | base64 -d | jq
```

{{< alert level="info" >}}
JWT and a Stronghold token have different purposes. JWT is used for authentication, and the Stronghold token is used for further Agent work.
{{< /alert >}}

## Best practices

When using JWT/OIDC, follow these recommendations:

- Use a short TTL for JWTs.
- Configure `bound_audiences`.
- Use `bound_subject` or `bound_claims` to further restrict access.
- Use OIDC discovery whenever possible.
- Log successful and failed authentication attempts.
- Do not store JWTs in Git or another version control system in plain text.
