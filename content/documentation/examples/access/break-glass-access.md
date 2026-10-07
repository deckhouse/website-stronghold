---
title: "Break-glass access"
linkTitle: "Break-glass access"
description: "Preparing emergency access to Stronghold for when the identity provider is unavailable: a backup account in a sealed envelope, the generate-root procedure, usage auditing, and rotation after use."
weight: 50
params:
  relatedLinks:
    - title: "Userpass auth method"
      url: ../../../user/auth/userpass/
    - title: "Tokens"
      url: ../../../concepts/tokens/
    - title: "Seal and unseal"
      url: ../../../concepts/seal/
    - title: "Audit"
      url: ../../../admin/audit/overview/
    - title: "Policies"
      url: ../../../concepts/policy/
---

People usually log in to Stronghold through OIDC. If the identity provider (in DP, Dex and an external IdP) is unavailable or misconfigured, no administrator can log in. Break-glass access is a pre-arranged, rarely used, and tightly controlled way to log in in such a situation.

## Goal

Prepare two levels of emergency access:

- **Level 1**: a separate `userpass` account with an administrative policy whose password is kept in a sealed envelope.
- **Level 2**: the `generate-root` procedure for when level 1 does not help.

Set up auditing of their use and a rotation procedure after use.

## Prerequisites

- A Stronghold token with administrative permissions.
- Storage for envelopes (a safe) and at least two people responsible for access to it.
- Information about the holders of unseal or recovery key shares and their quorum threshold.
- For auditing: Stronghold EE and a configured audit device (see [Audit](../../../admin/audit/overview/)).

## Step 1. Create a break-glass policy

The policy must allow restoring login, but does not have to grant full access to secrets. An example policy for restoring auth methods and policies:

```bash
d8 stronghold policy write break-glass - <<'POLICY'
path "sys/auth" {
  capabilities = ["read"]
}
path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "sudo"]
}
path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "sys/policies/acl/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "identity/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "sys/mounts" {
  capabilities = ["read"]
}
POLICY
```

Extend the policy only with what is actually needed in an emergency.

## Step 2. Create a break-glass account

1. Enable a separate instance of the `userpass` method so that its use is easy to distinguish in audit logs:

   ```bash
   d8 stronghold auth enable -path=breakglass userpass
   ```

1. Generate a long random password on an isolated workstation:

   ```bash
   openssl rand -base64 32
   ```

1. Create a user with a short token lifetime:

   ```bash
   d8 stronghold write auth/breakglass/users/breakglass-1 \
     password="<generated_password>" \
     token_policies=break-glass \
     token_ttl=1h \
     token_max_ttl=4h
   ```

1. Write the password on paper, put it in an envelope, seal and sign it, and place it in the safe. For a two-person rule, split the password into two parts and keep them in separate envelopes with different people.

1. Remove the password from the shell history and the clipboard.

Lockout after failed login attempts is enabled for `userpass` by default. An attacker can lock the break-glass account on purpose; take this into account in the procedure and, if needed, create a second `breakglass-2` account in a separate envelope.

{{< alert level="warning" >}}
Do not use long-lived tokens for break-glass access. Token lifetime is limited by the system maximum TTL, and such a token can expire unnoticed.
{{< /alert >}}

## Step 3. Prepare the generate-root procedure

If the break-glass account is unavailable, a root token can be generated only with a quorum of unseal key share holders (with auto-unseal, recovery key share holders).

1. The initiator starts the procedure. The response contains `nonce` and `otp`:

   ```bash
   d8 stronghold operator generate-root -init
   ```

1. Each share holder submits their share, specifying the `nonce`:

   ```bash
   d8 stronghold operator generate-root -nonce=<nonce>
   ```

   The command prompts for the key share interactively. When the quorum is reached, `encoded_token` is printed.

1. The initiator decodes the token:

   ```bash
   d8 stronghold operator generate-root -decode=<encoded_token> -otp=<otp>
   ```

If the procedure was started by mistake, cancel it with `d8 stronghold operator generate-root -cancel`. Describe the procedure in your runbook, including who holds the shares and how to gather them.

{{< alert level="warning" >}}
In DP in `Automatic` mode, the unseal key and the root token are stored in the `stronghold-keys` secret in the `d8-stronghold` namespace. Access to this secret equals full access to Stronghold: restrict it with RBAC and monitor it in Kubernetes audit logs. For details, see [Configuration](../../../install/dkp/configuration/).
{{< /alert >}}

<!-- TODO(verify): the generate-root procedure and revocation of the initial root token for Stronghold in DKP in Automatic mode -->

## Step 4. Set up usage monitoring

{{< alert level="info" >}}
Audit logging is available only in Stronghold EE.
{{< /alert >}}

1. Make sure an audit device is enabled:

   ```bash
   d8 stronghold audit list
   ```

1. Configure an alert in your SIEM for any login through the break-glass path. Indicators in the audit log:

   - `request.path` equals `auth/breakglass/login/breakglass-1`;
   - `request.path` starts with `sys/generate-root`;
   - `auth.policies` contains `root`.

   An example search in a file audit log:

   ```bash
   jq -c 'select(.request.path | startswith("auth/breakglass/login") or startswith("sys/generate-root"))' \
     /var/log/stronghold_audit.log
   ```

## Use in an emergency

1. Two responsible people open the envelope and record the time and reason in the incident log.
1. Log in:

   ```bash
   d8 stronghold login -method=userpass -path=breakglass username=breakglass-1
   ```

1. Restore the login configuration (for example, `auth/oidc_deckhouse/config`) and check that regular login works.
1. Revoke the break-glass token: `d8 stronghold token revoke -self`.

## Rotation after use

After every use, including drills:

1. Change the password and create a new envelope:

   ```bash
   d8 stronghold write auth/breakglass/users/breakglass-1/password password="<new_password>"
   ```

1. Revoke all tokens issued through the break-glass path:

   ```bash
   d8 stronghold lease revoke -prefix auth/breakglass/
   ```

1. If `generate-root` was used, revoke the root token as soon as the work is done:

   ```bash
   d8 stronghold token revoke <root_token>
   ```

1. Review the audit log for the period of use and attach an export to the incident report.

## Verification

Run a drill at least every six months:

1. Log in with the break-glass account on a test or production cluster.
1. Check the permissions: `d8 stronghold token capabilities sys/auth/oidc_deckhouse`.
1. Make sure the SIEM received an alert.
1. Perform the rotation after use.

## Cleanup

If break-glass access is no longer needed (for example, when decommissioning the cluster):

```bash
d8 stronghold auth disable breakglass
d8 stronghold policy delete break-glass
```

Destroy the password envelopes with a formal record.
