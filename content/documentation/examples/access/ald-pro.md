---
title: "Integration with ALD Pro"
linkTitle: "ALD Pro"
description: "Signing in to Stronghold with ALD Pro domain accounts: the LDAP auth method with LDAPS and group-to-policy mapping, Kerberos single sign-on, service account password rotation with the LDAP secrets engine, and common errors."
weight: 10
params:
  relatedLinks:
    - title: "LDAP auth method"
      url: ../../../user/auth/ldap/
    - title: "LDAP secrets engine"
      url: ../../../user/secrets-engines/ldap/
    - title: "Identity"
      url: ../../../concepts/identity/
    - title: "Auth methods API"
      url: ../../../reference/api/auth/
    - title: "Rotating service account passwords"
      url: ../../dynamic-credentials/static-credentials-rotation/
    - title: "Integration with Active Directory"
      url: ../active-directory/
---

ALD Pro is a Russian directory service built on FreeIPA: user and group data is stored in 389 Directory Server, and authentication is handled by MIT Kerberos. Stronghold connects to ALD Pro as a regular LDAP server and, optionally, as a Kerberos key distribution center.

## Goal

- ALD Pro domain users log in to Stronghold with their domain accounts through the `ldap` method.
- Stronghold permissions are assigned by membership in ALD Pro groups.
- Users with a valid Kerberos ticket log in without entering a password (the `kerberos` method, optional; Stronghold EE only).
- Stronghold rotates passwords of domain service accounts through the LDAP secrets engine.

![LDAP login flow with ALD Pro](../../../images/ex-ald-pro.en.png)

## Prerequisites

- A deployed ALD Pro domain, for example `ald.example.com` with domain controllers `dc01.ald.example.com` and `dc02.ald.example.com`.
- Network access from Stronghold nodes to the domain controllers on port 636 (LDAPS), and on port 88 for Kerberos. <!-- TODO(verify): set of ALD Pro ports open to external clients -->
- The domain root CA certificate. On any domain-joined host it is located at `/etc/ipa/ca.crt`.
- A Stronghold token with permissions to configure auth methods, secrets engines, policies, and Identity.
- The `ldapsearch` and `jq` utilities on the administrator's workstation.

The examples use the base DN `dc=ald,dc=example,dc=com`. Replace it with your domain's DN.

## Step 1. Prepare a service account for searches

Stronghold searches for users and groups as a separate read-only account. Do not use domain administrator accounts for this.

1. Create a system account (sysaccount) in ALD Pro. In FreeIPA, this is done with an LDIF file as `cn=Directory Manager`:

   ```text
   dn: uid=stronghold-bind,cn=sysaccounts,cn=etc,dc=ald,dc=example,dc=com
   changetype: add
   objectclass: account
   objectclass: simplesecurityobject
   uid: stronghold-bind
   userPassword: <password>
   passwordExpirationTime: 20380119031407Z
   nsIdleTimeout: 0
   ```

   ```bash
   ldapmodify -x -H ldaps://dc01.ald.example.com \
     -D "cn=Directory Manager" -W -f stronghold-bind.ldif
   ```

   <!-- TODO(verify): supported way to create a system account in ALD Pro (ALD Pro web UI or LDIF in cn=sysaccounts,cn=etc) -->

   If your installation does not use system accounts, create a regular domain user with a non-expiring password and use its DN `uid=stronghold-bind,cn=users,cn=accounts,dc=ald,dc=example,dc=com`.

1. Copy the domain CA certificate to the workstation you use to configure Stronghold:

   ```bash
   scp admin@dc01.ald.example.com:/etc/ipa/ca.crt ./ald-ca.crt
   ```

1. Check that the account can see users and groups over LDAPS:

   ```bash
   LDAPTLS_CACERT=./ald-ca.crt ldapsearch -x -H ldaps://dc01.ald.example.com \
     -D "uid=stronghold-bind,cn=sysaccounts,cn=etc,dc=ald,dc=example,dc=com" -W \
     -b "cn=users,cn=accounts,dc=ald,dc=example,dc=com" "(uid=ivanov)" uid memberOf
   ```

   The output must contain the user's `uid` attribute and the list of groups in `memberOf`.

