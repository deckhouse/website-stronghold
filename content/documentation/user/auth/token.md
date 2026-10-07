---
title: "Token"
description: "The built-in Token auth method: token login, creating service, batch, orphan, and periodic tokens, token store roles, and working with accessors."
weight: 60
---

## Token auth method

The `token` auth method is built-in and automatically available at `/auth/token`. It
allows users to authenticate using a token, as well to create new tokens, revoke
secrets by token, and more.

When any other auth method returns an identity, Stronghold core invokes the
token method to create a new unique token for that identity.

The token store can also be used to bypass any other auth method:
you can create tokens directly, as well as perform a variety of other
operations on tokens such as renewal and revocation.

## Authentication

### Via the CLI

{{< tabs name="stronghold_cmd_98272" >}}
{{% tab name="Stronghold in DKP" %}}

```shell-session
d8 stronghold login token=<token>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell-session
stronghold login token=<token>
```

{{% /tab %}}
{{< /tabs >}}

### Via the API

The token is set directly as a header for the HTTP API. The header should be
either `X-Vault-Token: <token>` or `Authorization: Bearer <token>`.

## Creating tokens

Tokens are created with `d8 stronghold token create` (the `auth/token/create` endpoint). By default, a new token becomes a child of the token that makes the request and can only receive a subset of its policies. For more on token properties, see [Tokens](../../../concepts/tokens/).

### Service token

A `service` token is the default. It is persisted in Stronghold, supports renewal, revocation, and child tokens, and has an accessor:

```bash
d8 stronghold token create \
  -policy="app-read" \
  -ttl=1h \
  -display-name="app-1"
```

### Batch token

A `batch` token is not persisted in Stronghold, cannot be renewed or manually revoked, and has no accessor. It fits large numbers of short-lived operations:

```bash
d8 stronghold token create \
  -type=batch \
  -policy="app-read" \
  -ttl=15m
```

### Orphan token

An orphan token has no parent and is not revoked when the token that created it is revoked. Creating one requires access to `auth/token/create-orphan` or the `sudo` capability on `auth/token/create`:

```bash
d8 stronghold token create -orphan -policy="app-read" -ttl=24h
```

### Periodic token

A periodic token has no maximum lifetime: on every renewal, its TTL is reset to the period. The token expires if it is not renewed within the period:

```bash
d8 stronghold token create -period=1h -policy="app-read"
```

### Explicit max TTL

The `-explicit-max-ttl` flag sets a hard limit on the token lifetime that renewal cannot exceed, including for periodic tokens:

```bash
d8 stronghold token create \
  -policy="app-read" \
  -ttl=1h \
  -explicit-max-ttl=8h
```

### Use limit

The `-use-limit` flag sets how many times the token can be used. Once the limit is reached, the token is revoked:

```bash
d8 stronghold token create -policy="app-read" -use-limit=3 -ttl=10m
```

### Other parameters

| CLI flag | API parameter | Description |
|----------|---------------|-------------|
| `-policy` | `policies` | Token policies. The flag can be repeated. |
| `-no-default-policy` | `no_default_policy` | Do not attach the `default` policy. |
| `-ttl` | `ttl` | Initial token TTL. |
| `-explicit-max-ttl` | `explicit_max_ttl` | Hard maximum TTL. |
| `-period` | `period` | Renewal period of a periodic token. |
| `-renewable` | `renewable` | Whether renewal is allowed. Defaults to `true`. |
| `-orphan` | `no_parent` | Create a token without a parent. |
| `-type` | `type` | Token type: `service` or `batch`. |
| `-use-limit` | `num_uses` | Maximum number of uses. |
| `-display-name` | `display_name` | Token display name. |
| `-metadata` | `meta` | Arbitrary `key=value` metadata. |
| `-entity-alias` | `entity_alias` | Entity alias to associate with the token. |
| `-role` | — | Create the token from a token store role. |

