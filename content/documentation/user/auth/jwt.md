---
title: "JWT authentication"
linkTitle: "JWT"
weight: 30
description: "JWT authentication in Deckhouse Stronghold."
---

JWT authentication is suitable for scenarios where a client already receives a JSON Web Token (JWT) from a trusted token provider and passes it to Deckhouse Stronghold for validation. In this scenario, Deckhouse Stronghold does not initiate an interactive browser-based login and only validates an already issued token.

{% alert level="info" %}
If you need an interactive browser-based login with an OIDC provider, use the OIDC authentication overview.
{% endalert %}

## How JWT authentication works

When using the `jwt` method, the client passes a JWT and a role name to Deckhouse Stronghold. Deckhouse Stronghold then performs the following checks:

- Validates the JWT signature.
- Validates the token expiration time.
- Validates the role-related parameters.
- Issues a Deckhouse Stronghold token if validation succeeds.

As a result, the user or application receives a regular Deckhouse Stronghold token that can be used for further operations.

## JWT validation

JWT signatures are validated using the public keys of the token issuer. For a single backend, you can choose one of the following validation methods:

- **Static keys** — a set of public keys is stored in the backend configuration.
- **JWKS** — a JSON Web Key Set URL is used, and the keys are retrieved during authentication.
- **OIDC Discovery** — an OIDC Discovery URL is used, the keys are retrieved from it, and additional OIDC checks such as `iss` and `aud` are applied.

If you need to use multiple validation methods, create multiple JWT authentication backends.

## Authentication via CLI

To authenticate via CLI, use the following command:

```shell
d8 stronghold write auth/<path-to-jwt-backend>/login role=demo jwt=...
```

The default path for the JWT authentication backend is `/jwt`. If you use the default backend, the command looks like this:

```shell
d8 stronghold write auth/jwt/login role=demo jwt=...
```

If the JWT backend is mounted at a different path, use that path instead of `jwt`.

## Authentication via API

By default, the `auth/jwt/login` endpoint is used. If the authentication method is enabled at a different path, replace `jwt` with the required value.

Request example:

```shell
curl \
  --request POST \
  --data '{"jwt": "your_jwt", "role": "demo"}' \
  http://127.0.0.1:8200/v1/auth/jwt/login
```

Deckhouse Stronghold returns the token in the `auth.client_token` field.

Response example:

```json
{
  "auth": {
    "client_token": "38fe9691-e623-7238-f618-c94d4e7bc674",
    "accessor": "78e87a38-84ed-2692-538f-ca8b9f400ab3",
    "policies": ["default"],
    "metadata": {
      "role": "demo"
    },
    "lease_duration": 2764800,
    "renewable": true
  }
}
```

## Enabling the method

Before authentication, enable and configure the JWT authentication backend. This is usually done by an administrator or a configuration management tool.

To enable the authentication method, run the following command:

```shell
d8 stronghold auth enable jwt
```

You can also mount the same method at a different path. For example, at the `oidc` path:

```shell
d8 stronghold auth enable -path=oidc jwt
```

The backend will be mounted at the selected path.

## Configuring the backend

Deckhouse Stronghold uses the `/config` endpoint for configuration. To support `jwt` roles, specify one of the following key sources:

- local keys.
- a JWKS URL.
- an OIDC Discovery URL.

For `oidc` roles, the `oidc_client_id` and `oidc_client_secret` parameters are also required. For the `jwt` scenario, these parameters can be left empty.

### Example configuration via OIDC Discovery

```shell
d8 stronghold write auth/jwt/config \
  oidc_discovery_url="https://myco.auth0.com/" \
  oidc_client_id="m5i8bj3iofytj" \
  oidc_client_secret="f4ubv72nfiu23hnsj" \
  default_role="demo"
```

### Example configuration for JWT validation only

If Deckhouse Stronghold should only validate JWTs, leave `oidc_client_id` and `oidc_client_secret` empty:

```shell
d8 stronghold write auth/jwt/config \
  oidc_discovery_url="https://MYDOMAIN.eu.auth0.com/" \
  oidc_client_id="" \
  oidc_client_secret=""
```

## Creating a role

After configuring the backend, create a named role:

```shell
d8 stronghold write auth/jwt/role/demo \
  bound_subject="r3qX9DljwFIWhsiqwFiu38209F10atW6@clients" \
  bound_audiences="https://vault.plugin.auth.jwt.test" \
  user_claim="https://vault/user" \
  groups_claim="https://vault/groups" \
  policies=webapps \
  ttl=1h
```

This role:

- Allows authentication with a JWT that has the specified `sub` and `aud` values.
- Assigns the `webapps` policy.
- Uses the specified user and group claims
  to configure identity aliases.

## Related role parameters

After a JWT successfully passes signature and expiration validation, Deckhouse Stronghold checks all role-related parameters.

### `bound_subject`

The `bound_subject` parameter must match the `sub` value in the JWT.

### `bound_claims`

The `bound_claims` parameter lets you define arbitrary claim-based constraints.
This is a JSON file in the form of a key-value map.

Example:

```json
{
  "division": "Europe",
  "department": "Engineering"
}
```

Only JWTs that contain the `division` and `department` claims with the `Europe` and `Engineering` values are authorized.

If a value is a list, the claim must match one of the list items.

For example:

```json
{
  "email": ["fred@example.com", "julie@example.com"]
}
```

## Claims as metadata

Claim data can be copied to the metadata of the authentication token and aliases using the `claim_mappings` parameter.

Example:

```json
{
  "division": "organization",
  "department": "department"
}
```

This means the following:

- The `division` claim value is copied to the `organization` metadata key.
- The `department` claim value is copied to the `department` metadata key.

{% alert level="info" %}
The `role` metadata key name is reserved and cannot be used for claim mapping.
{% endalert %}

## Claims and JSON Pointer

The `bound_claims`, `groups_claim`, `claim_mappings`, and `user_claim` parameters can reference both top-level claims and nested data inside the JWT.

If the required claim is located at the top level of the JWT, specify its name directly. If the claim is nested deeper, use JSON Pointer.

JWT example:

```json
{
  "division": "North America",
  "groups": {
    "primary": "Engineering",
    "secondary": "Software"
  }
}
```

In this case:

- `division` points to `North America`.
- `/groups/primary` points to `Engineering`.

## Practical recommendations

Use the following recommendations:

- Use JWT authentication if the client already receives a JWT from a trusted token provider.
- Choose only one signature validation method for a single backend.
- For complex scenarios, create multiple backends instead of overloading one.
- First make sure the basic authentication flow works, and then add claim-based constraints.
- Before enabling strict `bound_claims` and `bound_subject` constraints, verify the actual claim values received from the token issuer.
