---
title: "Configure Keycloak as an OIDC provider"
linkTitle: "Keycloak"
weight: 30
description: "Configure Keycloak as an OIDC provider for Deckhouse Stronghold."
---

Keycloak can be used as an OIDC provider for user authentication in Deckhouse Stronghold.

To configure it, prepare the `OIDC` auth method in Stronghold and define the `redirect URI` in advance. This address must match in both the Stronghold configuration and the Keycloak client settings.

## Configure a client in Keycloak

Perform the following steps:

1. Select an existing realm or create a new one.
1. Create a new client or select an existing one.
1. Go to the client settings page: “Settings”.
1. Specify the following parameters:
   - `Client Protocol` — `openid-connect`;
   - `Access Type` — `confidential`;
   - `Standard Flow Enabled` — `On`.
1. Configure the allowed redirect URIs in the `Valid Redirect URIs` parameter.
1. Save the changes.
1. Go to the “Credentials” page.
1. Save the `Client ID` and `Client Secret` values.

## What to do after configuration

After configuring the client in Keycloak, perform the following steps:

- Use `Client ID` and `Client Secret` in the OIDC configuration in Stronghold.
- Make sure that the `redirect URI` matches `allowed_redirect_uris` in Stronghold.
- Perform a test login through the Stronghold web interface or CLI.
