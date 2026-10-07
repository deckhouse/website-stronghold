---
title: "Single sign-on through a corporate IdP (OIDC)"
linkTitle: "SSO with OIDC"
description: "How to set up single sign-on to Stronghold through a corporate identity provider over OIDC: the general flow, a complete Keycloak example with group-to-policy mapping, login through Dex in DKP, and links to provider guides."
weight: 30
params:
  relatedLinks:
    - title: "OIDC method"
      url: ../../../user/auth/oidc/
    - title: "Keycloak OIDC provider"
      url: ../../../user/auth/oidc/keycloak/
    - title: "GitLab OIDC provider"
      url: ../../../user/auth/oidc/gitlab/
    - title: "Stronghold configuration in DKP"
      url: ../../../install/dkp/configuration/
    - title: "Identity"
      url: ../../../concepts/identity/
    - title: "Break-glass access"
      url: ../break-glass-access/
---

Single sign-on (SSO) lets users log in to Stronghold with their corporate account and lets administrators manage access through groups in the identity provider (IdP). Stronghold supports any OpenID Connect-compatible IdP.

## Goal

- Users log in to the Stronghold web UI and CLI through the corporate IdP.
- Policies are assigned by IdP groups through external Identity groups.
- Only members of specific groups are allowed to log in.

## General flow

![OIDC login flow](../../../images/ex-sso-oidc.en.png)

What to configure in any IdP:

1. An OIDC client (application) of the `confidential` type with a client secret.
1. Redirect URIs: for the web UI, `https://<Stronghold address>/ui/stronghold/auth/<method path>/oidc/callback`; for the CLI, `http://localhost:8250/oidc/callback`.
1. A claim with the list of the user's groups in the ID token.

## Choosing a provider

