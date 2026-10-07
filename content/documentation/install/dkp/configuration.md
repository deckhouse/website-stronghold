---
title: "Stronghold configuration"
description: "How to enable and disable the stronghold module in Deckhouse Kubernetes Platform, get access to the service, manage administrator access, and configure Ingress certificates."
weight: 60
---

## Enabling the module

The module runs in one of two editions, determined by whether a license key is present in ModuleConfig. Choose the edition before enabling the module.

**Base Stronghold**: no license key is required; simply enable the module:

```shell
d8 system module enable stronghold
```

**Stronghold EE**: create a ModuleConfig with the license key in the [`spec.settings.license`](/modules/stronghold/stable/configuration.html#parameters-license) parameter. No separate enable command is needed: the module is enabled by `spec.enabled: true` in the manifest.

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: stronghold
spec:
  enabled: true
  version: 1
  settings:
    license: <STRONGHOLD_EE_LICENSE_KEY>
```

Stronghold EE is licensed separately and is available only in commercial DP editions. For a comparison of editions, see [Editions](../../about/editions/).

{{< alert level="warning" >}}
If the module is already running and the storage contains data, add the key by following the [instructions for switching to Stronghold EE](../platform-management/switching-editions/ce-to-ee/).
{{< /alert >}}

{{< alert level="info" >}}
Enabling and initial configuration are covered in more detail [in the free "Deckhouse Stronghold capabilities overview" course](https://education.flant.ru/course/obzor-vozmozhnostej-deckhouse-stronghold/) by [Deckhouse Academy](https://deckhouse.ru/course-catalog/) (in Russian).
{{< /alert >}}

By default, the module starts in the `Automatic` mode (`management.mode`) with the `Ingress` inlet (`inlet`).

The following modes and inlets are available:

- `management.mode`: `Automatic` (default) or `Manual`.
- `inlet`: `Ingress` (default), `GatewayAPI`, `LoadBalancer`, `NodePort`, or `None`. For `LoadBalancer`, `NodePort`, and `None`, you must set `https.mode: CustomCertificate`.

## Disabling the module

To disable the module, run the following command:

```shell
d8 k annotate mc stronghold modules.deckhouse.io/allow-disabling=true
d8 system module disable stronghold
```

The module requires confirmation to be disabled, so the `modules.deckhouse.io/allow-disabling=true` annotation must be set first.

{{< alert level="danger" >}}
When the module is disabled, all Stronghold containers are deleted from the `d8-stronghold` namespace, as well as the `stronghold-keys` secret with the root and unseal keys. If `storageClass` is not set, the service data is not deleted from the node (see below); if `storageClass` is set, the data is stored in the `data-stronghold-N` PVCs, which are deleted together with the StatefulSet. You can enable the module again and put a saved copy of the `stronghold-keys` secret into the `d8-stronghold` namespace; access to the data will then be restored.
{{< /alert >}}

If `storageClass` is not set and you no longer need the old data, delete the `/var/lib/deckhouse/stronghold` directory
from all nodes where Stronghold data is stored (master nodes by default) beforehand.
If `storageClass` is set, the data resides in the `data-stronghold-N` PVCs rather than in this directory.

## Accessing the service

The service is accessed via inlets. An inlet is a source of input data for a pod. The following inlets are supported: `Ingress` (default), `GatewayAPI`, `LoadBalancer`, `NodePort`, and `None`. The examples below use `Ingress`.

The address of the Stronghold web interface is formed as follows: in the [`publicDomainTemplate`](/products/kubernetes-platform/documentation/v1/reference/api/global.html#parameters-modules-publicdomaintemplate) template of the global Deckhouse configuration parameter, the `%s` key is replaced with `stronghold`.
For example, if `publicDomainTemplate` is set to `%s-kube.mycompany.tld`, the Stronghold web interface is available at `stronghold-kube.mycompany.tld`.

## Using the data storage. Operating modes

The information stored in Stronghold is protected by encryption. To decrypt the storage data, an encryption key is required. This key is also stored together with the data (in the keyring), but it is encrypted with another key called the root key.

To decrypt the data, Stronghold first decrypts the encryption key using the root key. Access to the root key is obtained through a process called unsealing the storage. The root key is stored together with the rest of the storage data, but it is encrypted with yet another key: the unseal key.

The module supports two modes, set by the `management.mode` parameter: `Automatic` (default) and `Manual`.

In the `Automatic` mode, the storage is initialized automatically on the first start of the module. During initialization, the unseal key and the root token are placed into the `stronghold-keys` secret in the `d8-stronghold` Kubernetes namespace. After initialization, the module automatically unseals the Stronghold cluster nodes.
In the automatic mode, when Stronghold nodes restart, the storage is also unsealed automatically without user intervention.

In the `Manual` mode, automatic initialization is disabled, the Dex and Kubernetes integrations are not configured, the `stronghold-keys` secret is not created, and the `management.administrators` parameter is unavailable. Initialize and unseal the storage yourself.

## Access management

In the `Automatic` mode, after the storage is initialized, Stronghold creates the `deckhouse_administrators` role, which is granted access to the web interface via [Dex](/modules/user-authn/) OIDC authentication.
Automatic connection of the current Deckhouse cluster to Stronghold is also configured for the [`secrets-store-integration`](/modules/secrets-store-integration/stable/) module.

{{< alert level="info" >}}
A hands-on walkthrough of delivering secrets to an application is available [in the "Deckhouse Stronghold capabilities overview" course](https://education.flant.ru/course/obzor-vozmozhnostej-deckhouse-stronghold/) (in Russian).
{{< /alert >}}

To grant permissions to users in the `admins` group (group membership is passed from the IdP or LDAP in use via [Dex](/modules/user-authn/)), specify this group in the `management.administrators` array in `ModuleConfig` (available only when `management.mode: Automatic`; each item has a `type` of `Group` or `User` and a `name`):

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: stronghold
spec:
  enabled: true
  version: 1
  settings:
    management:
      mode: Automatic
      administrators:
      - type: Group
        name: admins
```

To grant `administrator` permissions to the `manager` and `securityoperator` users, you can use the following parameters in `ModuleConfig`:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: stronghold
spec:
  enabled: true
  version: 1
  settings:
    management:
      mode: Automatic
      administrators:
      - type: User
        name: manager@mycompany.tld
      - type: User
        name: securityoperator@mycompany.tld
```

Although access can be assigned to individual users, they must belong to at least one group to authenticate via OIDC.

Later, you can create users in Stronghold with different access permissions to secrets using the built-in storage mechanism.

## First launch

Before enabling the Stronghold module for the first time (with an empty `storageClass`), make sure that the `/var/lib/deckhouse/stronghold` directory does not exist on the file system of the nodes where Stronghold nodes will run (master nodes by default), and that the [`stronghold` module is disabled](#disabling-the-module).

{{< alert level="info" >}}
Experience with the `kubectl` utility is required.
{{< /alert >}}

Below are options for organizing access to the module via the [Ingress inlet](/modules/ingress-nginx/), followed by the process of enabling the module and checking that it works.

### Ways to organize access via the Ingress inlet

#### ClusterIssuer LetsEncrypt

This certificate issuance method is configured by default. However, it is suitable only for services reachable from the Internet (and not for internal networks). To check availability:

1. Get the address of the authentication platform:

    ```shell
    d8 k -n d8-user-authn get ing dex
    # Expected output:
    # NAME   CLASS   HOSTS               ADDRESS         PORTS     AGE
    # dex    nginx   dex.mycompany.tld   34.85.243.109   80, 443   4d20h
    ```

    The `HOSTS` column contains the domain to check, and the `ADDRESS` column contains its IP address. Now make sure that the domain name points to this IP address. To do this, run the command:

    ```shell
    nslookup dex.mycompany.tld 8.8.8.8
    # Expected output:
    # ...
    # Name: dex.mycompany.tld
    # Address: 34.85.243.109
    # ...

    # Or:
    dig @8.8.8.8 dex.mycompany.tld
    # Expected output:
    # ...
    # ;; ANSWER SECTION:
    # dex.mycompany.tld. 3600 IN A 34.85.243.109
    # ...
    ```

    If the response is an `NXDOMAIN` error, configure the user's DNS.

1. Open <https://dex.mycompany.tld/healthz> in a browser or run `curl -kL https://dex.mycompany.tld/healthz`. The response must be `Health check passed`.
1. Check that the Ingress controller handles requests to your `stronghold.mycompany.tld` subdomain. Open <https://stronghold.mycompany.tld> in a browser or with `curl -kL`. A 404 error must be returned.

#### ClusterIssuer with a self-signed certificate authority

This option is suitable if you want to use your own self-signed certificate authority. As an example, the already created **selfsigned** `ClusterIssuer` resource is used. To add an Issuer or ClusterIssuer with your own self-signed certificate authority, see the [official documentation](https://cert-manager.io/docs/configuration/ca/).

{{< alert level="info" >}}
This method works both with a public domain name and with access from the internal network only.
{{< /alert >}}

Edit the global module settings. To do this, run `d8 k edit mc global`.
Add the `settings.modules.https.certManager.clusterIssuerName: selfsigned` parameter. The resulting module configuration must look as follows:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: global
spec:
  settings:
    modules:
      https:
        certManager:
          clusterIssuerName: selfsigned    # The only parameter to add.
      publicDomainTemplate: '%s.mycompany.tld'
  version: 1
```

Next, edit the `user-authn` module settings. Run `d8 k edit mc user-authn` and change the `settings.controlPlaneConfigurator.dexCAMode` parameter to `FromIngressSecret`:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: user-authn
spec:
  enabled: true
  settings:
    controlPlaneConfigurator:
      dexCAMode: FromIngressSecret    # The parameter to change.
  ...
```

Before starting the module, make sure the key services are reachable from the **working network**.

1. Get the address of the authentication platform:

    ```shell
    d8 k -n d8-user-authn get ing dex
    # Expected output
    # NAME   CLASS   HOSTS               ADDRESS         PORTS     AGE
    # dex    nginx   dex.mycompany.tld   34.85.243.109   80, 443   4d20h
    ```

    The `HOSTS` column contains the domain to check, and the `ADDRESS` column contains its IP address. Now make sure that the domain name points to this IP address. To do this, run the command:

    ```shell
    nslookup dex.mycompany.tld
    # Expected output:
    # ...
    # Name: dex.mycompany.tld
    # Address: 34.85.243.109
    # ...

    # Or:
    dig dex.mycompany.tld
    # Expected output:
    # ...
    # ;; ANSWER SECTION:
    # dex.mycompany.tld. 3600 IN A 34.85.243.109
    # ...
    ```

    If the response is an `NXDOMAIN` error, configure the user's DNS. If the domain is not reachable from the Internet, perform an [additional step](#what-to-do-if-dexmycompanytld-does-not-resolve-via-dns).

    > As a temporary workaround, you can add the following line to the `/etc/hosts` file of your Unix system:
    >
    > ```shell
    > 34.85.243.109 dex.mycompany.tld stronghold.mycompany.tld
    > ```

1. Open <https://dex.mycompany.tld/healthz> in a browser or run `curl -kL https://dex.mycompany.tld/healthz`. The response must be `Health check passed`.
1. Check that the Ingress controller handles requests to your `stronghold.mycompany.tld` subdomain. Open <https://stronghold.mycompany.tld> in a browser or with `curl -kL`. A 404 error must be returned.

#### Using a certificate file

Create a CA and a certificate, and then sign the certificate with this CA. If a CA already exists, you can sign the certificate with the existing CA.
The certificate must be generated with the full chain (fullchain).

Below is the `createCertificate.sh` script, which uses openssl to create the required certificate and key pair
for the `mycompany.tld` domain (`*.mycompany.tld`).

```shell
#!/bin/bash

set -e
caName="MyOrg-RootCA"            # CA name (CN).
publicDomain="mycompany.tld"     # Cluster domain name (see publicDomainTemplate).
certName="kubernetes"            # Certificate name for the cluster (CN).

mkdir -p "${caName}"
cd "${caName}"

[ ! -f "${caName}.key" ] && openssl genrsa -out "${caName}.key" 4096

[ ! -f "${caName}.crt" ] &&  openssl req -x509 -new -nodes -key "${caName}.key" -sha256 -days 1826 -out "${caName}.crt" \
   -subj "/CN=${caName}/O=MyOrganisation"

openssl req -new -nodes -out ${certName}.csr -newkey rsa:4096 -keyout "${certName}.key" \
  -subj "/CN=${certName}/O=MyOrganisation"

# v3 ext file
cat > "${certName}.v3.ext" << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${publicDomain}
DNS.2 = *.${publicDomain}
EOF

openssl x509 -req -in "${certName}.csr" -CA "${caName}.crt" -CAkey "${caName}.key" -CAcreateserial -out "${certName}.crt" -days 730 -sha256 -extfile "${certName}.v3.ext"

cat "${certName}.crt" "${caName}.crt" > "${certName}_fullchain.crt"
```

Using the resulting `kubernetes.key` and `kubernetes_fullchain.crt` files, create a secret in the `d8-system` namespace.

```shell
d8 k -n d8-system create secret tls mycompany-wildcard-tls --cert=kubernetes_fullchain.crt --key=kubernetes.key
```

To use the resulting certificate in the cluster, bring the global configuration to the following form by running `d8 k edit mc global`:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: global
spec:
  settings:
    modules:
      https:
        customCertificate:
          secretName: mycompany-wildcard-tls    # The name of the Secret object containing the fullchain certificate and key.
        mode: CustomCertificate                 # Change the TLS mode for all modules.
      publicDomainTemplate: '%s.mycompany.tld'
```

You also need to configure the `user-authn` module by setting `controlPlaneConfigurator.dexCAMode` to `FromIngressSecret`.
In this case, the CA is taken from the chain placed in the `kubernetes_fullchain.crt` file.

Example:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: user-authn
spec:
  enabled: true
  settings:
    controlPlaneConfigurator:
      dexCAMode: FromIngressSecret
  ...
```

Before starting the module, make sure the key services are reachable from the **working network**.

1. Get the address of the authentication platform:

    ```shell
    d8 k -n d8-user-authn get ing dex
    # Expected output:
    # NAME   CLASS   HOSTS               ADDRESS         PORTS     AGE
    # dex    nginx   dex.mycompany.tld   34.85.243.109   80, 443   4d20h
    ```

    The `HOSTS` column contains the domain to check, and the `ADDRESS` column contains its IP address. Now make sure that the domain name points to this IP address. To do this, run the command:

    ```shell
    nslookup dex.mycompany.tld
    # Expected output:
    # ...
    # Name: dex.mycompany.tld
    # Address: 34.85.243.109
    # ...

    # Or:
    dig dex.mycompany.tld
    # Expected output:
    # ...
    # ;; ANSWER SECTION:
    # dex.mycompany.tld. 3600 IN A 34.85.243.109
    # ...
    ```

    If the response is an `NXDOMAIN` error, configure the user's DNS. If the domain is not reachable from the Internet, perform an [additional step](#what-to-do-if-dexmycompanytld-does-not-resolve-via-dns).

    > As a temporary workaround, you can add the following line to the `/etc/hosts` file of your Unix system:
    >
    > ```shell
    > 34.85.243.109 dex.mycompany.tld stronghold.mycompany.tld
    > ```

1. Open <https://dex.mycompany.tld/healthz> in a browser or run `curl -kL https://dex.mycompany.tld/healthz`. The response must be `Health check passed`.
1. Check that the Ingress controller handles requests to your `stronghold.mycompany.tld` subdomain. Open <https://stronghold.mycompany.tld> in a browser or with `curl -kL`. A 404 error must be returned.

### What to do if dex.mycompany.tld does not resolve via DNS

If the domain does not resolve via DNS and you plan to use the `hosts` file, add the load balancer address or the frontend node IP address to the cluster DNS for Dex to work.
You can use the [`kube-dns` module](/modules/kube-dns/) for this, so that pods can reach the `dex.mycompany.tld` domain by name.

Example of getting the IP address for the `nginx-load-balancer` Ingress of the `LoadBalancer` type:

```shell
d8 k -n d8-ingress-nginx get svc nginx-load-balancer -o jsonpath='{ .spec.clusterIP }'
```

Suppose the address is `34.85.243.109`; then the `kube-dns` module configuration looks like this:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: kube-dns
spec:
  version: 1
  enabled: true
  settings:
    hosts:
    - domain: dex.mycompany.tld
      ip: 34.85.243.109
```

### Enabling the module

After organizing access, enable the `stronghold` module. Initialization and configuration of the integration with `dex` happen automatically.

```shell
d8 system module enable stronghold
```

After the module starts:

1. Make sure that a certificate for the stronghold.* domain exists by running `d8 k -n d8-stronghold get secret ingress-tls` (or via the console).
   The [Troubleshooting](#troubleshooting) section describes solutions to possible problems with a missing certificate.
1. Make sure that <https://stronghold.mycompany.tld/v1/sys/health> is reachable.
1. Check that the certificate issuer matches the CA certificate (optional).

### Troubleshooting

#### Pods are in the ContainerCreating state and there is no Secret named ingress-tls

Check the status of the Stronghold pod:

```shell
d8 k -n d8-stronghold describe pod stronghold-0
```

Find the line:

```log
MountVolume.SetUp failed for volume "certificates" : secret "ingress-tls" not found
```

When using the [ClusterIssuer with LetsEncrypt](#clusterissuer-letsencrypt) method, there may be a problem with the Let's Encrypt certificate authority automatically creating a certificate for the `stronghold.mycompany.tld` domain.

Get the list of CertificateRequest objects:

```bash
d8 k -n d8-stronghold get certificaterequest
```

Find the object whose name starts with **stronghold-**; in this example, it is **stronghold-b5wc6**.

Check its status:

```bash
d8 k -n d8-stronghold describe certificaterequest stronghold-b5wc6
```

One possible cause is the `too many certificates already issued for mycompany.tld` error, especially if a free dynDNS service is used, for example, `sslip.io` or `getmoss.site`.
In this case, either wait until the rate limit expires or use another way to create a certificate for the `stronghold.mycompany.tld` domain (the **selfsigned** ClusterIssuer or manual certificate signing).

When the certificate is generated successfully, the CertificateRequest status must contain the following lines:

```log
Message:               Certificate fetched from issuer successfully
Reason:                Issued
Status:                True
Type:                  Ready
```

A Secret object of the `kubernetes.io/tls` type named **ingress-tls** must also appear.
