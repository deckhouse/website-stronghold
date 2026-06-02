---
title: "KV1/KV2 replication"
description: "Information about KV1/KV2 replication in Deckhouse Stronghold."
weight: 40
---

KV1/KV2 replication is a mechanism for automatically copying secrets between Stronghold instances in a master-slave setup by using the pull model.
Replication is supported only for KV1/KV2 secret engines.
Data synchronization runs periodically on a schedule or according to the settings of a specific KV1/KV2 secret engine.

To use replication, ensure network connectivity to the remote Stronghold cluster, configure a TLS connection, and obtain an access token.
The token must have the `list` and `read` permissions for the KV1/KV2 secret engine in the remote cluster.

Replication is configured when mounting a new KV1/KV2 secret engine.
The remote and local mount path names may differ.
You can also configure replication between different namespaces in the local and remote clusters.
It is allowed to replicate multiple local secret engines with different names from the same remote secret engine.

If replication is configured for a local KV1/KV2 secret engine, it becomes read-only.
Writing, updating, and deleting secrets in such a secret engine are unavailable.
Make all changes in the source secret engine.
After the next replication run, the data will be copied to the local secret engine.

If you disable replication, the read-only mode is removed.
After that, adding, updating, and deleting secrets become available.
If you enable replication again later, all local changes will be deleted or overwritten with data from the source secret engine.

## Configure replication

Configure replication on the consumer side, that is, in the slave Stronghold cluster.
To do this, specify the replication settings when mounting a new KV1/KV2 secret engine.

The following parameters are supported:

- remote Stronghold cluster address;
- access token for the remote Stronghold cluster;
- TLS certificate or path to the TLS certificate for connecting to the remote Stronghold cluster;
- namespace path where the KV1/KV2 secret engine is located in the remote Stronghold cluster. The default value is `root`;
- mount path name of the KV1/KV2 secret engine in the remote Stronghold cluster;
- list of secret paths to replicate. By default, all secrets are replicated;
- data replication interval. The default value is 1 minute;
- replication enable or disable setting. Replication is enabled by default for a new KV1/KV2 secret engine;
- KV secret engine version for mounting and replication.

{{< alert level="warning" >}}
The local and remote KV secret engine versions must match.
You cannot configure replication from `kv1` to `kv2` or from `kv2` to `kv1`.
{{< /alert >}}

## Create a token for replication

The token for accessing the remote cluster must have the `list` and `read` permissions for the replicated secrets.
If the token supports self-renewal, Stronghold automatically renews it for 30 days when the remaining token TTL becomes less than 7 days and the `maxTTL` limit is not exceeded.

The following example shows how to create a policy and a token for replication from the `dev-secrets` mount path located in the `ns_path_1` namespace.
Run these commands on the source server:

```shell
d8 stronghold policy write -namespace=ns_path_1 replicate-dev-secrets - <<EOF
# Allow token to list/read secrets from dev-secrets.
path "dev-secrets/*" {
  capabilities = ["read", "list"]
}

# Allow token to read info about dev-secrets.
path "sys/mounts/dev-secrets" {
  capabilities = ["read"]
}

# Allow token to look up own properties.
path "auth/token/lookup-self" {
  capabilities = ["read"]
}

# Allow token to renew self.
path "auth/token/renew-self" {
  capabilities = ["update"]
}
EOF

d8 stronghold token create \
  -namespace=ns_path_1 \
  -policy=replicate-dev-secrets \
  -orphan=true \
  -period=30d
```

## Configure replication via the Stronghold CLI

To configure replication via the Stronghold CLI, run one of the following commands.

Without a TLS connection:

```shell
d8 stronghold secrets enable \
  -path=<local_mount_path_name> \
  -src-address=<address_of_source_cluster> \
  -src-token=<token_of_source_cluster> \
  -src-namespace=<namespace_path_in_source_cluster> \
  -src-mount-path=<mount_path_in_source_cluster> \
  -sync-period-min=3 \
  -version=<1/2> \
  -namespace=<namespace_path_in_local_cluster> \
  kv
```

With a TLS connection:

```shell
d8 stronghold secrets enable \
  -path=<local_mount_path_name> \
  -src-address=<address_of_source_cluster> \
  -src-token=<token_of_source_cluster> \
  -src-namespace=<namespace_path_in_source_cluster> \
  -src-mount-path=<mount_path_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-period-min=3 \
  -version=<1/2> \
  -namespace=<namespace_path_in_local_cluster> \
  kv
```

The following parameters are used:

