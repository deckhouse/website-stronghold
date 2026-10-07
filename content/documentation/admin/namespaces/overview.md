---
title: "Namespaces"
linkTitle: "Introduction"
weight: 10
description: "Administrator guide for namespace isolation and management in Stronghold."
---

## What namespaces are

Namespaces in Stronghold let you split one server into several isolated logical areas. Each namespace works as a separate virtual Stronghold with its own:

- policies;
- mount paths;
- secrets;
- auth methods;
- tokens and sessions.

All namespaces are managed by a single Stronghold instance and form a hierarchy with the `root` namespace at the top.

## Key properties

- namespaces form a tree in which each namespace can have a parent and child namespaces;
- data and configuration are isolated between namespaces;
- an administrator of a parent namespace can manage its child namespaces;
- users and services work only within their own namespace and its child namespaces, if policies allow it.

To address a specific namespace:

- in the CLI, use the `-namespace=<namespace_path>` parameter;
- in the REST API, pass the `X-Vault-Namespace: <namespace_path>` header.

## Creating a namespace

To create a new namespace, you need read and write permissions on the `/namespaces` path in the current namespace.

The name of the new namespace must not overlap with existing paths. Once created, the namespace can be addressed by its full path in the hierarchy.

### Using the CLI

{{< tabs name="stronghold_cmd_55611" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace create \
  -namespace=<parent_namespace_name> \
  -custom-metadata=key="value" \
  <new_namespace_name>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace create \
  -namespace=<parent_namespace_name> \
  -custom-metadata=key="value" \
  <new_namespace_name>
```

{{% /tab %}}
{{< /tabs >}}

Example response:

```text
Key                Value
---                -----
custom_metadata    map[key:value]
id                 c88b1992-d9d3-4160-8283-77163a4e7fae
path               <new_namespace_name>/
```

Parameters:

- `-namespace`: the absolute path of the parent namespace in which the new namespace is created; if the parameter is not specified, the namespace is created in `root`;
- `-custom-metadata`: arbitrary namespace metadata in the `key=value` format.

### Using the REST API

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <parent_namespace_name>" \
  --request POST \
  --data @payload.json \
  $STRONGHOLD_ADDR/v1/sys/namespaces/<new_namespace_name> | jq -r ".data"
```

Parameters:

- `X-Vault-Namespace`: the absolute path of the parent namespace in which the new nested namespace is created;
- `payload.json`: JSON describing the namespace metadata.

## Reading and listing namespaces

### Reading using the CLI

{{< tabs name="stronghold_cmd_41027" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace lookup -namespace=<parent_namespace_name> <namespace_name>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace lookup -namespace=<parent_namespace_name> <namespace_name>
```

{{% /tab %}}
{{< /tabs >}}

Example response:

```text
Key                Value
---                -----
custom_metadata    map[key:value]
id                 c88b1992-d9d3-4160-8283-77163a4e7fae
path               <namespace_name>/
```

### Reading using the REST API

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <parent_namespace_name>" \
  --request GET \
  $STRONGHOLD_ADDR/v1/sys/namespaces/<namespace_name>
```

### Listing using the CLI

{{< tabs name="stronghold_cmd_58281" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace list -namespace=<parent_namespace_name>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace list -namespace=<parent_namespace_name>
```

{{% /tab %}}
{{< /tabs >}}

Example response:

```text
Keys
----
ns1/
ns2/
```

### Listing using the REST API

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <parent_namespace_name>" \
  --request LIST \
  $STRONGHOLD_ADDR/v1/sys/namespaces
```

## Deleting a namespace

### Using the CLI

{{< tabs name="stronghold_cmd_8806" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace delete -namespace=<parent_namespace_name> <namespace_name>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace delete -namespace=<parent_namespace_name> <namespace_name>
```

{{% /tab %}}
{{< /tabs >}}

The `-namespace` parameter sets the parent namespace that contains the namespace being deleted. If the parameter is not specified, the namespace is deleted from `root`.

### Using the REST API

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <parent_namespace_name>" \
  --request DELETE \
  $STRONGHOLD_ADDR/v1/sys/namespaces/<namespace_name>
```

## Namespace API lock

`Namespace API Lock` lets you temporarily block all API requests to the selected namespace and all its child namespaces.

When a namespace is locked, Stronghold returns a one-time `unlock_key`. Save it to remove the lock later. Unlocking can also be performed with a root token without `unlock_key`.

### Locking using the CLI

Lock the current namespace and all its child namespaces:

{{< tabs name="stronghold_cmd_69466" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace lock
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace lock
```

{{% /tab %}}
{{< /tabs >}}

Lock a specific child namespace, for example, `ns1/ns2/`:

{{< tabs name="stronghold_cmd_52362" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace lock ns1/ns2
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace lock ns1/ns2
```

{{% /tab %}}
{{< /tabs >}}

Example response:

```text
Key            Value
---            -----
unlock_key     7Hk3xQ9mR2pN5vL8wY4tZa
```

{{< alert level="warning" >}}
Save the `unlock_key` value. It is required for unlocking if the operation is not performed with a root token.
{{< /alert >}}

### Locking using the REST API

Lock the `<namespace_name>` namespace from the root namespace:

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --request POST \
  $STRONGHOLD_ADDR/v1/sys/namespaces/api-lock/lock/<namespace_name> | jq -r ".data"
```

Lock the current namespace defined by the `X-Vault-Namespace` header:

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --header "X-Vault-Namespace: <namespace_name>" \
  --request POST \
  $STRONGHOLD_ADDR/v1/sys/namespaces/api-lock/lock | jq -r ".data"
```

Example response:

```json
{
  "unlock_key": "7Hk3xQ9mR2pN5vL8wY4tZa"
}
```

### Unlocking using the CLI

Unlock the current namespace using the key:

{{< tabs name="stronghold_cmd_22637" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace unlock -unlock-key=<key>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace unlock -unlock-key=<key>
```

{{% /tab %}}
{{< /tabs >}}

Unlock the current namespace using a root token:

{{< tabs name="stronghold_cmd_76641" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace unlock
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace unlock
```

{{% /tab %}}
{{< /tabs >}}

Unlock a specific child namespace:

{{< tabs name="stronghold_cmd_32282" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold namespace unlock -unlock-key=<key> ns1/ns2
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold namespace unlock -unlock-key=<key> ns1/ns2
```

{{% /tab %}}
{{< /tabs >}}

### Unlocking using the REST API

Using `unlock_key`:

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_TOKEN" \
  --request POST \
  --data '{"unlock_key": "<key>"}' \
  $STRONGHOLD_ADDR/v1/sys/namespaces/api-lock/unlock/<namespace_name>
```

Using a root token:

```shell
curl \
  --header "X-Vault-Token: $STRONGHOLD_ROOT_TOKEN" \
  --request POST \
  $STRONGHOLD_ADDR/v1/sys/namespaces/api-lock/unlock/<namespace_name>
```

### Lock API parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `<namespace_name>` | string | no | Path of the namespace to lock or unlock. If not specified, the operation applies to the current namespace defined by the `X-Vault-Namespace` header. |
| `unlock_key` | string | no | The unlock key received when locking. Required for unlocking if the token is not a root token. |

## Lock and unlock example

{{< tabs name="stronghold_cmd_46189" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
# Create a namespace.
d8 stronghold namespace create production

# Lock it.
d8 stronghold namespace lock production
# Key            Value
# ---            -----
# unlock_key     7Hk3xQ9mR2pN5vL8wY4tZa

# Any requests to production are now blocked.
d8 stronghold secrets list -namespace=production
# Error: API access to this namespace has been locked by an administrator...

# Unlock it with the key.
d8 stronghold namespace unlock -unlock-key=7Hk3xQ9mR2pN5vL8wY4tZa production

# Access is restored.
d8 stronghold secrets list -namespace=production
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
# Create a namespace.
stronghold namespace create production

# Lock it.
stronghold namespace lock production
# Key            Value
# ---            -----
# unlock_key     7Hk3xQ9mR2pN5vL8wY4tZa

# Any requests to production are now blocked.
stronghold secrets list -namespace=production
# Error: API access to this namespace has been locked by an administrator...

# Unlock it with the key.
stronghold namespace unlock -unlock-key=7Hk3xQ9mR2pN5vL8wY4tZa production

# Access is restored.
stronghold secrets list -namespace=production
```

{{% /tab %}}
{{< /tabs >}}

## Usage examples

Ready-made examples that use this feature:

- [Multi-tenancy with namespaces](../../../examples/operations/multi-tenancy-namespaces/)

See all examples in [Usage examples](../../../examples/).
