---
title: "Authentication methods overview"
linkTitle: "Overview"
weight: 10
description: "Overview of authentication methods in Deckhouse Stronghold and recommendations for choosing the right scenario."
---
**Authentication methods in Deckhouse Stronghold** define how a user, application, or service proves its identity and gets access to the secrets storage. After successful authentication, Deckhouse Stronghold issues a token that the client uses for subsequent requests.

Deckhouse Stronghold supports multiple authentication methods at the same time. This makes it possible to use different access scenarios within a single installation: user sign-in through a browser or CLI, authentication for applications and services, integration with corporate directories, and additional authentication factors.

When choosing an authentication method, consider not only the sign-in mechanism but also who exactly gets access to Deckhouse Stronghold:

- a user.
- an application.
- a service.
- an application in Kubernetes.
- an external tool or integration.

{{< alert level="info" >}}
A specific authentication method must be enabled and configured by an administrator in advance. If the required method is unavailable in your installation, contact your Deckhouse Stronghold administrator.
{{< /alert >}}

## Choosing an authentication method

In general, authentication methods can be divided into three groups:

- **user authentication methods** — for interactive sign-in through a browser, CLI, or web interface.
- **machine authentication methods** — for applications, services, and automation.
- **additional authentication factors** — for strengthening an existing sign-in flow.

If you are not sure which option to use, follow these general recommendations:

- if you sign in as a user through a browser or a corporate account, [`OIDC`](./oidc/overview), [`SAML`](./saml), [`LDAP`](./ldap), [`Userpass`](./userpass), or [`WebAuthn`](./web-auth-n) is usually appropriate.
- if access is required for an application or a service, [`AppRole`](./approle), [`Kubernetes`](./kubernetes), [`JWT`](./jwt), or [`Token`](./token) is more commonly used.
- if sign-in requires an additional confirmation step, use [`MFA`](./mfa).

## Methods for users

### OIDC

[OpenID Connect (OIDC)](./oidc/overview) is one of the main sign-in methods for users. It allows users to authenticate through an external OIDC provider by using a browser and supports both UI and CLI scenarios.

This method is typically chosen in the following cases:

- your organization already uses an OIDC provider.
- you need single sign-on with a corporate account.
- you need a convenient user flow for both browser and CLI access.

If you need to understand how Deckhouse Stronghold works with specific OIDC providers, see [OIDC providers](./oidc/oidc-providers/overview).

### SAML

[Security Assertion Markup Language (SAML)](./saml) allows users to authenticate through an external `SAML 2.0` identity provider by using the `Web SSO` profile.

This method is suitable in the following cases:

- your organization already uses SAML infrastructure.
- you need browser-based sign-in through a corporate identity provider.
- you need integration with systems that rely on SAML.

### LDAP

[Lightweight Directory Access Protocol (LDAP)](./ldap) is used for authentication through an existing LDAP directory or Active Directory.

This method is suitable in the following cases:

- user accounts are already managed in LDAP.
- you need to map users and groups from the directory to Deckhouse Stronghold policies.
- you need integration with a corporate directory service.

### Userpass

[`Userpass`](./userpass) is a built-in authentication method based on a username and password.

It is suitable in the following cases:

- you need a simple local sign-in method.
- no external identity provider is used.
- you need a basic authentication scenario for testing, internal environments, or a limited number of users.

### WebAuthn

[`WebAuthn`](./web-auth-n) allows users to authenticate with `FIDO2`-compatible authenticators and `passkeys`.

This method is especially useful in the following cases:

- you need passwordless sign-in.
- you need a modern browser-based user flow.
- you want to strengthen account protection by using a hardware or platform authenticator.

### Token

[`Token`](./token) is a built-in token-based authentication method. It is available in Deckhouse Stronghold by default and can be used both by users and automation tools.

This method is commonly used in the following cases:

- the user already has a valid token.
- you need to sign in without repeating authentication through an external method.
- you use an automation scenario or work directly with the Deckhouse Stronghold API.

## Methods for applications and services

### AppRole

[`AppRole`](./approle) is designed for authentication of applications and services. It is especially well suited for machine scenarios where an application must receive access based on a role and additional restrictions.

This method is typically chosen in the following cases:

- access is needed for an application rather than a person.
- you need a flexible and manageable service authentication flow.
- you need to separate the delivery of `role_id` and `secret_id`.
- you use automation or service-to-service communication.

### JWT

[JSON Web Token (JWT)](./jwt) is used when a client can already provide a JWT that Deckhouse Stronghold must validate.

This method is suitable in the following cases:

- the application already receives a JWT from a trusted issuer.
- you need authentication without a browser-based OIDC flow.
- you need to use claim checks and claim-based bindings.

### Kubernetes

[`Kubernetes`](./kubernetes) is designed for authentication of applications in Kubernetes by using a `ServiceAccount` token.

This method is especially useful in the following cases:

- the application runs in Kubernetes.
- Deckhouse Stronghold is used to deliver secrets into pods.
- you need to issue Deckhouse Stronghold tokens based on Kubernetes identity.

## Additional protection layer

### MFA

[Multi-Factor Authentication (MFA)](./mfa) is not a separate primary sign-in method. It is an additional verification factor that strengthens an already configured authentication method.

Deckhouse Stronghold supports different MFA scenarios, including the following:

- [`MULTIFACTOR Ldap Adapter`](./mfa/mfla) — for two-factor protection of users in LDAP-based scenarios.
- [`Time-Based One-Time Password (TOTP)`](./mfa/totp) — for validating one-time codes.

MFA should be used in the following cases:

- you need to strengthen user sign-in protection.
- your organization has increased security requirements.
- you need to add a second factor for methods such as LDAP, Userpass, or other user authentication scenarios.

## Typical scenarios

The following simplified recommendations can help you choose a method:

- For user sign-in with a corporate account, [`OIDC`](./oidc/overview) or [`SAML`](./saml) is usually appropriate.
- For user sign-in through LDAP or Active Directory, use [`LDAP`](./ldap).
- For a simple local username-and-password sign-in flow, use [`Userpass`](./userpass).
- For passwordless sign-in, use [`WebAuthn`](./web-auth-n).
- For application or service authentication, use [`AppRole`](./approle).
- For application authentication in Kubernetes, use [`Kubernetes`](./kubernetes).
- For authentication with an already issued JWT, use [`JWT`](./jwt).
- For work with an existing Deckhouse Stronghold token, use [`Token`](./token).
- To strengthen an existing sign-in flow, use [`MFA`](./mfa).

## External authentication methods

In many scenarios, Deckhouse Stronghold delegates authentication to an external method such as `JWT`, `OIDC`, `Kubernetes`, or `LDAP`. External systems such as Deckhouse Kubernetes Platform, Keycloak, Blitz Identity Provider, or Active Directory can be used as identity data sources for users and applications.

When an external authentication method is used, Deckhouse Stronghold calls the external service during sign-in and during subsequent token renewal operations. If the entity status changes in the external system, for example if the account is disabled or expires, Deckhouse Stronghold rejects token renewal requests associated with that entity. At the same time, already issued tokens remain valid until their original expiration time unless they are explicitly revoked in Deckhouse Stronghold.

{{< alert level="warning" >}}
When you use external authentication methods, set an appropriate token TTL. This helps limit the lifetime of already issued tokens if access in the external system is changed or revoked.
{{< /alert >}}

If an authentication method is disabled, all users and applications that used this method lose the ability to sign in again.
