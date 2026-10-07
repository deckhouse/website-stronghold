---
title: "KV1/KV2 replication"
linkTitle: "KV1/KV2 replication"
weight: 45
description: "Administrator guide for KV1/KV2 replication in Stronghold."
---

## How KV1/KV2 replication works in Stronghold

Replication means automatic copying of secrets between several Stronghold instances in the `master`/`slave` mode using a pull model.

Replication is supported only for `KV1` and `KV2` stores.

Data is synchronized periodically on a schedule or according to the individual settings of a specific `KV1/KV2` store.

For replication to work, you need to:

- ensure network connectivity with the remote Stronghold cluster;
- configure a correct TLS connection, if TLS is used;
- obtain a token to access the remote cluster;
- grant this token `list` and `read` permissions on the `KV1/KV2` stores in the remote cluster.

To enable replication, set the replication parameters when mounting a new `KV1/KV2` store.

Keep in mind:

- the remote and local `mount` paths may have different names;
- replication can be configured between different namespaces in the local and remote clusters;
- you can configure replication of several local stores with different names from one remote store;
- if replication is configured for a local store, it works only in `read-only` mode.

Writing, modifying, and deleting secrets in a local replicated store is not possible. All changes must be made in the source `master` store.

After the next synchronization run, changes from the remote store are applied to the local one.

When replication is disabled, the `read-only` status is removed, and adding, modifying, and deleting secrets becomes available locally.

When replication is enabled again, all local changes are deleted or overwritten with data from the source store.

## Configuring KV1/KV2 replication

Replication is configured on the consumer side, that is, on the `slave` Stronghold cluster, when mounting a new `KV1/KV2` store.

The settings include the following parameters:

- the address of the remote Stronghold cluster;
- a token to access the remote cluster;
- a `wrapping token` for secure transfer of the access token;
- a TLS certificate or a path to the certificate for connecting to the remote cluster;
- the `namespace path` where the remote `KV1/KV2` store is located;
- the `mount path` of the remote `KV1/KV2` store;
- a list of `secret path` values to replicate;
- the replication run period;
- enabling or disabling replication;
- the KV store version.

{{< alert level="warning" >}}
The local and remote KV store versions must match. You cannot configure replication from `kv1` to `kv2` or from `kv2` to `kv1`.
{{< /alert >}}

## Creating a replication token

The token for accessing the remote cluster must have `list` and `read` permissions on the replicated secrets.

If the issued token supports self-renewal, Stronghold automatically renews it for 30 days when the remaining TTL drops below 7 days and the `maxTTL` parameter is not exceeded.

Below is an example of creating a policy and a token for replication from the `dev-secrets` `mount` located in the `ns_path_1` namespace:

