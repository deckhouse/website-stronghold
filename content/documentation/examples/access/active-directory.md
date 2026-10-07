---
title: "Integration with Active Directory"
linkTitle: "Active Directory"
description: "Signing in to Stronghold with Active Directory accounts through the LDAP method with LDAPS, nested groups, and UPN, and rotating AD service account passwords with the LDAP secrets engine and the ad schema."
weight: 20
params:
  relatedLinks:
    - title: "LDAP auth method"
      url: ../../../user/auth/ldap/
    - title: "LDAP secrets engine"
      url: ../../../user/secrets-engines/ldap/
    - title: "Identity"
      url: ../../../concepts/identity/
    - title: "Rotating service account passwords"
      url: ../../dynamic-credentials/static-credentials-rotation/
    - title: "Integration with ALD Pro"
      url: ../ald-pro/
---

Stronghold connects to Active Directory (AD) over LDAP: it verifies domain user passwords, resolves their groups, including nested ones, and assigns policies by group. The LDAP secrets engine with the `ad` schema changes passwords of domain service accounts.

## Goal

- AD users log in to Stronghold with their domain accounts through the `ldap` method over LDAPS.
- Stronghold permissions are assigned by AD group membership, including nested groups.
- Stronghold rotates AD service account passwords and hands them out to applications.

## Prerequisites

- An AD domain, for example `corp.example.com` with domain controllers `dc01.corp.example.com` and `dc02.corp.example.com`.
- LDAPS (port 636) enabled on the domain controllers with a certificate issued by the corporate CA.
- The root CA certificate in PEM (Base-64) format, for example `corp-ca.pem`.
- Network access from Stronghold nodes to the domain controllers on port 636.
- A Stronghold token with permissions to configure auth methods, secrets engines, policies, and Identity.

The examples assume users are in `OU=Users,DC=corp,DC=example,DC=com`, groups in `OU=Groups,DC=corp,DC=example,DC=com`, and service accounts in `OU=Service Accounts,DC=corp,DC=example,DC=com`. Replace the DNs with your own.

## Step 1. Prepare AD accounts

1. Create a `svc-stronghold-bind` account for searching users and groups. Regular domain user permissions to read the directory are enough. Set a non-expiring password.

1. If you plan to rotate passwords, create a `svc-stronghold-rotator` account and delegate the **Reset password** permission on the service account OU to it. Do not grant it domain administrator rights.

1. Check search over LDAPS:

   ```bash
   LDAPTLS_CACERT=./corp-ca.pem ldapsearch -x -H ldaps://dc01.corp.example.com \
     -D "svc-stronghold-bind@corp.example.com" -W \
     -b "OU=Users,DC=corp,DC=example,DC=com" "(sAMAccountName=ivanov)" \
     sAMAccountName userPrincipalName memberOf
   ```

## Step 2. Configure the LDAP method

1. Enable the auth method:

   ```bash
   d8 stronghold auth enable ldap
   ```

1. Configure the connection. Users log in with `sAMAccountName`; groups are resolved including nesting through the `LDAP_MATCHING_RULE_IN_CHAIN` rule (OID `1.2.840.113556.1.4.1941`):

   ```bash
   d8 stronghold write auth/ldap/config \
     url="ldaps://dc01.corp.example.com:636,ldaps://dc02.corp.example.com:636" \
     certificate=@corp-ca.pem \
     insecure_tls=false \
     starttls=false \
     binddn="CN=svc-stronghold-bind,OU=Service Accounts,DC=corp,DC=example,DC=com" \
     bindpass='<password>' \
     userdn="OU=Users,DC=corp,DC=example,DC=com" \
     userattr="sAMAccountName" \
     groupdn="OU=Groups,DC=corp,DC=example,DC=com" \
     groupfilter="(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))" \
     groupattr="cn" \
     username_as_alias=true
   ```

   Key parameters:

   | Parameter | Value for AD |
   | --- | --- |
   | `url` | Comma-separated list of domain controllers, port 636. |
   | `certificate` | Certificate of the root CA that issued the controller certificates. |
   | `insecure_tls`, `starttls` | `false` for `ldaps://`. When connecting via `ldap://` on port 389, set `starttls=true`. |
   | `userattr` | `sAMAccountName`: the short login name, for example `ivanov`. |
   | `groupfilter` | Finds all groups the user belongs to directly or through nested groups. |
   | `groupattr` | `cn`: the group name in Stronghold. |

   Instead of `groupfilter`, you can set `use_token_groups=true`: Stronghold reads the user's computed `tokenGroups` attribute, which contains all security groups, including nested ones.

1. If users must log in with a UPN (`ivanov@corp.example.com`), replace `userattr` with the `upndomain` parameter:

   ```bash
   d8 stronghold write auth/ldap/config \
     url="ldaps://dc01.corp.example.com:636,ldaps://dc02.corp.example.com:636" \
     certificate=@corp-ca.pem \
     insecure_tls=false \
     binddn="CN=svc-stronghold-bind,OU=Service Accounts,DC=corp,DC=example,DC=com" \
     bindpass='<password>' \
     userdn="OU=Users,DC=corp,DC=example,DC=com" \
     upndomain="corp.example.com" \
     groupdn="OU=Groups,DC=corp,DC=example,DC=com" \
     groupfilter="(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))" \
     groupattr="cn"
   ```

   The user still enters the short name, and Stronghold binds as `ivanov@corp.example.com` and searches for the user by `userPrincipalName`.

   The `write` command replaces the whole configuration. When changing individual parameters, repeat all the others in the command.