## Step 2. Configure the LDAP method

1. Enable the auth method:

   ```bash
   d8 stronghold auth enable ldap
   ```

1. Configure the connection to ALD Pro. Groups are searched by the `member` attribute of group objects:

   ```bash
   d8 stronghold write auth/ldap/config \
     url="ldaps://dc01.ald.example.com:636,ldaps://dc02.ald.example.com:636" \
     certificate=@ald-ca.crt \
     insecure_tls=false \
     starttls=false \
     binddn="uid=stronghold-bind,cn=sysaccounts,cn=etc,dc=ald,dc=example,dc=com" \
     bindpass='<password>' \
     userdn="cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     userattr="uid" \
     groupdn="cn=groups,cn=accounts,dc=ald,dc=example,dc=com" \
     groupfilter="(&(objectClass=groupOfNames)(member={{.UserDN}}))" \
     groupattr="cn" \
     username_as_alias=true
   ```

   <!-- TODO(verify): objectClass of user groups in ALD Pro (groupOfNames / ipausergroup) -->

   Key parameters:

   | Parameter | Value for ALD Pro |
   | --- | --- |
   | `url` | Comma-separated list of domain controllers. Stronghold tries them in order if a connection fails. |
   | `certificate` | Domain CA certificate from `/etc/ipa/ca.crt`. |
   | `insecure_tls` | `false`: server certificate verification is mandatory. |
   | `starttls` | `false` for `ldaps://`. If you use `ldap://` on port 389, set it to `true`. |
   | `userdn`, `userattr` | FreeIPA user container and the `uid` login attribute. |
   | `groupdn`, `groupfilter`, `groupattr` | Group container and a filter that finds groups where the user is listed in `member`. The group name is taken from `cn`. |
   | `username_as_alias` | The entity alias name matches the domain username. |

   Alternatively, read groups from the `memberOf` attribute of the user object:

   ```bash
   d8 stronghold write auth/ldap/config \
     groupdn="cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     groupfilter="(&(objectClass=person)(uid={{.Username}}))" \
     groupattr="memberOf"
   ```

   In this case the group list also includes roles, privileges, HBAC rules, and other objects the user is a member of. The `member` variant returns only groups from `cn=groups,cn=accounts`. <!-- TODO(verify): contents of memberOf for ALD Pro users -->

   The `write` command replaces the whole configuration. When changing individual parameters, repeat all the others in the command.

1. Check login with a domain account:

   ```bash
   d8 stronghold login -method=ldap username=ivanov
   ```

   Specify the username without the domain. If no policies are assigned to groups yet, the token gets only the `default` policy.

## Step 3. Map ALD Pro groups to policies

Use one of the two approaches. Do not assign policies to the same group in both ways: it complicates auditing permissions.

### With LDAP method groups

The group name in Stronghold must match the `cn` of the group in ALD Pro:

```bash
d8 stronghold write auth/ldap/groups/stronghold-admins policies=admin
d8 stronghold write auth/ldap/groups/developers policies=dev-read
```

This approach suits setups where Stronghold uses only one login method.

### With external Identity groups

External groups let you assign policies once and use them for several login methods (for example, `ldap` and `kerberos`), and also apply MFA rules to the group.

1. Create an external group and save its ID:

   ```bash
   GROUP_ID=$(d8 stronghold write -field=id identity/group \
     name="ald-stronghold-admins" \
     type="external" \
     policies="admin")
   ```

1. Get the accessor of the `ldap/` method:

   ```bash
   LDAP_ACCESSOR=$(d8 stronghold auth list -format=json | jq -r '."ldap/".accessor')
   ```

1. Create a group alias. The `name` value must match the group name in ALD Pro:

   ```bash
   d8 stronghold write identity/group-alias \
     name="stronghold-admins" \
     mount_accessor="$LDAP_ACCESSOR" \
     canonical_id="$GROUP_ID"
   ```

Stronghold updates the entity's membership in external groups on every login and token renewal. Group changes in ALD Pro do not affect tokens that have already been issued.

## Step 4. Configure Kerberos single sign-on (optional)

