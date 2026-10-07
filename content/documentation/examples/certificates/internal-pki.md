---
title: "Internal PKI with Stronghold"
linkTitle: "Internal PKI"
description: "Building a two-tier PKI: a root CA outside Stronghold or in a separate mount, an intermediate CA in Stronghold, roles, certificate issuance, CRL, OCSP, cert-manager, and ACME."
weight: 10
params:
  relatedLinks:
    - title: "PKI secrets engine"
      url: ../../../user/secrets-engines/pki/
    - title: "cert-manager integration"
      url: ../cert-manager/
    - title: "Secrets engines API"
      url: ../../../reference/api/secrets/
---

A two-tier setup protects the root CA: it signs only intermediate CAs and is unavailable most of the time. Certificates for services are issued by the intermediate CA in Stronghold.

## Goal

Deploy a root CA, an intermediate CA in Stronghold, a role for issuing server certificates, CRL and OCSP publishing, and connect consumers: cert-manager and ACME clients.

## Prerequisites

- A Stronghold token with permissions to enable secrets engines and write to `pki*/`.
- The `jq` and `openssl` utilities on your workstation.
- An external Stronghold address reachable by clients: `https://stronghold.example.com` in the examples.
- A domain for certificates: `example.com` in the examples.

## Step 1. Prepare the root CA

Choose one of the options.

### Option A. Offline root CA (recommended)

The root CA is kept outside Stronghold (for example, on an isolated machine or in an HSM) and is used only to sign intermediate CAs. In this option, there is no `pki_root` mount in Stronghold, and step 3 is performed with the offline CA tools (`openssl ca`, `openssl x509 -req`, or your HSM software). Save the root CA certificate to the `root-ca.pem` file.

### Option B. Root CA in a separate mount

If you do not use an offline CA, keep the root CA in a separate mount that only PKI administrators can access.

1. Enable the secrets engine and set the maximum lifetime:

   ```bash
   d8 stronghold secrets enable -path=pki_root pki
   d8 stronghold secrets tune -max-lease-ttl=87600h pki_root
   ```

1. Generate the root CA. The private key (`internal`) never leaves Stronghold:

   ```bash
   d8 stronghold write -field=certificate pki_root/root/generate/internal \
     common_name="Example Root CA" \
     issuer_name="root-2026" \
     ttl=87600h > root-ca.pem
   ```

1. Publish the issuing certificate and CRL addresses of the root CA:

   ```bash
   d8 stronghold write pki_root/config/urls \
     issuing_certificates="https://stronghold.example.com/v1/pki_root/ca" \
     crl_distribution_points="https://stronghold.example.com/v1/pki_root/crl"
   ```

For a GOST CA, set the `key_type` parameter, for example `key_type=gost3410-256-paramset-a`. The list of allowed values is given in the [API reference](../../../reference/api/secrets/).

## Step 2. Create an intermediate CA

1. Enable a separate mount for the intermediate CA:

   ```bash
   d8 stronghold secrets enable -path=pki_int pki
   d8 stronghold secrets tune -max-lease-ttl=43800h pki_int
   ```

1. Generate a key and a certificate signing request (CSR):

   ```bash
   d8 stronghold write -format=json pki_int/intermediate/generate/internal \
     common_name="Example Intermediate CA" \
     | jq -r '.data.csr' > pki_int.csr
   ```

## Step 3. Sign the intermediate CA with the root

For option B, run:

```bash
d8 stronghold write -format=json pki_root/root/sign-intermediate \
  csr=@pki_int.csr \
  format=pem_bundle \
  ttl=43800h \
  | jq -r '.data.certificate' > pki_int.pem
```

For option A, sign `pki_int.csr` on the offline CA and build a `pki_int.pem` file that contains the intermediate CA certificate and the root CA certificate.

## Step 4. Import the signed certificate

```bash
d8 stronghold write pki_int/intermediate/set-signed certificate=@pki_int.pem
```

Set the CRL, OCSP, and issuing certificate addresses that will be written into issued certificates:

