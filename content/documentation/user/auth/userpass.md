---
title: "Username and password authentication method"
linkTitle: "userpass"
weight: 50
---

The `userpass` auth method allows users to authenticate in Deckhouse Stronghold using a username and password.

## Method features

When using the `userpass` method, keep the following features in mind:

- Usernames and passwords are specified directly in the authentication method at the path `auth/userpass/users/`.
- The `userpass` method cannot read usernames and passwords from an external source.
- Entered usernames are converted to lowercase. For example, `Mary` and `mary` are equivalent entries.

## Configuration

To allow users to authenticate, configure the `userpass` method.
These steps are usually completed by an operator or configuration management tool.

To configure authentication using the `userpass` method, follow these steps:

1. Enable the `userpass` auth method:

   ```shell
   d8 stronghold auth enable userpass
   ```

   The method will be enabled at the `auth/userpass` path.

   To enable the method at a different path, use the `-path` flag:

   ```shell
   d8 stronghold auth enable -path=<userpass_mount_path> userpass
   ```

1. Create a user who is allowed to authenticate:

   ```shell
   d8 stronghold write auth/<userpass_mount_path>/users/alice \
     password=Pass-123! \
     token_policies=admins
   ```

This creates user `alice` with password `Pass-123!` and the `admins` [policy](../../concepts/policy/).

### User authentication using the userpass method

Example command for user authentication using the `userpass` method:

```shell
d8 stronghold login -method=userpass username=alice password="Pass-123!"
```

### User lockout

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

## Changing your own password

You can allow a user to change only their own password in the `userpass` method.
To do this, create a policy where the path to the user's password is formed using an entity alias.

### Creating a policy

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

The `{{identity.entity.aliases.<accessor>.name}}` template automatically substitutes the authenticated user's name.
Therefore, the path always points only to the current user's password.

The template works after logging in via the `userpass` method.

### Allow a user to change their own password

To allow a user to change their own password using the `userpass` method, follow these steps:

{{< tabs >}}
{{% tab "If the userpass method is already enabled" %}}

1. Get the unique method identifier:

   ```shell
   ACCESSOR=$(d8 stronghold read -field=accessor sys/auth/userpass)
   ```

1. Create a policy that allows a user authenticated via `userpass` to change their password:

   ```shell
   d8 stronghold policy write self-change-password - <<EOF
   path "auth/userpass/users/{{identity.entity.aliases.${ACCESSOR}.name}}/password" {
     capabilities = ["update"]
   }
   EOF
   ```

1. Assign the policy to the user you want to allow to change their password:

{{< tabs >}}
{{% tab "If the user exists" %}}

```shell
d8 stronghold write auth/userpass/users/alice/policies \
  token_policies="self-change-password"
```

This example will apply the `self-change-password` policy to the existing user `alice`.

{{% /tab %}}
{{% tab "If the user does not exist" %}}

```shell
d8 stronghold write auth/userpass/users/alice \
  password="OldPass-123!" \
  token_policies="self-change-password"
```

This example will create a user named `alice`, grant them authentication via `userpass`, and assign the `self-change-password` policy to them.

{{% /tab %}}
{{< /tabs >}}

{{% /tab %}}
{{% tab "If the userpass method is not enabled" %}}

1. Enable the `userpass` method:

   ```shell
   d8 stronghold auth enable userpass
   ```

1. Get the unique method identifier:

   ```shell
   ACCESSOR=$(d8 stronghold read -field=accessor sys/auth/userpass)
   ```

1. Create a policy that allows a user authenticated via `userpass` to change their password:

   ```shell
   d8 stronghold policy write self-change-password - <<EOF
   path "auth/userpass/users/{{identity.entity.aliases.${ACCESSOR}.name}}/password" {
     capabilities = ["update"]
   }
   EOF
   ```

1. Assign the policy to the user you want to allow to change their password:

{{< tabs >}}
{{% tab "Если пользователь существует" %}}

```shell
d8 stronghold write auth/userpass/users/alice/policies \
  token_policies="self-change-password"
```

This example will apply the `self-change-password` policy to the existing user `alice`

{{% /tab %}}
{{% tab "Если пользователя не существует" %}}

```shell
d8 stronghold write auth/userpass/users/alice \
  password="OldPass-123!" \
  token_policies="self-change-password"
```

This example will create a user named `alice`, grant them authentication via `userpass`, and assign the `self-change-password` policy to them.

{{% /tab %}}
{{< /tabs >}}

{{% /tab %}}
{{< /tabs >}}

### Changing password as a user

After [authentication](#user-authentication-using-the-userpass-method), a user can change their password if they are [allowed](#allow-a-user-to-change-their-own-password) to do so.

Example command for a user to change their password:

```shell
d8 stronghold write auth/userpass/users/alice/password password="NewPass-456!"
```

If the user tries to change someone else's password, Stronghold returns a `permission denied` error.
If a user enters incorrect login credentials when changing their password, their account may be [locked](#user-lockout).

### Default password policy

If no custom [password policy](../../concepts/password-policy/) is assigned to the `userpass` method, Stronghold uses the default policy when creating a user or changing a user's password.

The default policy for the `userpass` method requires:

- `8` characters in the password;
- at least one uppercase letter;
- at least one lowercase letter;
- at least one digit;
- at least one hyphen (`-`).

To use a custom policy instead of the default password policy for the `userpass` method, run the following command (replace `policy_name` with the name of the desired policy):

```shell
d8 stronghold write auth/userpass/password-policy/{policy_name}
```
