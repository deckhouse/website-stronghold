---
title: "TLS certificate for a web server on a VM from Stronghold PKI"
linkTitle: "TLS for a web server on a VM"
description: "Automatic issuance and renewal of an Nginx TLS certificate on a virtual machine: a PKI role, a policy, a Stronghold Agent template with pki/issue, and an Nginx reload after renewal; notes for Apache."
weight: 60
params:
  relatedLinks:
    - title: "PKI secrets engine"
      url: ../../../user/secrets-engines/pki/
    - title: "Internal PKI with Stronghold"
      url: ../../certificates/internal-pki/
    - title: "Stronghold Agent: key features"
      url: ../../../user/agent/key-features/
    - title: "Application on a VM with Stronghold Agent"
      url: ../legacy-app-on-vm/
    - title: "Agent launch and management"
      url: ../../../user/agent/launch-and-control/
---

A web server on a virtual machine does not need a certificate issued manually for a year. Stronghold Agent requests a certificate from the PKI secrets engine, writes it to disk, reissues it before it expires, and reloads Nginx. The private key is generated in Stronghold and is passed only to Agent on the VM.

![Web server TLS certificate issuance and renewal flow](../../../images/ex-web-server-tls.en.png)

## Goal

Configure Stronghold Agent on a VM so that it issues a 72-hour `www.example.com` certificate for Nginx from the `pki_int` intermediate CA, reissues it automatically, and reloads Nginx without dropping connections.

## Prerequisites

- An intermediate CA in Stronghold mounted at `pki_int`, as in the [Internal PKI with Stronghold](../../certificates/internal-pki/) guide.
- A Linux virtual machine with systemd and Nginx.
- Stronghold Agent installed on the VM and logging in to Stronghold with AppRole, as in the [Application on a VM with Stronghold Agent](../legacy-app-on-vm/) guide.
- A Stronghold token with permissions to configure PKI, AppRole, and policies.

## Step 1. Create a PKI role

The role restricts the names and lifetime of certificates that the web server can obtain:

```bash
d8 stronghold write pki_int/roles/web-server \
  allowed_domains="example.com" \
  allow_subdomains=true \
  max_ttl=720h
```

A short certificate lifetime reduces the impact of a key compromise: Agent renews the certificate automatically, so manual reissuance is not needed.

## Step 2. Create a policy and an AppRole role

1. Create a policy that allows issuing certificates only through the `web-server` role:

   ```bash
   d8 stronghold policy write web-server-tls - <<'POLICY'
   path "pki_int/issue/web-server" {
     capabilities = ["update"]
   }
   POLICY
   ```

1. Create an AppRole role for the VM:

   ```bash
   d8 stronghold write auth/approle/role/web-server \
     token_policies=web-server-tls \
     token_ttl=1h \
     token_max_ttl=24h \
     secret_id_ttl=720h \
     secret_id_bound_cidrs="10.0.10.20/32"
   ```

