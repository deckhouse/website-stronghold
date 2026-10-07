---
title: "MFA for administrators"
linkTitle: "MFA for administrators"
description: "Mandatory TOTP second factor for administrator logins to Stronghold: a TOTP method, a login enforcement rule for the administrators group, QR code enrollment, login in the CLI and web UI, recovery after a lost device, and audit log monitoring."
weight: 40
params:
  relatedLinks:
    - title: "TOTP"
      url: ../../../user/auth/mfa/totp/
    - title: "Multifactor"
      url: ../../../user/auth/mfa/multifactor/
    - title: "Identity"
      url: ../../../concepts/identity/
    - title: "Identity API"
      url: ../../../reference/api/identity/
    - title: "Audit"
      url: ../../../admin/audit/overview/
    - title: "Break-glass access"
      url: ../break-glass-access/
---

A Stronghold administrator account grants access to policies, auth methods, and secrets, so a password or an IdP session alone is not enough to protect it. This guide enables a mandatory TOTP one-time code check at login for all administrators.

## MFA mechanisms in Stronghold

Stronghold checks the second factor at login (login MFA): after the primary auth method succeeds, a token is issued only after the second factor is confirmed. MFA methods and the rules for applying them are configured in the Identity subsystem at the `identity/mfa/method/*` and `identity/mfa/login-enforcement/*` paths.

The documentation describes two MFA methods:

| Method | How the login is confirmed | When to choose |
| --- | --- | --- |
| [TOTP](../../../user/auth/mfa/totp/) | One-time code from an authenticator app | No external service is needed; suitable for air-gapped environments |
| [Multifactor](../../../user/auth/mfa/multifactor/) | Push notification, Telegram, or a phone call through the Multifactor service | Your organization already uses Multifactor |

A login enforcement rule can apply to entities (`identity_entity_ids`), Identity groups (`identity_group_ids`), specific auth method mounts (`auth_method_accessors`), or all mounts of a given auth method type (`auth_method_types`).

