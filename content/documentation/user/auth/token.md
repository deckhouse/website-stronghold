---
title: "Token"
linkTitle: "Token"
weight: 60
description: "Authenticate to Deckhouse Stronghold using a token."
---

Token authentication is built in and automatically available in Deckhouse Stronghold at `/auth/token`.

This method allows you to:

- authenticate with an existing token .
- use tokens for further work with Deckhouse Stronghold .
- create new tokens .
- revoke existing tokens if you have the required permissions.

## When to use Token auth

Use the `Token` method in the following cases:

- a user or application already has a valid Deckhouse Stronghold token .
- you need to log in without repeating authentication through another backend .
- you use automation or API integration .
- you need to work with a token directly as the main access method.

{{< alert level="info" >}}
When any other authentication method completes successfully, Deckhouse Stronghold still creates a token. Because of this, Token auth can be considered the basic built-in access mechanism that is used both on its own and as the result of other authentication methods.
{{< /alert >}}

## How Token auth works

When another authentication method returns a user or application identity, Deckhouse Stronghold creates a new unique token for it.

The token store can also be used without any other authentication method. You can create tokens directly and perform other token operations, such as renewal and revocation.

This means that in some scenarios an already issued token is enough to perform further actions in Deckhouse Stronghold without repeating external authentication.

## Authenticate using the CLI

To log in with a token, run the following command:

```shell
d8 stronghold login token=<token>
```

After that, the CLI uses the provided token for subsequent requests.

The following example shows authentication with the `userpass` authentication method:

```shell
d8 stronghold login -method=userpass \
  username=mitchellh \
  password=foo
```

## Authenticate using the API

When you work with the HTTP API, pass the token in the request header. The following options are supported:

- `X-Vault-Token: <token>` .
- `Authorization: Bearer <token>`.

The following example shows a typical API request:

```shell
curl \
  --request POST \
  --data '{"password": "foo"}' \
  http://127.0.0.1:8200/v1/auth/userpass/login/mitchellh
```

In the response, Deckhouse Stronghold returns the token in the `auth.client_token` field.

Example response:

```json
{
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": null,
  "auth": {
    "client_token": "c4f280f6-fdb2-18eb-89d3-589e2e834cdb",
    "policies": ["admins"],
    "metadata": {
      "username": "mitchellh"
    },
    "lease_duration": 0,
    "renewable": false
  }
}
```

## Important notes

When you use Token auth, consider the following:

- Token auth does not require separate enablement because this method is built in to Deckhouse Stronghold .
- a token grants access within the policies assigned to it .
- if a token has expired or has been revoked, authentication with it fails .
- in user scenarios, a token is often the result of logging in with another authentication method, for example `OIDC`, `LDAP`, or `userpass`.

## Best practices

Use the following recommendations:

- Use Token auth if you already have a valid Deckhouse Stronghold token.
- Do not transfer tokens over insecure channels.
- Do not store tokens in plain text in scripts, notes, or systems without access control.
- If a token is used in automation, consider its TTL, renewal capability, and revocation procedure.
