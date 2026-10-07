---
title: "AppRole for an application and scripts"
linkTitle: "AppRole"
description: "Authenticating an application without a user: the AppRole method, a role limited by network and secret_id usage count, obtaining a token, and reading a secret."
weight: 90
params:
  relatedLinks:
    - title: "AppRole authentication method"
      url: ../../../user/auth/approle/
    - title: "Policies"
      url: ../../../concepts/policy/
    - title: "Application on a virtual machine with Stronghold Agent"
      url: ../legacy-app-on-vm/
---

AppRole is intended for applications and scripts that need access to secrets without a human. The application presents a pair of identifiers: `role_id` (changes rarely and is not a secret) and `secret_id` (issued for a limited time and number of uses), and receives a token with the role's policies.

## Goal

Configure a role for the `myapp` application, obtain its `role_id` and `secret_id`, log in with them, and read a secret, while limiting the token lifetime, the number of `secret_id` uses, and the networks from which login is allowed.

## Prerequisites

- A Stronghold token with permission to enable authentication methods and write policies and roles.
- The KV version 2 secrets engine mounted at `secret/` (it is enabled by default in development mode).

## Step 1. Create a secret and a policy

1. Write the secret that the application will read:

   ```bash
   d8 stronghold kv put -mount=secret myapp/config db_password=s3cr3t
   ```

1. Create a policy that allows only reading this path:

   ```bash
   d8 stronghold policy write myapp - <<'POLICY'
   path "secret/data/myapp/*" {
     capabilities = ["read"]
   }
   POLICY
   ```

## Step 2. Enable the method and create a role

```bash
d8 stronghold auth enable approle

d8 stronghold write auth/approle/role/myapp \
  token_policies=myapp \
  token_ttl=15m \
  token_max_ttl=1h \
  secret_id_ttl=24h \
  secret_id_num_uses=1 \
  token_bound_cidrs="127.0.0.1/32,10.0.0.0/8"
```

Role parameters:

| Parameter | Purpose |
| --- | --- |
| `token_policies` | Policies of the token that the application receives |
| `token_ttl`, `token_max_ttl` | Token lifetime and the limit of its renewal |
| `secret_id_ttl` | The period during which an issued `secret_id` can be used |
| `secret_id_num_uses` | How many times a single `secret_id` can be used to log in. `1` makes `secret_id` single-use |
| `token_bound_cidrs` | Networks from which the issued token can be used |
| `secret_id_bound_cidrs` | Networks from which login with `secret_id` is allowed |

Specify the actual networks of the application servers in `token_bound_cidrs`. The example adds `127.0.0.1/32` so that it can be checked on a single machine.

## Step 3. Get the identifiers

```bash
ROLE_ID=$(d8 stronghold read -field=role_id auth/approle/role/myapp/role-id)
SECRET_ID=$(d8 stronghold write -f -field=secret_id auth/approle/role/myapp/secret-id)
```

You can put `role_id` into the application configuration. Deliver `secret_id` over a secure channel, preferably [wrapped](../../operations/secret-sharing-wrapping/): a wrapped `secret_id` can be retrieved only once.

## Step 4. Log in and read the secret

1. Log in with `role_id` and `secret_id`:

   ```bash
   APP_TOKEN=$(d8 stronghold write -field=token auth/approle/login \
     role_id="$ROLE_ID" secret_id="$SECRET_ID")
   ```

1. Read the secret with the application token:

   ```bash
   STRONGHOLD_TOKEN="$APP_TOKEN" d8 stronghold kv get -mount=secret myapp/config
   ```

## Verification

1. Make sure the token received only the role's policies:

   ```bash
   STRONGHOLD_TOKEN="$APP_TOKEN" d8 stronghold token lookup
   ```

   The `policies` field must contain `default` and `myapp`.

1. Make sure `secret_id` is single-use: a second login must fail:

   ```bash
   # repeated login with the same secret_id
   d8 stronghold write auth/approle/login role_id="$ROLE_ID" secret_id="$SECRET_ID"
   ```

1. Make sure the application cannot read other secrets:

   ```bash
   # reading a path outside the policy
   STRONGHOLD_TOKEN="$APP_TOKEN" d8 stronghold kv get -mount=secret other/config
   ```

   The request must fail with `permission denied`.

## Cleanup

```bash
d8 stronghold token revoke "$APP_TOKEN"
d8 stronghold delete auth/approle/role/myapp
d8 stronghold policy delete myapp
d8 stronghold kv metadata delete -mount=secret myapp/config
```
