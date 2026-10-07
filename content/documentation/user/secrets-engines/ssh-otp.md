---
title: "SSH one-time passwords (OTP)"
linkTitle: "SSH OTP"
description: "The OTP mode of the SSH secrets engine: roles with key_type=otp, requesting a one-time password for a host, and verification by an agent on the remote host."
weight: 45
---

In addition to [SSH certificate signing](../signed-ssh-certificates/) (CA mode), the SSH secrets engine supports the one-time password (OTP) mode. In this mode, Stronghold issues a client a one-time password for logging in to a specific host, and an agent on the remote host verifies the password against Stronghold when the user logs in.

OTP mode properties:

- a password is valid for a single login and a single IP address;
- every remote host needs an agent that verifies OTPs via the `verify` endpoint of the SSH secrets engine;
- Stronghold does not store or distribute SSH keys.

{{< alert level="warning" >}}
OTP mode requires an agent on remote hosts that is compatible with the `ssh/verify` endpoint (for example, an SSH helper integrated with PAM). Installing and configuring the agent is out of scope of this documentation.
{{< /alert >}}

## Setup

1. Enable the SSH secrets engine if it is not enabled yet:

   ```bash
   d8 stronghold secrets enable ssh
   ```

1. Create a role with the `otp` key type:

   ```bash
   d8 stronghold write ssh/roles/otp-role \
     key_type=otp \
     default_user=ubuntu \
     cidr_list=10.0.0.0/24 \
     exclude_cidr_list=10.0.0.1/32 \
     allowed_users="ubuntu,deploy" \
     port=22
   ```

   Role parameters for OTP mode:

   | Parameter | Description |
   |-----------|-------------|
   | `key_type` | Key type: `otp` or `ca`. Use `otp` for this mode. |
   | `default_user` | Default username on the remote host. Required for the `otp` type. |
   | `cidr_list` | Comma-separated list of CIDR blocks the role applies to. |
   | `exclude_cidr_list` | Comma-separated list of CIDR blocks whose IP addresses the role does not accept. |
   | `allowed_users` | Users an OTP can be requested for. `*` allows any user. |
   | `port` | SSH port returned to the client together with the OTP. It does not affect OTP generation. Defaults to `22`. |

1. If the role must issue OTPs for any IP address, add it to the zero-address role list. CIDR blocks previously set for these roles are ignored:

   ```bash
   d8 stronghold write ssh/config/zeroaddress roles="otp-role"
   ```

1. Configure an agent on the remote hosts that verifies OTPs against Stronghold.

## Usage

1. Request a one-time password for a host:

   ```bash
   d8 stronghold write ssh/creds/otp-role ip=10.0.0.15 username=ubuntu
   ```

   If `username` is omitted, the role's `default_user` is used. The response contains the one-time password, the username, the IP address, and the port.

1. Connect to the host over SSH and enter the OTP as the password. The agent on the host verifies it via the `ssh/verify` endpoint, after which the password becomes invalid.

To find out which roles apply to an IP address, use the `lookup` endpoint:

```bash
d8 stronghold write ssh/lookup ip=10.0.0.15
```

## Policy example

A policy that only allows requesting OTPs for the `otp-role` role:

```hcl
path "ssh/creds/otp-role" {
  capabilities = ["create", "update"]
}
```

For the full list of endpoints and parameters, see the [secrets engines API reference](../../../reference/api/secrets/).
