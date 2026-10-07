---
title: "Rotating service account passwords"
linkTitle: "Static credentials rotation"
description: "Automatic password rotation for existing service accounts in databases and LDAP with static roles and service account libraries."
weight: 60
params:
  relatedLinks:
    - title: "Database secrets engine"
      url: ../../../user/secrets-engines/databases/overview/
    - title: "LDAP secrets engine"
      url: ../../../user/secrets-engines/ldap/
    - title: "Secrets engines API"
      url: ../../../reference/api/secrets/
    - title: "Stronghold Agent: templates"
      url: ../../../user/agent/key-features/
---

An application cannot always work with dynamic credentials: some systems require a pre-created account with a fixed name. In this case, Stronghold takes over the password of an existing account and changes it on a schedule. Consumers always get the current password from Stronghold.

## Goal

Set up automatic password rotation for a PostgreSQL service account and for an LDAP or Active Directory account, and set up checking out accounts from a shared pool (library) with check-in.

## Prerequisites

- A Stronghold token with permissions to configure the `database` and `ldap` secrets engines.
- A configured PostgreSQL connection in the `database` engine (for example, `database/config/myapp-postgres` from [Dynamic PostgreSQL credentials](../postgresql/)).
- An existing `billing_svc` account in PostgreSQL.
- An LDAP or Active Directory server and an account for Stronghold that can change service account passwords.

{{< alert level="warning" >}}
Do not assign a static role to the account that Stronghold uses to connect in `config/`. After rotation, the password in `config/` becomes invalid, and all roles of this connection stop working. To change the password of the Stronghold service account, use `rotate-root`.
{{< /alert >}}

## Part 1. Database static role

1. Allow the connection to use the new role by adding it to `allowed_roles`:

   ```bash
   d8 stronghold write database/config/myapp-postgres allowed_roles="myapp,billing-svc"
   ```

   On update, the passed parameters are merged with the stored ones, so you do not need to pass the other connection parameters again.

1. Create a static role. Stronghold changes the account password immediately and then every 24 hours:

   ```bash
   d8 stronghold write database/static-roles/billing-svc \
     db_name="myapp-postgres" \
     username="billing_svc" \
     rotation_period=24h
   ```

1. Create a policy for consumers:

   ```bash
   d8 stronghold policy write billing-svc-creds - <<'POLICY'
   path "database/static-creds/billing-svc" {
     capabilities = ["read"]
   }
   POLICY
   ```

1. Get the current credentials:

   ```bash
   d8 stronghold read database/static-creds/billing-svc
   ```

   The `ttl` field shows the time until the next rotation, and `last_vault_rotation` shows when the password was last changed.

1. If needed, change the password immediately, for example if you suspect a compromise:

   ```bash
   d8 stronghold write -f database/rotate-role/billing-svc
   ```

The application must request the password again after each rotation. For applications that read the password from a file, use [Stronghold Agent templates](../../../user/agent/key-features/): Agent re-renders the file after rotation and runs a reload command.

## Part 2. LDAP static role

1. Enable the secrets engine and configure the connection. For Active Directory, set `schema=ad`:

   ```bash
   d8 stronghold secrets enable ldap
   d8 stronghold write ldap/config \
     binddn="CN=stronghold,OU=Service,DC=example,DC=com" \
     bindpass="<password>" \
     url=ldaps://dc01.example.com \
     schema=ad
   ```

1. Rotate the Stronghold account password so that only Stronghold knows it:

   ```bash
   d8 stronghold write -f ldap/rotate-root
   ```

1. Create a static role for the service account:

   ```bash
   d8 stronghold write ldap/static-role/svc-backup \
     dn="CN=svc-backup,OU=Service,DC=example,DC=com" \
     username="svc-backup" \
     rotation_period=24h
   ```

1. Get the current password:

   ```bash
   d8 stronghold read ldap/static-cred/svc-backup
   ```

1. Rotate the password manually if needed:

   ```bash
   d8 stronghold write -f ldap/rotate-role/svc-backup
   ```

The LDAP engine does not hash the password before writing it. Make sure the LDAP server has a password policy that stores passwords hashed. For details, see [LDAP secrets engine](../../../user/secrets-engines/ldap/).

## Part 3. Service account library

A library is a pool of service accounts that are lent out temporarily. When an account is checked in, Stronghold changes its password, so the password given to the previous user stops working.

1. Create a library:

   ```bash
   d8 stronghold write ldap/library/ops-team \
     service_account_names="ops-svc1@example.com,ops-svc2@example.com" \
     ttl=4h \
     max_ttl=8h
   ```

1. Check out an account:

   ```bash
   d8 stronghold write -f ldap/library/ops-team/check-out
   ```

   The response contains `service_account_name`, `password`, and `lease_id`.

1. Check the account back in after work:

   ```bash
   d8 stronghold write ldap/library/ops-team/check-in \
     service_account_names="ops-svc1@example.com"
   ```

   If the account is not checked in, Stronghold checks it in automatically when the lease `ttl` expires.

## Verification

1. Read `database/static-creds/billing-svc`, run `rotate-role`, and read the credentials again: the password must change.
1. Connect to PostgreSQL with the new password and make sure the old password is rejected.
1. Check the library status:

   ```bash
   d8 stronghold read ldap/library/ops-team/status
   ```

   Checked-out accounts are marked `available:false`.

## Cleanup

Deleting a static role does not change the password. Before deleting, rotate the password manually so that the last value issued by Stronghold stops working:

```bash
d8 stronghold write -f database/rotate-role/billing-svc
d8 stronghold delete database/static-roles/billing-svc

d8 stronghold write -f ldap/rotate-role/svc-backup
d8 stronghold delete ldap/static-role/svc-backup

d8 stronghold delete ldap/library/ops-team
d8 stronghold policy delete billing-svc-creds
```

After deletion, hand password management back to the account owners or disable the accounts.
