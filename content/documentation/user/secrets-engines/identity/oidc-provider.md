---
title: "OIDC identity provider"
description: "Information about OIDC identity provider in Deckhouse Stronghold."
weight: 20
---

## Overview

Stronghold can act as an OpenID Connect (OIDC) identity provider for client applications that use this protocol.
This allows applications to use Stronghold as a source of [identity](../../../../concepts/identity/) and to apply the available [authentication methods](../../../../concepts/auth/) to authenticate end users.

After this capability is enabled, Stronghold can act as an intermediary between a client application and external identity providers through already configured authentication methods.
In addition, client applications can retrieve end-user identity data from Stronghold.

The OIDC provider system in Stronghold is built on the identity secrets engine.
This engine is enabled by default and cannot be disabled or moved.

In each Stronghold namespace, an OIDC provider and key are available by default.
This built-in configuration lets you start using Stronghold as an identity provider with minimal initial setup.

To allow a client application to use Stronghold as an OIDC provider, you usually need to do the following:

- Enable an authentication method.
- Create a user.
- Create a client application.
- Retrieve the `client_id` and `client_secret`.
- Retrieve the `issuer` value from the OIDC discovery configuration.

## Setup

Below is a minimal configuration example that allows a client application to use Stronghold as an OIDC provider.

1. Enable the `userpass` authentication method.

   ```console
   $ d8 stronghold auth enable userpass
   Success! Enabled userpass auth method at: userpass/
   ```

   In OIDC mode, you can use any Stronghold authentication method.
   For simplicity, this example uses the `userpass` method.

1. Create a user.

   ```console
   $ d8 stronghold write auth/userpass/users/end-user password="securepassword"
   Success! Data written to: auth/userpass/users/end-user
   ```

   This user will authenticate in Stronghold through the client application, that is, through the OIDC relying party.

1. Create a client application.

   ```console
   $ d8 stronghold write identity/oidc/client/my-webapp \
     redirect_uris="https://localhost:9702/auth/oidc-callback" \
     assignments="allow_all"
   Success! Data written to: identity/oidc/client/my-webapp
   ```

   This command creates a client application that you can use when configuring the OIDC relying party.

   The `assignments` parameter limits which Stronghold entities and groups are allowed to authenticate through this client application.
   By default, authentication is not allowed for any entity.
   To allow authentication for all Stronghold entities, use the built-in `allow_all` assignment.

1. Read the client credentials.

   ```console
   $ d8 stronghold read identity/oidc/client/my-webapp
   Key                 Value
   ---                 -----
   access_token_ttl    24h
   assignments         [allow_all]
   client_id           GSDTnn3KaOrLpNlVGlYLS9TVsZgOTweO
   client_secret       hvo_secret_gBKHcTP58C4aq7FqPWsuqKgpiiegd7ahpifGae9WGkHRCwFEJTZA9KGdNVpzE0r8
   client_type         confidential
   id_token_ttl        24h
   key                 default
   redirect_uris       [https://localhost:9702/auth/oidc-callback]
   ```

   The `client_id` and `client_secret` parameters are the client application credentials.
   They are usually required when configuring the OIDC relying party.

1. Read the OIDC discovery configuration.

   ```console
   $ curl -s http://127.0.0.1:8200/v1/identity/oidc/provider/default/.well-known/openid-configuration
   {
     "issuer": "http://127.0.0.1:8200/v1/identity/oidc/provider/default",
     "jwks_uri": "http://127.0.0.1:8200/v1/identity/oidc/provider/default/.well-known/keys",
     "authorization_endpoint": "http://127.0.0.1:8200/ui/vault/identity/oidc/provider/default/authorize",
     "token_endpoint": "http://127.0.0.1:8200/v1/identity/oidc/provider/default/token",
     "userinfo_endpoint": "http://127.0.0.1:8200/v1/identity/oidc/provider/default/userinfo",
     "request_parameter_supported": false,
     "request_uri_parameter_supported": false,
     "id_token_signing_alg_values_supported": [
       "RS256",
       "RS384",
       "RS512",
       "ES256",
       "ES384",
       "ES512",
       "EdDSA"
     ],
     "response_types_supported": [
       "code"
     ],
     "scopes_supported": [
       "openid"
     ],
     "subject_types_supported": [
       "public"
     ],
     "grant_types_supported": [
       "authorization_code"
     ],
     "token_endpoint_auth_methods_supported": [
       "none",
       "client_secret_basic",
       "client_secret_post"
     ]
   }
   ```

   Every Stronghold OIDC provider publishes discovery metadata.
   The `issuer` value is usually required when configuring the OIDC relying party.

## Usage

After you configure the Stronghold authentication method and the client application, use the following values to configure the OIDC relying party:

- `client_id` — the client application identifier;
- `client_secret` — the client application secret;
- `issuer` — the issuer of the Stronghold OIDC provider.

Further configuration depends on the specific client application.
For details, refer to the documentation for the corresponding OIDC relying party.

## Supported flows

The OIDC provider feature in Stronghold currently supports the following authentication flow:

- [Authorization Code Flow](https://openid.net/specs/openid-connect-core-1_0.html#CodeFlowAuth).
