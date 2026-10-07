---
title: "OS configuration"
description: "Configuring the operating system on Deckhouse Kubernetes Platform nodes with NodeGroupConfiguration: installing utilities, sysctl parameters, kernel versions, and root certificates."
weight: 20
---

## Installing the cert-manager kubectl plugin on master nodes

You can use NodeGroupConfiguration to install the required utilities on master nodes.

For example, you can install the cmctl utility from the cert-manager project. This command can also be used as a kubectl plugin.

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: kubectl-plugin-cert-manager.sh
spec:
  weight: 100
  bundles:
    - "*"
  nodeGroups:
    - "master"
  content: |
    # See https://github.com/cert-manager/cmctl/releases/tag/v2.1.0
    version=v2.1.1

    if [ -x /usr/local/bin/kubectl-cert_manager ]; then
      exit 0
    fi
    curl -L https://github.com/cert-manager/cmctl/releases/download/${version}/cmctl_linux_amd64.tar.gz -o - | tar zxf - cmctl
    mv cmctl /usr/local/bin
    ln -s /usr/local/bin/cmctl /usr/local/bin/kubectl-cert_manager
```

## Setting a sysctl parameter

Some tasks require changing sysctl parameters on nodes.

For example, applications that use mmapfs may require increasing the number of memory map areas a process is allowed to have. This number is set by the `vm.max_map_count` parameter and can be configured via NodeGroupConfiguration:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: sysctl-tune.sh
spec:
  weight: 100
  bundles:
  - "*"
  nodeGroups:
  - "worker"
  content: |
    sysctl -w vm.max_map_count=262144
```

## Installing a specific kernel version

