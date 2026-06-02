---
title: "PKI secrets engine"
description: "Information about PKI secrets engine in Deckhouse Stronghold."
weight: 20
---

The PKI secrets engine allows Stronghold to work as a certificate authority (CA).
It can generate dynamic root and intermediate certificates and sign certificates on request.

The Public Key Infrastructure (PKI) secrets engine generates dynamic X.509 certificates.
With this secrets engine, services can obtain certificates without manually creating a private key and certificate signing request (CSR), submitting a request to the CA, and waiting for verification and signing.
The built-in authentication and authorization mechanisms in Stronghold provide the required verification.

If you use a short TTL, certificate revocation is needed less often.
This helps keep the certificate revocation list (CRL) short and reduces the load on the secrets engine.
As a result, each application instance can use its own certificate instead of sharing one with other instances, which simplifies rotation and revocation.

The PKI secrets engine also supports ephemeral certificates.
Applications can retrieve and store them in memory at startup and delete them on shutdown without writing them to disk.

## Configuration

Most secrets engines must be configured in advance before they can perform their functions.
These steps are usually performed by an operator or a configuration management system.

1. Enable the PKI secrets engine:

   ```shell
   d8 stronghold secrets enable pki
   ```

   Example output:

   ```console
   Success! Enabled the pki secrets engine at: pki/
   ```

   By default, the secrets engine is mounted at the engine name.
   To enable it at a different path, use the `-path` argument.

1. Increase the TTL for the secrets engine.
   The default value of 30 days may be too short.
   For example, you can increase it to one year:

   ```shell
   d8 stronghold secrets tune -max-lease-ttl=8760h pki
   ```

   Example output:

   ```console
   Success! Tuned the secrets engine at: pki/
   ```

   Individual roles can still limit this value for specific certificates.
   This parameter only sets the global maximum TTL for the secrets engine.

1. Configure the CA certificate and private key.
   Stronghold can use an existing key pair or generate its own self-signed root certificate.
   In most cases, it is recommended to keep the root CA outside Stronghold and provide Stronghold with a signed intermediate CA.

   ```shell
   d8 stronghold write pki/root/generate/internal \
     common_name=my-website.ru \
     ttl=8760h
   ```

   Example output:

   ```console
   Key              Value
   ---              -----
   certificate      -----BEGIN CERTIFICATE-----...
   expiration       1756317679
   issuing_ca       -----BEGIN CERTIFICATE-----...
   serial_number    fc:f1:fb:2c:6d:4d:99:1e:82:1b:08:0a:81:ed:61:3e:1d:fa:f5:29
   ```

   The returned certificate is provided for informational purposes only.
   The private key is stored securely inside Stronghold.

1. Update the URLs for the CRL and issuing certificates.
   You can change these values later if needed:

   ```shell
   d8 stronghold write pki/config/urls \
     issuing_certificates="http://127.0.0.1:8200/v1/pki/ca" \
     crl_distribution_points="http://127.0.0.1:8200/v1/pki/crl"
   ```

   Example output:

   ```console
   Success! Data written to: pki/config/urls
   ```

1. Configure a role that maps a Stronghold role name to certificate generation rules.
   When users or machines request credentials, certificates are issued according to this role:

   ```shell
   d8 stronghold write pki/roles/example-dot-ru \
     allowed_domains=my-website.ru \
     allow_subdomains=true \
     max_ttl=72h
   ```

   Example output:

   ```console
   Success! Data written to: pki/roles/example-dot-ru
   ```

## Usage

After you configure the secrets engine and obtain a Stronghold token with the required permissions, you can generate credentials.

1. Generate new credentials by writing to the `/issue` path with the role name:

   ```shell
   d8 stronghold write pki/issue/example-dot-ru \
     common_name=www.my-website.ru
   ```

   Example output:

   ```console
   Key                 Value
   ---                 -----
   certificate         -----BEGIN CERTIFICATE-----...
   issuing_ca          -----BEGIN CERTIFICATE-----...
   private_key         -----BEGIN RSA PRIVATE KEY-----...
   private_key_type    rsa
   serial_number       1d:2e:c6:06:45:18:60:0e:23:d6:c5:17:43:c0:fe:46:ed:d1:50:be
   ```

   The response includes a dynamically generated private key and certificate.
   The certificate matches the specified role and expires after 72 hours, as defined in the role configuration.
   The issuing CA and trust chain are also returned to simplify automation.