- `-path` — local KV1/KV2 secret engine mount path where the data from the source will be copied. Required parameter. Example: `my-mount-kv2`;
- `-src-address` — remote Stronghold cluster address. Required parameter. Examples: `127.0.0.1:8200`, `vault.mycompany.tld:8200`, `stronghold.mycompany.tld:443`;
- `-src-token` — access token for the remote Stronghold cluster. Required parameter. Example: `z6VXjAi6F3vjaclHu99FLOcr`;
- `-src-namespace` — namespace path where the KV1/KV2 secret engine is located in the remote Stronghold cluster. Optional parameter. The default value is `root`;
- `-src-mount-path` — mount path name of the KV1/KV2 secret engine in the remote Stronghold cluster. Required parameter. Example: `remote-mount-kv2`;
- `-src-secret-path` — list of secret paths to replicate. Optional parameter;
- `-src-ca-cert` — CA certificate for establishing a TLS connection. If the certificate is stored in a file, use the `-src-ca-cert=@ca-cert.pem` format. Optional parameter;
- `-sync-period-min` — interval in minutes between replication runs for the secret engine. Optional parameter. The default value is `60`;
- `-version` — KV secret engine version for mounting and replication. Required parameter;
- `-namespace` — namespace path where the local KV1/KV2 secret engine is created. Optional parameter. The default value is `root`.

{{< alert level="warning" >}}
The local and remote KV secret engine versions must match.
{{< /alert >}}

## Change replication settings via the Stronghold CLI

You can change the following parameters:

- access token for the remote Stronghold cluster;
- TLS certificate or path to the TLS certificate for connecting to the remote Stronghold cluster;
- list of secret paths to replicate. This parameter is currently not used, so all secrets are replicated by default;
- replication interval;
- replication enable or disable setting.

{{< alert level="warning" >}}
If you change `secret path` in the replication configuration, the old path in the local cluster remains unchanged, and the new path is added.
If the old and new paths overlap, some data may be overwritten.

For example, suppose the configuration initially contains `-src-secret-path=[first-secret/one, second-sercet/two]`,
and after the change it contains `-src-secret-path=[first-secret/two, second-sercet/two]`.
In this case, the data in `first-secret/one` remains unchanged and is no longer updated.
{{< /alert >}}

To change replication settings via the Stronghold CLI, run the following command:

```shell
d8 stronghold secrets tune \
  -src-token=<token_of_source_cluster> \
  -src-secret-path=<list_of_secret_paths_in_source_cluster> \
  -src-ca-cert=@<path_to_file_with_certificate> \
  -sync-enable=true \
  -sync-period-min=3 \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

The following parameters are used:

- `-src-token` — access token for the remote Stronghold cluster. Required parameter. Example: `z6VXjAi6F3vjaclHu99FLOcr`;
- `-src-secret-path` — list of secret paths to replicate. Optional parameter;
- `-src-ca-cert` — CA certificate for establishing a TLS connection. If the certificate is stored in a file, use the `-src-ca-cert=@ca-cert.pem` format. Optional parameter;
- `-sync-enable` — enables or disables replication for the local mount path. Required parameter;
- `-sync-period-min` — interval in minutes between replication runs for the secret engine. Optional parameter;
- `-namespace` — namespace path where the local KV1/KV2 secret engine is created. Optional parameter. The default value is `root`.

To disable replication, run the following command:

```shell
d8 stronghold secrets tune \
  -sync-enable=false \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

If you pass other replication parameters together with `-sync-enable=false`, they are ignored.

To enable replication, run the following command:

```shell
d8 stronghold secrets tune \
  -sync-enable=true \
  -namespace=<namespace_path_in_local_cluster> \
  <local_mount_path_name>
```

In this case, you can also pass other replication parameters.
They are applied.

To read the replication settings, run the following command:

```shell
d8 stronghold read \
  -namespace=<namespace_path_in_local_cluster> \
  sys/mounts/<mount_path>/tune
```

## Configure replication via the Stronghold API

To configure replication via the Stronghold API, call the mount creation API and add the replication configuration to the request body:

```shell
curl --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "type": "<kv-v1>/<kv-v2>",
    "config": {
      "replication_config": {
        "src_address": "<address_of_source_cluster>",
        "src_token": "<token_of_source_cluster>",
        "src_ca_cert": "<tls_cert_for_source_cluster>",
        "src_namespace": "<namespace_path_in_source_cluster>",
        "src_mount_path": "<mount_path_in_source_cluster>",
        "src_secret_path": ["<list_of_secret_paths_in_source_cluster>"],
        "sync_period_min": <interval_in_minutes_for_synchronization_period>
      }
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>
```

