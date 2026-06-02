---
title: "SSH secrets engine"
description: "Information about SSH secrets engine in Deckhouse Stronghold."
weight: 40
---

The SSH secrets engine lets you use SSH certificates to access servers.
This is one of the simplest and most convenient ways to organize SSH access with minimal platform dependency.

Stronghold can act as a certificate authority (CA) for SSH.
Together with built-in OpenSSH capabilities, this allows clients to connect to remote hosts over SSH by using their own local SSH keys.

On this page, the term “client” refers to a user or machine that initiates an SSH connection.
The term “host” refers to a remote machine.
This page provides a quick-start guide for configuring and using the SSH secrets engine.

## Sign client keys

First, configure the SSH secrets engine in Stronghold.
After that, clients can sign their SSH keys.
These tasks are usually performed by a Stronghold administrator, a security team, or a configuration management system.

### Create a signing key and configure a role

The following steps are performed by a Stronghold administrator, a security team, or a configuration management system.

1. Mount the SSH secrets engine.
   Without this step, the secrets engine does not work:

   ```shell
   stronghold secrets enable -path=ssh-client-signer ssh
   ```

   Example output:

   ```console
   Successfully mounted 'ssh' at 'ssh-client-signer'!
   ```

   This command enables the SSH secrets engine at the `ssh-client-signer` path.
   You can mount the same secrets engine multiple times by using different `-path` values.
   The name `ssh-client-signer` is not special.
   It is used in the examples on this page.

1. Configure Stronghold with a CA for signing client keys by using the `/config/ca` endpoint.
   If you do not have your own CA, Stronghold can generate the public and private keys for you:

   ```shell
   stronghold write ssh-client-signer/config/ca generate_signing_key=true
   ```

   Example output:

   ```console
   Key             Value
   ---             -----
   public_key      ssh-rsa AAAAB3NzaC1yc2EA...
   ```

   If you already have an SSH key pair, pass the public and private keys in the request:

   ```shell
   stronghold write ssh-client-signer/config/ca \
     private_key="..." \
     public_key="..."
   ```

   The SSH secrets engine supports configuring multiple CA certificates in a single mount.
   This is useful for CA rotation.
   When you configure a CA, one issuer is assigned as the default issuer.
   It is used for all operations unless another issuer is specified when creating a role.
   You can change the default issuer at any time by creating a new CA or updating the existing one through a configuration endpoint.

   Whether the key is generated or uploaded, the public key is available through the `/public_key` API endpoint and through the CLI.

1. Add the public key to all SSH host configurations.
   You can do this manually or automate it through a configuration management system.
   The public key is available through the API and does not require authentication:

   ```shell
   curl -o /etc/ssh/trusted-user-ca-keys.pem http://127.0.0.1:8200/v1/ssh-client-signer/public_key
   ```

   Or by using the CLI:

   ```shell
   stronghold read -field=public_key ssh-client-signer/config/ca > /etc/ssh/trusted-user-ca-keys.pem
   ```

   Add the path to the file that stores the public key to the SSH configuration as the `TrustedUserCAKeys` value:

   ```text
   # /etc/ssh/sshd_config
   TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem
   ```

   Restart the SSH service to apply the changes.

1. Create a role in Stronghold for signing client keys.
   Because of SSH certificate implementation details, some options are passed as key-value pairs.
   The following example adds the `permit-pty` extension to the certificate and allows the user to specify custom values for `permit-pty` and `permit-port-forwarding` when requesting a certificate:

   ```shell
   stronghold write ssh-client-signer/roles/my-role -<<"EOH"
   {
     "algorithm_signer": "rsa-sha2-256",
     "allow_user_certificates": true,
     "allowed_users": "*",
     "allowed_extensions": "permit-pty,permit-port-forwarding",
     "default_extensions": {
       "permit-pty": ""
     },
     "key_type": "ca",
     "default_user": "ubuntu",
     "ttl": "30m0s"
   }
   EOH
   ```

### Authenticate a client over SSH

The following steps are performed by the client that wants to authenticate to machines configured to work with Stronghold.
These commands are usually run on the client’s local workstation.

1. Find or generate a public SSH key.
   It is usually located at `~/.ssh/id_rsa.pub`.
   If you do not already have an SSH key pair, create one:

   ```shell
   ssh-keygen -t rsa -C "user@example.com"
   ```

1. Ask Stronghold to sign your public key.
   This is usually a `.pub` file, and its contents start with `ssh-rsa ...`:

   ```shell
   stronghold write ssh-client-signer/sign/my-role \
     public_key=@$HOME/.ssh/id_rsa.pub
   ```

   Example output:

   ```console
   Key             Value
   ---             -----
   serial_number   c73f26d2340276aa
   signed_key      ssh-rsa-cert-v01@openssh.com AAAAHHNzaC1...
   ```

   The response includes the serial number, which is the unique certificate ID, and the signed key.
   The signed key is also a public key.

   To further customize signing parameters, use a JSON request:

   ```shell
   stronghold write ssh-client-signer/sign/my-role -<<"EOH"
   {
     "public_key": "ssh-rsa AAA...",
     "valid_principals": "my-user",
     "key_id": "custom-prefix",
     "extensions": {
       "permit-pty": "",
       "permit-port-forwarding": ""
     }
   }
   EOH
   ```

