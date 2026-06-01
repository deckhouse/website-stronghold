---
title: "SAML"
linkTitle: "SAML"
weight: 75
description: "Authenticate to Deckhouse Stronghold through an external SAML 2.0 Identity Provider."
---

The `saml` authentication method lets you authenticate users in Deckhouse Stronghold through an external SAML 2.0 Identity Provider using the Web SSO profile.

In this flow, Stronghold acts as a SAML Service Provider. It validates the response from the Identity Provider and then issues a Stronghold token according to the selected role.

Use this method in the following cases:

- to sign in to the Stronghold web interface through a browser.
- to integrate external applications that start a SAML flow through the HTTP API.

## How it works

The sign-in flow consists of three stages:

1. The client calls `auth/<mount>/sso_service_url` and receives a URL to redirect the user to the Identity Provider, along with a temporary `token_poll_id`.
1. The user authenticates with the Identity Provider. The provider then sends a signed SAML response to `auth/<mount>/callback`.
1. The client exchanges `token_poll_id` and `client_verifier` for a token through `auth/<mount>/token`.

Stronghold supports two client modes:

- `browser` — for web interface scenarios.
- `cli` — for external tools that open the IdP page in a browser and then poll the token issuance endpoint.

## Enable the method

Enable the SAML authentication method:

```shell
d8 stronghold auth enable saml
```

By default, the method is mounted at `auth/saml`.

If needed, use a different mount path:

```shell
d8 stronghold auth enable -path=corp-saml saml
```

## Configuration

Before you begin, configure Stronghold as a SAML Service Provider.

The main configuration parameters are:

- `entity_id` — the Service Provider identifier. It must match the application settings on the Identity Provider side.
- `acs_urls` — the list of allowed callback URLs for Assertion Consumer Service.
- `default_role` — the optional default role. If it is set, you do not need to pass the role explicitly when starting the sign-in flow.
- `idp_metadata_url` — the URL of the Identity Provider metadata.
- `idp_sso_url`, `idp_entity_id`, and `idp_cert` — a manual alternative to `idp_metadata_url`.
- `validate_response_signature` and `validate_assertion_signature` — signature validation parameters. In a production environment, enable both if the IdP can sign both the response and the assertion.
- `verbose_logging` — extended logging for the SAML exchange. Use it only for troubleshooting.

### Configure with IdP metadata

```shell
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://stronghold.example.com/v1/auth/saml/callback" \
  idp_metadata_url="https://idp.example.com/app/stronghold/sso/saml/metadata" \
  default_role="employees" \
  validate_response_signature=true \
  validate_assertion_signature=true
```

### Configure manually

If IdP metadata is unavailable, set the parameters manually:

```shell
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://stronghold.example.com/v1/auth/saml/callback" \
  idp_sso_url="https://idp.example.com/sso" \
  idp_entity_id="https://idp.example.com/entity" \
  idp_cert=@idp-signing-cert.pem \
  validate_response_signature=true \
  validate_assertion_signature=true
```

If `acs_urls` contains multiple addresses, the client must explicitly pass the required `acs_url` when starting the sign-in flow.

## Assertion Consumer Service URLs

The `acs_urls` parameter defines which addresses the Identity Provider can use to return the SAML response after successful user authentication.

When you configure `acs_urls`, make sure each address:

- matches one of the callback URLs allowed in the SAML application on the Identity Provider side.
- points to a Stronghold callback endpoint in the form `.../v1/auth/<mount>/callback` for the selected mount path.
- uses `https://` in the production environment.

If Stronghold is available through multiple public addresses, you can specify multiple ACS URLs:

```shell
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://primary.example.com/v1/auth/saml/callback,https://secondary.example.com/v1/auth/saml/callback" \
  idp_metadata_url="https://idp.example.com/app/stronghold/sso/saml/metadata"
```

If you use `namespaces`, include the namespace path in the API URL or pass the `X-Vault-Namespace` header so that the callback URL matches the actual location of the auth mount.

## Roles

SAML roles define which subjects and attributes from the SAML assertion are allowed to authenticate, and which token parameters apply after sign-in.

The main role parameters are:

- `bound_subjects` — restricts allowed SAML subject values.
- `bound_subjects_type` — the matching type: `string` or `glob`.
- `bound_attributes` — the list of required assertion attributes and expected values.
- `bound_attributes_type` — the attribute value matching type: `string` or `glob`.
- `groups_attribute` — the attribute from which Stronghold creates Identity group aliases.
- `alias_metadata` — static metadata that will be written to the entity alias.
- token parameters such as `token_policies`, `token_ttl`, `token_max_ttl`, `token_period`, and `token_bound_cidrs`.

