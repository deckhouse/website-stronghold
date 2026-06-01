---
title: "AppRole"
linkTitle: "AppRole"
weight: 40
description: "Application and service authentication in Deckhouse Stronghold using AppRole."
---

The AppRole auth method is designed to authenticate machines and applications in Deckhouse Stronghold. It is especially useful in automated scenarios where access to Stronghold is required by a service rather than a user.

AppRole is a set of policies and authentication constraints that must be satisfied to obtain a Stronghold token. Its scope can be either narrow or broad: a role can be configured for a single machine, a specific service, or a service running on multiple machines.

The credentials required for successful authentication depend on the AppRole constraints associated with the role.

{% alert level="info" %}
For AppRole, it is recommended to use `batch` tokens unless your scenario requires child token creation or other advanced `service` token capabilities.
{% endalert %}

## How AppRole works

In a typical login flow, two values are used to log in with AppRole:

- `role_id`.
- `secret_id`.

The client passes them to Stronghold, after which Stronghold:

1. Checks that the role exists.
1. Verifies the role constraints.
1. If the checks succeed, issues a Stronghold token with the policies attached to the role.

## When to use AppRole

AppRole is typically chosen if:

- you need to authenticate an application or service.
- access to Stronghold must be granted to a workload rather than a user.
- you need flexible control over token TTL, number of uses, and other constraints.
- you need to deliver the role identifier and role secret through separate channels.

## Enable the method

Auth methods can be enabled and disabled through the Deckhouse web UI, CLI, or API. When enabled, an auth method is mounted into the Stronghold mount table and becomes available through the standard read and write API. By default, auth methods are mounted under the `auth/` directory and receive a path in the `auth/<type>` form.

For AppRole, the basic way to enable it through the CLI is as follows:

```shell
d8 stronghold auth enable approle
```

By default, the method is mounted at the following path:

```text
auth/approle
```

If needed, you can enable it at a different path:

```shell
d8 stronghold auth enable -path=my-login approle
```

This is useful if you need to use multiple independent AppRole instances with different settings in a single installation.

### Enable and disable through the Deckhouse web UI

To enable the auth method through the Deckhouse web UI, follow these steps:

1. Open the auth methods management page.
1. Select AppRole from the list of available methods.
1. Configure the mount settings and confirm that you want to enable the method.

![Enabling the auth method](/images/stronghold/admin-guide-image1.png)
![Selecting the auth method](/images/stronghold/admin-guide-image2.png)
![Configuring and confirming the auth method](/images/stronghold/admin-guide-image3.png)

To disable the auth method through the Deckhouse web UI, follow these steps:

1. Select the previously enabled auth method.
1. Confirm its removal.

![Selecting the auth method](/images/stronghold/admin-guide-image4.png)
![Confirming auth method removal](/images/stronghold/admin-guide-image5.png)

## Authenticate through the CLI

By default, the `auth/approle/login` path is used. If the auth method is enabled at a different path, specify that path instead of the default one.

Authentication example:

```shell
d8 stronghold write auth/approle/login \
  role_id=db02de05-fa49-4055-059b-67221c5c2f63 \
  secret_id=6a174c20-f6de-a63c-74d2-6018fcceff64
```

Example result:

```text
Key                Value
---                -----
token              75b74ffd-842c-fd43-1386-f7d7006e520a
token_accessor     4c29bc22-5c72-11a6-f778-2bc8f48cea0e
token_duration     20m0s
token_renewable    true
token_policies     [default]
```

After that, the client receives a Stronghold token and can use it for subsequent requests.

## Authenticate through the API

By default, the following path is used:

```text
auth/approle/login
```

If the method is enabled at a different path, use that path instead of the standard one.

Request example:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request POST \
  --data '{"role_id":"988a9df-...","secret_id":"37b74931..."}' \
  ${STRONGHOLD_ADDR}/v1/auth/approle/login
```

Response example:

```json
{
  "auth": {
    "renewable": true,
    "lease_duration": 2764800,
    "metadata": {},
    "policies": ["default", "dev-policy", "test-policy"],
    "accessor": "5d7fb475-07cb-4060-c2de-1ca3fcbf0c56",
    "client_token": "98a4c7ab-b1fe-361b-ba0b-e307aacfd587"
  }
}
```

The Stronghold token is returned in the `auth.client_token` field.

## Configure AppRole

Before an application can authenticate, the AppRole method must be configured in advance. This is typically done by an administrator or a configuration management tool.

### Step 1. Enable the method

```shell
d8 stronghold auth enable approle
```

### Step 2. Create a role

Role creation example:

```shell
d8 stronghold write auth/approle/role/my-role \
  token_type=batch \
  secret_id_ttl=10m \
  token_num_uses=10 \
  token_ttl=20m \
  token_max_ttl=30m \
  secret_id_num_uses=40
