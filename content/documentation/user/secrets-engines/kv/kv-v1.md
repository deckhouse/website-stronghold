---
title: "KV v1"
weight: 20
---

The version 1 `kv` secrets engine is intended for storing arbitrary secrets in the Stronghold storage.
When writing to a key, the previous value is replaced with the new one.

Key names must be strings.
If you write non-string values directly via the CLI, Stronghold converts them to strings.
To store non-string values, pass key-value pairs from a JSON file or use the HTTP API.

The `kv` secrets engine distinguishes between `create` and `update` operations in ACL policies.

{{< alert level="warning" >}}
Paths and key names are not obfuscated or encrypted.
Only key values are encrypted.
Do not store sensitive data in the secret path or key name.
{{< /alert >}}

## Enable

To enable version 1 of the `kv` storage, run the following command:

```shell
d8 stronghold secrets enable -version=1 kv
```

## Usage

The `kv` secrets engine lets you write keys with arbitrary values.
A token with the appropriate permissions is required.

Perform the following steps:

1. Write arbitrary data:

   ```console
   $ d8 stronghold kv put kv/my-secret my-value=s3cr3t
   Success! Data written to: kv/my-secret
   ```

1. Read the data:

   ```console
   $ d8 stronghold kv get kv/my-secret
   Key                 Value
   ---                 -----
   my-value            s3cr3t
   ```

1. Get the list of keys:

   ```console
   $ d8 stronghold kv list kv/
   Keys
   ----
   my-secret
   ```

1. Delete the key:

   ```console
   $ d8 stronghold kv delete kv/my-secret
   Success! Data deleted (if it existed) at: kv/my-secret
   ```

You can also use the password policy mechanism to generate values.

1. Create a password policy:

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

1. Generate a password using the `example` policy:

   ```console
   $ d8 stronghold kv put kv/my-generated-secret \
       password=$(d8 stronghold read -field password sys/policies/password/example/generate)
   ```

1. Read the generated secret value:

   ```console
   $ d8 stronghold kv get kv/my-generated-secret
   ====== Data ======
   Key         Value
   ---         -----
   password    ^dajd609Xf8Zhac$dW24
   ```

## Key lifetime

Unlike other secrets engines, `kv` does not apply TTL for automatic data expiration.
The `lease_duration` value here is informational and shows how often it is recommended to check whether the value needs to be refreshed.
If the `ttl` parameter is set for a key, the `kv` secrets engine uses it as the lease duration:

```console
$ d8 stronghold kv put kv/my-secret ttl=5s my-value=s3cr3t
Success! Data written to: kv/my-secret
```

Even if `ttl` is set, the secrets engine never deletes data automatically.
The `ttl` parameter is advisory only.

When reading a secret with a `ttl` value, both the `ttl` key itself and the refresh interval reflect that value:

```console
$ d8 stronghold kv get kv/my-secret
Key                 Value
---                 -----
my-value            s3cr3t
ttl                 5s
```

```console
$ curl -X 'GET' \
    'https://stronghold.example.com/v1/kv/my-secret' \
    -H 'X-Vault-Token: ***'
{
  "request_id": "3879d849-cb78-725a-c2eb-3ba9dfe8a1d3",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 5,
  "data": {
    "my-value": "s3cr3t",
    "ttl": "5s"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}
```
