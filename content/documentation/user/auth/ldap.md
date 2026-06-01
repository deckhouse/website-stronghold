---
title: "LDAP"
linkTitle: "LDAP"
weight: 70
description: "LDAP authentication in Deckhouse Stronghold."
---

The `ldap` authentication method lets you authenticate users against an existing LDAP server by using their corporate credentials. This simplifies integrating Deckhouse Stronghold into environments where LDAP or Active Directory is already in use. Mapping LDAP users and LDAP groups to Stronghold policies is configured through the `users/` and `groups/` paths. The page follows an overview-style structure: it explains the purpose of the method, its place in the system, the main operating principles, and practical configuration details.

## When to use LDAP

The `ldap` method is suitable in the following cases:

- LDAP or Active Directory is already used in the organization.
- users must authenticate with corporate credentials.
- LDAP groups must be mapped to Stronghold policies.
- Stronghold must be integrated into an existing LDAP infrastructure.

## Authentication

### Via CLI

```bash
d8 stronghold login -method=ldap username=mitchellh
```

After you run the command, Stronghold prompts for the password. If authentication succeeds, the service returns a token with the assigned policies.

### Via API

```bash
curl \
  --request POST \
  --data '{"password":"foo"}' \
  http://127.0.0.1:8200/v1/auth/ldap/login/mitchellh
```

Example response:

```json
{
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": null,
  "auth": {
    "client_token": "c4f280f6-fdb2-18eb-89d3-589e2e834cdb",
    "policies": [
      "admins"
    ],
    "metadata": {
      "username": "mitchellh"
    },
    "lease_duration": 0,
    "renewable": false
  }
}
```

## Enable the method

Before users can authenticate, enable the method:

```bash
d8 stronghold auth enable ldap
```

## Configuration

After enabling the method, configure the LDAP server connection, user lookup settings, and the way group membership is determined.

### Connection parameters

Use the following main parameters:

- `url` — LDAP server address. You can specify a single URL, for example `ldap://ldap.example.com` or `ldaps://ldap.example.com:636`, or a comma-separated list of URLs.
- `starttls` — enables `StartTLS` after an unencrypted connection is established.
- `insecure_tls` — disables LDAP server certificate verification.
- `certificate` — CA certificate used to verify the LDAP server certificate in x509 PEM format.
- `client_tls_cert` — client certificate in x509 PEM format.
- `client_tls_key` — client certificate key in x509 PEM format.

### User lookup

Stronghold supports two ways to identify the LDAP user object:

- searching for the user in the directory.
- using a UPN (User Principal Name).

#### Authenticated search

For authenticated search, use the following parameters:

- `binddn` — object name used for binding when searching for users and groups.
- `bindpass` — password for `binddn`.
- `userdn` — base DN for user lookup.
- `userattr` — LDAP user attribute that corresponds to the username.
- `userfilter` — Go template used to build the user lookup filter.

By default, the filter is `({{.UserAttr}}={{.Username}})`. If `upndomain` is set, the default filter becomes `(userPrincipalName={{.Username}}@UPNDomain)`.

{% alert level="warning" %}
If you specify `userfilter`, include either the templated value `{{.UserAttr}}` or a literal value that matches `userattr`. This ensures the lookup returns a unique result and prevents collisions during login.
{% endalert %}

#### Anonymous search

For anonymous search, use the following parameters:

- `discoverdn` — enables anonymous connection to determine the user's bind DN.
- `userdn` — base DN for user lookup.
- `userattr` — LDAP user attribute that corresponds to the username.
- `userfilter` — Go template used to build the user lookup filter.
- `deny_null_bind` — prevents authentication bypass with an empty password. The default value is `true`.
- `anonymous_group_search` — enables anonymous connections for group lookup. The default value is `false`.

{% alert level="warning" %}
If you specify `userfilter`, include either the templated value `{{.UserAttr}}` or a literal value that matches `userattr`. This ensures the lookup returns a unique result and prevents collisions during login.
{% endalert %}

#### Using UPN in Active Directory

The `upndomain` parameter sets the domain used to build a UPN in the `[username]@UPNDomain` format. This is especially useful when working with Active Directory.

### Determine group membership

After the user is authenticated, Stronghold must determine which LDAP groups the user belongs to.

Two approaches are supported:

- searching for groups that include the user.
- searching for the user object and reading the attribute
  that contains the list of groups.

Use the following parameters:

- `groupfilter` — Go template used to build the group membership query.
- `groupdn` — base DN for group lookup.
- `groupattr` — LDAP attribute used to determine group names.

By default, `groupfilter` is compatible with several common directory schemas:

`(|(memberUid={{.Username}})(member={{.UserDN}})(uniqueMember={{.UserDN}}))`.

