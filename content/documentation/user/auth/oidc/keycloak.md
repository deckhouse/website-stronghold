---
title: "Keycloak OIDC provider"
linkTitle: "Keycloak"
description: "Configuring Stronghold login through Keycloak over OIDC: the Keycloak client and group mapper, the oidc method configuration, a role, mapping Keycloak groups to Identity groups, and login verification."
weight: 30
---

This page describes how to configure Stronghold login through Keycloak using OIDC. For general information about the method, see [OIDC auth method](../../oidc/).

## Keycloak client setup

1. Select or create a Realm and a Client. Open the client **Settings**.
1. **Client Protocol**: `openid-connect`.
1. **Access Type**: `confidential`. In newer Keycloak versions, enable **Client authentication**.
1. Enable **Standard Flow Enabled**.
1. Set **Valid Redirect URIs**:

   ```text
   https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback
   http://localhost:8250/oidc/callback
   ```

   The first URI is used for web interface login, the second one for CLI login. If the method is enabled at a different path, replace `oidc` in the first URI with that path.

1. Click **Save**.
1. Open the **Credentials** tab and note the Client ID and Client Secret.
1. To make Keycloak pass user groups, add a **Group Membership** mapper with the `groups` claim name to the client (or its client scope). Disable **Full group path** if Stronghold needs group names without a path.

## Stronghold setup

1. Enable the OIDC auth method:

   ```bash
   d8 stronghold auth enable oidc
   ```

   By default, the method is enabled at `auth/oidc/`. To use a different path, specify `-path`.

1. Configure the connection to Keycloak. Set `oidc_discovery_url` to the Realm address without `/.well-known/openid-configuration`:

   ```bash
   d8 stronghold write auth/oidc/config \
     oidc_discovery_url="https://keycloak.example.com/realms/myrealm" \
     oidc_client_id="<Client ID>" \
     oidc_client_secret="<Client Secret>" \
     default_role="keycloak"
   ```

   In older Keycloak versions, the Realm address contains the `/auth` prefix, for example `https://keycloak.example.com/auth/realms/myrealm`.

1. Create a role. In `allowed_redirect_uris`, list the same URIs as in the Keycloak client:

   ```bash
   d8 stronghold write auth/oidc/role/keycloak -<<EOF
   {
     "role_type": "oidc",
     "allowed_redirect_uris": [
       "https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback",
       "http://localhost:8250/oidc/callback"
     ],
     "oidc_scopes": ["openid", "profile", "email"],
     "user_claim": "preferred_username",
     "groups_claim": "groups",
     "bound_claims": { "groups": ["stronghold-users", "stronghold-admins"] },
     "token_policies": ["default"],
     "token_ttl": "1h"
   }
   EOF
   ```

   Key role parameters:

   | Parameter | Description |
   |-----------|-------------|
   | `allowed_redirect_uris` | Allowed redirect URIs. They must exactly match the URIs in Keycloak. |
   | `user_claim` | Claim whose value becomes the entity alias name, for example `preferred_username` or `sub`. |
   | `groups_claim` | Claim with the list of user groups (`groups` from the Group Membership mapper). |
   | `bound_claims` | Claims and values that must match for login to succeed. In the example, only members of the `stronghold-users` or `stronghold-admins` groups can log in. |
   | `oidc_scopes` | Requested OIDC scopes. |
   | `token_policies`, `token_ttl` | Policies and TTL of the issued token. |

## Mapping Keycloak groups to Stronghold groups

To assign policies based on Keycloak group membership, create external Identity groups and group aliases. For more on external groups, see [Identity](../../../../concepts/identity/).

1. Create an external group with the required policies and save its ID:

   ```bash
   d8 stronghold write -field=id identity/group \
     name="keycloak-admins" \
     type="external" \
     policies="admin"
   ```

1. Get the accessor of the `oidc/` auth method:

   ```bash
   d8 stronghold auth list -format=json | jq -r '."oidc/".accessor'
   ```

1. Create a group alias. The `name` value must match the group name in the `groups` claim:

   ```bash
   d8 stronghold write identity/group-alias \
     name="stronghold-admins" \
     mount_accessor="<oidc method accessor>" \
     canonical_id="<group ID>"
   ```

On every login and token renewal, Stronghold updates the entity's external group membership based on Keycloak data.

## Verifying login

1. Log in via the CLI:

   ```bash
   d8 stronghold login -method=oidc -path=oidc role=keycloak
   ```

   The CLI opens a browser for Keycloak login and receives the response at `http://localhost:8250/oidc/callback`.

1. Check the policies and metadata of the issued token:

   ```bash
   d8 stronghold token lookup
   ```

1. To verify web interface login, select the **OIDC** method, specify the `keycloak` role if needed, and complete the Keycloak login.

If login fails, check that the redirect URIs match and see the troubleshooting tips in [OIDC auth method](../../oidc/).

## Usage examples

Ready-made examples that use this feature:

- [Single sign-on through a corporate IdP (OIDC)](../../../../examples/access/sso-oidc/)

See all examples in [Usage examples](../../../../examples/).
