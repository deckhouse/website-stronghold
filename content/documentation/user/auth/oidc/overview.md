---
title: "OIDC"
linkTitle: "OIDC"
weight: 20
description: "Authentication in Stronghold using OIDC"
---

The `jwt auth` method can be used for authentication in Stronghold in two ways:
- via OIDC;
- by directly providing a JWT.

This section covers the OIDC scenario, in which a user authenticates through an external OIDC provider using the web interface or CLI. The specific scenario depends on the role configuration.

{{< alert level="info" >}}
If you need a scenario in which the client directly provides a JWT without browser-based login, use the [JWT](../jwt) section. In the JWT scenario, Stronghold cryptographically validates the provided token using locally configured keys or keys obtained through OIDC Discovery.
{{< /alert >}}

## How OIDC works in Stronghold

The OIDC authentication flow in Stronghold is based on a standard browser-based login. Stronghold redirects the user to the configured OIDC provider, receives the authentication result, and issues its own token.

For OIDC, Stronghold uses the Authorization Code flow with the PKCE extension.

Stronghold supports two built-in OIDC authentication flows:

- login through the Stronghold web interface;
- login through the CLI using the `d8 stronghold login` command.

## Redirect URI configuration

**Correct redirect URI configuration is one of the most important parts of OIDC setup.**

The redirect URI must:

- be configured in Stronghold in the role's `allowed_redirect_uris` parameter;
- be configured in the external OIDC provider;
- match exactly in both systems.

When verifying that they match, pay attention to the following:

- protocol — `http` or `https`;
- host — `127.0.0.1` or `localhost`;
- port number;
- presence or absence of a trailing slash.

## OIDC for CLI

If you plan to use CLI login with the following command:

```shell
d8 stronghold login -method=oidc
```

allow a redirect URI on `localhost`.

A typical example:

```text
http://localhost:8250/oidc/callback
```

If a different host or port is used for CLI login, that URI must also match one of the values in `allowed_redirect_uris` and must be allowed by the OIDC provider.

### Login through CLI

By default, CLI uses the `/oidc_deckhouse` path. If the authentication method is enabled at a different path, specify it explicitly:

```shell
d8 stronghold login -method=oidc -path=oidc_deckhouse role=test
```

Example of CLI behavior:

```text
Complete the login via your OIDC provider. Launching browser to:
https://myco.auth0.com/authorize?redirect_uri=http%3A%2F%2Flocalhost%3A8400%2Foidc%2Fcallback&client_id=r3qXc2bix9eF...
```

After that, the browser opens the OIDC provider page to complete the login. If the browser cannot be opened automatically, open the URL manually or use the `skip_browser` parameter.

### Useful CLI parameters

You can use the following parameters for the CLI login flow:

- `mount` — defaults to `oidc`;
- `listenaddress` — defaults to `localhost`;
- `port` — defaults to `8250`;
- `callbackhost` — defaults to `localhost`;
- `callbackmethod` — defaults to `http`;
- `callbackport` — defaults to the same value as `port`;
- `skip_browser` — defaults to `false`.

In most cases, you do not need to change these parameters. In some scenarios, `callbackport` and `port` can differ: `callbackport` is used in the `redirect_uri` parameter, while `port` is used by the local server that accepts the incoming request.

## OIDC for the web interface

The Stronghold web interface supports direct OIDC login and does not require separate configuration. After the authentication method is enabled, the interface automatically uses the available login flow.

### Login through the web interface

A typical user flow:

1. Select the OIDC login method.
1. Enter the role name if required.
1. Click the login button for the OIDC provider.
1. Complete authentication with the configured provider.

## What is configured in an OIDC role

An OIDC role defines how Stronghold processes the authentication result.

Pay attention to the following parameters:

- `allowed_redirect_uris` defines the allowed redirect URIs;
- `user_claim` is the required and key role parameter;
- `oidc_scopes` defines the list of scopes that Stronghold requests from the provider;
- if needed, you can use additional claim bindings such as `bound_claims`.

The `bound_audiences` parameter is usually not required for OIDC roles. In most cases, the OIDC provider uses `client_id` to define the audience, and the standard validation is sufficient.

If a role parameter requires a map value, define it as a complete JSON object instead of configuring it piece by piece.

Example:

```shell
d8 stronghold write auth/oidc/role/demo -<<EOF
{
  "user_claim": "sub",
  "bound_audiences": "abc123",
  "role_type": "oidc",
  "policies": "demo",
  "ttl": "1h",
  "bound_claims": { "groups": ["mygroup/mysubgroup"] }
}
EOF
```

## Practical configuration recommendations

When configuring OIDC, follow this order:

- first, make sure that basic login works successfully;
- then add claim restrictions and additional checks;
- separately verify redirect URIs in Stronghold and at the provider;
- request only the scopes that are actually needed from the provider.

For example, if Stronghold must receive the user profile and group information, configure the following value for the role:

```text
oidc_scopes="profile,groups"
```

## OIDC troubleshooting

If OIDC authentication does not work, first do the following:

- check the Stronghold logs;
- make sure that redirect URIs match in Stronghold and at the provider;
- verify that `user_claim` is configured correctly;
- temporarily avoid making the role more complex with additional `bound_claims`;
- confirm with the provider what claim structure is actually present in the token.

If you need to inspect the contents of a JWT manually, you can decode the token payload.

For example:

```shell
cat jwt.json | jq -r .access_token | cut -d. -f2 | base64 -D
```

This is useful for troubleshooting claims and comparing them with the expected role configuration.

## verbose_oidc_logging

Stronghold provides the `verbose_oidc_logging` role option.

If it is enabled and logging is running at the `debug` level, Stronghold writes the received OIDC token to the server logs.

{{< alert level="warning" >}}
Do not use `verbose_oidc_logging` in a production environment. Claims can contain sensitive data.
{{< /alert >}}

## OIDC providers

Stronghold works with a number of OIDC providers. Detailed setup steps for specific providers are documented in a separate section.

Continue to the [OIDC providers](./oidc-providers/overview) section.
