---
title: "AppRole"
linkTitle: "AppRole"
description: "Stronghold Agent authentication using AppRole."
weight: 50
---

AppRole is the recommended Stronghold Agent authentication method for applications that run on virtual machines and bare metal. It is suitable when an application or service must access Stronghold without user involvement.

This page describes how AppRole works, which entities it uses, and how to configure Stronghold Agent to authenticate with this method.

## How AppRole works

AppRole uses two entities:

- `Role ID`: Role identifier.
- `Secret ID`: Secret identifier.

For successful authentication, Stronghold Agent must receive both values. After that, the Agent calls the AppRole authentication method and obtains a Stronghold token, which it then uses to work with secrets.

## When to use AppRole

Use AppRole in the following cases:

- The application runs on a virtual machine or bare metal.
- Authentication must happen automatically without user involvement.
- Access must be limited by a set of policies.
- The lifetime and number of `Secret ID` uses must be controlled.

For production environments, AppRole is usually a better choice than a static token because it provides flexible lifetime, policy, and access restriction settings.

## Benefits

AppRole provides the following benefits:

- Separates a public identifier from a secret value.
- Allows access to be limited by policies.
- Supports CIDR restrictions.
- Supports single-use and temporary `Secret ID` values.
- Fits automation on virtual machines and bare metal.

## Configuring AppRole in Stronghold

Enable and configure AppRole:

```shell
stronghold auth enable approle

stronghold write auth/approle/role/myapp \
  token_ttl=1h \
  token_max_ttl=4h \
  policies="myapp-policy"
```

Get the `Role ID`:

```shell
stronghold read auth/approle/role/myapp/role-id
```

Create a `Secret ID`:

```shell
stronghold write -f auth/approle/role/myapp/secret-id
```

## Stronghold Agent configuration

Example Stronghold Agent configuration for AppRole:

```hcl
auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
      remove_secret_id_file_after_reading = true
    }
  }
}
```

In this example, the Agent:

- reads `Role ID` from `/etc/stronghold-agent/role-id`;
- reads `Secret ID` from `/etc/stronghold-agent/secret-id`;
- removes the `Secret ID` file after reading it.

## Storing Role ID and Secret ID

Consider the following storage specifics.

### Role ID

`Role ID` is a role identifier. It is usually stored in `/etc/stronghold-agent/role-id`.

`Role ID` has the following properties:

- it can be delivered through a configuration management system;
- it can be stored in a virtual machine image;
- it is usually not removed after use;
- by itself, it is not considered a critical secret.

### Secret ID

`Secret ID` is a sensitive secret. It is usually stored in `/etc/stronghold-agent/secret-id`.

`Secret ID` has the following properties:

- it must be delivered through a protected channel;
- it must not be stored in Git in plain text;
- it can be removed after reading;
- it must be reissued if compromised.

## Secret ID types

You can use the following `Secret ID` types.

### Single-use Secret ID

Use this option to allow only one use of the secret:

```shell
stronghold write auth/approle/role/myapp \
  secret_id_num_uses=1 \
  policies="myapp-policy"
```

Or create a single-use `Secret ID` during issuance:

```shell
stronghold write -f auth/approle/role/myapp/secret-id num_uses=1
```

This option is usually recommended for production environments.

### Multi-use Secret ID

Use this option if `Secret ID` must be used repeatedly:

```shell
stronghold write auth/approle/role/myapp \
  secret_id_num_uses=0 \
  policies="myapp-policy"
```

This option is suitable for development and testing, but requires manual rotation if compromised.

### Secret ID with limited TTL

Use this option to limit the secret lifetime:

```shell
stronghold write auth/approle/role/myapp \
  secret_id_ttl=24h \
  policies="myapp-policy"
```

After the TTL expires, a new `Secret ID` must be issued.

## Best practices

When using AppRole, follow these recommendations:

- Use single-use `Secret ID` values in production environments.
- Deliver `Role ID` and `Secret ID` through separate channels.
- Limit access by CIDR if your scenario supports it.
- Log `Secret ID` usage for audit purposes.
- Restrict access to `/etc/stronghold-agent/role-id` and `/etc/stronghold-agent/secret-id`.

{{< alert level="warning" >}}
Do not store `Secret ID` in Git or another version control system in plain text.
{{< /alert >}}