1. Save the signed public key to disk.
   If needed, restrict file permissions:

   ```shell
   stronghold write -field=signed_key ssh-client-signer/sign/my-role \
     public_key=@$HOME/.ssh/id_rsa.pub > signed-cert.pub
   ```

   If you store the certificate next to the SSH key pair, use the `-cert.pub` suffix, for example `~/.ssh/id_rsa-cert.pub`.
   In this case, OpenSSH uses the certificate automatically during authentication.

1. If needed, inspect the extensions, users, hosts, and metadata of the signed key:

   ```shell
   ssh-keygen -Lf ~/.ssh/signed-cert.pub
   ```

1. Connect over SSH by using the signed public key and the corresponding private key:

   ```shell
   ssh -i signed-cert.pub -i ~/.ssh/id_rsa username@10.0.23.5
   ```

## Sign host keys

For additional security, it is recommended to enable host key signing.
This capability complements client key signing and increases the integrity of SSH connections.

If host key signing is configured, the SSH client can verify that the remote host is trusted before the connection is established.
This reduces the risk of accidentally connecting to a malicious machine.

### Configure host key signing

1. Mount the SSH secrets engine at a different path from the client signing path:

   ```shell
   stronghold secrets enable -path=ssh-host-signer ssh
   ```

   Example output:

   ```console
   Successfully mounted 'ssh' at 'ssh-host-signer'!
   ```

1. Configure Stronghold with a CA for signing host keys by using the `/config/ca` endpoint.
   If you do not have your own CA, Stronghold can generate a key pair:

   ```shell
   stronghold write ssh-host-signer/config/ca generate_signing_key=true
   ```

   Example output:

   ```console
   Key             Value
   ---             -----
   public_key      ssh-rsa AAAAB3NzaC1yc2EA...
   ```

   If you already have an SSH key pair, pass it in the request:

   ```shell
   stronghold write ssh-host-signer/config/ca \
     private_key="..." \
     public_key="..."
   ```

   The public key of the CA used for host key signing is available through the `/public_key` API endpoint.

1. Increase the TTL of the host key certificate:

   ```shell
   stronghold secrets tune -max-lease-ttl=87600h ssh-host-signer
   ```

1. Create a role for signing host keys.
   Be sure to set the list of allowed domains, `allow_bare_domains`, or both:

   ```shell
   stronghold write ssh-host-signer/roles/hostrole \
     key_type=ca \
     algorithm_signer=rsa-sha2-256 \
     ttl=87600h \
     allow_host_certificates=true \
     allowed_domains="localdomain,example.com" \
     allow_subdomains=true
   ```

1. Sign the host’s public SSH key:

   ```shell
   stronghold write ssh-host-signer/sign/hostrole \
     cert_type=host \
     public_key=@/etc/ssh/ssh_host_rsa_key.pub
   ```

   Example output:

   ```console
   Key             Value
   ---             -----
   serial_number   3746eb17371540d9
   signed_key      ssh-rsa-cert-v01@openssh.com AAAAHHNzaC1y...
   ```

1. Save the signed certificate and configure it as the `HostCertificate` value in the SSH configuration on the host:

   ```shell
   stronghold write -field=signed_key ssh-host-signer/sign/hostrole \
     cert_type=host \
     public_key=@/etc/ssh/ssh_host_rsa_key.pub > /etc/ssh/ssh_host_rsa_key-cert.pub
   ```

   Set the file permissions to `0640`:

   ```shell
   chmod 0640 /etc/ssh/ssh_host_rsa_key-cert.pub
   ```

   Add the host key and host certificate to the SSH configuration file:

   ```text
   # /etc/ssh/sshd_config
   TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem
   HostKey /etc/ssh/ssh_host_rsa_key
   HostCertificate /etc/ssh/ssh_host_rsa_key-cert.pub
   ```

   Restart the SSH service to apply the changes.

### Verify the host on the client side

1. Get the host CA public key for signature verification:

   ```shell
   curl http://127.0.0.1:8200/v1/ssh-host-signer/public_key
   ```

   Or by using the CLI:

   ```shell
   stronghold read -field=public_key ssh-host-signer/config/ca
   ```

1. Add the returned public key to the `known_hosts` file:

   ```text
   # ~/.ssh/known_hosts
   @cert-authority *.example.com ssh-rsa AAAAB3NzaC1yc2EAAA...
   ```

1. After that, you can connect to remote machines over SSH.

## Troubleshooting

To simplify configuration and debugging of the key signing process, enable the `VERBOSE` log level in the SSH configuration:

```text
# /etc/ssh/sshd_config
LogLevel VERBOSE
```

After updating the configuration, restart the SSH service.

By default, SSH writes logs to `/var/log/auth.log`.
Because this file can contain entries from other services, use the following command to filter only SSH logs:

```shell
tail -f /var/log/auth.log | grep --line-buffered "sshd"
```

If you cannot connect to the host, SSH server logs can help identify the cause.

