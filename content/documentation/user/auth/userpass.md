---
title: "Userpass method"
linkTitle: "Userpass"
weight: 50
---

## Username and password authentication method

The `userpass` auth method allows users to authenticate in Deckhouse Stronghold using a username and password.

Usernames and passwords are configured directly in the auth method via the `users/` path.
The `userpass` method cannot read usernames and passwords from an external source.

All submitted usernames are stored in lowercase.
For example, `Mary` and `mary` refer to the same entry.

## Configuration

Before authenticating users, configure the `userpass` method.
These steps are usually completed by an operator or configuration management tool.

Complete the following steps:

1. Enable the `userpass` auth method:

   ```shell
   d8 stronghold auth enable userpass
   ```

   The method will be enabled at the `auth/userpass` path.

   To enable the method at a different path, use the `-path` flag:

   ```shell
   d8 stronghold auth enable -path=<path> userpass
   ```

1. Create a user who is allowed to authenticate:

   ```shell
   d8 stronghold write auth/<userpass:path>/users/mitchellh \
     password=foo \
     policies=admins
   ```

This creates user `mitchellh` with password `foo` and the `admins` policy.
This is the only required configuration.

## Changing your own password

You can allow a user to change only their own password in the `userpass` method.
To do this, create a policy where the path to the user's password is formed using an entity alias.

### Policy

Use the following policy template:

```hcl
path "auth/userpass/users/{{identity.entity.aliases.<accessor>.name}}/password" {
  capabilities = ["update"]
}
```

Get the `<accessor>` value with the command:

```shell
d8 stronghold read -field=accessor sys/auth/userpass
```

### Admin setup

Enable the `userpass` method, create the policy, and assign it to the user:

```shell
d8 stronghold auth enable userpass
ACCESSOR=$(d8 stronghold read -field=accessor sys/auth/userpass)

d8 stronghold policy write self-change-password - <<EOF
path "auth/userpass/users/{{identity.entity.aliases.${ACCESSOR}.name}}/password" {
  capabilities = ["update"]
}
EOF

d8 stronghold write auth/userpass/users/alice \
  password="OldPass-123!" \
  token_policies="self-change-password"
```

### Changing password as a user

After logging in, the user can change their password:

```shell
d8 stronghold login -method=userpass username=alice password="OldPass-123!"
d8 stronghold write auth/userpass/users/alice/password password="NewPass-456!"
```

If the user tries to change someone else's password, Stronghold returns a `permission denied` error.

### How it works

The `{{identity.entity.aliases.<accessor>.name}}` template automatically substitutes the authenticated user's name.
Therefore, the path always points only to the current user's password.

The template works after logging in via the `userpass` method.

### Verification

To verify the configuration, run the `userpass_self_password_verify.sh` script if it is available in your environment:

```shell
VAULT_ADDR=http://127.0.0.1:8200 VAULT_TOKEN=<root-token> \
  ./userpass_self_password_verify.sh
```

## User lockout

If a user provides incorrect credentials several times in a row, Stronghold stops validating them for a while and immediately returns an access denied error.
This behavior is called user lockout (`user_lockout`).

The time for which a user is locked out is called lockout duration (`lockout_duration`).
After this time expires, the user can log in again.

The number of failed login attempts after which a user is locked out is called lockout threshold (`lockout_threshold`).
The lockout threshold counter resets after a few minutes without login attempts or after a successful login.
The interval after which the counter resets when there are no login attempts is called lockout counter reset (`lockout_counter_reset`).

User lockout helps reduce the risk of password guessing attacks.

The user lockout feature is enabled by default.
Default values:

- `lockout_threshold` — 5 attempts;
- `lockout_duration` — 15 minutes;
- `lockout_counter_reset` — 15 minutes.

You can disable user lockout with the `auth tune` command by setting the `disable_lockout` parameter to `true`.

{{< alert level="warning" >}}
User lockout is supported only by the `userpass`, `ldap`, and `approle` auth methods.
{{< /alert >}}
