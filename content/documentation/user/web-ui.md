---
title: "Web interface"
description: "An overview of the Stronghold web interface: where it is available, how to log in, and which operations on secrets, secrets engines, auth methods, and policies it supports."
weight: 30
---

Stronghold provides a built-in web interface. It works on top of the same HTTP API as the CLI, so the web interface allows only the operations permitted by your token's policies. The web interface is available in all Stronghold editions.

## Where the web interface is available

- **Stronghold in Deckhouse Platform.** The address is built from the `publicDomainTemplate` template: the `%s` key is replaced with `stronghold`. For example, with the `%s.example.com` template, the web interface is available at `https://stronghold.example.com`. See [Accessing the service](../../install/dkp/configuration/#accessing-the-service).
- **Stronghold on Linux.** The web interface is enabled with the `ui = true` parameter in the server configuration and is available on all listeners at the `/ui` path, for example `https://10.0.1.35:8200/ui/`. See [UI section](../../install/standalone/configuration/#ui-section).

Ask your administrator for the exact web interface address of your installation.

## Logging in

1. Open the web interface address.
1. Select an auth method, for example OIDC, LDAP, Userpass, Token, WebAuthn, or SAML.
1. If needed, specify the method path, role name, or other parameters provided by the administrator.
1. Complete authentication. If multi-factor authentication is configured, confirm the login with the second factor.

For step-by-step instructions, see [Configuring access and first login](../get-started/access/). Method-specific notes:

- in Stronghold in DP, the `deckhouse_administrators` role with web interface access through Dex OIDC authentication is created after initialization;
- [WebAuthn](../auth/webauthn/) supports passkey registration and passwordless login directly in the web interface;
- [SAML](../auth/saml/) supports browser login through an external SAML 2.0 identity provider;
- with access to `identity/mfa/method/totp/generate`, users can get their [TOTP MFA](../auth/mfa/totp/) settings in the web interface themselves.

## What you can do in the web interface

- **Secrets.** Create, view, update, and delete secrets, for example in the [KV](../secrets-engines/kv/overview/) version 1 and 2 secrets engines.
- **Secrets engines.** Enable and configure secrets engines, including [LDAP](../secrets-engines/ldap/) and [ClickHouse](../secrets-engines/databases/clickhouse/) in the databases secrets engine, and [KV replication](../secrets-engines/kv/kv-replication/) parameters.
- **Auth methods.** Manage `OIDC` and `AppRole` roles, change the password for the `userpass` method, and configure user lockout for the `ldap`, `userpass`, and `approle` methods.
- **Policies.** Manage [password policies](../../concepts/password-policy/). Managing roles and access policies in the web interface is available in Stronghold EE and Stronghold CSE.
- **Leases.** Revoke leases on the **Access** tab. See [Lease](../../concepts/lease/).
- **System log.** View the server system log.
- **Unsealing.** Unseal storage manually if auto-unseal is not used. For Stronghold on Linux, initialization can be performed in the web interface right after installation.

The web interface supports Russian localization and a dark theme.

The web interface also has a **Tools** page with the wrap, unwrap, lookup, rewrap, hash, and random tools, management of Raft snapshots (including automatic ones), and a namespace management page. Managing audit devices is not provided in the web interface; use the CLI or API.

{{< alert level="info" >}}
Working with secrets in the web interface is covered [in the "Deckhouse Stronghold features overview" course](https://education.flant.ru/course/obzor-vozmozhnostej-deckhouse-stronghold/).
{{< /alert >}}
