---
title: "KV v2"
description: "Information about KV v2 in Deckhouse Stronghold."
weight: 30
---

The version 2 `kv` secrets engine is designed to store arbitrary secrets in the Stronghold storage.
Unlike version 1 of `kv`, it supports data versioning and partial secret updates.
Key names must be strings.
If you write non-string values directly through the CLI, Stronghold converts them to strings.
To store non-string values, pass key-value pairs from a JSON file or use the HTTP API.
The version 2 `kv` secrets engine distinguishes between the `create` and `update` operations in ACL policies.
It also supports the `patch` operation for partial secret updates.
The `update` operation, in contrast, fully overwrites the value.

## How to enable

Most secrets engines must be configured in advance.
This is usually done by an operator or a configuration management system such as Terraform.
To enable version 2 of the `kv` store, run one of the following commands:

```shell
d8 stronghold secrets enable -version=2 kv
```

```shell
d8 stronghold secrets enable kv-v2
```

## Upgrading from version 1 to version 2

An existing version 1 `kv` store can be upgraded to version 2 of `kv` using the CLI or API.
During migration, the store is unavailable.
The upgrade can take a long time, so plan it in advance.
After upgrading to version 2, the previous data access paths no longer work.
Update the ACL policies to restore access.
Also update the paths in applications and user scenarios that work with `kv` data.
To enable versioning for an existing store, run the following command:

```console
$ d8 stronghold kv enable-versioning secret/
Success! Tuned the secrets engine at: secret/
```

## ACL rules

The version 2 `kv` store uses API prefixes that differ from those in version 1 of the API.
Before upgrading from version 1 of `kv`, change the ACL policies.
Different version 2 API paths can be protected by different ACL rules.
Read and write paths use the `data/` prefix.
For example, replace the following policy for version 1 of `kv`:

```text
path "secret/dev/team-1/*" {
  capabilities = ["create", "update", "read"]
}
```

with the following policy for version 2 of `kv`:

```text
path "secret/data/dev/team-1/*" {
  capabilities = ["create", "update", "read"]
}
```

Different data deletion levels are available for version 2 of `kv`.
To allow deleting the latest version of a key, use the following policy:

```text
path "secret/data/dev/team-1/*" {
  capabilities = ["delete"]
}
```

To allow deleting an arbitrary version of a key, use the following policy:

```text
path "secret/delete/dev/team-1/*" {
  capabilities = ["update"]
}
```

To allow restoring deleted versions, use the following policy:

```text
path "secret/undelete/dev/team-1/*" {
  capabilities = ["update"]
}
```

To allow permanently destroying values without the possibility of recovery, use the following policy:

```text
path "secret/destroy/dev/team-1/*" {
  capabilities = ["update"]
}
```

To allow listing keys, use the following policy:

```text
path "secret/metadata/dev/team-1/*" {
  capabilities = ["list"]
}
```

To allow viewing key metadata, use the following policy:

```text
path "secret/metadata/dev/team-1/*" {
  capabilities = ["read"]
}
```

To allow permanent deletion of all versions and metadata of a key, use the following policy:

```text
path "secret/metadata/dev/team-1/*" {
  capabilities = ["delete"]
}
```

The `allowed_parameters`, `denied_parameters`, and `required_parameters` fields are not supported in policies for the version 2 `kv` store.

## Usage

After you enable the secrets engine and obtain a Stronghold token with the required permissions, you can work with secrets.
For version 2 of `kv`, you can still use version 1 style syntax, for example the `secret/foo` path.
However, it is preferable to use the `-mount=secret` flag to avoid confusing the logical secret path with the actual API path.
In this case, the actual path is `secret/data/foo`.

### Writing and reading arbitrary data

Write a secret:

```console
$ d8 stronghold kv put -mount=secret my-secret foo=a bar=b
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          1
```

Read a secret:

```console
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          1
====== Data ======
Key         Value
---         -----
foo         a
bar         b
```

Write a new version of the secret:

```console
$ d8 stronghold kv put -mount=secret -cas=1 my-secret foo=aa bar=bb
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          2
```