Nodes may require a specific Linux kernel version, and NodeGroupConfiguration can help in this case. To simplify the script, it is better to use [bashbooster](http://www.bashbooster.net/) constructs.

Different operating systems require different operations to change the kernel version, so examples for Debian and CentOS are provided below.

Both examples use the `bb-deckhouse-get-disruptive-update-approval` construct, a Deckhouse extension of the bashbooster command set. This construct prevents the node from rebooting if the reboot must be approved by adding an annotation to the node.

In addition, the following bashbooster constructs are used:

- [bb-apt-install](http://www.bashbooster.net/#apt) to install an apt package and send the "bb-package-installed" event if the package was installed;
- [bb-yum-install](http://www.bashbooster.net/#yum) to install a yum package and send the "bb-package-installed" event if the package was installed;
- [bb-event-on](http://www.bashbooster.net/#event) to signal that a node reboot is required if the "bb-package-installed" event was sent;
- [bb-log-info](http://www.bashbooster.net/#log) for logging;
- [bb-flag-set](http://www.bashbooster.net/#flag) to signal that a node restart is required.

### For Debian-based distributions

Create a NodeGroupConfiguration resource, specifying the desired kernel version in the `desired_version` variable:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: install-kernel.sh
spec:
  bundles:
    - '*'
  nodeGroups:
    - '*'
  weight: 32
  content: |
    desired_version="5.15.0-53-generic"

    bb-event-on 'bb-package-installed' 'post-install'
    post-install() {
      bb-log-info "Setting reboot flag due to kernel was updated"
      bb-flag-set reboot
    }

    version_in_use="$(uname -r)"

    if [[ "$version_in_use" == "$desired_version" ]]; then
      exit 0
    fi

    bb-deckhouse-get-disruptive-update-approval
    bb-apt-install "linux-image-${desired_version}"
```

### For CentOS-based distributions

Create a NodeGroupConfiguration resource, specifying the desired kernel version in the `desired_version` variable:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: install-kernel.sh
spec:
  bundles:
    - '*'
  nodeGroups:
    - '*'
  weight: 32
  content: |
    desired_version="3.10.0-1160.42.2.el7.x86_64"

    bb-event-on 'bb-package-installed' 'post-install'
    post-install() {
      bb-log-info "Setting reboot flag due to kernel was updated"
      bb-flag-set reboot
    }

    version_in_use="$(uname -r)"

    if [[ "$version_in_use" == "$desired_version" ]]; then
      exit 0
    fi

    bb-deckhouse-get-disruptive-update-approval
    bb-yum-install "kernel-${desired_version}"
```

## Adding a root certificate

<span id="adding-a-ca-certificate"></span>

In some cases, an additional root certificate may be required, for example, to access internal resources of an organization. Adding root certificates can be implemented as a NodeGroupConfiguration.

{{< alert level="warning" >}}
This example is provided for Ubuntu.  
The method of updating the certificate store may differ depending on the OS.

When adapting the script for another OS, change the [`bundles`](/modules/node-manager/cr.html#nodegroupconfiguration-v1alpha1-spec-bundles) parameter.
{{< /alert >}}

The script uses the following bashbooster constructs:

- [bb-sync-file](http://www.bashbooster.net/#sync) to synchronize the file contents and send the "ca-file-updated" event if the file has changed;
- [bb-event-on](http://www.bashbooster.net/#event) to start updating certificates if the "ca-file-updated" event was sent;
- [bb-tmp-file](http://www.bashbooster.net/#tmp) to create temporary files and delete them after the script runs.

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: add-custom-ca.sh
spec:
  weight: 31
  nodeGroups:
  - '*'  
  bundles:
  - 'ubuntu-lts'
  content: |-
    CERT_FILE_NAME=example_ca
    CERTS_FOLDER="/usr/local/share/ca-certificates"
    CERT_CONTENT=$(cat <<EOF
    -----BEGIN CERTIFICATE-----
    MIIDSjCCAjKgAwIBAgIRAJ4RR/WDuAym7M11JA8W7D0wDQYJKoZIhvcNAQELBQAw
    JTEjMCEGA1UEAxMabmV4dXMuNTEuMjUwLjQxLjIuc3NsaXAuaW8wHhcNMjQwODAx
    MTAzMjA4WhcNMjQxMDMwMTAzMjA4WjAlMSMwIQYDVQQDExpuZXh1cy41MS4yNTAu
    NDEuMi5zc2xpcC5pbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAL1p
    WLPr2c4SZX/i4IS59Ly1USPjRE21G4pMYewUjkSXnYv7hUkHvbNL/P9dmGBm2Jsl
    WFlRZbzCv7+5/J+9mPVL2TdTbWuAcTUyaG5GZ/1w64AmAWxqGMFx4eyD1zo9eSmN
    G2jis8VofL9dWDfUYhRzJ90qKxgK6k7tfhL0pv7IHDbqf28fCEnkvxsA98lGkq3H
    fUfvHV6Oi8pcyPZ/c8ayIf4+JOnf7oW/TgWqI7x6R1CkdzwepJ8oU7PGc0ySUWaP
    G5bH3ofBavL0bNEsyScz4TFCJ9b4aO5GFAOmgjFMMUi9qXDH72sBSrgi08Dxmimg
    Hfs198SZr3br5GTJoAkCAwEAAaN1MHMwDgYDVR0PAQH/BAQDAgWgMAwGA1UdEwEB
    /wQCMAAwUwYDVR0RBEwwSoIPbmV4dXMuc3ZjLmxvY2FsghpuZXh1cy41MS4yNTAu
    NDEuMi5zc2xpcC5pb4IbZG9ja2VyLjUxLjI1MC40MS4yLnNzbGlwLmlvMA0GCSqG
    SIb3DQEBCwUAA4IBAQBvTjTTXWeWtfaUDrcp1YW1pKgZ7lTb27f3QCxukXpbC+wL
    dcb4EP/vDf+UqCogKl6rCEA0i23Dtn85KAE9PQZFfI5hLulptdOgUhO3Udluoy36
    D4WvUoCfgPgx12FrdanQBBja+oDsT1QeOpKwQJuwjpZcGfB2YZqhO0UcJpC8kxtU
    by3uoxJoveHPRlbM2+ACPBPlHu/yH7st24sr1CodJHNt6P8ugIBAZxi3/Hq0wj4K
    aaQzdGXeFckWaxIny7F1M3cIWEXWzhAFnoTgrwlklf7N7VWHPIvlIh1EYASsVYKn
    iATq8C7qhUOGsknDh3QSpOJeJmpcBwln11/9BGRP
    -----END CERTIFICATE-----
    EOF
    )

    bb-event-on "ca-file-updated" "update-certs"
    
    update-certs() {          # Function with commands for adding a certificate to the store
      update-ca-certificates
    }

    CERT_TMP_FILE="$( bb-tmp-file )"
    echo -e "${CERT_CONTENT}" > "${CERT_TMP_FILE}"  

    bb-sync-file \
      "${CERTS_FOLDER}/${CERT_FILE_NAME}.crt" \
      ${CERT_TMP_FILE} \
      ca-file-updated   
```

A root certificate for containerd is configured in a similar way; see the example in the [containerd settings](./containerd/#adding-a-certificate-for-an-additional-registry) section.