```

{% alert level="warning" %}
If the token issued by the AppRole role must be able to create child tokens, set the `token_num_uses` parameter to `0`.
{% endalert %}

### Step 3. Get the RoleID

```shell
d8 stronghold read auth/approle/role/my-role/role-id
```

Example result:

```text
role_id     db02de05-fa49-4055-059b-67221c5c2f63
```

### Step 4. Get the SecretID

```shell
d8 stronghold write -f auth/approle/role/my-role/secret-id
```

Example result:

```text
secret_id               6a174c20-f6de-a63c-74d2-6018fcceff64
secret_id_accessor      c454f7e5-996e-7230-6074-6ef26b7bcf86
secret_id_ttl           10m
secret_id_num_uses      40
```

## Configure through the API

### Enable the method

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request POST \
  --data '{"type":"approle"}' \
  ${STRONGHOLD_ADDR}/v1/sys/auth/approle
```

### Create a role

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --data '{"policies":"dev-policy,test-policy","token_type":"batch"}' \
  ${STRONGHOLD_ADDR}/v1/auth/approle/role/my-role
```

### Get the RoleID

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  ${STRONGHOLD_ADDR}/v1/auth/approle/role/my-role/role-id
```

Response example:

```json
{
  "data": {
    "role_id": "888a9dfd-ea69-4a53-6cb6-9d6b86474bba"
  }
}
```

### Create a SecretID

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request POST \
  ${STRONGHOLD_ADDR}/v1/auth/approle/role/my-role/secret-id
```

Response example:

```json
{
  "data": {
    "secret_id_accessor": "65946873-1d96-a9d4-678c-9229f74386a5",
    "secret_id": "37b24931-c4cd-d49a-9246-ccc62d682a25",
    "secret_id_ttl": 600,
    "secret_id_num_uses": 40
  }
}
```

## Credentials and constraints

### RoleID

RoleID is the identifier assigned to an AppRole role and used as a required login parameter. By default, it is a UUID, but it can be replaced with a custom value if your scenario requires it.

### SecretID

SecretID is a secret credential that is also required for login by default. It must remain confidential.

For advanced scenarios, you can disable the `secret_id` requirement using the `bind_secret_id` parameter. In this case, tokens can be obtained by clients that only know the `role_id` or satisfy other constraints configured for the role.

A SecretID can be created for AppRole in two ways:

- automatically by the role itself in Pull mode.
- by setting a custom value in Push mode.

A SecretID can have its own parameters:

- TTL.
- number of uses.
- expiration period.
- constraints related to the issuance scenario.

## Pull and Push modes

For SecretID, Stronghold supports two approaches.

### Pull

In Pull mode, the SecretID is retrieved from AppRole by the client itself or by a trusted party. This is the primary and preferred scenario.

### Push

In Push mode, a custom SecretID is assigned to the client manually. This mode can be used for compatibility or specific scenarios, including compatibility with the App-ID workflow, but in most cases it is considered less preferable.

The reason is that in Push mode, the full set of credentials must be delivered to an external system, whereas in Pull mode, RoleID and SecretID can be distributed more securely through separate channels.

{% alert level="info" %}
To deliver SecretID in Pull mode, it is often convenient to use [response wrapping](../../concepts/response-wrapping/) so that the secret is not transmitted in plain text.
{% endalert %}

## Additional constraints

In addition to `role_id` and `secret_id`, an AppRole role can have extra constraints.

For example:

- `bind_secret_id` requires `secret_id` to be provided at login.
- `secret_id_bound_cidrs` allows login only from IP addresses that belong to the specified CIDR blocks.

Some constraints do not require additional credentials, but they still impose restrictions on the login process.

This makes it possible to use AppRole not just as a pair of identifiers, but as a managed machine authentication model with additional admission rules.

## Best practices

Use the following recommendations:

- Use AppRole for applications and services rather than for interactive user login.
- Prefer Pull mode for SecretID whenever possible.
- Deliver RoleID and SecretID through separate channels.
- Use `batch` tokens unless your scenario requires advanced `service` token capabilities.
- Plan TTL, number of uses, and possible role constraints in advance.
- If SecretID must be transmitted through an untrusted environment, use [response wrapping](../../concepts/response-wrapping/).
