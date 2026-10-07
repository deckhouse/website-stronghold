---
title: "Service mesh"
description: "Using Stronghold PKI as an external certificate authority for Istio via cert-manager and istio-csr."
weight: 50
---

{{< alert level="warning" >}}
This scenario describes integrating third-party components and has not been verified as part of Deckhouse Platform. <!-- TODO(verify): external CA support in the DKP istio module and istio-csr compatibility with Stronghold -->
{{< /alert >}}

By default, Istio issues workload certificates for mTLS with its own certificate authority in istiod. To have certificates issued by an intermediate CA from Stronghold, use this chain:

1. The Stronghold [PKI secrets engine](../../../user/secrets-engines/pki/) holds the intermediate CA for the mesh and a role for issuing SPIFFE certificates.
1. cert-manager calls Stronghold through a `vault` Issuer (see [cert-manager](../cert-manager/)).
1. The istio-csr component of the cert-manager project accepts certificate requests from Istio proxies and issues them through cert-manager.

## Configuring Stronghold

Create a PKI role that allows SPIFFE URI SANs used by Istio:

```bash
d8 stronghold write pki/roles/istio-mesh \
  allowed_uri_sans="spiffe://cluster.local/*" \
  allow_any_name=true \
  require_cn=false \
  max_ttl=24h
```

Allow the cert-manager Issuer to sign requests for this role (`pki/sign/istio-mesh`), as described in [cert-manager](../cert-manager/). If a ClusterIssuer is used, the token audience in the Kubernetes auth role is `vault://<ClusterIssuer name>`.

## Configuring cert-manager and Istio

1. Create a `stronghold-mesh` ClusterIssuer of type `vault` with `path: pki/sign/istio-mesh`.
1. Install istio-csr and set the `stronghold-mesh` ClusterIssuer in its parameters.
1. Configure Istio to use istio-csr as its CA: set the istio-csr service address in `global.caAddress` and disable the built-in istiod CA (`ENABLE_CA_SERVER=false`).

Example istio-csr Helm chart values:

```yaml
app:
  certmanager:
    issuer:
      name: stronghold-mesh
      kind: ClusterIssuer
      group: cert-manager.io
```

The chain root certificate that proxies must trust can be read from Stronghold:

```bash
d8 stronghold read -field=certificate pki/cert/ca
```

Set a short `max_ttl` for workload certificates: they are renewed automatically.
