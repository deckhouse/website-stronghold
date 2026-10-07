---
title: "Authentication methods"
linkTitle: "Overview"
description: Auth methods are mountable methods that perform authentication for Stronghold.
weight: 5
---

Auth methods are the components in Stronghold that perform authentication and are
responsible for assigning identity and a set of policies to a user. In all cases,
Stronghold will enforce authentication as part of the request processing. In most cases,
Stronghold will delegate the authentication administration and decision to the relevant configured
external auth method (e.g., Kubernetes).

Stronghold also supports `WebAuthn` for passwordless authentication with `FIDO2` authenticators and `passkeys`.

Stronghold also supports `SAML` for browser-based Web SSO through an external `SAML 2.0` identity provider.

Having multiple auth methods enables you to use an auth method that makes the
most sense for your use case of Stronghold and your organization.

For example, on developer machines, the [Userpass](./userpass/)
is easiest to use. But for servers the [AppRole](./approle/)
method is the recommended choice.

To learn more about authentication, see the
[authentication concepts page](../../concepts/auth/).

## Available auth methods

| Method | Type (`auth enable`) | Purpose | Editions |
|--------|----------------------|---------|----------|
| [AppRole](../approle/) | `approle` | Authentication of applications and automation with `role_id` and `secret_id`. | Not specified |
| [JWT](../jwt/) | `jwt` | Login with a JWT whose signature is verified with keys or JWKS. | Stronghold, Stronghold EE, Stronghold CSE |
| [OIDC](../oidc/) | `oidc` (`jwt` plugin) | Browser login through an OIDC provider: [GitLab](../oidc/gitlab/), [Keycloak](../oidc/keycloak/), [Kubernetes](../oidc/kubernetes/). | Stronghold, Stronghold EE, Stronghold CSE |
| [Kubernetes](../kubernetes/) | `kubernetes` | Login with a Kubernetes service account token. | Stronghold, Stronghold EE, Stronghold CSE |
| [LDAP](../ldap/) | `ldap` | Login with an LDAP or Active Directory account. | Stronghold, Stronghold EE, Stronghold CSE |
| [SAML](../saml/) | `saml` | Browser-based Web SSO through an external SAML 2.0 identity provider. | Stronghold EE |
| [Token](../token/) | `token` | Login with a Stronghold token. Built in and cannot be disabled. | Stronghold, Stronghold EE, Stronghold CSE |
| [Userpass](../userpass/) | `userpass` | Login with a username and password stored in Stronghold. | Not specified |
| [WebAuthn](../webauthn/) | `webauthn` | Passwordless login with FIDO2 authenticators and passkeys. | Stronghold, Stronghold EE |
| Kerberos | `kerberos` | Kerberos (SPNEGO) login. There is no dedicated guide yet; see the [auth methods API reference](../../../reference/api/auth/). | Not specified |

Multi-factor authentication ([TOTP](../mfa/totp/), [Multifactor](../mfa/multifactor/)) is not a separate login method: it adds a second factor to the methods above. See [2FA / MFA](../mfa/).

Editions are listed according to the [Editions](../../../about/editions/) page.

<!-- TODO(verify): edition availability of approle, userpass, kerberos, and MFA (not listed on the editions page); whether the kerberos method is shipped (it appears only in the API reference). -->

## Enabling/Disabling auth methods

Auth methods can be enabled/disabled using the CLI or the API.

{{< tabs name="stronghold_cmd_16299" >}}
{{% tab name="Stronghold in DKP" %}}

```shell-session
d8 stronghold auth enable userpass
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell-session
stronghold auth enable userpass
```

{{% /tab %}}
{{< /tabs >}}

When enabled, auth methods are similar to [secrets engines](../secrets-engines/):
they are mounted within the Stronghold mount table and can be accessed
and configured using the standard read/write API. All auth methods are mounted underneath the `auth/` prefix.

By default, auth methods are mounted to `auth/<type>`. For example, if you
enable "ldap", then you can interact with it at `auth/ldap`. However, this
path is customizable, allowing users with advanced use cases to mount a single
auth method multiple times.

{{< tabs name="stronghold_cmd_73169" >}}
{{% tab name="Stronghold in DKP" %}}

```shell-session
d8 stronghold auth enable -path=my-login userpass
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell-session
stronghold auth enable -path=my-login userpass
```

{{% /tab %}}
{{< /tabs >}}

When an auth method is disabled, all users authenticated via that method are
automatically logged out.

## External auth method considerations

When using an external auth method (e.g., Kubernetes), Stronghold will call the external service
at the time of authentication and for subsequent token renewals. If the status
of an entity changes in the external system (e.g., an account expires or is
disabled), Stronghold denies requests to **renew** tokens associated with the entity.
However, any existing token remain valid for the original grant period unless
they are explicitly revoked within Stronghold. Operators should set appropriate
[token TTLs](../../concepts/tokens#general-case) when using external
authN methods.