The `-cas` flag enables the Check-and-Set check.
If the flag is not specified, the write operation is performed without validation.
If the flag is specified, its value must match the current version of the secret.
A value of `0` allows writing only if the key does not yet exist.
Note that deleting a version does not remove version information from the store.
Therefore, when writing to a secret that has deleted versions, the `cas` value must match the current version of the secret.
By default, read returns the latest version:

```console
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          2
====== Data ======
Key         Value
---         -----
foo         aa
bar         bb
```

To partially update a secret, use the `d8 stronghold kv patch` command:

```console
$ d8 stronghold kv patch -mount=secret -cas=2 my-secret bar=bbb
Key              Value
---              -----
created_time     2024-06-19T17:23:49.199802Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          3
```

The command first attempts to send an HTTP `PATCH` request.
For this, the token must have the `patch` ACL capability.
If this capability is unavailable, the command performs a read, a local update, and then a write.
In this case, the `read` and `update` ACL capabilities are required.
The `-cas` flag can also be used here.
For a direct `PATCH`, it is applied immediately.
For the read-and-write-back scenario, the command uses the `version` value obtained during the read to perform the `cas` check during the write operation.
The `d8 stronghold kv patch` command also supports the `-method` flag.
It defines the update method: `patch` or `rw`.

Update the secret using `patch`:

```console
$ d8 stronghold kv patch -mount=secret -method=patch -cas=2 my-secret bar=bbb
Key              Value
---              -----
created_time     2024-06-19T17:23:49.199802Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          3
```

Update the secret using `rw`, that is, by reading and writing it again:

```console
$ d8 stronghold kv patch -mount=secret -method=rw my-secret bar=bbb
Key              Value
---              -----
created_time     2024-06-19T17:23:49.199802Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          3
```

Read the updated secret:

```console
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:23:49.199802Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          3
====== Data ======
Key         Value
---         -----
foo         aa
bar         bbb
```

To read a previous version, use the `-version` flag:

```console
$ d8 stronghold kv get -mount=secret -version=1 my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          1
====== Data ======
Key         Value
---         -----
foo         a
bar         b
```

You can also use a password policy to generate values.

Create a policy:

```console
$ d8 stronghold write sys/policies/password/example policy=-<<EOF
  length=20
  rule "charset" {
    charset = "abcdefghij0123456789"
    min-chars = 1
  }
  rule "charset" {
    charset = "!@#$%^&*STUVWXYZ"
    min-chars = 1
  }
EOF
```

Create a secret using the `example` policy:

```console
$ d8 stronghold kv put -mount=secret my-generated-secret \
    password=$(d8 stronghold read -field password sys/policies/password/example/generate)
```

```text
========= Secret Path =========
secret/data/my-generated-secret
======= Metadata =======
Key                Value
---                -----
created_time       2024-06-10T14:32:32.37354939Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1
```

Read the created secret:

```console
$ d8 stronghold kv get -mount=secret my-generated-secret
========= Secret Path =========
secret/data/my-generated-secret
======= Metadata =======
Key                Value
---                -----
created_time       2024-06-10T14:32:32.37354939Z
custom_metadata    <nil>
deletion_time      n/a
destroyed          false
version            1
====== Data ======
Key         Value
---         -----
password    !hh&be1e4j16dVc0ggae
```

### Deleting and destroying secrets

The `d8 stronghold kv delete` command performs a soft delete.
It marks the version as deleted and sets the `deletion_time` value in the secret metadata.
With a soft delete, the version data is not removed from the store.
Such a version can be restored with the `d8 stronghold kv undelete` command.
A secret version is permanently deleted in two cases:

- if the number of versions exceeds the `max-versions` value;
- if the `d8 stronghold kv destroy` command is used.

The `destroy` command removes version data without the possibility of recovery.
At the same time, the version metadata is marked as destroyed.
If a version is removed because the number of versions is exceeded, its metadata is also deleted.
You can delete the latest version of a key using the `delete` command.
The command also supports the `-versions` flag for deleting previous versions:

```console
$ d8 stronghold kv delete -mount=secret my-secret
Success! Data deleted (if it existed) at: secret/data/my-secret
```

Deleted versions can be restored:

