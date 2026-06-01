---
title: "Userpass Method"
linkTitle: "Userpass"
weight: 50
description: "Authentication in Deckhouse Stronghold using a username and password."
---

The `userpass` authentication method allows users to authenticate to Deckhouse Stronghold using a username and password.
This method is suitable when you need a simple local sign-in mechanism, do not use an external identity provider, and want to grant access quickly.

## How the Userpass method works

The `userpass` method stores username and password combinations directly in Stronghold.
It does not read users or passwords from an external source.

This means that:

- user accounts are created and managed inside Stronghold .
- access permissions are defined through policies assigned to the user .
- the method does not depend on an external directory such as `LDAP`, `OIDC`, or `SAML`.

All entered usernames are stored in lowercase.
For example, `Mary` and `mary` are treated as the same account.

## When to use the Userpass method

Use `userpass` in the following cases:

- you need a simple username-and-password login .
- you do not need to connect an external authentication system .
- an administrator wants to manage the user list manually .
- you need a basic sign-in method for a local environment, pilot stand, or demo environment.

If your organization already uses a corporate identity system, an external authentication method is usually a better choice.

## Method configuration

Before users can authenticate, enable the method and create users.

### Enable at the default path

Run the following command:

```shell
d8 stronghold auth enable userpass
```

After that, the method becomes available at the `auth/userpass` path.

### Enable at a custom path

If you need to mount the method at a different path, use the `-path` parameter:

```shell
d8 stronghold auth enable -path=<path> userpass
```

This can be useful if a single installation needs several independent sign-in entry points with different settings.

## Create a user

After enabling the method, create a user that is allowed to sign in.

```shell
d8 stronghold write auth/<userpass:path>/users/mitchellh \
  password=foo \
  policies=admins
```

As a result, a user named `mitchellh` is created with the password `foo` and the `admins` policy assigned.

{% alert level="info" %}
The `<userpass:path>` value must match the actual mount path of the method.
If the method is enabled at the default path, use `userpass`, and the full write path will be `auth/userpass/users/<username>`.
{% endalert %}

## What the user receives after sign-in

After successful authentication, Stronghold issues a token with the user's policies attached.
This token is used for all subsequent requests to Stronghold.
The available actions are determined by the assigned policies.

## User lockout

The `userpass` method supports the `user_lockout` mechanism, which protects against password guessing.
If a user enters invalid credentials several times in a row, Stronghold temporarily stops verifying the credentials and immediately returns an access denied response.

The following parameters control user lockout behavior:

- `lockout_threshold` — Number of failed sign-in attempts after which the user is locked out. The default value is `5`.
- `lockout_duration` — Amount of time the user remains locked out. The default value is `15 minutes`.
- `lockout_counter_reset` — Amount of time after which the failed attempt counter is reset if there are no new attempts. The default value is `15 minutes`.

The user lockout feature is enabled by default.
The failed attempt counter is also reset after a successful sign-in.

### Disable lockout

You can disable user lockout with the `auth tune` command by setting `disable_lockout=true`.

{% alert level="warning" %}
The `user_lockout` mechanism is supported only by the `userpass`, `ldap`, and `approle` methods.
{% endalert %}

## Best practices

Keep the following recommendations in mind:

- use `userpass` when you need simple local sign-in without an external identity provider .
- do not use weak passwords .
- assign only the required policies to users .
- consider the risks of local account storage if the method is used in a production environment .
- if you have many users, consider switching to an external authentication method.
