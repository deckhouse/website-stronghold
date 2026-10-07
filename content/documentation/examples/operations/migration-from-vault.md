---
title: "Migrating from HashiCorp Vault"
linkTitle: "Migrating from Vault"
description: "Moving secrets, policies, auth methods, PKI, and dynamic secrets from HashiCorp Vault to Stronghold, with a client cutover and rollback plan."
weight: 10
params:
  relatedLinks:
    - title: "KV1/KV2 replication"
      url: ../../../admin/replication/kv-replication/
    - title: "Policies"
      url: ../../../concepts/policy/
    - title: "KV secrets engine"
      url: ../../../user/secrets-engines/kv/overview/
    - title: "PKI secrets engine"
      url: ../../../user/secrets-engines/pki/
    - title: "Namespaces"
      url: ../../../admin/namespaces/overview/
---

Stronghold is compatible with the HashiCorp Vault API, so a migration comes down to moving configuration and data and pointing clients to a new address. The encrypted Vault storage cannot be attached to Stronghold directly: data is moved through the API.

## Goal

Move KV secrets, policies, auth methods, PKI, and dynamic secrets from a running HashiCorp Vault to Stronghold, switch clients over, and keep the ability to roll back.

## Prerequisites

- A deployed and unsealed Stronghold cluster.
- A HashiCorp Vault token with `read` and `list` permissions on the paths being migrated, and on `sys/mounts`, `sys/auth`, and `sys/policies/acl`.
- A Stronghold token with administrative permissions.
- Network connectivity between Stronghold and Vault if you use KV replication.
- The `vault`, `d8`, and `jq` utilities on your workstation.

The examples below set addresses with environment variables:

```bash
export VAULT_ADDR=https://vault.example.com:8200
export VAULT_TOKEN=<vault_token>
export STRONGHOLD_ADDR=https://stronghold.example.com
```

## Step 1. Take an inventory

List the objects to migrate. Run the following commands against Vault:

```bash
vault secrets list -detailed -format=json > vault-mounts.json
vault auth list -detailed -format=json > vault-auth.json
vault policy list -format=json > vault-policies.json
vault namespace list -format=json > vault-namespaces.json
```

The `vault namespace list` command is available only in Vault Enterprise. If Vault uses namespaces, repeat the inventory for each of them with the `-namespace` flag.

For each `mount`, record the type (`kv` version 1 or 2, `pki`, `database`, `transit`, and so on), the path, and the parameters (`default_lease_ttl`, `max_lease_ttl`). For each auth method, record the type, path, roles, and attached policies.

{{< alert level="info" >}}
Namespaces are supported only in Stronghold EE. If Vault uses namespaces and the target installation is base Stronghold, move the content of each namespace to separate mount paths.
{{< /alert >}}

## Step 2. Recreate namespaces and secrets engines

1. If you need namespaces (Stronghold EE), create them:

   ```bash
   d8 stronghold namespace create team-a
   ```

1. Enable secrets engines at the same paths as in Vault. For example, for KV version 2:

   ```bash
   d8 stronghold secrets enable -path=secret -version=2 kv
   ```

   Keep the original mount paths so you do not have to change policies and client code.

## Step 3. Move KV secrets

Choose one of the strategies.

### Strategy A. Copying with a script

This works with base Stronghold and for a one-time migration. The script walks the secret tree in Vault and writes the latest version of each secret to Stronghold. KV2 version history is not migrated.

```bash
#!/usr/bin/env bash
set -euo pipefail

SRC_MOUNT=secret
DST_MOUNT=secret

copy_tree() {
  local path="$1"
  for key in $(vault kv list -mount="$SRC_MOUNT" -format=json "$path" | jq -r '.[]'); do
    if [[ "$key" == */ ]]; then
      copy_tree "${path}${key}"
    else
      vault kv get -mount="$SRC_MOUNT" -format=json "${path}${key}" \
        | jq '.data.data' > /tmp/secret.json
      d8 stronghold kv put -mount="$DST_MOUNT" "${path}${key}" @/tmp/secret.json
      echo "copied ${path}${key}"
    fi
  done
  rm -f /tmp/secret.json
}

copy_tree ""
```

For KV version 1, use the `.data` field instead of `.data.data`. Before running the script, make sure the temporary file is created in a directory only you can access, or replace it with passing data through standard input.

### Strategy B. KV replication from an external Vault

{{< alert level="info" >}}
KV replication is available only in Stronghold EE.
{{< /alert >}}

[KV1/KV2 replication](../../../admin/replication/kv-replication/) uses a pull model: Stronghold periodically fetches secrets from the source through the API. The source can be Vault CE, Vault Enterprise, or another store with a compatible KV/KV2 API. This option keeps data in sync for the whole client cutover period.

1. On the Vault side, create a policy and a token for reading the source store:

   ```bash
   vault policy write replicate-secret - <<'POLICY'
   path "secret/*" {
     capabilities = ["read", "list"]
   }
   path "sys/mounts/secret" {
     capabilities = ["read"]
   }
   path "auth/token/lookup-self" {
     capabilities = ["read"]
   }
   path "auth/token/renew-self" {
     capabilities = ["update"]
   }
   POLICY

   vault token create -policy=replicate-secret -orphan=true -period=30d \
     -wrap-ttl=5m -field=wrapping_token
   ```

1. In Stronghold, mount the replicated store:

   ```bash
   d8 stronghold secrets enable \
     -path=secret \
     -src-address="$VAULT_ADDR" \
     -src-wrapping-token=<wrapping_token> \
     -src-mount-path=secret \
     -src-ca-cert=@vault-ca.pem \
     -sync-period-min=5 \
     -version=2 \
     kv
   ```

   The KV versions in the source and in Stronghold must match.

