---
title: "GitLab OIDC provider"
linkTitle: "Gitlab"
description: "Configuring Stronghold login through GitLab over OIDC: the GitLab application, the oidc method configuration, a role, mapping GitLab groups to Identity groups, and login verification."
weight: 20
---

This page describes how to configure Stronghold login through GitLab using OIDC. For general information about the method, see [OIDC auth method](../../oidc/).

## GitLab application setup

1. Go to **Settings > Applications** of a GitLab user, group, or instance.
1. Fill out the application name and **Redirect URI**. Specify one URI per line:

   ```text
   https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback
   http://localhost:8250/oidc/callback
   ```

   The first URI is used for web interface login, the second one for CLI login (`d8 stronghold login -method=oidc`). If the method is enabled at a different path, replace `oidc` in the first URI with that path.

1. Make sure the `openid` scope is selected. To pass groups and the username, also select `profile` and `email`.
1. Save the application and copy the client **Application ID** and **Secret**.

## Stronghold setup

1. Enable the OIDC auth method:

   ```bash
   d8 stronghold auth enable oidc
   ```

   By default, the method is enabled at `auth/oidc/`. To use a different path, specify `-path`.

1. Configure the connection to GitLab. Set `oidc_discovery_url` to the GitLab address without `/.well-known/openid-configuration`:

   ```bash
   d8 stronghold write auth/oidc/config \
     oidc_discovery_url="https://gitlab.example.com" \
     oidc_client_id="<Application ID>" \
     oidc_client_secret="<Secret>" \
     default_role="gitlab"
   ```

   The `default_role` parameter sets the role used when no role is specified at login.

1. Create a role. In `allowed_redirect_uris`, list the same URIs as in the GitLab application:

   ```bash
   d8 stronghold write auth/oidc/role/gitlab -<<EOF
   {
     "role_type": "oidc",
     "allowed_redirect_uris": [
       "https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback",
       "http://localhost:8250/oidc/callback"
     ],
     "oidc_scopes": ["openid", "profile", "email"],
     "user_claim": "sub",
     "groups_claim": "groups",
     "bound_claims": { "groups": ["devops", "security"] },
     "token_policies": ["default"],
     "token_ttl": "1h"
   }
   EOF
   ```

   Key role parameters:

   | Parameter | Description |
   |-----------|-------------|
   | `allowed_redirect_uris` | Allowed redirect URIs. They must exactly match the URIs in GitLab. |
   | `user_claim` | Claim whose value becomes the entity alias name. For GitLab, you can use `sub` or `nickname`. |
   | `groups_claim` | Claim with the list of user groups. The values are used as Identity group alias names. |
   | `bound_claims` | Claims and values that must match for login to succeed. In the example, only members of the `devops` or `security` groups can log in. |
   | `oidc_scopes` | Requested OIDC scopes. |
   | `token_policies`, `token_ttl` | Policies and TTL of the issued token. |

   <!-- TODO(verify): GitLab claims in the ID token/userinfo (groups, nickname) for the current GitLab version. -->

## Mapping GitLab groups to Stronghold groups

To assign policies based on GitLab group membership, create external Identity groups and group aliases. For more on external groups, see [Identity](../../../../concepts/identity/).

1. Create an external group with the required policies and save its ID:

   ```bash
   d8 stronghold write -field=id identity/group \
     name="gitlab-devops" \
     type="external" \
     policies="devops"
   ```

1. Get the accessor of the `oidc/` auth method:

   ```bash
   d8 stronghold auth list -format=json | jq -r '."oidc/".accessor'
   ```

1. Create a group alias. The `name` value must match the group name in the `groups` claim:

   ```bash
   d8 stronghold write identity/group-alias \
     name="devops" \
     mount_accessor="<oidc method accessor>" \
     canonical_id="<group ID>"
   ```

On every login and token renewal, Stronghold updates the entity's external group membership based on GitLab data.

## Verifying login

1. Log in via the CLI:

   ```bash
   d8 stronghold login -method=oidc -path=oidc role=gitlab
   ```

   The CLI opens a browser for GitLab login and receives the response at `http://localhost:8250/oidc/callback`.

1. Check the policies and metadata of the issued token:

   ```bash
   d8 stronghold token lookup
   ```

1. To verify web interface login, select the **OIDC** method, specify the `gitlab` role if needed, and complete the GitLab login.

If login fails, check that the redirect URIs match and see the troubleshooting tips in [OIDC auth method](../../oidc/).

## Usage examples

Ready-made examples that use this feature:

- [Single sign-on through a corporate IdP (OIDC)](../../../../examples/access/sso-oidc/)

See all examples in [Usage examples](../../../../examples/).