## Token store roles

A token store role (`auth/token/roles/<name>`) defines the properties of tokens issued through it: allowed policies, token type, TTL, period, orphan flag, and more. Roles let you allow creating tokens with policies the calling token does not have, without granting it `sudo`.

1. Create a role:

   ```bash
   d8 stronghold write auth/token/roles/ci-deploy \
     allowed_policies="deploy" \
     disallowed_policies="admin" \
     orphan=true \
     token_type=service \
     token_period=1h \
     token_explicit_max_ttl=24h \
     renewable=true
   ```

   Key role parameters:

   | Parameter | Description |
   |-----------|-------------|
   | `allowed_policies`, `allowed_policies_glob` | Policies that can be assigned to the token. |
   | `disallowed_policies`, `disallowed_policies_glob` | Policies that must not be requested. |
   | `orphan` | Issue orphan tokens. |
   | `token_type` | Token type: `service` or `batch`. |
   | `token_period` | Period for periodic tokens. |
   | `token_explicit_max_ttl` | Explicit maximum TTL. |
   | `token_num_uses` | Token use limit. |
   | `token_bound_cidrs` | CIDR blocks the token can be used from. |
   | `renewable` | Whether tokens can be renewed. |
   | `path_suffix` | Token path suffix for later prefix-based revocation. |
   | `allowed_entity_aliases` | Allowed entity aliases. |

1. Issue a token from the role:

   ```bash
   d8 stronghold token create -role=ci-deploy
   ```

   The calling token needs the `update` capability on `auth/token/create/ci-deploy`.

1. View roles:

   ```bash
   d8 stronghold list auth/token/roles
   d8 stronghold read auth/token/roles/ci-deploy
   ```

## Looking up, renewing, and revoking tokens

- Look up the current token:

  ```bash
  d8 stronghold token lookup
  ```

- Look up another token:

  ```bash
  d8 stronghold token lookup <token>
  ```

- Renew the current token, or another token requesting a new TTL:

  ```bash
  d8 stronghold token renew
  d8 stronghold token renew -increment=1h <token>
  ```

- Revoke a token together with all its child tokens and leases:

  ```bash
  d8 stronghold token revoke <token>
  ```

- Revoke only the token itself; its direct children become orphans:

  ```bash
  d8 stronghold token revoke -mode=orphan <token>
  ```

- Revoke the current token:

  ```bash
  d8 stronghold token revoke -self
  ```

- Check token capabilities on a path:

  ```bash
  d8 stronghold token capabilities <token> secret/data/app
  ```

## Working with accessors

An accessor is a reference to a token that lets you look up, renew, and revoke the token without knowing its value. The accessor is returned when a token is created (the `token_accessor` field) and in `token lookup` output (the `accessor` field). Batch tokens have no accessor.

- Look up a token by accessor:

  ```bash
  d8 stronghold token lookup -accessor <accessor>
  ```

- Renew a token by accessor:

  ```bash
  d8 stronghold token renew -accessor <accessor>
  ```

- Revoke a token by accessor:

  ```bash
  d8 stronghold token revoke -accessor <accessor>
  ```

- Check token capabilities by accessor:

  ```bash
  d8 stronghold token capabilities -accessor <accessor> secret/data/app
  ```

- List the accessors of all tokens (requires `sudo`):

  ```bash
  d8 stronghold list auth/token/accessors
  ```

{{< alert level="warning" >}}
The accessor list allows revoking any token. Grant access to `auth/token/accessors` to administrators only.
{{< /alert >}}

## Tidying the token store

The `auth/token/tidy` endpoint removes stale entries from the token store, such as accessors of expired tokens:

```bash
d8 stronghold write -force auth/token/tidy
```

## Usage examples

Ready-made examples that use this feature:

- [Break-glass access](../../../examples/access/break-glass-access/)

See all examples in [Usage examples](../../../examples/).