{{< alert level="warning" >}}
The Stronghold documentation does not yet have a separate guide for the `kerberos` method; the parameters below come from the [auth methods API reference](../../../reference/api/auth/). Test the scenario in a lab environment before rolling it out.
{{< /alert >}}

{{< alert level="info" >}}
The `kerberos` auth method (like `saml`) is available only in Stronghold EE.
{{< /alert >}}

The `kerberos` method accepts an SPNEGO header from a client with a valid Kerberos ticket and gets the user's groups from LDAP.

1. Create a service principal in ALD Pro for the Stronghold HTTP address and get a keytab:

   ```bash
   kinit admin
   ipa service-add HTTP/stronghold.ald.example.com
   ipa-getkeytab -s dc01.ald.example.com \
     -p HTTP/stronghold.ald.example.com \
     -k stronghold.keytab
   ```

   <!-- TODO(verify): creating a service principal and keytab with ALD Pro tools; whether a host entry for stronghold.ald.example.com is required in the domain -->

1. Enable the method and pass the Base64-encoded keytab:

   ```bash
   d8 stronghold auth enable kerberos
   base64 -w0 stronghold.keytab > stronghold.keytab.b64
   d8 stronghold write auth/kerberos/config \
     keytab=@stronghold.keytab.b64 \
     service_account="HTTP/stronghold.ald.example.com"
   ```

   <!-- TODO(verify): format of service_account for a FreeIPA principal; whether remove_instance_name=true is needed -->

1. Configure group lookup in LDAP. The `auth/kerberos/config/ldap` parameters are the same as those of `auth/ldap/config`:

   ```bash
   d8 stronghold write auth/kerberos/config/ldap \
     url="ldaps://dc01.ald.example.com:636,ldaps://dc02.ald.example.com:636" \
     certificate=@ald-ca.crt \
     insecure_tls=false \
     binddn="uid=stronghold-bind,cn=sysaccounts,cn=etc,dc=ald,dc=example,dc=com" \
     bindpass='<password>' \
     userdn="cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     userattr="uid" \
     groupdn="cn=groups,cn=accounts,dc=ald,dc=example,dc=com" \
     groupfilter="(&(objectClass=groupOfNames)(member={{.UserDN}}))" \
     groupattr="cn"
   ```

1. Assign policies to groups:

   ```bash
   d8 stronghold write auth/kerberos/groups/stronghold-admins policies=admin
   ```

   If you use external Identity groups, create a second alias for each group with the `kerberos/` method accessor instead of this step.

1. Check login from a domain-joined machine:

   ```bash
   kinit ivanov
   curl --negotiate -u : https://stronghold.ald.example.com/v1/auth/kerberos/login
   ```

   <!-- TODO(verify): SPNEGO login with curl and support for the kerberos method in d8 stronghold login and the web UI -->

   The response contains a Stronghold token in the `auth.client_token` field.

## Step 5. Rotate service account passwords

The LDAP secrets engine changes passwords of existing domain accounts and hands them out to applications. For ALD Pro, use the `openldap` schema: the password is stored in the `userPassword` attribute.

{{< alert level="warning" >}}
In FreeIPA, a password set by another account is marked as expired by default, and the user must change it at the next login. The account that Stronghold uses to change passwords must be listed in `passSyncManagersDNs` of the `cn=ipa_pwd_extop,cn=plugins,cn=config` entry; otherwise rotated passwords expire immediately.
{{< /alert >}}

<!-- TODO(verify): password change behavior and passSyncManagersDNs in ALD Pro; permissions the stronghold-rotator account needs on userPassword -->

1. Create a `stronghold-rotator` account in ALD Pro with permission to change passwords of the target service accounts.

1. Enable the secrets engine and configure the connection:

   ```bash
   d8 stronghold secrets enable ldap
   d8 stronghold write ldap/config \
     url="ldaps://dc01.ald.example.com:636" \
     certificate=@ald-ca.crt \
     insecure_tls=false \
     binddn="uid=stronghold-rotator,cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     bindpass='<password>' \
     userdn="cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
     userattr="uid" \
     schema=openldap
   ```