```console
$ d8 stronghold kv undelete -mount=secret -versions=2 my-secret
Success! Data written to: secret/undelete/my-secret
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:23:21.834403Z
custom_metadata  <nil>
deletion_time    n/a
destroyed        false
version          2
====== Data ======
Key         Value
---         -----
my-value    short-lived-s3cr3t
```

To permanently destroy a version, run the following command:

```console
$ d8 stronghold kv destroy -mount=secret -versions=2 my-secret
Success! Data written to: secret/destroy/my-secret
```

### Metadata

All versions and metadata of a key can be viewed with the `metadata` command or through the API.
If you delete a key with `metadata delete`, all metadata and all versions of this key are removed without the possibility of recovery.

View key metadata and versions:

```console
$ d8 stronghold kv metadata get -mount=secret my-secret
========== Metadata ==========
Key                     Value
---                     -----
cas_required            false
created_time            2024-06-19T17:20:22.985303Z
current_version         2
custom_metadata         <nil>
delete_version_after    0s
max_versions            0
oldest_version          0
updated_time            2024-06-19T17:22:23.369372Z
====== Version 1 ======
Key              Value
---              -----
created_time     2024-06-19T17:20:22.985303Z
deletion_time    n/a
destroyed        false
====== Version 2 ======
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
deletion_time    n/a
destroyed        true
```

Configure version retention settings:

```console
$ d8 stronghold kv metadata put -mount=secret -max-versions=2 -delete-version-after=3h25m19s my-secret
Success! Data written to: secret/metadata/my-secret
```

The `delete-version-after` parameter applies only to new versions.
The `max-versions` parameter is applied during the next write operation.

```console
$ d8 stronghold kv put -mount=secret my-secret my-value=newer-s3cr3t
Key              Value
---              -----
created_time     2024-06-19T17:31:16.662563Z
custom_metadata  <nil>
deletion_time    2024-06-19T20:56:35.662563Z
destroyed        false
version          4
```

If the number of versions exceeds `max-versions`, the oldest versions are destroyed:

```console
$ d8 stronghold kv metadata get -mount=secret my-secret
========== Metadata ==========
Key                     Value
---                     -----
cas_required            false
created_time            2024-06-19T17:20:22.985303Z
current_version         4
custom_metadata         <nil>
delete_version_after    3h25m19s
max_versions            2
oldest_version          3
updated_time            2024-06-19T17:31:16.662563Z
====== Version 3 ======
Key              Value
---              -----
created_time     2024-06-19T17:23:21.834403Z
deletion_time    n/a
destroyed        true
====== Version 4 ======
Key              Value
---              -----
created_time     2024-06-19T17:31:16.662563Z
deletion_time    2024-06-19T20:56:35.662563Z
destroyed        false
```

Secret metadata can include custom metadata as key-value pairs.
The `-custom-metadata` flag can be specified multiple times.
The `d8 stronghold kv metadata put` command fully overwrites the `custom_metadata` value:

```console
$ d8 stronghold kv metadata put -mount=secret -custom-metadata=foo=abc -custom-metadata=bar=123 my-secret
Success! Data written to: secret/metadata/my-secret
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  map[bar:123 foo:abc]
deletion_time    n/a
destroyed        false
version          2
====== Data ======
Key         Value
---         -----
foo         aa
bar         bb
```

The `d8 stronghold kv metadata patch` command partially updates the `custom_metadata` value.
For example, the following command updates the `foo` field but leaves the `bar` field unchanged:

```console
$ d8 stronghold kv metadata patch -mount=secret -custom-metadata=foo=def my-secret
Success! Data written to: secret/metadata/my-secret
```

```console
$ d8 stronghold kv get -mount=secret my-secret
====== Metadata ======
Key              Value
---              -----
created_time     2024-06-19T17:22:23.369372Z
custom_metadata  map[bar:123 foo:def]
deletion_time    n/a
destroyed        false
version          2
====== Data ======
Key         Value
---         -----
foo         aa
bar         bb
```

To delete all metadata and all versions of a key, run the following command:

```console
$ d8 stronghold kv metadata delete -mount=secret my-secret
Success! Data deleted (if it existed) at: secret/metadata/my-secret
```