| Provider | When to use | Guide |
| --- | --- | --- |
| Dex in DP | Stronghold is installed in DP, and users already log in to the cluster through the `user-authn` module | [Login through Dex in DP](#login-through-dex-in-dp) on this page |
| Keycloak | A corporate IdP based on Keycloak, including federation from LDAP or AD | [Keycloak OIDC provider](../../../user/auth/oidc/keycloak/), example below |
| GitLab | Developer accounts and groups are managed in GitLab | [GitLab OIDC provider](../../../user/auth/oidc/gitlab/) |
| Another OIDC provider | Any IdP with a `/.well-known/openid-configuration` discovery document | [OIDC method](../../../user/auth/oidc/) |

## Prerequisites

- Stronghold is available to users over HTTPS, for example `https://stronghold.example.com`.
- The IdP is available to Stronghold nodes over HTTPS, and its certificate is issued by a trusted CA.
- A Stronghold token with permissions to configure auth methods, policies, and Identity.
- Prepared [break-glass access](../break-glass-access/) in case the IdP is unavailable.

## Example: Keycloak

In this example, users from the Keycloak group `stronghold-admins` get the `admin` policy, and users from `stronghold-users` get the `user-read` policy.

### Step 1. Configure the client in Keycloak

1. In the `corp` realm, create a `stronghold` client with the `openid-connect` protocol, and enable **Client authentication** and **Standard flow**.
1. In **Valid Redirect URIs**, specify:

   ```text
   https://stronghold.example.com/ui/stronghold/auth/oidc/oidc/callback
   http://localhost:8250/oidc/callback
   ```

1. Add a **Group Membership** mapper with the `groups` claim name to the client and disable **Full group path**.
1. On the **Credentials** tab, copy the client secret.

For details, see [Keycloak OIDC provider](../../../user/auth/oidc/keycloak/).

### Step 2. Configure the OIDC method

1. Enable the method and configure the connection to the realm:

   ```bash
   d8 stronghold auth enable oidc
   d8 stronghold write auth/oidc/config \
     oidc_discovery_url="https://keycloak.example.com/realms/corp" \
     oidc_client_id="stronghold" \
     oidc_client_secret="<client secret>" \
     default_role="corp"
   ```

1. Create a role. `bound_claims` admits only members of the two groups, and `groups_claim` passes the groups to Identity:

   ```bash
   d8 stronghold write auth/oidc/role/corp -<<EOF
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
     "token_ttl": "1h",
     "token_max_ttl": "8h"
   }
   EOF
   ```

### Step 3. Map groups to policies

For each group, create an external Identity group and an alias. The alias `name` must match the group name in the `groups` claim:

```bash
OIDC_ACCESSOR=$(d8 stronghold auth list -format=json | jq -r '."oidc/".accessor')

ADMINS_ID=$(d8 stronghold write -field=id identity/group \
  name="corp-admins" type="external" policies="admin")
d8 stronghold write identity/group-alias \
  name="stronghold-admins" mount_accessor="$OIDC_ACCESSOR" canonical_id="$ADMINS_ID"

USERS_ID=$(d8 stronghold write -field=id identity/group \
  name="corp-users" type="external" policies="user-read")
d8 stronghold write identity/group-alias \
  name="stronghold-users" mount_accessor="$OIDC_ACCESSOR" canonical_id="$USERS_ID"
```

The `admin` and `user-read` policies must exist beforehand. See [Policy examples](../../../concepts/policy-examples/).

### Step 4. Check login

1. Log in through the CLI; a browser opens the Keycloak login page:

   ```bash
   d8 stronghold login -method=oidc role=corp
   ```

1. Check the token policies:

   ```bash
   d8 stronghold token lookup
   ```

   For a member of `stronghold-admins`, `identity_policies` must contain the `admin` policy.

1. In the web UI, select the **OIDC** method, enter the `corp` role, and log in.

## Login through Dex in DP

If Stronghold is installed as a DP module in `Automatic` mode, after Stronghold is initialized the module itself configures login through [Dex](/products/kubernetes-platform/documentation/v1/modules/user-authn/), the DP authentication provider that gets users and groups from a connected external IdP or LDAP. <!-- TODO(verify): path to the user-authn module documentation -->

- The OIDC method is enabled at the `oidc_deckhouse` path, and the `deckhouse_administrators` role is created for administrators.
- Administrators are set in the `settings.management.administrators` parameter of the `stronghold` ModuleConfig, as groups (`type: Group`) or individual users (`type: User`). A user must belong to at least one group.

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: stronghold
spec:
  enabled: true
  version: 1
  settings:
    management:
      mode: Automatic
      administrators:
      - type: Group
        name: admins
```

User login:

```bash
d8 stronghold login -path=oidc_deckhouse -method=oidc -no-print
```

To grant permissions to users other than administrators, create external Identity groups with aliases on the `oidc_deckhouse/` method accessor, as in step 3 of the Keycloak example. Alias names must match the group names that Dex passes.

<!-- TODO(verify): whether the stronghold module overwrites manual changes in auth/oidc_deckhouse; which roles other than deckhouse_administrators regular users can log in with -->

For details, see [Stronghold configuration](../../../install/dkp/configuration/#access-management) and [Configuring access and first login](../../../user/get-started/access/).

## Troubleshooting

| Symptom | Likely cause | What to do |
| --- | --- | --- |
| The IdP or Stronghold rejects `redirect_uri` | The URIs in the IdP client and in the role's `allowed_redirect_uris` differ | Specify identical URIs in both places, including the method path. |
| Login is denied because of `bound_claims`, or the group list is empty | The IdP does not include groups in the ID token | Check the group mapper in the IdP and the `groups_claim` value. |
| Login succeeds but group policies are missing | The alias name does not match the group name in the claim (for example, the full path `/stronghold-admins` is passed) | Compare `groups` in the ID token with the alias names. In Keycloak, disable **Full group path**. |
| `x509: certificate signed by unknown authority` when writing `auth/oidc/config` | Stronghold does not trust the IdP certificate | Pass the CA chain in the `oidc_discovery_ca_pem` parameter. |
| The IdP is unavailable and nobody can log in | IdP outage or configuration error | Use [break-glass access](../break-glass-access/). |

## Cleanup

```bash
d8 stronghold auth disable oidc
```

Delete the `corp-admins` and `corp-users` external groups via `identity/group/name/<name>` and the `stronghold` client in Keycloak.
