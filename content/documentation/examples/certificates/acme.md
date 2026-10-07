---
title: "Automatic certificates with ACME"
linkTitle: "ACME"
description: "Automatic issuance and renewal of TLS certificates for internal servers through the ACME server of the Stronghold PKI secrets engine: config/cluster, config/acme, EAB, certbot, lego, and Caddy."
weight: 30
params:
  relatedLinks:
    - title: "Internal PKI with Stronghold"
      url: ../internal-pki/
    - title: "PKI secrets engine"
      url: ../../../user/secrets-engines/pki/
    - title: "Secrets engines API"
      url: ../../../reference/api/secrets/
---

The Stronghold PKI secrets engine implements the ACME protocol (RFC 8555). Internal servers obtain and renew certificates with standard ACME clients such as certbot, lego, and Caddy, without a Stronghold token and without custom issuance scripts.

![ACME certificate issuance flow](../../../images/ex-acme.en.png)

## Goal

Enable ACME on the intermediate CA mount, restrict issuance to a single role, issue external account binding (EAB) keys, and set up automatic certificate issuance and renewal on a server.

## Prerequisites

- The `pki_int` mount with an intermediate CA and the `example-com` role, configured as in [Internal PKI with Stronghold](../internal-pki/) (steps 1–5).
- A Stronghold token with permissions to write to `pki_int/config/*`, `pki_int/roles/*`, `pki_int/eab/*`, and to tune the mount (`sys/mounts/pki_int/tune`).
- An external Stronghold address reachable by ACME clients: `https://stronghold.example.com` in the examples.
- A server named `api.example.com` that runs the ACME client. Stronghold must have network access to this server to validate challenges (port `80`/TCP for `http-01`). The `http-01`, `dns-01`, and `tls-alpn-01` challenge types are supported.
- The certificate that protects the Stronghold API must be trusted on the client server.

## Step 1. Set the external mount address

The ACME server builds directory links from the `path` parameter in `config/cluster`. Set the address clients use to reach the mount:

```bash
d8 stronghold write pki_int/config/cluster \
  path="https://stronghold.example.com/v1/pki_int" \
  aia_path="https://stronghold.example.com/v1/pki_int"
```

In a multi-node cluster, set the address of the load balancer or of any node in this cluster. The address must not point to another replication cluster.

## Step 2. Allow ACME headers

The ACME protocol carries service data in HTTP headers. Allow them for the mount:

```bash
d8 stronghold secrets tune \
  -passthrough-request-headers=If-Modified-Since \
  -allowed-response-headers=Last-Modified \
  -allowed-response-headers=Location \
  -allowed-response-headers=Replay-Nonce \
  -allowed-response-headers=Link \
  pki_int
```

## Step 3. Enable ACME

Enable ACME, allow only the `example-com` role, restrict the default directory, and require EAB:

```bash
d8 stronghold write pki_int/config/acme \
  enabled=true \
  allowed_roles="example-com" \
  default_directory_policy="role:example-com" \
  eab_policy="always-required" \
  max_ttl=720h
```

Parameters:

- `allowed_roles`: roles whose directories may issue certificates. The default value `*` allows all roles, including `sign-verbatim`.
- `default_directory_policy`: the policy for directories without a role (`pki_int/acme/directory`). The default is `sign-verbatim`, which means issuance without role restrictions. The `role:<name>` value applies the restrictions of the specified role; the role must be listed in `allowed_roles`.
- `eab_policy`: whether EAB is required to register an ACME account. Valid values: `not-required`, `new-account-required`, `always-required`. Set the parameter explicitly (the example below uses `always-required`).
- `max_ttl`: the maximum validity period of certificates issued through ACME. The default is `2160h` (90 days).
- `challenge_permitted_ip_ranges`, `challenge_excluded_ip_ranges`: CIDRs Stronghold may or may not connect to when validating challenges.
- `dns_resolver`: a DNS resolver in the `<host>:<port>` format used to resolve names during challenge validation. Set it if internal names are not resolved by the system resolver of the Stronghold nodes.

Check the configuration:

```bash
d8 stronghold read pki_int/config/acme
```

The ACME directory of the role is available without authentication at:

```text
https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory
```

## Step 4. Issue an EAB key

EAB ties an ACME account to a Stronghold operator: only someone who has received a key can register an account. The key is single-use and is consumed when the account is first registered.

1. Create an EAB key for the role directory:

   ```bash
   d8 stronghold write -f -format=json pki_int/roles/example-com/acme/new-eab > eab.json
   jq -r '.data.id, .data.key' eab.json
   ```

   The response contains `id` (the key identifier, `kid`) and `key` (the HMAC key). Pass them to the server administrator over a secure channel.