1. Deliver `role_id` and a wrapped `secret_id` to the `/etc/stronghold-agent/` directory on the VM, as described in step 3 of the [Application on a VM with Stronghold Agent](../legacy-app-on-vm/#step-3-deliver-role_id-and-a-wrapped-secret_id-to-the-vm) guide.

## Step 3. Create a template

Create the `/etc/stronghold-agent/templates/www.pem.ctmpl` file. The template writes the private key, the certificate, and the chain of intermediate CAs into a single file, so the key and the certificate always match each other:

```text
{{ with secret "pki_int/issue/web-server" "common_name=www.example.com" "alt_names=example.com" "ttl=72h" }}
{{ .Data.private_key }}
{{ .Data.certificate }}
{{ range .Data.ca_chain }}{{ . }}
{{ end }}{{ end }}
```

Each `secret "pki_int/issue/..."` call issues a certificate with a new key. If you need separate files for the certificate and the key, as in the example from the [Key features](../../../user/agent/key-features/) section, use a call with identical arguments in both templates.

Identical calls (same path and same arguments) are executed by Agent once and the result is reused across all templates, so the key and the certificate in separate files form a pair.

The `pkiCert` template function is also available for issuing certificates: it writes the certificate, key, and chain to files itself and does not require parsing the `secret` response manually.

## Step 4. Configure Agent

Add a `template` block to `/etc/stronghold-agent/agent.hcl`:

```hcl
template {
  source          = "/etc/stronghold-agent/templates/www.pem.ctmpl"
  destination     = "/etc/nginx/tls/www.pem"
  perms           = "0600"
  command         = "sudo /usr/sbin/nginx -t && sudo /usr/sbin/nginx -s reload"
  command_timeout = "30s"
  error_on_missing_key = true
}
```

- `perms = "0600"` — the file contains the private key; the Nginx master process reads it with root privileges;
- `command` — runs after each write of the file: `nginx -t` checks the configuration, and `nginx -s reload` rereads the certificate without dropping current connections.

If `command` is a single string containing spaces, Agent runs it through `sh -c`, so `&&` and other shell features work.

Allow the `stronghold-agent` user to run only these commands, for example through sudoers:

```text
stronghold-agent ALL=(root) NOPASSWD: /usr/sbin/nginx -t, /usr/sbin/nginx -s reload
```

Create a directory for the certificate and add it to `ReadWritePaths` in the Agent unit file:

```bash
sudo install -d -o stronghold-agent -g root -m 0750 /etc/nginx/tls
```

## Step 5. Configure Nginx

Specify the same file in `ssl_certificate` and `ssl_certificate_key`:

```nginx
server {
    listen 443 ssl;
    server_name www.example.com example.com;

    ssl_certificate     /etc/nginx/tls/www.pem;
    ssl_certificate_key /etc/nginx/tls/www.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;

    location / {
        root /var/www/html;
    }
}
```

Start Agent before the first start of Nginx with this configuration so that the certificate file already exists:

```bash
sudo systemctl restart stronghold-agent
sudo systemctl reload nginx
```

## Step 6. How renewal works

A PKI certificate is issued without a renewable lease. Agent tracks the certificate validity period and, before it expires, runs `pki_int/issue/web-server` again, overwrites the file, and runs `command`. Choose the `ttl` in the template so that reissuance happens no more often than you are ready to reload Nginx and no later than you can notice a failure.

Agent reissues the certificate after about 90% of its remaining validity period has elapsed (with a random spread of ±5% so that clients do not arrive at the same time). For `ttl = 24h` this is roughly 21–22 hours.

## Apache HTTP Server

For Apache, use the same template and the same `template` block with the following differences:

- in the virtual host configuration, specify the file in the `SSLCertificateFile` and `SSLCertificateKeyFile` directives. Apache 2.4.8 and later reads intermediate certificates from `SSLCertificateFile`:

  ```apacheconf
  <VirtualHost *:443>
      ServerName www.example.com
      SSLEngine on
      SSLCertificateFile    /etc/apache2/tls/www.pem
      SSLCertificateKeyFile /etc/apache2/tls/www.pem
  </VirtualHost>
  ```

- in `destination`, specify `/etc/apache2/tls/www.pem`, and in `command`, a graceful restart: `sudo /usr/sbin/apachectl -t && sudo /usr/sbin/apachectl -k graceful` (or `systemctl reload apache2`/`systemctl reload httpd`, depending on the distribution).

## Verification

1. Check the Agent log — it should contain messages about rendering the template and running the command:

   ```bash
   sudo journalctl -u stronghold-agent -n 50
   ```

1. Check the certificate that Nginx serves:

   ```bash
   openssl s_client -connect www.example.com:443 -servername www.example.com </dev/null 2>/dev/null \
     | openssl x509 -noout -subject -issuer -serial -dates
   ```

   The issuer must match the `pki_int` intermediate CA, and the validity period must match the `ttl` from the template.

1. Check the chain of trust against the root CA certificate:

   ```bash
   curl --cacert root-ca.pem -sSI https://www.example.com
   ```

1. Check reissuance: temporarily set `ttl=10m` in the template, restart Agent, and make sure that the certificate serial number from verification step 2 changes without manual actions.

## Cleanup

1. Stop Agent and remove the `template` block and the template file.
1. Delete the `/etc/nginx/tls/www.pem` file and remove the `ssl_certificate` directives from the Nginx configuration.
1. If necessary, revoke the issued certificate and delete the roles and the policy:

   ```bash
   d8 stronghold write pki_int/revoke serial_number=<serial_number>
   d8 stronghold delete pki_int/roles/web-server
   d8 stronghold delete auth/approle/role/web-server
   d8 stronghold policy delete web-server-tls
   ```