```bash
d8 stronghold write pki_int/config/urls \
  issuing_certificates="https://stronghold.example.com/v1/pki_int/ca" \
  crl_distribution_points="https://stronghold.example.com/v1/pki_int/crl" \
  ocsp_servers="https://stronghold.example.com/v1/pki_int/ocsp"
```

## Step 5. Create a role

The role restricts domains and certificate lifetime:

```bash
d8 stronghold write pki_int/roles/example-com \
  allowed_domains="example.com" \
  allow_subdomains=true \
  max_ttl=720h
```

Keep `max_ttl` short: the shorter the lifetime, the less often revocation is needed and the smaller the CRL.

## Step 6. Issue a certificate

```bash
d8 stronghold write pki_int/issue/example-com \
  common_name="api.example.com" \
  ttl=72h
```

The response contains `certificate`, `issuing_ca`, `ca_chain`, `private_key`, and `serial_number`. If the private key is generated on the client side, send a CSR to the `pki_int/sign/example-com` endpoint.

## Step 7. Configure CRL and OCSP

1. Enable automatic CRL rebuilding:

   ```bash
   d8 stronghold write pki_int/config/crl \
     auto_rebuild=true \
     expiry=72h
   ```

1. Enable automatic cleanup of expired and revoked certificates:

   ```bash
   d8 stronghold write pki_int/config/auto-tidy \
     enabled=true \
     tidy_cert_store=true \
     tidy_revoked_certs=true \
     safety_buffer=72h
   ```

1. Revoke a certificate by its serial number:

   ```bash
   d8 stronghold write pki_int/revoke serial_number=<serial_number>
   ```

## Step 8. Connect cert-manager

To issue certificates for Ingress and Kubernetes pods, use cert-manager with a `vault` Issuer pointing to `pki_int/sign/example-com`. Setting up the Stronghold role and cert-manager resources is described in [cert-manager integration](../cert-manager/).

## Step 9. Enable ACME (optional)

The PKI engine supports the ACME protocol, so certificates can be obtained with standard ACME clients.

1. Set the external mount address used in ACME links:

   ```bash
   d8 stronghold write pki_int/config/cluster \
     path="https://stronghold.example.com/v1/pki_int" \
     aia_path="https://stronghold.example.com/v1/pki_int"
   ```

1. Allow the headers required by the ACME protocol:

   ```bash
   d8 stronghold secrets tune \
     -passthrough-request-headers=If-Modified-Since \
     -allowed-response-headers=Last-Modified \
     -allowed-response-headers=Location \
     -allowed-response-headers=Replay-Nonce \
     -allowed-response-headers=Link \
     pki_int
   ```

1. Enable ACME and restrict roles:

   ```bash
   d8 stronghold write pki_int/config/acme \
     enabled=true \
     allowed_roles="example-com" \
     max_ttl=720h
   ```

1. In the ACME client, use the role directory `https://stronghold.example.com/v1/pki_int/roles/example-com/acme/directory`. By default, account registration requires external account binding (EAB, the `eab_policy` parameter).

## Verification

1. Check the chain of an issued certificate:

   ```bash
   d8 stronghold write -format=json pki_int/issue/example-com common_name=test.example.com ttl=1h > test.json
   jq -r '.data.certificate' test.json > test.pem
   jq -r '.data.issuing_ca' test.json > int.pem
   openssl verify -CAfile root-ca.pem -untrusted int.pem test.pem
   ```

1. Check that the CRL is available:

   ```bash
   curl -s https://stronghold.example.com/v1/pki_int/crl/pem | openssl crl -noout -text | head
   ```

1. Check the certificate status via OCSP:

   ```bash
   openssl ocsp -issuer int.pem -VAfile int.pem -cert test.pem \
     -url https://stronghold.example.com/v1/pki_int/ocsp -text
   ```

## Cleanup

1. Revoke test certificates: `d8 stronghold write pki_int/revoke serial_number=<serial_number>`.
1. Delete the local files `test.json`, `test.pem`, and `int.pem`.
1. If the setup was for testing, disable the mounts: `d8 stronghold secrets disable pki_int` and `d8 stronghold secrets disable pki_root`. Disabling a mount deletes the CA keys irreversibly.