{{< tabs name="stronghold_cmd_42837" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold policy write -namespace=ns_path_1 replicate-dev-secrets - <<'EOF'
# Allow token to list/read secrets from dev-secrets
path "dev-secrets/*" {
  capabilities = ["read", "list"]
}

# Allow token to read info about dev-secrets
path "sys/mounts/dev-secrets" {
  capabilities = ["read"]
}

# Allow token to look up own properties
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

# Allow token to renew self
path "auth/token/renew-self" {
  capabilities = ["update"]
}
EOF

d8 stronghold token create -namespace=ns_path_1 -policy=replicate-dev-secrets -orphan=true -period=30d
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold policy write -namespace=ns_path_1 replicate-dev-secrets - <<'EOF'
# Allow token to list/read secrets from dev-secrets
path "dev-secrets/*" {
  capabilities = ["read", "list"]
}

# Allow token to read info about dev-secrets
path "sys/mounts/dev-secrets" {
  capabilities = ["read"]
}

# Allow token to look up own properties
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

# Allow token to renew self
path "auth/token/renew-self" {
  capabilities = ["update"]
}
EOF

stronghold token create -namespace=ns_path_1 -policy=replicate-dev-secrets -orphan=true -period=30d
```

{{% /tab %}}
{{< /tabs >}}

## Creating a wrapping token

It is recommended to configure replication using a `wrapping token`.

A `wrapping token` is a single-use token with a limited TTL that carries the real access token. When configuring replication, the wrapping token is passed to the consumer cluster, which then performs `unwrap` on the source cluster and obtains the real replication token.

Example of creating a wrapping token on the source cluster:

{{< tabs name="stronghold_cmd_90364" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold token create \
  -namespace=ns_path_1 \
  -policy=replicate-dev-secrets \
  -orphan=true \
  -period=30d \
  -wrap-ttl=5m \
  -field=wrapping_token
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold token create \
  -namespace=ns_path_1 \
  -policy=replicate-dev-secrets \
  -orphan=true \
  -period=30d \
  -wrap-ttl=5m \
  -field=wrapping_token
```

{{% /tab %}}
{{< /tabs >}}

Pass the resulting `wrapping token` when configuring replication on the consumer cluster.

## Configuring replication using the CLI

### Without TLS

{{< tabs name="stronghold_cmd_6304" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold secrets enable \
  -path=<local_mount_path_name> \
  -src-address=<address_of_source_cluster> \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-namespace=<namespace_path_in_source_cluster> \
  -src-mount-path=<mount_path_in_source_cluster> \
  -sync-period-min=<interval_in_minutes> \
  -version=<1|2> \
  -namespace=<namespace_path_in_local_cluster> \
  kv
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold secrets enable \
  -path=<local_mount_path_name> \
  -src-address=<address_of_source_cluster> \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-namespace=<namespace_path_in_source_cluster> \
  -src-mount-path=<mount_path_in_source_cluster> \
  -sync-period-min=<interval_in_minutes> \
  -version=<1|2> \
  -namespace=<namespace_path_in_local_cluster> \
  kv
```

{{% /tab %}}
{{< /tabs >}}

### With TLS

{{< tabs name="stronghold_cmd_40961" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold secrets enable \
  -path=<local_mount_path_name> \
  -src-address=<address_of_source_cluster> \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-namespace=<namespace_path_in_source_cluster> \
  -src-mount-path=<mount_path_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-period-min=<interval_in_minutes> \
  -version=<1|2> \
  -namespace=<namespace_path_in_local_cluster> \
  kv
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold secrets enable \
  -path=<local_mount_path_name> \
  -src-address=<address_of_source_cluster> \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-namespace=<namespace_path_in_source_cluster> \
  -src-mount-path=<mount_path_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-period-min=<interval_in_minutes> \
  -version=<1|2> \
  -namespace=<namespace_path_in_local_cluster> \
  kv
```

{{% /tab %}}
{{< /tabs >}}

Parameters:

- `-path`: the `mount path` name of the local KV store where the data is copied;
- `-src-address`: the address of the remote Stronghold cluster;
- `-src-token`: a token to access the remote cluster; required if `-src-wrapping-token` is not specified;
- `-src-wrapping-token`: a wrapping token to be unwrapped on the remote cluster; required if `-src-token` is not specified;
- `-src-namespace`: the `namespace path` in the remote cluster; `root` by default;
- `-src-mount-path`: the `mount path` name of the remote KV store;
- `-src-secret-path`: a list of `secret path` values to replicate;
- `-src-ca-cert`: the CA certificate for the TLS connection; if the certificate is in a file, use the `@ca-cert.pem` form;
- `-sync-period-min`: the synchronization period in minutes; `1` by default;
- `-version`: the KV store version (`1` or `2`);
- `-namespace`: the `namespace path` in the local cluster; `root` by default.

## Changing replication settings using the CLI

You can edit the following:

- the token for accessing the remote cluster;
- the wrapping token for updating the token;
- the TLS certificate;
- the list of `secret path` values;
- the replication run period;
- enabling and disabling replication.

{{< alert level="warning" >}}
When you change `secret path`, the old path in the local cluster remains unchanged, and the new one is added. If the old and new paths overlap, the new data may partially overwrite the existing data.
{{< /alert >}}

Example of changing the settings:

{{< tabs name="stronghold_cmd_91012" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold secrets tune \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-secret-path=<list_of_secret_paths_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-enable=true \
  -sync-period-min=<interval_in_minutes> \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold secrets tune \
  -src-wrapping-token=<wrapping_token_from_source_cluster> \
  -src-secret-path=<list_of_secret_paths_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-enable=true \
  -sync-period-min=<interval_in_minutes> \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

{{% /tab %}}
{{< /tabs >}}

To disable replication:

{{< tabs name="stronghold_cmd_83685" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold secrets tune \
  -sync-enable=false \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold secrets tune \
  -sync-enable=false \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

{{% /tab %}}
{{< /tabs >}}

To enable it again:

{{< tabs name="stronghold_cmd_53688" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold secrets tune \
  -sync-enable=true \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold secrets tune \
  -sync-enable=true \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

{{% /tab %}}
{{< /tabs >}}

To read the current settings:

{{< tabs name="stronghold_cmd_50008" >}}
{{% tab name="Stronghold in DKP" %}}

```shell
d8 stronghold read \
  -namespace=<namespace_path_in_local_cluster> \
  sys/mounts/<mount_path>/tune
```

{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}

```shell
stronghold read \
  -namespace=<namespace_path_in_local_cluster> \
  sys/mounts/<mount_path>/tune
```

{{% /tab %}}
{{< /tabs >}}

## Usage examples

Ready-made examples that use this feature:

- [Migrating from HashiCorp Vault](../../../examples/operations/migration-from-vault/)

See all examples in [Usage examples](../../../examples/).
