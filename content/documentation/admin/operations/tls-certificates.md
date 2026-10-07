---
title: "TLS certificates"
description: "Replacing Stronghold listener TLS certificates in Linux without a service restart, and updating Stronghold certificates in DKP."
weight: 60
---

This page describes replacing Stronghold's own TLS certificates: the API listener certificate and the certificates nodes use to communicate within the Raft cluster. Issuing certificates for applications is covered in [PKI](../../../user/secrets-engines/pki/).

Monitor certificate expiry with the alerts from [Monitoring](../monitoring/#alerting-rule-examples) and replace certificates in advance.

## Stronghold in Linux

The listener certificate and key are set with the `tls_cert_file` and `tls_key_file` parameters of the `listener "tcp"` block (see [Configuration](../../../install/standalone/configuration/#listener)). The certificate file must contain the full chain: the server certificate first, followed by the intermediate CA certificates.

### Checking the current certificate

```shell
openssl x509 -in /opt/stronghold/tls/node-1-cert.pem -noout -subject -issuer -dates -ext subjectAltName
```

### Replacing a certificate without a restart

Stronghold rereads the listener certificate and key files on `SIGHUP`. Changes to file paths are not applied this way, so place the new files at the same paths.

Perform on each node in turn:

1. Issue a new certificate with the same SANs (the node DNS names and IP addresses, and the load balancer address, if used).
1. Check that the certificate matches the key:

   ```shell
   openssl x509 -in node-1-cert.pem -noout -pubkey | sha256sum
   openssl pkey -in node-1-key.pem -pubout | sha256sum
   ```

   The hashes must match.

1. Back up the current files and replace them, preserving the owner and permissions:

   ```shell
   cp -a /opt/stronghold/tls /opt/stronghold/tls.bak-$(date +%F)
   install -o stronghold -g stronghold -m 0600 node-1-key.pem /opt/stronghold/tls/node-1-key.pem
   install -o stronghold -g stronghold -m 0644 node-1-cert.pem /opt/stronghold/tls/node-1-cert.pem
   ```

1. Send `SIGHUP` to the process:

   ```shell
   systemctl reload stronghold
   ```

1. Check that the server presents the new certificate:

   ```shell
   openssl s_client -connect raft-node-1.demo.tld:8200 </dev/null 2>/dev/null \
     | openssl x509 -noout -dates
   ```

1. Check the logs for certificate reload errors:

   ```shell
   journalctl -u stronghold.service --since "5 minutes ago"
   ```

If the key is password-protected, the password must match the original one when reloading on `SIGHUP`.

The `retry_join` certificates (`leader_client_cert_file`, `leader_ca_cert_file`, `leader_client_key_file`) are not reread on `SIGHUP`: they are loaded when the node attempts to join the cluster, so new files take effect after the node restarts. Only the TLS listener certificates are reloaded on `SIGHUP`.

### Replacing the CA

When changing the CA that issued node certificates, follow this order so that nodes and clients keep trusting each other:

1. Add the new CA to the trust store on all clients and to the `leader_ca_cert_file` files on all nodes (the file may contain several CA certificates).
1. Restart the nodes one at a time, starting with standby nodes, and wait until each node rejoins the cluster (`d8 stronghold operator raft list-peers`). Nodes with Shamir seal must be unsealed after a restart.
1. Replace node certificates with ones issued by the new CA as described above.
1. After all nodes are updated, remove the old CA from the trust stores.

A restart is required: the `leader_ca_cert_file` files are read only when the node joins.

## Stronghold in DP

In DP, the following external access types (inlets) are supported: `Ingress`, `GatewayAPI`, `LoadBalancer`, `NodePort` and `None`. For `LoadBalancer`, `NodePort` and `None`, the `https.mode: CustomCertificate` mode is required. The certificate for the `stronghold.<domain>` domain is issued and stored according to the `https` settings (see [Configuring Stronghold](../../../install/dkp/configuration/#ways-to-organize-access-via-the-ingress-inlet)).

Stronghold uses two kinds of certificates:

- **External** certificate: the `ingress-tls` secret in `CertManager` mode or the `ingress-tls-customcertificate` secret in `CustomCertificate` mode, in the `d8-stronghold` namespace. It is mounted into the Pod at `/stronghold/tls` and used by the API listener on port `8200`.
- **Internal** certificate: the `stronghold-tls` secret (`ca.crt`, `tls.crt`, `tls.key`). It is issued by a module hook as a self-signed one and mounted at `/stronghold/tls-internal`. It is used on ports `8300`, `8301` and `8500` (communication between Pods and with the proxy). Its SANs include `127.0.0.1`, `stronghold`, `*.stronghold-internal`, `stronghold.d8-stronghold` and `stronghold.d8-stronghold.svc`. It is managed by the module; you do not need to replace it manually.

| Method | Certificate renewal |
| --- | --- |
| `CertManager` with a ClusterIssuer (Let's Encrypt or a private CA) | Automatic, by cert-manager |
| `CustomCertificate` | Manual, by updating the secret in the `d8-system` namespace |

The `stronghold` Certificate resource exists only with the `Ingress` inlet and the `CertManager` mode.

### Checking the certificate

```shell
d8 k -n d8-stronghold get certificate
d8 k -n d8-stronghold get secret ingress-tls -o jsonpath='{.data.tls\.crt}' \
  | base64 -d | openssl x509 -noout -subject -issuer -dates
```

The commands are for the `CertManager` mode. In `CustomCertificate` mode, use the `ingress-tls-customcertificate` secret (there is no Certificate resource).

### cert-manager certificates

cert-manager renews the certificate automatically before it expires. If the certificate was not renewed, check the CertificateRequest resources:

```shell
d8 k -n d8-stronghold get certificaterequest
d8 k -n d8-stronghold describe certificaterequest <NAME>
```

To force reissuing, delete the `ingress-tls` secret: cert-manager issues it again.

<!-- TODO(verify): whether deleting ingress-tls is safe for forced reissue (pods get stuck in ContainerCreating without the secret), or cmctl renew is recommended. -->

### Custom certificate

In `CustomCertificate` mode, replace the contents of the secret set in `settings.modules.https.customCertificate.secretName`:

```shell
d8 k -n d8-system create secret tls mycompany-wildcard-tls \
  --cert=kubernetes_fullchain.crt --key=kubernetes.key \
  --dry-run=client -o yaml | d8 k apply -f -
```

Deckhouse copies the updated certificate to module namespaces. Check that Stronghold presents the new certificate:

```shell
openssl s_client -connect stronghold.mycompany.tld:443 </dev/null 2>/dev/null \
  | openssl x509 -noout -dates
```

If the CA has changed and `user-authn` uses `dexCAMode: FromIngressSecret`, make sure Dex has received the new CA, otherwise OIDC login stops working.

<!-- TODO(verify): timing and mechanism of CustomCertificate propagation to d8-stronghold; whether Stronghold pods restart; how internal certificates between pods (port 8300) are rotated. -->