### The username is not listed in valid principals

If the following messages appear in `/var/log/auth.log`:

```text
key_cert_check_authority: invalid certificate
Certificate invalid: name is not a listed principal
```

This means the certificate does not allow the specified username to be used as a valid principal for system authentication.
The cause is most likely related to an OpenSSH behavior.
For details, see [Known issues](#known-issues).

This error does not take `allowed_users="*"` into account.
Use one of the following workarounds:

- Set `default_user` if you always authenticate as the same user:

  ```shell
  stronghold write ssh/roles/my-role -<<"EOH"
  {
    "default_user": "YOUR_USER"
  }
  EOH
  ```

- Set `valid_principals` when signing the key if different users authenticate over SSH through Stronghold:

  ```shell
  stronghold write ssh-client-signer/sign/my-role -<<"EOH"
  {
    "valid_principals": "my-user"
  }
  EOH
  ```

### No shell prompt after login

If no shell prompt appears after authentication on the host, the signed certificate may be missing the `permit-pty` extension.

You can add this extension in two ways:

- When creating the role:

  ```shell
  stronghold write ssh-client-signer/roles/my-role -<<"EOH"
  {
    "default_extensions": {
      "permit-pty": ""
    }
  }
  EOH
  ```

- When signing the key:

  ```shell
  stronghold write ssh-client-signer/sign/my-role -<<"EOH"
  {
    "extensions": {
      "permit-pty": ""
    }
  }
  EOH
  ```

### Port forwarding does not work

If port forwarding does not work, the certificate may be missing the `permit-port-forwarding` extension.

Add it when creating the role or when signing the key:

```json
{
  "default_extensions": {
    "permit-port-forwarding": ""
  }
}
```

### X11 forwarding does not work

If X11 forwarding does not work, the certificate may be missing the `permit-X11-forwarding` extension.

Add it when creating the role or when signing the key:

```json
{
  "default_extensions": {
    "permit-X11-forwarding": ""
  }
}
```

### SSH agent forwarding does not work

If SSH agent forwarding does not work, the certificate may be missing the `permit-agent-forwarding` extension.

Add it when creating the role or when signing the key:

```json
{
  "default_extensions": {
    "permit-agent-forwarding": ""
  }
}
```

### Key comments

If you need to preserve [comment attributes](https://www.rfc-editor.org/rfc/rfc4716#section-3.3.2) in keys, this operation may require additional steps.

Private and public keys can contain comments.
For example, you can define them with `ssh-keygen` by using the `-C` option:

```shell
ssh-keygen -C "...Comments" -N "" -t rsa -b 4096 -f host-ca
```

Key values that contain comments must be passed together with the parameters associated with that key.
Examples for the Stronghold CLI and API are shown below.

CLI example:

```shell
stronghold secrets enable -path=hosts-ca ssh
KEY_PRI=$(cat ~/.ssh/id_rsa | sed -z 's/\n/\\n/g')
KEY_PUB=$(cat ~/.ssh/id_rsa.pub | sed -z 's/\n/\\n/g')
stronghold write ssh-client-signer/config/ca \
  generate_signing_key=false \
  private_key="${KEY_PRI}" \
  public_key="${KEY_PUB}"
```

API example:

```shell
curl -X POST -H "X-Vault-Token: ..." -d '{"type":"ssh"}' http://127.0.0.1:8200/v1/sys/mounts/hosts-ca
KEY_PRI=$(cat ~/.ssh/id_rsa | sed -z 's/\n/\\n/g')
KEY_PUB=$(cat ~/.ssh/id_rsa.pub | sed -z 's/\n/\\n/g')
tee payload.json <<EOF
{
  "generate_signing_key" : false,
  "private_key"          : "${KEY_PRI}",
  "public_key"           : "${KEY_PUB}"
}
EOF
curl -X POST -H "X-Vault-Token: ..." -d @payload.json http://127.0.0.1:8200/v1/hosts-ca/config/ca
```

{{< alert level="warning" >}}
Do not add a password to the private key because Stronghold cannot decrypt it.
Delete the public key, the private key, and the `payload.json` file from the host immediately after you confirm that the upload succeeded.
{{< /alert >}}

### Known issues

- On systems with SELinux, you may need to configure the appropriate types so that the SSH daemon can read the required files.
  For example, the signed host certificate may require the `sshd_key_t` type.

- In some SSH versions, the following error may occur:

  ```text
  no separate private key for certificate
  ```

  This error appeared in OpenSSH 7.2 and was fixed in OpenSSH 7.5.
  For details, see [OpenSSH bug 2617](https://bugzilla.mindrot.org/show_bug.cgi?id=2617).

- In some SSH versions, the following error may occur on the host:

  ```text
  userauth_pubkey: certificate signature algorithm ssh-rsa: signature algorithm not supported [preauth]
  ```

  To fix it, add the following line to `/etc/ssh/sshd_config`:

  ```text
  CASignatureAlgorithms ^ssh-rsa
  ```

  The `ssh-rsa` algorithm is no longer supported in [OpenSSH 8.2](https://www.openssh.com/txt/release-8.2).