To support nested groups in Active Directory, use the following filter:

`(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))`

### Additional parameters

The following additional parameters are also supported:

- `dereference_aliases` — controls how alias objects are dereferenced during search. Supported values are `never`, `finding`, `searching`, and `always`.
- `username_as_alias` — if set to `true`, uses the username passed during login as the alias name.
- `max_page_size` — if the value is greater than `0`, Stronghold uses paged search on the LDAP server.
- `upndomain` — sets the UPN domain used for authentication.
- `discoverdn` — enables automatic detection of the user's bind DN.

Terminology on this page uses canonical forms from the glossary. For example, `alias` is used as the canonical term [1].

## Configuration examples

### Active Directory with StartTLS and nested groups

```bash
d8 stronghold write auth/ldap/config \
  url="ldap://ldap.example.com" \
  userdn="ou=Users,dc=example,dc=com" \
  groupdn="ou=Groups,dc=example,dc=com" \
  groupfilter="(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))" \
  groupattr="cn" \
  upndomain="example.com" \
  certificate=@ldap_ca_cert.pem \
  insecure_tls=false \
  starttls=true
```

### Authenticated search with a service account

```bash
d8 stronghold write auth/ldap/config \
  url="ldap://ldap.example.com" \
  userattr="sAMAccountName" \
  userdn="ou=Users,dc=example,dc=com" \
  groupdn="ou=Users,dc=example,dc=com" \
  groupfilter="(&(objectClass=person)(uid={{.Username}}))" \
  groupattr="memberOf" \
  binddn="cn=stronghold,ou=users,dc=example,dc=com" \
  bindpass='My$ecrt3tP4ss' \
  certificate=@ldap_ca_cert.pem \
  insecure_tls=false \
  starttls=true
```

### Anonymous search with automatic bind DN discovery

```bash
d8 stronghold write auth/ldap/config \
  url="ldaps://ldap.example.com" \
  userattr="uid" \
  userdn="ou=Users,dc=example,dc=com" \
  discoverdn=true \
  groupdn="ou=Groups,dc=example,dc=com" \
  certificate=@ldap_ca_cert.pem \
  insecure_tls=false \
  starttls=true
```

## Map LDAP groups to Stronghold policies

After you configure the connection, map LDAP groups to Stronghold policies.

Example:

```bash
d8 stronghold write auth/ldap/groups/scientists policies=foo,bar
```

This command maps the `scientists` LDAP group to the `foo` and `bar` Stronghold policies.

You can also add a specific LDAP user to an additional Stronghold group and assign separate policies to that user:

```bash
d8 stronghold write auth/ldap/groups/engineers policies=foobar
d8 stronghold write auth/ldap/users/tesla groups=engineers policies=zoobar
```

## Verify the result

After configuration, verify that the user can log in:

```bash
d8 stronghold login -method=ldap username=tesla
```

If everything is configured correctly, the user receives a token with policies assigned through LDAP groups and additional mappings.

## How policy application works

The `user → policies` mapping is evaluated when the token is created. If the user's LDAP group membership changes later, this does not affect tokens that have already been issued.

To apply the changes, do the following:

- revoke the old tokens.
- authenticate again.

## User lockout

The `ldap` method supports the `user_lockout` mechanism. If a user enters invalid credentials several times in a row, Stronghold temporarily stops checking the password and immediately denies access. This helps reduce the risk of password guessing. User lockout is enabled by default.

Default values:

- `lockout_threshold` — 5 attempts.
- `lockout_duration` — 15 minutes.
- `lockout_counter_reset` — 15 minutes.

You can disable this feature with `auth tune` by passing `disable_lockout=true`.

{% alert level="warning" %}
This mechanism is supported only by the `userpass`, `ldap`, and `approle` methods.
{% endalert %}

## DN escaping

It is important to correctly escape the user DN, search DNs, and other DN values. The `ldap` method automatically escapes the username passed during login when substituting it into the bind DN according to RFC 4514.

When using Active Directory, take additional escaping rules into account. For example, the `#` character may need to be escaped regardless of its position in the DN. If the username or configured DNs contain special characters, verify that escaping is correct.

For more information, see [RFC 4514](https://www.ietf.org/rfc/rfc4514.txt) and [Microsoft recommendations for escaping characters in Active Directory](http://social.technet.microsoft.com/wiki/contents/articles/5312.active-directory-characters-to-escape.aspx).

## Practical recommendations

Use the following recommendations:

- For a production environment, use `ldaps` or `StartTLS`.
- do not enable `insecure_tls` unless absolutely necessary.
- before using complex `userfilter` and `groupfilter` values, make sure the lookup returns a unique and expected result.
- if you use Active Directory, account for DN escaping specifics.
- after configuration, always verify which policies the user actually receives at login.
