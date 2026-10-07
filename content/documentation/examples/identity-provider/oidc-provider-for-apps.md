---
title: "Stronghold as an OIDC provider for internal applications"
linkTitle: "OIDC provider for applications"
description: "Setting up single sign-on for internal applications with the Stronghold OIDC provider: signing key, assignments, templated scopes, client, and provider."
weight: 120
params:
  relatedLinks:
    - title: "OIDC identity provider"
      url: ../../../user/secrets-engines/identity/oidc-provider/
    - title: "Identity tokens"
      url: ../../../user/secrets-engines/identity/token/
    - title: "Identity"
      url: ../../../concepts/identity/
    - title: "Identity API"
      url: ../../../reference/api/identity/
---

Stronghold can act as an OIDC provider (OpenID Provider) for applications that support OpenID Connect login. A user logs in to Stronghold with any configured auth method, and the application receives an ID token with Stronghold entity data: name, email address, groups.

## Goal

Connect an internal application (Grafana in the example) to the Stronghold OIDC provider so that only members of the `grafana-users` group can log in and the application receives the user's email address and group list.

## Prerequisites

- A Stronghold token with permissions on `identity/*`.
- A configured auth method for users (for example, OIDC through Dex in DP, or [`userpass`](../../../user/auth/userpass/)).
- An Identity group `grafana-users` that contains user entities, and `email` metadata on the entities.
- An external Stronghold address reachable by user browsers and the application: `https://stronghold.example.com` in the examples.

Entities, groups, and aliases are described in [Identity](../../../concepts/identity/).

## Step 1. Create a signing key

```bash
d8 stronghold write identity/oidc/key/apps \
  algorithm=RS256 \
  rotation_period=24h \
  verification_ttl=24h
```

The key rotates automatically once a day, and the previous public part stays available for signature verification for `verification_ttl`.

## Step 2. Restrict who can log in

An assignment defines the entities and groups allowed to log in through the client application:

```bash
GROUP_ID=$(d8 stronghold read -field=id identity/group/name/grafana-users)

d8 stronghold write identity/oidc/assignment/grafana-users group_ids="$GROUP_ID"
```

The built-in `allow_all` assignment allows all entities to log in: do not use it for production applications.

## Step 3. Create scopes

Scopes add claims to the ID token based on a template. The template syntax is the same as for [templated policies](../../../concepts/policy/).

```bash
d8 stronghold write identity/oidc/scope/email \
  description="User email address" \
  template='{"email": {{identity.entity.metadata.email}}}'

d8 stronghold write identity/oidc/scope/groups \
  description="User groups" \
  template='{"groups": {{identity.entity.groups.names}}}'
```

The `template` parameter accepts both a JSON string and a base64-encoded value.

## Step 4. Create a client application

```bash
d8 stronghold write identity/oidc/client/grafana \
  redirect_uris="https://grafana.example.com/login/generic_oauth" \
  assignments="grafana-users" \
  key="apps" \
  id_token_ttl=30m \
  access_token_ttl=1h
```

Allow the client to use the `apps` key:

```bash
CLIENT_ID=$(d8 stronghold read -field=client_id identity/oidc/client/grafana)

d8 stronghold write identity/oidc/key/apps allowed_client_ids="$CLIENT_ID"
```

## Step 5. Create a provider

```bash
d8 stronghold write identity/oidc/provider/internal \
  issuer="https://stronghold.example.com" \
  allowed_client_ids="$CLIENT_ID" \
  scopes_supported="email,groups"
```

Get the parameters for configuring the application:

```bash
d8 stronghold read identity/oidc/client/grafana
curl -s https://stronghold.example.com/v1/identity/oidc/provider/internal/.well-known/openid-configuration | jq
```

From the responses, you need `client_id`, `client_secret`, `issuer`, `authorization_endpoint`, `token_endpoint`, and `userinfo_endpoint`.

## Step 6. Configure the application

An example Grafana configuration (the `auth.generic_oauth` section in `grafana.ini`):

```ini
[auth.generic_oauth]
enabled = true
name = Stronghold
client_id = <client_id>
client_secret = <client_secret>
scopes = openid email groups
auth_url = https://stronghold.example.com/ui/stronghold/identity/oidc/provider/internal/authorize
token_url = https://stronghold.example.com/v1/identity/oidc/provider/internal/token
api_url = https://stronghold.example.com/v1/identity/oidc/provider/internal/userinfo
email_attribute_path = email
groups_attribute_path = groups
```

Store `client_secret` in Stronghold and deliver it to the application like other secrets, for example with [Stronghold Agent](../../../user/agent/overview/) or the [Kubernetes integration](../../delivery/kubernetes-workloads/).

<!-- TODO(verify): whether compatibility of Grafana and other specific applications with the Stronghold OIDC provider is confirmed by tests -->

## Verification

1. Open the application and choose to log in with Stronghold. The browser redirects you to the Stronghold login page and, after login, back to the application.
1. Log in as a user who is not in the `grafana-users` group: login must be rejected.
1. Check the ID token contents in the application log or decode it: it must contain `iss`, `aud` (equal to `client_id`), `email`, and `groups`.
1. Check that the public keys are available:

   ```bash
   curl -s https://stronghold.example.com/v1/identity/oidc/provider/internal/.well-known/keys | jq '.keys[].kid'
   ```

## Cleanup

```bash
d8 stronghold delete identity/oidc/provider/internal
d8 stronghold delete identity/oidc/client/grafana
d8 stronghold delete identity/oidc/scope/email
d8 stronghold delete identity/oidc/scope/groups
d8 stronghold delete identity/oidc/assignment/grafana-users
d8 stronghold delete identity/oidc/key/apps
```