1. Rotate the `stronghold-rotator` password so that only Stronghold knows it:

   ```bash
   d8 stronghold write -f ldap/rotate-root
   ```

   The new password cannot be retrieved after rotation. If Stronghold loses access to ALD Pro, reset the password with domain tools.

### Static role

A static role binds one domain account to a name in Stronghold and changes its password on a schedule:

```bash
d8 stronghold write ldap/static-role/svc-gitlab \
  dn="uid=svc-gitlab,cn=users,cn=accounts,dc=ald,dc=example,dc=com" \
  username="svc-gitlab" \
  rotation_period="24h"
```

The application reads the current password:

```bash
d8 stronghold read ldap/static-cred/svc-gitlab
```

The `ttl` field in the response shows the time until the next rotation. To rotate immediately:

```bash
d8 stronghold write -f ldap/rotate-role/svc-gitlab
```

### Account library (check-out)

A library lends an account from a shared pool for a limited time and changes its password after it is returned. It suits shared accounts used by on-call engineers.

```bash
d8 stronghold write ldap/library/ops-shared \
  service_account_names="svc-ops1,svc-ops2" \
  ttl=1h \
  max_ttl=4h
```

The `service_account_names` values are matched against the attribute set in `userattr` (here `uid`), so specify the `uid` value, not the full DN.

Check-out, status, and check-in:

```bash
d8 stronghold write -f ldap/library/ops-shared/check-out
d8 stronghold read ldap/library/ops-shared/status
d8 stronghold write ldap/library/ops-shared/check-in service_account_names="svc-ops1"
```

Example policy for the application and on-call engineers:

```hcl
path "ldap/static-cred/svc-gitlab" {
  capabilities = ["read"]
}

path "ldap/library/ops-shared/check-out" {
  capabilities = ["update"]
}

path "ldap/library/ops-shared/check-in" {
  capabilities = ["update"]
}

path "ldap/library/ops-shared/status" {
  capabilities = ["read"]
}
```

## Verification

1. Log in as a user from the `stronghold-admins` group and check the token policies:

   ```bash
   d8 stronghold login -method=ldap username=ivanov
   d8 stronghold token lookup
   ```

   `policies` (or `identity_policies` when you use external groups) must contain the `admin` policy.

1. Make sure a user outside the group gets only `default`.

1. Read the static role password and run `ldapsearch -D "uid=svc-gitlab,..."` with it: the bind must succeed.

## Troubleshooting

| Symptom | Likely cause | What to do |
| --- | --- | --- |
| `ldap operation failed: failed to bind as user` or `Invalid Credentials` at login | Wrong user password, wrong `userdn`/`userattr`, or wrong `binddn` password | Check search and bind with the `ldapsearch` command from step 1. Check that the user is in `cn=users,cn=accounts`. |
| `x509: certificate signed by unknown authority` | The domain CA is missing from `certificate`, or a different certificate is specified | Pass `/etc/ipa/ca.crt` in `certificate`. Do not enable `insecure_tls`. |
| `x509: certificate is valid for ..., not ...` | The address in `url` does not match the name in the controller certificate | Use the controllers' FQDNs, not IP addresses. |
| Login succeeds but the token gets only `default` | `groupfilter` finds no groups, or group names do not match `auth/ldap/groups/<name>` and the aliases | Run `ldapsearch` with the `groupfilter` filter, substituting the user DN. Compare the groups' `cn` with the names in Stronghold. |
| A user cannot log in after several failed attempts, although the password is correct | Stronghold lockout was triggered (`lockout_threshold`, 5 attempts for 15 minutes by default), or the account is locked in ALD Pro | Wait for `lockout_duration` to pass or check the account status in ALD Pro. Lockout settings are described in [LDAP auth method](../../../user/auth/ldap/#user-lockout). |
| An application cannot log in with the password from `static-cred` | ALD Pro marked the password as expired | Add the Stronghold account to `passSyncManagersDNs` and run `rotate-role`. |

## Cleanup

```bash
d8 stronghold auth disable kerberos
d8 stronghold auth disable ldap
d8 stronghold secrets disable ldap
```

Disabling the secrets engine does not change account passwords. Change them with ALD Pro tools, then delete the `stronghold-bind` and `stronghold-rotator` accounts.