1. Check login:

   ```bash
   d8 stronghold login -method=ldap username=ivanov
   ```

## Step 3. Map AD groups to policies

For a single login method, LDAP method groups are enough. The group name must match the group's `cn` in AD:

```bash
d8 stronghold write auth/ldap/groups/SG-Stronghold-Admins policies=admin
d8 stronghold write auth/ldap/groups/SG-Developers policies=dev-read
```

If AD groups are also used by other login methods, or you need to apply MFA to them, create external Identity groups:

```bash
GROUP_ID=$(d8 stronghold write -field=id identity/group \
  name="ad-stronghold-admins" type="external" policies="admin")
LDAP_ACCESSOR=$(d8 stronghold auth list -format=json | jq -r '."ldap/".accessor')
d8 stronghold write identity/group-alias \
  name="SG-Stronghold-Admins" \
  mount_accessor="$LDAP_ACCESSOR" \
  canonical_id="$GROUP_ID"
```

For details on external groups, see [Identity](../../../concepts/identity/#external-vs-internal-groups).

## Step 4. Rotate service account passwords

{{< alert level="info" >}}
AD allows changing a password (the `unicodePwd` attribute) only over a secure connection. Use `ldaps://` or `starttls=true`.
{{< /alert >}}

1. Enable the secrets engine and configure it with the `ad` schema:

   ```bash
   d8 stronghold secrets enable ldap
   d8 stronghold write ldap/config \
     url="ldaps://dc01.corp.example.com:636" \
     certificate=@corp-ca.pem \
     insecure_tls=false \
     binddn="CN=svc-stronghold-rotator,OU=Service Accounts,DC=corp,DC=example,DC=com" \
     bindpass='<password>' \
     userdn="OU=Service Accounts,DC=corp,DC=example,DC=com" \
     schema=ad
   ```

1. Rotate the `svc-stronghold-rotator` password so that only Stronghold knows it:

   ```bash
   d8 stronghold write -f ldap/rotate-root
   ```

1. Create a static role for the application service account:

   ```bash
   d8 stronghold write ldap/static-role/svc-jenkins \
     dn="CN=svc-jenkins,OU=Service Accounts,DC=corp,DC=example,DC=com" \
     username="svc-jenkins" \
     rotation_period="24h"
   ```

1. The application reads the current password:

   ```bash
   d8 stronghold read ldap/static-cred/svc-jenkins
   ```

For shared accounts lent to people for a limited time, use a library:

```bash
d8 stronghold write ldap/library/helpdesk \
  service_account_names="svc-helpdesk1@corp.example.com,svc-helpdesk2@corp.example.com" \
  ttl=2h \
  max_ttl=8h
d8 stronghold write -f ldap/library/helpdesk/check-out
d8 stronghold write ldap/library/helpdesk/check-in service_account_names="svc-helpdesk1@corp.example.com"
```

The `service_account_names` values are matched against the attribute set in the `userattr` parameter; for the `ad` schema it defaults to `userPrincipalName`, so use the full name such as `svc-helpdesk1@corp.example.com`.

Access policies for `static-cred` and `library` are similar to those in [Integration with ALD Pro](../ald-pro/#step-5-rotate-service-account-passwords).

## Verification

1. Log in as a user from a group nested in `SG-Stronghold-Admins` and check that the token got the `admin` policy:

   ```bash
   d8 stronghold login -method=ldap username=ivanov
   d8 stronghold token lookup
   ```

1. Read the static role password and check the bind:

   ```bash
   LDAPTLS_CACERT=./corp-ca.pem ldapsearch -x -H ldaps://dc01.corp.example.com \
     -D "svc-jenkins@corp.example.com" -w '<password from static-cred>' \
     -b "DC=corp,DC=example,DC=com" -s base
   ```

## Troubleshooting

On a bind error, AD returns code `49` and a subcode in the `data` field of the error message. The subcode indicates the cause.

| Symptom | Likely cause | What to do |
| --- | --- | --- |
| `Invalid Credentials`, `data 52e` | Wrong user password or `bindpass` | Check the password. After `rotate-root`, only Stronghold knows the `binddn` password. |
| `data 525` | User not found | Check `userdn`, `userattr`, or `upndomain`. |
| `data 775` | The account is locked in AD | Unlock it in AD. Also check the Stronghold lockout (`lockout_threshold`) described in [LDAP auth method](../../../user/auth/ldap/#user-lockout). |
| `data 533`, `data 532`, `data 701` | The account is disabled, or the password or account has expired | Enable the account or change the password in AD. |
| `x509: certificate signed by unknown authority` | The CA certificate is missing or wrong | Pass the corporate root CA certificate in `certificate`. |
| Login succeeds but group policies are missing | `groupfilter` finds no groups, or group names do not match | Run `ldapsearch` with the `(member:1.2.840.113556.1.4.1941:=<user DN>)` filter. Group names in Stronghold must match `cn`. |
| `rotate-role` returns `Unwilling To Perform` or `Insufficient Access Rights` | No secure connection, `binddn` lacks the **Reset password** permission, or the password does not meet the domain password policy | Use LDAPS, check the delegation, and configure a [Stronghold password policy](../../../concepts/password-policy/) with the `password_policy` parameter. |

## Cleanup

```bash
d8 stronghold auth disable ldap
d8 stronghold secrets disable ldap
```

Disabling the secrets engine does not change passwords. Change the service account passwords with AD tools.