Example role:

```shell
d8 stronghold write auth/saml/role/employees \
  bound_subjects="*@example.com" \
  bound_subjects_type="glob" \
  bound_attributes=department=platform \
  groups_attribute="memberOf" \
  token_policies="default,developers" \
  token_ttl="1h"
```

If the IdP returns multivalue attributes, `bound_attributes` can match any of the expected values. Attribute names are matched case-insensitively.

### Attribute constraints

After the Identity Provider authenticates the user, Stronghold checks the role constraints against the contents of the SAML assertion:

- `bound_subjects` is matched against the SAML subject.
- `bound_attributes` is matched against the required assertion attributes and allowed values.

This makes it possible to grant access only to selected users or groups from the Identity Provider.

For example:

```shell
d8 stronghold write auth/saml/role/support \
  bound_subjects="*@example.com" \
  bound_subjects_type="glob" \
  bound_attributes=groups="support,engineering" \
  token_policies="support-ro"
```

This role allows sign-in only for users whose subject ends with `@example.com` and whose `groups` attribute contains `support` or `engineering`.

On Microsoft platforms, group membership often appears in the `http://schemas.microsoft.com/ws/2008/06/identity/claims/groups` attribute. In this case, use that exact attribute name in `bound_attributes` and, if needed, in `groups_attribute`.

### Map SAML groups to Identity

If you want SAML group membership to grant Stronghold policies through Identity, set `groups_attribute` in the role and create the corresponding external Identity groups and aliases.

The general flow looks like this:

1. Create an external Identity group with the required policies.
1. Get the `mount accessor` for the SAML auth method.
1. Create a `group alias` whose name matches the value received in the SAML attribute.

Example commands:

```shell
d8 stronghold write identity/group \
  name="SamlDevelopers" \
  type="external" \
  policies="developers"
```

```shell
d8 stronghold auth list -format=json
```

```shell
d8 stronghold write identity/group-alias \
  name="engineering" \
  mount_accessor="<saml-mount-accessor>" \
  canonical_id="<identity-group-id>"
```

With this configuration, a SAML sign-in that returns the value `engineering` in the attribute specified by `groups_attribute` is linked to the external Identity group.

## Sign in through the web interface

The Stronghold web interface supports authentication through SAML.

The typical flow is as follows:

1. Select the SAML sign-in method on the sign-in form.
1. Enter the role name if `default_role` is not set for the mount path.
1. Click `Sign In`.
1. Authenticate with the external Identity Provider.
1. After a successful callback, Stronghold issues a token and signs you in.

## Start the sign-in flow through the API

This flow is useful for custom portals, wrappers, and external CLI tools.

### Generate the verifier and challenge

The `client_challenge` parameter must be a Base64-encoded SHA-256 hash of `client_verifier`.

```shell
verifier="$(uuidgen)"
challenge="$(printf '%s' "$verifier" | openssl dgst -sha256 -binary | base64)"
```

### Request the SSO URL

```shell
curl \
  --request POST \
  --data "{\"role\":\"employees\",\"client_challenge\":\"$challenge\",\"client_type\":\"browser\"}" \
  https://stronghold.example.com/v1/auth/saml/sso_service_url
```

Stronghold returns the following fields in the response:

- `sso_service_url` — the URL to redirect the user to the Identity Provider.
- `token_poll_id` — the identifier for the final token exchange.

If multiple ACS URLs are configured, add the `acs_url` parameter to the request.

### Exchange the temporary state for a token

After the IdP redirects the user back and Stronghold accepts the SAML response, exchange `token_poll_id` for a token:

```shell
curl \
  --request POST \
  --data "{\"token_poll_id\":\"<poll-id>\",\"client_verifier\":\"$verifier\"}" \
  https://stronghold.example.com/v1/auth/saml/token
```

The token is returned in `auth.client_token`.

## Practical notes

Keep the following recommendations in mind:

- Use only `https://` in ACS URLs in the production environment.
- if the IdP can sign both the response and the assertion, enable both signature checks.
- do not leave `verbose_logging=true` enabled in the production environment, because logs may contain sensitive SAML data.
- if the role has `token_bound_cidrs` configured, the final request to `token` must come from an allowed client address.
- a SAML role must define at least one admission condition: `bound_subjects` or `bound_attributes`.
- if multiple `acs_urls` are configured, the client must pass the correct `acs_url` when starting the sign-in flow.
