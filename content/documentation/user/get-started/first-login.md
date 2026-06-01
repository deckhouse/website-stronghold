---
title: "First login"
linkTitle: "First login"
description: "First login to Deckhouse Stronghold via the CLI and web interface"
weight: 20
---

After configuring access to Deckhouse Stronghold, perform your first login. This helps verify that authentication works correctly. The command-line interface (CLI) and the `d8 stronghold login` command are typically used for this. If a web interface is available in your installation, you can also log in through it.

## What must be ready before login

Access to Stronghold is expected to already be configured as described in the [Access to a project](../access/) section. For the CLI scenario, the `STRONGHOLD_ADDR` environment variable must be set.

## What happens during login

Before starting work with Stronghold, a user or service must authenticate.

After successful authentication, Stronghold:

- verifies the provided credentials;
- creates an access token;
- attaches policies to it;
- uses this token for subsequent requests.

The token is what determines which actions the user can perform in Stronghold.

## First login via the CLI

In a typical OIDC scenario, run the following command:

```shell
d8 stronghold login -path=oidc_deckhouse -method=oidc -no-print
```

Depending on your installation configuration, the authentication path may differ. If the administrator provided a different path, use that value.

### What happens after running the command

After you run the command, the CLI initiates login through the configured OIDC provider.
Typically, the following happens:

- a browser opens to complete the login,
  or the CLI shows a URL that you need to open manually;
- after successful authentication, Stronghold issues an access token;
- the CLI uses this token for subsequent commands.

If the browser does not open automatically, copy the URL and open it manually.

{% alert level="info" %}
OIDC-based CLI authentication usually uses a redirect URI on `localhost`, for example, `http://localhost:8250/oidc/callback`. These parameters are defined by the OIDC role configuration and the provider settings.
{% endalert %}

## First login via the web interface

If the Stronghold web interface is available in your installation, you can log in through it.

Use the following typical flow:

1. Open the Stronghold web interface address.
1. Select a login method, for example OIDC.
1. If required, specify the role name.
1. Click the button to log in via the authentication provider.
1. Complete authentication with the external provider.

After a successful login, Stronghold opens the user interface with the permissions that correspond to your account and the assigned policies.

## How to tell whether login was successful

After logging in via the CLI, make sure that `d8 stronghold` commands run with a valid token.
For example, use:

```shell
d8 stronghold status
```

When logging in via the web interface, a successful login is indicated by the Stronghold interface opening without an authentication error.

## What is important to know about the token after login

After authentication, Stronghold creates a token that is used for subsequent requests.
Keep the following in mind:

- a token is created even when you log in through OIDC;
- the token has an expiration time;
- the set of available operations is determined by the attached policies;
- in the CLI, the token is automatically used for subsequent commands.

{% alert level="warning" %}
Do not rely on the internal token structure in automation.Stronghold tokens are opaque values, and their format is not intended to be parsed on the client side.
{% endalert %}

## If login does not work

If the first login fails, check the following:

- whether the authentication path is correct;
- whether the external OIDC provider is reachable;
- whether the browser is blocking the redirect URI from opening;
- whether the issued token has expired.

If OIDC is used, also verify the following:

- the redirect URI is configured correctly;
- the addresses in Stronghold and at the provider match;
- the correct `http` / `https`, host, and port are used;
- the administrator configured the corresponding authentication role.

## Practical recommendations

Use the following recommendations:

- Do not store tokens in plain text in scripts or notes.
- If you are logging in through OIDC for the first time, make sure the browser can open the provider page and return to the local callback URI.