1. Check the replication settings:

   ```bash
   d8 stronghold read sys/mounts/secret/tune
   ```

While replication is enabled, the store in Stronghold is read-only. After switching clients, disable replication to allow writes:

```bash
d8 stronghold secrets tune -sync-enable=false secret
```

{{< alert level="warning" >}}
Do not re-enable replication after disabling it: all local changes will be overwritten with data from the source.
{{< /alert >}}

## Step 4. Move policies

The ACL policy syntax is compatible, so policies are moved as is:

```bash
for p in $(vault policy list | grep -v -E '^(root|default)$'); do
  vault policy read "$p" > "policy-$p.hcl"
  d8 stronghold policy write "$p" "policy-$p.hcl"
done
```

Review policies that reference auth method accessors (`{{identity.entity.aliases.<mount accessor>...}}` templates). After you recreate auth methods, the accessor changes: replace it in policies with the new value. For details, see [Templated policies](../../../concepts/policy/#templated-policies).

## Step 5. Recreate auth methods

Move auth method configuration and roles manually or with an IaC tool: Vault does not return secret parameters (for example, OIDC `client_secret` or LDAP `bindpass`) on read.

1. Enable methods at the original paths:

   ```bash
   d8 stronghold auth enable -path=oidc oidc
   d8 stronghold auth enable -path=kubernetes kubernetes
   d8 stronghold auth enable -path=approle approle
   ```

1. Read non-secret role parameters in Vault and write them to Stronghold:

   ```bash
   vault read -format=json auth/approle/role/my-app | jq '.data' > role-my-app.json
   d8 stronghold write auth/approle/role/my-app @role-my-app.json
   ```

1. For AppRole, issue new `secret_id` values: existing ones cannot be moved from Vault. You can set `role_id` explicitly by writing it to `auth/approle/role/<role>/role-id` so that client configuration does not change.

1. For OIDC, add the Stronghold address to the list of allowed redirect URIs in the identity provider.

Method setup is described in [OIDC](../../../user/auth/oidc/overview/), [Kubernetes](../../../user/auth/kubernetes/), [AppRole](../../../user/auth/approle/), and [LDAP](../../../user/auth/ldap/).

## Step 6. Move PKI

Choose an option depending on how important the continuity of the trust chain is.

- **Reissue.** Create a new intermediate CA in Stronghold, sign it with the existing root CA, and issue new certificates from Stronghold. Old certificates remain valid until they expire. This is the preferred option: the CA private key never leaves Vault. The procedure is described in [Internal PKI](../../certificates/internal-pki/).
- **Import the CA.** If the CA key is exportable (generated as `exported` or stored outside Vault), import the certificate and key bundle into Stronghold:

  ```bash
  d8 stronghold secrets enable -path=pki pki
  d8 stronghold write pki/issuers/import/bundle pem_bundle=@ca-bundle.pem
  ```

  Then recreate the roles (`pki/roles/<name>`) and the `pki/config/urls` settings. If CRL and OCSP addresses in issued certificates point to Vault, keep the old addresses available until those certificates expire.

## Step 7. Reconfigure dynamic secrets

Dynamic credentials (`database`, `ldap`, `kubernetes`) are not migrated: they are bound to leases of the source cluster.

1. Recreate connection settings and roles in Stronghold. For a database, use a separate service account rather than the one Vault uses, otherwise a root password rotation in one system breaks the other.
1. Run `rotate-root` for the new account so that only Stronghold knows its password:

   ```bash
   d8 stronghold write -force database/rotate-root/my-database
   ```

1. After switching clients, revoke leases in Vault: `vault lease revoke -prefix database/creds/`.

## Step 8. Switch clients

Most Vault clients (CLI, SDKs, Vault Agent, External Secrets Operator, the Terraform provider) take the server address from the `VAULT_ADDR` variable or a configuration parameter. Set the Stronghold address there:

```bash
export VAULT_ADDR=https://stronghold.example.com
```

For the `d8 stronghold` utility, use the `STRONGHOLD_ADDR` variable.

Switch clients in stages: test environments first, then non-critical services, then the rest. For services that use a Vault DNS name, you can point the CNAME to Stronghold if the Stronghold certificate includes this name in SAN.

<!-- TODO(verify): list of Vault ecosystem clients whose compatibility with Stronghold is confirmed by tests -->

## Verification

1. Compare the number of secrets in Vault and Stronghold for each KV store.
1. Read several secrets through the application client and through the CLI:

   ```bash
   d8 stronghold kv get -mount=secret myapp/config
   ```

1. Log in with each migrated auth method and check the policies in the issued token: `d8 stronghold token lookup`.
1. Issue a test certificate and dynamic database credentials.
1. Make sure from the Vault logs that production clients no longer access it.

## Rollback plan

Keep Vault running until the observation period ends (usually one to two weeks).

- With strategy A, do not delete data in Vault. If secrets were changed in Stronghold after the cutover, move the changes back with the same script, swapping source and destination.
- With strategy B, rollback is simple while replication is enabled: Vault remains the source, and clients return to the old address.
- Point client `VAULT_ADDR` or the DNS record back to Vault.
- Revoke dynamic credentials issued by Stronghold: `d8 stronghold lease revoke -prefix database/creds/`.

## Cleanup

After the observation period ends:

1. Disable KV replication (`-sync-enable=false`) and revoke the replication token in Vault.
1. Delete temporary policy and role files (`policy-*.hcl`, `role-*.json`).
1. Decommission Vault according to your organization's procedures: take a final snapshot, revoke tokens and leases, and stop the servers.