If the remote source cluster does not support TLS, do not pass the `"src_ca_cert"` parameter.
By default, the `"src_secret_path"` parameter is set to `"*"`, which means that all secret paths are replicated.

The following parameters are used:

- `local_stronghold_address` — address of the local Stronghold cluster where replication is configured;
- `token_for_local_cluster` — token for the local cluster that has permission to create a mount;
- `namespace_path_in_local_cluster` — namespace path where the local KV1/KV2 secret engine is created. Optional parameter. The default value is `root`;
- `local_mount_path_name` — mount path name of the local KV1/KV2 secret engine where data from the source is copied. Required parameter. Example: `my-mount-kv2`;
- `src_address` — remote Stronghold cluster address. Required parameter. Examples: `127.0.0.1:8200`, `vault.mycompany.tld:8200`, `stronghold.mycompany.tld:443`;
- `src_token` — access token for the remote Stronghold cluster. Required parameter. Example: `z6VXjAi6F3vjaclHu99FLOcr`;
- `src_namespace` — namespace path where the KV1/KV2 secret engine is located in the remote Stronghold cluster. Optional parameter. The default value is `root`;
- `src_mount_path` — mount path name of the KV1/KV2 secret engine in the remote Stronghold cluster. Required parameter. Example: `remote-mount-kv2`;
- `src_secret_path` — list of secret paths to replicate. Optional parameter;
- `src_ca_cert` — CA certificate for establishing a TLS connection. Optional parameter;
- `sync_period_min` — interval in minutes between replication runs for the secret engine. Optional parameter. The default value is `1`;
- `type` — KV secret engine version for mounting and replication. Required parameter.

{{< alert level="warning" >}}
The local and remote KV secret engine versions must match.
{{< /alert >}}

## Change replication settings via the Stronghold API

You can change the following parameters:

- access token for the remote Stronghold cluster;
- TLS certificate or path to the TLS certificate for connecting to the remote Stronghold cluster;
- list of secret paths to replicate. By default, all secrets are replicated;
- replication interval;
- replication enable or disable setting.

{{< alert level="warning" >}}
If you change `secret path` in the replication configuration, the old path in the local cluster remains unchanged, and the new path is added.
If the old and new paths overlap, some data may be overwritten.

For example, suppose the configuration initially contains `"src_secret_path"=["first-secret/one", "second-sercet/two"]`,
and after the change it contains `"src_secret_path"=["first-secret/two", "second-sercet/two"]`.
In this case, the data in `"first-secret/one"` remains unchanged and is no longer updated.
{{< /alert >}}

To change replication settings via the Stronghold API, call the mount tuning API and pass the new replication configuration in the request body:

```shell
curl --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "src_token": "<token_of_source_cluster>",
      "src_ca_cert": "<tls_cert_for_source_cluster>",
      "src_secret_path": ["<list_of_secret_paths_in_source_cluster>"],
      "sync_period_min": <interval_in_minutes_for_synchronization_period>,
      "sync_enable": true
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

The following parameters are used:

- `local_stronghold_address` — address of the local Stronghold cluster where replication is configured;
- `token_for_local_cluster` — token for the local cluster that has permission to update a mount;
- `namespace_path_in_local_cluster` — namespace path where the local KV1/KV2 secret engine is created. Optional parameter. The default value is `root`;
- `local_mount_path_name` — mount path name of the local KV1/KV2 secret engine where data from the source is copied. Required parameter. Example: `my-mount-kv2`;
- `src_token` — access token for the remote Stronghold cluster. Required parameter. Example: `z6VXjAi6F3vjaclHu99FLOcr`;
- `src_ca_cert` — CA certificate for establishing a TLS connection. Optional parameter;
- `sync_period_min` — interval in minutes between replication runs for the secret engine. Optional parameter;
- `sync_enable` — enables or disables replication for the local mount path. Required parameter;
- `src_secret_path` — list of secret paths to replicate. Optional parameter.

To disable replication, run the following request:

```shell
curl --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "sync_enable": false
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

If you pass other replication parameters together with `"sync_enable": false`, they are ignored.

To enable replication, run the following request:

```shell
curl --header "X-Vault-Token: <token_for_local_cluster>" \
  --header "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  --request POST \
  --data '{
    "replication_config": {
      "sync_enable": true
    }
  }' \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```

In this case, you can also pass other replication parameters.
They are applied.

To read the replication settings, run the following request:

```shell
curl -X GET \
  -H "X-Vault-Token: <token_for_local_cluster>" \
  -H "X-Vault-Namespace: <namespace_path_in_local_cluster>" \
  <local_stronghold_address>/v1/sys/mounts/<local_mount_path_name>/tune
```
