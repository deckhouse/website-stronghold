---
title: "KV/v1"
weight: 20
---

The `kv` secrets engine is used to store arbitrary secrets within the
configured physical storage for Stronghold.

Writing to a key in the `kv` backend will replace the old value; sub-fields are
not merged together.

Key names must always be strings. If you write non-string values directly via
the CLI, they will be converted into strings. However, you can preserve
non-string values by writing the key/value pairs to Stronghold from a JSON file or
using the HTTP API.

This secrets engine honors the distinction between the `create` and `update`
capabilities inside ACL policies.

{{< alert level="warning" >}}

**Note**: Path and key names are _not_ obfuscated or encrypted; only the
values set on keys are. You should not store sensitive information as part of a
secret's path.

{{< /alert >}}

## Setup

To enable a version 1 kv store:

{{< tabs name="stronghold_cmd_35718" >}}
{{% tab name="Stronghold in DKP" %}}
```shell-session
d8 stronghold secrets enable -version=1 kv
```
{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}
```shell-session
stronghold secrets enable -version=1 kv
```
{{% /tab %}}
{{< /tabs >}}

## Usage

After the secrets engine is configured and a user/machine has an Stronghold token with
the proper permission, it can generate credentials. The `kv` secrets engine
allows for writing keys with arbitrary values.

1. Write arbitrary data:

   {{< tabs name="stronghold_cmd_81605" >}}
   {{% tab name="Stronghold in DKP" %}}
   ```shell-session
   $ d8 stronghold kv put kv/my-secret my-value=s3cr3t
   Success! Data written to: kv/my-secret
   ```
   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}
   ```shell-session
   $ stronghold kv put kv/my-secret my-value=s3cr3t
   Success! Data written to: kv/my-secret
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Read arbitrary data:

   {{< tabs name="stronghold_cmd_34392" >}}
   {{% tab name="Stronghold in DKP" %}}
   ```shell-session
   $ d8 stronghold kv get kv/my-secret
   Key                 Value
   ---                 -----
   my-value            s3cr3t
   ```
   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}
   ```shell-session
   $ stronghold kv get kv/my-secret
   Key                 Value
   ---                 -----
   my-value            s3cr3t
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. List the keys:

   {{< tabs name="stronghold_cmd_87945" >}}
   {{% tab name="Stronghold in DKP" %}}
   ```shell-session
   $ d8 stronghold kv list kv/
   Keys
   ----
   my-secret
   ```
   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}
   ```shell-session
   $ stronghold kv list kv/
   Keys
   ----
   my-secret
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Delete a key:

   {{< tabs name="stronghold_cmd_3924" >}}
   {{% tab name="Stronghold in DKP" %}}
   ```shell-session
   $ d8 stronghold kv delete kv/my-secret
   Success! Data deleted (if it existed) at: kv/my-secret
   ```
   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}
   ```shell-session
   $ stronghold kv delete kv/my-secret
   Success! Data deleted (if it existed) at: kv/my-secret
   ```
   {{% /tab %}}
   {{< /tabs >}}

You can also use Stronghold's password policy feature to generate arbitrary values.

1. Write a password policy:

   {{< tabs name="stronghold_cmd_15069" >}}
   {{% tab name="Stronghold in DKP" %}}
   ```shell-session
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
   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}
   ```shell-session
   $ stronghold write sys/policies/password/example policy=-<<EOF

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
   {{% /tab %}}
   {{< /tabs >}}

1. Write data using the `example` policy:

   {{< tabs name="stronghold_cmd_95485" >}}
   {{% tab name="Stronghold in DKP" %}}
   ```shell-session
   $ d8 stronghold kv put kv/my-generated-secret \
       password=$(d8 stronghold read -field password sys/policies/password/example/generate)
   ```
   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}
   ```shell-session
   $ stronghold kv put kv/my-generated-secret \
       password=$(stronghold read -field password sys/policies/password/example/generate)
   ```
   {{% /tab %}}
   {{< /tabs >}}

1. Read the generated data:

   {{< tabs name="stronghold_cmd_86790" >}}
   {{% tab name="Stronghold in DKP" %}}
   ```shell-session
   $ d8 stronghold kv get kv/my-generated-secret
   ====== Data ======
   Key         Value
   ---         -----
   password    ^dajd609Xf8Zhac$dW24
   ```
   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}
   ```shell-session
   $ stronghold kv get kv/my-generated-secret
   ====== Data ======
   Key         Value
   ---         -----
   password    ^dajd609Xf8Zhac$dW24
   ```
   {{% /tab %}}
   {{< /tabs >}}

## TTLs

Unlike other secrets engines, the KV secrets engine does not enforce TTLs
for expiration. Instead, the `lease_duration` is a hint for how often consumers
should check back for a new value.

If provided a key of `ttl`, the KV secrets engine will utilize this value
as the lease duration:

{{< tabs name="stronghold_cmd_62205" >}}
{{% tab name="Stronghold in DKP" %}}
```shell-session
$ d8 stronghold kv put kv/my-secret ttl=30m my-value=s3cr3t
Success! Data written to: kv/my-secret
```
{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}
```shell-session
$ stronghold kv put kv/my-secret ttl=30m my-value=s3cr3t
Success! Data written to: kv/my-secret
```
{{% /tab %}}
{{< /tabs >}}

Even with a `ttl` set, the secrets engine _never_ removes data on its own. The
`ttl` key is merely advisory.

When reading a value with a `ttl`, both the `ttl` key _and_ the refresh interval
will reflect the value:

{{< tabs name="stronghold_cmd_85875" >}}
{{% tab name="Stronghold in DKP" %}}
```shell-session
$ d8 stronghold kv get kv/my-secret
Key                 Value
---                 -----
my-value            s3cr3t
ttl                 30m
```
{{% /tab %}}
{{% tab name="Stronghold in Linux" %}}
```shell-session
$ stronghold kv get kv/my-secret
Key                 Value
---                 -----
my-value            s3cr3t
ttl                 30m
```
{{% /tab %}}
{{< /tabs >}}
