---
title: "SSH access with signed certificates"
linkTitle: "SSH certificates"
description: "Setting up SSH access to servers with short-lived certificates that users obtain from Stronghold after an OIDC login."
weight: 40
params:
  relatedLinks:
    - title: "Signed SSH certificates"
      url: ../../../user/secrets-engines/signed-ssh-certificates/
    - title: "OIDC auth method"
      url: ../../../user/auth/oidc/overview/
    - title: "Templated policies"
      url: ../../../concepts/policy/
    - title: "Identity"
      url: ../../../concepts/identity/
---

Instead of distributing public keys to `authorized_keys` on every server, servers trust a single CA in Stronghold. A user logs in to Stronghold through OIDC, signs their public key, and gets a certificate valid for 30 minutes. You do not need to revoke access on servers: the certificate expires on its own.

## Goal

Set up an SSH CA in Stronghold, a role that lets users log in only under their own name, CA trust on servers, and certificate issuance after an OIDC login.

## Prerequisites

- A Stronghold token with permissions to configure secrets engines, policies, and Identity.
- A configured OIDC auth method that passes user groups (in DP, the `oidc_deckhouse` method).
- Servers with OpenSSH where you can change `sshd_config`.
- User names in the IdP that match account names on the servers (or another mapping rule, see step 3).

## Step 1. Set up the SSH CA

1. Enable the SSH secrets engine:

   ```bash
   d8 stronghold secrets enable -path=ssh-client-signer ssh
   ```

1. Generate the CA key:

   ```bash
   d8 stronghold write ssh-client-signer/config/ca generate_signing_key=true
   ```

## Step 2. Configure trust on servers

1. Save the CA public key on each server. The `public_key` endpoint does not require authentication:

   ```bash
   curl -o /etc/ssh/trusted-user-ca-keys.pem \
     https://stronghold.example.com/v1/ssh-client-signer/public_key
   ```

1. Add the following to `/etc/ssh/sshd_config`:

   ```text
   TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem
   ```

1. Restart `sshd`:

   ```bash
   sudo systemctl restart sshd
   ```

Distribute the key and the setting with a configuration management system (Ansible, Puppet, Salt).

## Step 3. Create a role with a templated user name

The role allows signing a key only for the principal that matches the user name in OIDC. This uses an Identity template with the OIDC method accessor.

1. Get the auth method accessor:

   ```bash
   OIDC_ACCESSOR=$(d8 stronghold read -field=accessor sys/auth/oidc_deckhouse)
   ```

1. Create the role:

   ```bash
   d8 stronghold write ssh-client-signer/roles/ops - <<EOF
   {
     "key_type": "ca",
     "algorithm_signer": "rsa-sha2-256",
     "allow_user_certificates": true,
     "allowed_users_template": true,
     "allowed_users": "{{identity.entity.aliases.${OIDC_ACCESSOR}.name}}",
     "default_user_template": true,
     "default_user": "{{identity.entity.aliases.${OIDC_ACCESSOR}.name}}",
     "allowed_extensions": "permit-pty,permit-port-forwarding",
     "default_extensions": {
       "permit-pty": ""
     },
     "ttl": "30m",
     "max_ttl": "1h"
   }
   EOF
   ```

   If the IdP name differs from the server account name (for example, it includes a domain), store the required name in entity metadata and use the `{{identity.entity.metadata.<key>}}` template.

## Step 4. Grant signing permissions

1. Create a policy:

   ```bash
   d8 stronghold policy write ssh-ops - <<'POLICY'
   path "ssh-client-signer/sign/ops" {
     capabilities = ["update"]
   }
   POLICY
   ```

1. Bind the policy to an IdP group through an external Identity group:

   ```bash
   GROUP_ID=$(d8 stronghold write -field=id identity/group \
     name=ops type=external policies=ssh-ops)

   d8 stronghold write identity/group-alias \
     name=ops \
     mount_accessor="$OIDC_ACCESSOR" \
     canonical_id="$GROUP_ID"
   ```

   Here, `name=ops` in `group-alias` is the group name that the IdP passes in the groups claim.

## Step 5. Get a certificate and log in to a server

The user performs these actions on their workstation.

1. Log in to Stronghold:

   ```bash
   d8 stronghold login -method=oidc -path=oidc_deckhouse
   ```

1. If you do not have an SSH key, create one:

   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519
   ```

1. Sign the public key and save the certificate next to the key:

   ```bash
   d8 stronghold write -field=signed_key ssh-client-signer/sign/ops \
     public_key=@$HOME/.ssh/id_ed25519.pub > ~/.ssh/id_ed25519-cert.pub
   ```

1. Connect to the server. OpenSSH uses the `*-cert.pub` file automatically:

   ```bash
   ssh <username>@server.example.com
   ```

Repeat steps 1 and 3 after the certificate expires.

## Verification

1. View the certificate contents:

   ```bash
   ssh-keygen -Lf ~/.ssh/id_ed25519-cert.pub
   ```

   The `Principals` field must contain your user name, and the `Valid` field must show an interval of no more than 30 minutes.

1. Try to sign a key for another principal; the request must be rejected:

   ```bash
   d8 stronghold write ssh-client-signer/sign/ops \
     public_key=@$HOME/.ssh/id_ed25519.pub valid_principals=root
   ```

1. After 30 minutes, make sure the server rejects login with the same certificate.

For login errors, see [Troubleshooting](../../../user/secrets-engines/signed-ssh-certificates/#troubleshooting).

## Cleanup

1. Delete the external group and its alias: `d8 stronghold delete identity/group/name/ops`.
1. Delete the policy and the role:

   ```bash
   d8 stronghold policy delete ssh-ops
   d8 stronghold delete ssh-client-signer/roles/ops
   ```

1. For a test setup, disable the mount: `d8 stronghold secrets disable ssh-client-signer`. Then remove the `TrustedUserCAKeys` line on the servers.