MFA checks on access to individual paths (step-up MFA through `sys/mfa/method/*`) are not described for Stronghold: the API reference marks these paths as Enterprise feature stubs. <!-- TODO(verify): whether step-up MFA (sys/mfa/method/*) is supported in Stronghold EE -->

{{< alert level="info" >}}
Audit logging, which is used in the [Audit log monitoring](#audit-log-monitoring) section, is available only in Stronghold EE.
{{< /alert >}}

<!-- TODO(verify): MFA (TOTP and Multifactor) availability by edition — the editions page does not mention MFA -->

## Goal

- All Stronghold administrators enter a TOTP code at login.
- The check applies to the `stronghold-admins` Identity group, so new administrators get it automatically.
- A recovery procedure is documented for an administrator who has lost the device with the authenticator.
- Operations on MFA settings are visible in the audit log.

![Admin login flow with TOTP](../../../images/ex-admin-mfa.en.png)

## Prerequisites

- A Stronghold token with permissions to configure Identity and policies.
- The `stronghold-admins` Identity group that contains the administrator entities. For directories and IdPs, use an external group with a directory group alias, as in the [Active Directory integration](../active-directory/) guide.
- An authenticator app with TOTP support for each administrator.
- Prepared [break-glass access](../break-glass-access/) that the MFA rule does not apply to.
- The `jq` utility on the administrator workstation.
- For audit log monitoring, Stronghold EE and an enabled audit device.

## Step 1. Create a TOTP method

Create a TOTP MFA method and save its ID:

```bash
TOTP_METHOD_ID=$(d8 stronghold write -format=json identity/mfa/method/totp \
  method_name=admin-totp \
  generate=true \
  issuer=Stronghold \
  period=30 \
  algorithm=SHA256 \
  digits=6 \
  max_validation_attempts=5 | jq -r '.data.method_id')
echo $TOTP_METHOD_ID
```

Main method parameters:

| Parameter | Description |
| --- | --- |
| `method_name` | Unique MFA method name |
| `issuer` | Name that the authenticator app shows next to the code |
| `period` | Code rotation period in seconds. Default: `30` |
| `algorithm` | Hash algorithm: `SHA1` (default), `SHA256`, or `SHA512` |
| `digits` | Number of digits in the code: `6` or `8` |
| `skew` | Allowed deviation in periods when validating a code: `0` or `1`. Default: `1` |
| `max_validation_attempts` | Maximum number of code validation attempts |

The full list of parameters is available in the [Identity API reference](../../../reference/api/identity/).

Not all authenticator apps support `SHA256` and 8-digit codes. If administrators use different apps, set `algorithm=SHA1` and `digits=6`. <!-- TODO(verify): recommended TOTP parameters for compatibility with common authenticator apps -->

## Step 2. Enroll administrator authenticators

Enroll TOTP for all administrators **before** enabling the mandatory check. If an entity has no secret, the second-factor check fails with the error `MFA secret … not present in entity`, and the login does not succeed.

An administrator entity is created at the first login to Stronghold. An administrator can find the ID of their entity as follows:

```bash
d8 stronghold token lookup -format=json | jq -r '.data.entity_id'
```

Choose one of the enrollment methods.

### QR code issued by a security administrator

1. Get the entity ID by name:

   ```bash
   ENTITY_ID=$(d8 stronghold read -field=id identity/entity/name/<entity_name>)
   ```

1. Generate a TOTP secret and a QR code:

   ```bash
   d8 stronghold write -field=barcode \
     identity/mfa/method/totp/admin-generate \
     method_id=$TOTP_METHOD_ID entity_id=$ENTITY_ID \
     | base64 -d > /tmp/qr-code.png
   ```

1. Pass the QR code to the administrator over a secure channel, wait until they scan it in the authenticator app, and delete the file:

   ```bash
   shred -u /tmp/qr-code.png
   ```

Calling `admin-generate` again for an entity that already has a secret for this method does not create a new one: Stronghold returns the warning `Entity already has a secret for MFA method`. To re-enroll TOTP, first delete the secret with `admin-destroy` (see the "Recovery after a lost device" section below).

### Self-enrollment

1. Add a permission to generate one's own secret to the administrator policy:

   ```hcl
   path "identity/mfa/method/totp/generate" {
     capabilities = ["update"]
   }
   ```

1. The administrator gets a QR code for their entity in the Stronghold web UI or through the CLI:

   ```bash
   d8 stronghold write -field=barcode identity/mfa/method/totp/generate \
     method_id=$TOTP_METHOD_ID | base64 -d > /tmp/qr-code.png
   ```

If the entity already has a secret for the method, `generate` does not overwrite it and returns a warning, so a stolen session cannot be used to re-enroll TOTP; resetting requires `admin-destroy`.

## Step 3. Enable the mandatory check

1. Get the administrators group ID:

   ```bash
   ADMINS_GROUP_ID=$(d8 stronghold read -field=id identity/group/name/stronghold-admins)
   ```

1. Create a login enforcement rule:

   ```bash
   d8 stronghold write identity/mfa/login-enforcement/admin-totp \
     mfa_method_ids="$TOTP_METHOD_ID" \
     identity_group_ids="$ADMINS_GROUP_ID"
   ```

If administrators log in through a separate auth method mount, for example `userpass` at the `admin/` path, you can bind the rule to that mount:

```bash
ADMIN_ACCESSOR=$(d8 stronghold auth list -format=json -detailed \
  | jq -r '."admin/".accessor')

d8 stronghold write identity/mfa/login-enforcement/admin-totp \
  mfa_method_ids="$TOTP_METHOD_ID" \
  auth_method_accessors="$ADMIN_ACCESSOR"
```

{{< alert level="warning" >}}
Do not include the break-glass account and its auth method mount (for example, `breakglass/`) in the rule. Otherwise, if the authenticators are lost, nobody will be able to log in to Stronghold.
{{< /alert >}}

## Logging in with MFA

### CLI

After the primary factor is checked, the CLI asks for a TOTP code:

```bash
d8 stronghold login -method=oidc
```

Example output:

```console
Initiating Interactive MFA Validation...
Enter the passphrase for methodID "22c35aa4-bf37-cf31-4187-c5a676c19aca" of type "totp":
```

Enter the current code from the authenticator app. After a successful check, the CLI saves the token.

When logging in through the API, the auth method response contains `mfa_request_id`. Pass it together with the code to `sys/mfa/validate`:

```bash
d8 stronghold write -format=json sys/mfa/validate - <<EOF
{
  "mfa_request_id": "<mfa_request_id>",
  "mfa_payload": {
    "$TOTP_METHOD_ID": ["<totp_code>"]
  }
}
EOF
```

The identifier is in the `auth.mfa_requirement.mfa_request_id` field. To make the CLI skip the interactive prompt and only print `mfa_request_id`, add the `-non-interactive` flag to `d8 stronghold login`.

### Web UI

1. Open the Stronghold web UI and log in with the chosen method.
1. After the primary factor is checked, enter the TOTP code from the authenticator app.

## Recovery after a lost device

If an administrator has lost the device with the authenticator or it may have been compromised:

1. Confirm the administrator's identity through an independent channel, for example in person or through their manager. Record the request in the incident log.
1. Log in with a security administrator account that has permissions for `admin-destroy` and `admin-generate`. Example policy:

   ```hcl
   path "identity/entity/name/*" {
     capabilities = ["read"]
   }

   path "identity/mfa/method/totp/admin-destroy" {
     capabilities = ["update"]
   }

   path "identity/mfa/method/totp/admin-generate" {
     capabilities = ["update"]
   }
   ```

1. Delete the TOTP secret of the administrator entity:

   ```bash
   ENTITY_ID=$(d8 stronghold read -field=id identity/entity/name/<entity_name>)

   d8 stronghold write identity/mfa/method/totp/admin-destroy \
     method_id=$TOTP_METHOD_ID entity_id=$ENTITY_ID
   ```

1. Issue a new QR code as in step 2 and pass it to the administrator over a secure channel.
1. If the device may have fallen into the wrong hands together with the password, change the password or the credentials of the primary login method.

If you need to restore access for the only administrator and there are no other accounts with `admin-destroy` permissions, use [break-glass access](../break-glass-access/).

## Audit log monitoring

{{< alert level="info" >}}
Audit logging is available only in Stronghold EE.
{{< /alert >}}

1. Make sure that an audit device is enabled:

   ```bash
   d8 stronghold audit list
   ```

1. Configure SIEM alerts for changes to MFA settings. Indicators in the audit log:

   - `request.path` equals `identity/mfa/method/totp/admin-destroy` or `identity/mfa/method/totp/admin-generate`;
   - `request.path` starts with `identity/mfa/login-enforcement/` or `identity/mfa/method/`, and `request.operation` equals `update` or `delete`.

   Example search in a file log:

   ```bash
   jq -c 'select(.type == "request" and (.request.path | startswith("identity/mfa/")))
     | {time, path: .request.path, op: .request.operation, entity: .auth.entity_id}' \
     /var/log/stronghold_audit.log
   ```

1. Monitor second factor checks — requests to `sys/mfa/validate`. A large number of failed checks for one entity may indicate an attempt to brute-force the code. <!-- TODO(verify): whether MFA checks during interactive login in the CLI and web UI are recorded in the audit log -->

## Verification

1. Make sure that the rule exists:

   ```bash
   d8 stronghold read identity/mfa/login-enforcement/admin-totp
   ```

1. Log in with an administrator account and make sure that Stronghold asks for a TOTP code.
1. Enter a wrong code and make sure that no token is issued.
1. Log in with the break-glass account and make sure that no TOTP code is requested.

## Cleanup

1. Delete the login enforcement rule:

   ```bash
   d8 stronghold delete identity/mfa/login-enforcement/admin-totp
   ```

1. Delete the TOTP method:

   ```bash
   d8 stronghold delete identity/mfa/method/totp/$TOTP_METHOD_ID
   ```
