---
title: "SAML method"
linkTitle: "SAML"
weight: 75
---

## SAML auth method

The `saml` auth method lets users authenticate to Deckhouse Stronghold through an external `SAML 2.0` Identity Provider using the `Web SSO` profile.

This method is suitable for browser-based sign-in through the Stronghold UI and for custom integrations that start the SAML login flow through the HTTP API. Stronghold acts as a SAML Service Provider, validates the response from the Identity Provider, and then issues a Stronghold token according to the matched role.

## How it works

The login flow has three stages:

1. A client requests `auth/<mount>/sso_service_url` and receives an IdP redirect URL plus a temporary `token_poll_id`.
1. The user completes authentication at the Identity Provider, which sends a signed SAML response to `auth/<mount>/callback`.
1. The client exchanges `token_poll_id` and `client_verifier` at `auth/<mount>/token` and receives a Stronghold token.

Stronghold supports two client modes:

- `browser` for UI-style login flows;
- `cli` for external tools that open the IdP page in a browser and then poll for the token.

## Enable the method

```shell
d8 stronghold auth enable saml
```

By default the method is mounted at `auth/saml`. You can mount it at a custom path if needed:

```shell
d8 stronghold auth enable -path=corp-saml saml
```

## Configuration

Before users can log in, configure Stronghold as a SAML Service Provider.

Main configuration parameters:

- `entity_id` is the Service Provider entity ID that must match the application settings in the Identity Provider.
- `acs_urls` is the list of allowed Assertion Consumer Service callback URLs.
- `default_role` is optional. If set, users can start login without explicitly passing a role.
- `idp_metadata_url` lets Stronghold fetch IdP settings from metadata.
- `idp_sso_url`, `idp_entity_id`, and `idp_cert` can be used instead of `idp_metadata_url` for manual configuration.
- `validate_response_signature` and `validate_assertion_signature` control signature validation. In production, enable both when the Identity Provider supports signing both objects.
- `verbose_logging` adds SAML exchange details to logs and should only be used for troubleshooting.

### Configure via IdP metadata

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

Use manual configuration when metadata is unavailable:

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

If multiple `acs_urls` are configured, the client must explicitly choose which one to use when starting the login flow.

### Assertion consumer service URLs

The `acs_urls` parameter defines where the Identity Provider is allowed to send the SAML response after the user completes authentication.

When configuring `acs_urls`, make sure that each value:

- matches one of the callback URLs configured for the SAML application on the Identity Provider side;
- points to the Stronghold SAML callback endpoint `.../v1/auth/<mount>/callback` for the chosen mount path;
- uses `https://` in production environments.

If Stronghold is exposed through multiple public addresses, you can configure several ACS URLs:

```shell
d8 stronghold write auth/saml/config \
  entity_id="https://stronghold.example.com/v1/auth/saml" \
  acs_urls="https://primary.example.com/v1/auth/saml/callback,https://secondary.example.com/v1/auth/saml/callback" \
  idp_metadata_url="https://idp.example.com/app/stronghold/sso/saml/metadata"
```

If you use namespaces, include the namespace path in the API URL or pass the `X-Vault-Namespace` header so that the resulting callback URL matches the real auth mount location.

## Roles

SAML roles define which SAML subjects and attributes are allowed to authenticate and which token settings are applied after login.

Main role parameters:

- `bound_subjects` restricts allowed SAML subjects.
- `bound_subjects_type` defines matching mode: `string` or `glob`.
- `bound_attributes` requires specific assertion attributes and values.
- `bound_attributes_type` defines matching mode for attribute values: `string` or `glob`.
- `groups_attribute` maps a SAML attribute to Stronghold Identity group aliases.
- `alias_metadata` copies static metadata into the generated entity alias.
- token parameters such as `token_policies`, `token_ttl`, `token_max_ttl`, `token_period`, and `token_bound_cidrs` work the same way as in other auth methods.

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

If your IdP sends multi-value attributes, `bound_attributes` can match any of the expected values. Matching is case-insensitive for attribute names.

### Bound attributes

After the user is authenticated by the Identity Provider, Stronghold checks the role restrictions against the SAML assertion:

- `bound_subjects` matches the SAML subject;
- `bound_attributes` matches required assertion attributes and allowed values.

This lets you allow access only to selected users or groups from the Identity Provider. For example:

```shell
d8 stronghold write auth/saml/role/support \
  bound_subjects="*@example.com" \
  bound_subjects_type="glob" \
  bound_attributes=groups="support,engineering" \
  token_policies="support-ro"
```

The role above authorizes users whose subject ends with `@example.com` and whose `groups` attribute contains either `support` or `engineering`.

For Microsoft identity platforms, group membership is often sent in the attribute `http://schemas.microsoft.com/ws/2008/06/identity/claims/groups`. In that case, use that attribute name in `bound_attributes` and, if needed, in `groups_attribute`.

### Link SAML groups to Stronghold Identity groups

If you want SAML group membership to grant Stronghold policies through Identity, use `groups_attribute` in the role and create matching external Identity groups and aliases.

Example flow:

1. Create an external Identity group with the required policies.
1. Read the auth mount accessor for the SAML auth method.
1. Create a group alias whose name matches the value coming from the SAML assertion.

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

With this setup, a SAML login that returns `engineering` in the attribute referenced by `groups_attribute` will be linked to the external Identity group.

## Sign in through the UI

The Stronghold UI supports SAML login directly.

Typical flow:

1. Select the `SAML` auth method in the login form.
1. Enter the role name if your mount does not use `default_role`.
1. Click `Sign In`.
1. Complete authentication at the external Identity Provider.
1. After the callback completes, Stronghold issues the token and signs the user in.

## Start login through the API

The API is useful for custom portals, wrappers, and external CLI tools.

### 1. Generate a verifier and challenge

`client_challenge` must be a base64-encoded SHA-256 digest of `client_verifier`.

```shell
verifier="$(uuidgen)"
challenge="$(printf '%s' "$verifier" | openssl dgst -sha256 -binary | base64)"
```

### 2. Request the SSO URL

```shell
curl \
  --request POST \
  --data "{\"role\":\"employees\",\"client_challenge\":\"$challenge\",\"client_type\":\"browser\"}" \
  https://stronghold.example.com/v1/auth/saml/sso_service_url
```

The response contains:

- `sso_service_url` for redirecting the user to the Identity Provider;
- `token_poll_id` for the final token exchange.

If several ACS URLs are configured, include `acs_url` in this request.

### 3. Exchange the temporary flow state for a token

After the IdP sends the user back and Stronghold accepts the SAML response, exchange the poll ID for a token:

```shell
curl \
  --request POST \
  --data "{\"token_poll_id\":\"<poll-id>\",\"client_verifier\":\"$verifier\"}" \
  https://stronghold.example.com/v1/auth/saml/token
```

The token is returned in `auth.client_token`.

## Practical notes

- Use `https://` ACS URLs in production.
- If your IdP can sign both the SAML response and the assertion, enable both signature validation options.
- Do not leave `verbose_logging=true` enabled in production, because logs may contain sensitive SAML data.
- If a role has `token_bound_cidrs`, the final token exchange must come from an allowed client address.
- A SAML role must define at least one bound condition: `bound_subjects` or `bound_attributes`.
- If you configure multiple `acs_urls`, make sure the client sends the correct `acs_url` when starting the login flow.