1. List unused keys:

   ```bash
   d8 stronghold list pki_int/eab
   ```

1. If needed, delete an unused key:

   ```bash
   d8 stronghold delete pki_int/eab/<key_id>
   ```

## Step 5. Configure the ACME client

Perform this step on the `api.example.com` server. Set the variables:

```bash
ACME_DIR="https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory"
EAB_KID="<id from eab.json>"
EAB_HMAC="<key from eab.json>"
```

If the Stronghold API certificate is issued by an internal CA, point the client to a file with that CA. The examples use `/etc/ssl/stronghold-ca.pem`.

### certbot

1. Obtain a certificate. The `--standalone` mode starts a temporary web server on port `80` for `http-01` validation:

   ```bash
   REQUESTS_CA_BUNDLE=/etc/ssl/stronghold-ca.pem \
   certbot certonly --standalone \
     --server "$ACME_DIR" \
     --eab-kid "$EAB_KID" \
     --eab-hmac-key "$EAB_HMAC" \
     --agree-tos -m admin@example.com \
     -d api.example.com
   ```

   If a web server already runs on the host, use `--webroot -w <directory>` or a web server plugin.

1. The certificate and key are saved to `/etc/letsencrypt/live/api.example.com/`: `fullchain.pem` and `privkey.pem`. Reference them in the service configuration.

### lego

```bash
LEGO_CA_CERTIFICATES=/etc/ssl/stronghold-ca.pem \
lego --server "$ACME_DIR" \
  --eab --kid "$EAB_KID" --hmac "$EAB_HMAC" \
  --email admin@example.com --accept-tos \
  --domains api.example.com \
  --http \
  run
```

Certificates are saved to the `.lego/certificates/` directory.

### Caddy

Caddy obtains and renews certificates for sites in the `Caddyfile` on its own. Set Stronghold as the ACME server in the global options:

```caddyfile
{
  email admin@example.com
  acme_ca https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory
  acme_ca_root /etc/ssl/stronghold-ca.pem
  acme_eab {
    key_id <id from eab.json>
    mac_key <key from eab.json>
  }
}

api.example.com {
  reverse_proxy 127.0.0.1:8080
}
```

## Step 6. Configure renewal

The certificate validity is limited by the role `max_ttl` and by `max_ttl` in `config/acme`. Renewal is a new order through the already registered account; no new EAB key is needed.

- **certbot**. certbot packages usually install a systemd timer or a cron job that runs `certbot renew`. For short-lived certificates, set the renewal threshold in `/etc/letsencrypt/renewal/api.example.com.conf`, for example `renew_before_expiry = 3 days`, and reload the service with `--deploy-hook`:

  ```bash
  certbot renew --deploy-hook "systemctl reload nginx"
  ```

  The CA file for `REQUESTS_CA_BUNDLE` must also be available for scheduled runs, for example through `Environment=` in the timer unit.

- **lego**. Run renewal on a schedule, for example daily:

  ```bash
  LEGO_CA_CERTIFICATES=/etc/ssl/stronghold-ca.pem \
  lego --server "$ACME_DIR" --email admin@example.com \
    --domains api.example.com --http \
    renew --days 7 --renew-hook "systemctl reload nginx"
  ```

- **Caddy** renews certificates automatically; no extra configuration is needed.

## Verification

1. Make sure the ACME directory responds:

   ```bash
   curl -s --cacert /etc/ssl/stronghold-ca.pem \
     https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory | jq
   ```

   The response must contain the `newNonce`, `newAccount`, and `newOrder` links with the address from `config/cluster`.

1. Check the issued certificate: issuer, name, and validity:

   ```bash
   openssl x509 -in /etc/letsencrypt/live/api.example.com/fullchain.pem \
     -noout -issuer -subject -dates
   ```

1. Check that the certificate is present in the mount storage:

   ```bash
   d8 stronghold list pki_int/certs
   ```

1. Test renewal without issuing a certificate:

   ```bash
   REQUESTS_CA_BUNDLE=/etc/ssl/stronghold-ca.pem certbot renew --dry-run
   ```

## Cleanup

1. Revoke the test certificate: `certbot revoke --cert-name api.example.com` (with `REQUESTS_CA_BUNDLE`) or `d8 stronghold write pki_int/revoke serial_number=<serial_number>`.
1. Delete unused EAB keys: `d8 stronghold delete pki_int/eab/<key_id>`.
1. Delete the local `eab.json` file.
1. If ACME is no longer needed, disable it: `d8 stronghold write pki_int/config/acme enabled=false`.
