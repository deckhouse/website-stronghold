---
title: "LDAP secrets engine"
weight: 80
---

The LDAP secrets engine manages LDAP credentials and supports dynamic credential generation.
It integrates with LDAP v3 implementations, including OpenLDAP, [ALD Pro](https://www.aldpro.ru/), Active Directory, and IBM Resource Access Control Facility (RACF).

The LDAP secrets engine provides three main capabilities:

- [Managing static credentials](#static-roles);
- [Managing dynamic credentials](#dynamic-roles);
- [Rotating passwords for account libraries](#password-rotation-for-account-libraries).

## Configuration

Enable the LDAP secrets engine:

```shell
d8 stronghold secrets enable ldap
```

By default, the engine is enabled at the `ldap` path.
To use a different path, specify the `-path` argument.

Configure the credentials that Stronghold uses to connect to LDAP when generating passwords:

```shell
d8 stronghold write ldap/config \
  binddn=$USERNAME \
  bindpass=$PASSWORD \
  url=ldaps://138.91.247.105
```

{{< alert level="info" >}}
Create a dedicated account specifically for Stronghold.
{{< /alert >}}

Rotate the password so that it is stored only in Stronghold:

```shell
d8 stronghold write -f ldap/rotate-root
```

{{< alert level="warning" >}}
After rotation, you cannot retrieve the generated password from Stronghold.
{{< /alert >}}

### LDAP schemas

{: #schemas .anchored}

The LDAP secrets engine supports three schemas:

- `openldap` — the default schema;
- `racf`;
- `ad`.

#### OpenLDAP

By default, the LDAP secrets engine assumes that the account password is stored in the `userPassword` field.

For example, the following object classes support the `userPassword` field:

- `organization`;
- `organizationalUnit`;
- `organizationalRole`;
- `inetOrgPerson`;
- `person`;
- `posixAccount`.

#### Resource Access Control Facility

To manage IBM Resource Access Control Facility (RACF), configure the LDAP secrets engine to use the `racf` schema.

For RACF support, generated passwords must be no longer than eight characters.
You can configure the password length with a password policy:

```shell
d8 stronghold write ldap/config \
  binddn=$USERNAME \
  bindpass=$PASSWORD \
  url=ldaps://138.91.247.105 \
  schema=racf \
  password_policy=racf_password_policy
```

#### Active Directory

To manage passwords in Active Directory, configure the LDAP secrets engine to use the `ad` schema:

```shell
d8 stronghold write ldap/config \
  binddn=$USERNAME \
  bindpass=$PASSWORD \
  url=ldaps://138.91.247.105 \
  schema=ad
```

### Static roles

{: #static-roles .anchored}

#### Configuration

Configure a static role that maps a Stronghold name to an LDAP entry.
This role manages the password rotation settings.

```shell
d8 stronghold write ldap/static-role/lf-edge \
  dn='uid=lf-edge,ou=users,dc=lf-edge,dc=com' \
  username='stronghold' \
  rotation_period="24h"
```

Read the credentials for the `lf-edge` role:

```shell
d8 stronghold read ldap/static-cred/lf-edge
```

### Password rotation

You can manage passwords in two ways:

- automatic time-based rotation;
- manual rotation.

### Automatic password rotation

Passwords rotate automatically according to the `rotation_period` configured in the static role.
The minimum value is `5s`.

When you read credentials for a static role, the response includes the time until the next rotation in the `ttl` field.

At the moment, automatic rotation is supported only for static roles.
Rotate the `binddn` account used by Stronghold with the `rotate-root` call to generate a password known only to Stronghold.

### Manual rotation

You can rotate static role passwords manually with the `rotate-role` call.
Manual rotation restarts the rotation period.

### Deleting static roles

When you delete a static role, passwords do not change.
Before deleting the role or revoking access to the static role, rotate the password manually.

### Dynamic roles

{: #dynamic-roles .anchored}

#### Configuration

Configure a dynamic role by calling `/role/:role_name`:

```shell
d8 stronghold write ldap/role/dynamic-role \
  creation_ldif=@/path/to/creation.ldif \
  deletion_ldif=@/path/to/deletion.ldif \
  rollback_ldif=@/path/to/rollback.ldif \
  default_ttl=1h \
  max_ttl=24h
```

{{< alert level="warning" >}}
The `rollback_ldif` argument is optional, but recommended.
Stronghold executes the operations defined in `rollback_ldif` if creation fails.
This helps ensure that all objects are removed when an error occurs.
{{< /alert >}}

To generate credentials, run:

```shell
d8 stronghold read ldap/creds/dynamic-role
```

Example output:

```console
Key                    Value
---                    -----
lease_id               ldap/creds/dynamic-role/HFgd6uKaDomVMvJpYbn9q4q5
lease_duration         1h
lease_renewable        true
distinguished_names    [cn=v_token_dynamic-role_FfH2i1c4dO_1611952635,ou=users,dc=learn,dc=example]
password               xWMjkIFMerYttEbzfnBVZvhRQGmhpAA0yeTya8fdmDB3LXDzGrjNEPV2bCPE9CW6
username               v_token_testrole_FfH2i1c4dO_1611952635
```

The `distinguished_names` field contains an array of DNs created from `creation_ldif`.
If multiple LDIF entries are used, this field includes DNs from each entry.
Each element in this field corresponds to one LDIF statement.
No deduplication is performed, and the original order is preserved.

### LDIF entries

User account management is performed with LDIF entries.
LDIF entries can be provided as a Base64-encoded version of an LDIF string.
Stronghold parses the string and validates it against LDIF syntax.
For details about LDIF syntax, see the [LDAP.com LDIF reference](https://ldap.com/ldif-the-ldap-data-interchange-format/).

When creating LDIF entries, keep the following in mind:

- Do not leave trailing spaces at the end of lines.
- Precede each `modify` block with an empty line.
- You can define multiple modifications for a `dn` in a single `modify` block.
  End each modification with a single hyphen `-`.

### Active Directory

Active Directory has several additional specifics.

To create a user in AD programmatically, first perform an `add` operation for the user object.
Then perform a `modify` operation for that user to set the password and enable the account.

Passwords in AD are set through the `unicodePwd` field.
Place two colons `::` before the value.

When setting a password in AD programmatically, all of the following requirements apply:

- The password must be enclosed in double quotation marks `""`;
- The password must use the [UTF16LE format](https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/6e803168-f140-4d23-b2d3-c3a8ab5917d2);
- The password must be Base64-encoded;
- For more details, see the [Microsoft documentation](https://docs.microsoft.com/en-us/troubleshoot/windows-server/identity/set-user-password-with-ldifde).

After you set the user password, you can enable the user.
In AD, this is done with the `userAccountControl` field:

- To enable the account, set `userAccountControl` to `512`;
- To disable password expiration for this dynamic user account, set `userAccountControl` to `65536`;
- The `userAccountControl` flags are cumulative, so set the value to `66048` to apply both flags (`512 + 65536 = 66048`);
- For more information about `userAccountControl` flags, see the [Microsoft documentation](https://docs.microsoft.com/en-us/troubleshoot/windows-server/identity/useraccountcontrol-manipulate-account-properties#property-flag-descriptions).

The `sAMAccountName` field is commonly used when working with AD users.
It exists for compatibility with legacy Microsoft Windows NT systems and is limited to 20 characters.
Keep this in mind when defining the `username_template`.
For more details, see the [Microsoft documentation](https://docs.microsoft.com/en-us/windows/win32/adschema/a-samaccountname).

Because the default `username_template` is longer than 20 characters and uses the `v_{{.DisplayName}}_{{.RoleName}}_{{random 10}}_{{unix_time}}` pattern, configure `username_template` in the role configuration so that generated account names stay under 20 characters.

AD does not allow direct modification of the user's `memberOf` attribute.
The group's `member` attribute and the user's `memberOf` attribute are [linked attributes](https://docs.microsoft.com/en-us/windows/win32/ad/linked-attributes).
These attributes form a forward-link/back-link pair, and only the forward link can be modified.
For AD group membership, the group's `member` attribute is the forward link.
To add a newly created dynamic user to a group, send a `modify` request to the target group and add the user there.

#### Active Directory LDIF example

The `*_ldif` parameters are templates that use the [Go template language](https://golang.org/pkg/text/template/).
The following example shows an LDIF template for creating an Active Directory user account:

```ldif
dn: CN={{.Username}},OU=Stronghold,DC=adtesting,DC=lab
changetype: add
objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: user
userPrincipalName: {{.Username}}@adtesting.lab
sAMAccountName: {{.Username}}

dn: CN={{.Username}},OU=Stronghold,DC=adtesting,DC=lab
changetype: modify
replace: unicodePwd
unicodePwd::{{ printf "%q" .Password | utf16le | base64 }}
-
replace: userAccountControl
userAccountControl: 66048
-

dn: CN=test-group,OU=Stronghold,DC=adtesting,DC=lab
changetype: modify
add: member
member: CN={{.Username}},OU=Stronghold,DC=adtesting,DC=lab
-
```

## Password rotation for account libraries

{: #rotation .anchored}

Stronghold can automatically rotate passwords for a group of accounts.
You can trigger the rotation manually, or Stronghold does it after the TTL from the previous rotation expires.

This functionality works with different schemas, including OpenLDAP, Active Directory, and RACF.
The following example uses Active Directory.

First, enable the LDAP secrets engine and configure how it connects to the AD server.

Example:

```console
$ d8 stronghold secrets enable ldap
Success! Enabled the ad secrets engine at: ldap/
$ d8 stronghold write ldap/config \
    binddn=$USERNAME \
    bindpass=$PASSWORD \
    url=ldaps://138.91.247.105 \
    userdn='dc=example,dc=com'
```

Then configure the account library for which password rotation is required:

```console
d8 stronghold write ldap/library/accounting-team \
  service_account_names=fizz@example.com,buzz@example.com \
  ttl=10h \
  max_ttl=20h \
  disable_check_in_enforcement=false
```

In this example, the `fizz@example.com` and `buzz@example.com` service accounts already exist on the remote AD server.
The `ttl` parameter defines how long Stronghold waits before rotating the account password again.
The `max_ttl` parameter defines the maximum time that the password can remain valid after rotation.
By default, both parameters are set to `24h`.

By default, the same Stronghold entity or client token that checks out the account must also check it back in.
If this behavior causes issues, set `disable_check_in_enforcement=true`.

After you create the account library, you can view its status at any time.

Example:

```console
d8 stronghold read ldap/library/accounting-team/status
```

Example output:

```console
Key                 Value
---                 -----
buzz@example.com    map[available:true]
fizz@example.com    map[available:true]
```

To rotate passwords, run:

```console
d8 stronghold write -f ldap/library/accounting-team/check-out
```

Example output:

```console
Key                     Value
---                     -----
lease_id                ldap/library/accounting-team/check-out/EpuS8cX7uEsDzOwW9kkKOyGW
lease_duration          10h
lease_renewable         true
password                ?@09AZKh03hBORZPJcTDgLfntlHqxLy29tcQjPVThzuwWAx/Twx4a2ZcRQRqrZ1w
service_account_name    fizz@example.com
```

If the default `ttl` value is longer than necessary, set a shorter value:

```console
d8 stronghold write ldap/library/accounting-team/check-out ttl=30m
```

Example output:

```console
Key                     Value
---                     -----
lease_id                ldap/library/accounting-team/check-out/gMonJ2jB6kYs6d3Vw37WFDCY
lease_duration          30m
lease_renewable         true
password                ?@09AZerLLuJfEMbRqP+3yfQYDSq6laP48TCJRBJaJu/kDKLsq9WxL9szVAvL/E1
service_account_name    buzz@example.com
```

You can renew the password lease for the account library:

```console
d8 stronghold lease renew ldap/library/accounting-team/check-out/0C2wmeaDmsToVFc0zDiX9cMq
```

Example output:

```console
Key                Value
---                -----
lease_id           ldap/library/accounting-team/check-out/0C2wmeaDmsToVFc0zDiX9cMq
lease_duration     10h
lease_renewable    true
```

In this case, the current account passwords remain valid longer because the next rotation is postponed.

## LDAP password policy

The LDAP secrets engine does not hash or encrypt passwords before updating values in LDAP.
As a result, LDAP may store passwords in plain text.

To avoid storing passwords in plain text, configure an LDAP password policy (`ppolicy`) on the LDAP server.
Do not confuse it with a Stronghold password policy.
A `ppolicy` can enforce password handling rules, such as hashing passwords by default.

The following example shows an LDAP password policy that enables hashing for `dc=example,dc=com`:

```console
dn: cn=module{0},cn=config
changetype: modify
add: olcModuleLoad
olcModuleLoad: ppolicy

dn: olcOverlay={2}ppolicy,olcDatabase={1}mdb,cn=config
changetype: add
objectClass: olcPPolicyConfig
objectClass: olcOverlayConfig
olcOverlay: {2}ppolicy
olcPPolicyDefault: cn=default,ou=pwpolicies,dc=example,dc=com
olcPPolicyForwardUpdates: FALSE
olcPPolicyHashCleartext: TRUE
olcPPolicyUseLockout: TRUE
```
