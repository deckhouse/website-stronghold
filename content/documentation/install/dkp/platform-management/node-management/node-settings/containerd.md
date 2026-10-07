---
title: "Containerd configuration"
description: "Additional containerd configuration on Deckhouse Kubernetes Platform nodes using NodeGroupConfiguration: metrics, private registries, and registry certificates."
weight: 30
---

## Overview

You can configure containerd additionally by creating configuration files using the NodeGroupConfiguration resource.

The containerd settings are handled by the built-in [`032_configure_containerd.sh`](https://github.com/deckhouse/deckhouse/blob/main/candi/bashible/common-steps/all/032_configure_containerd.sh.tpl) script. It merges all `containerd` service configuration files located at `/etc/containerd/conf.d/*.toml` and restarts the service.

When developing a `NodeGroupConfiguration`, keep in mind the following:

1. The `/etc/containerd/conf.d/` directory is not created automatically;
1. Files in this directory must be created before `032_configure_containerd.sh` runs, that is, with a priority lower than `32`.

## Additional containerd settings

{{< alert level="danger" >}}
Adding custom settings causes the `containerd` service to restart.
{{< /alert >}}

{{< alert level="warning" >}}
You can override the values of parameters defined in the `/etc/containerd/deckhouse.toml` file, but you are then responsible for their correct operation. Avoid making changes that may affect master nodes.
{{< /alert >}}

## Enabling containerd metrics

The simplest example of adding `containerd` settings is enabling metrics.

Note the following:

1. The script creates a directory for configuration files.
1. The script creates a file in the `/etc/containerd/conf.d` directory.
1. The script has priority 31 (`weight: 31`).
1. The configuration is not changed on master nodes, only on nodes of the `worker` group.
1. Metrics collection must be configured separately; this only enables the metrics.
1. The script uses the bashbooster [bb-sync-file](http://www.bashbooster.net/#sync) construct to synchronize the file contents.

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: containerd-enable-metrics.sh
spec:
  bundles:
    - '*'
  content: |
    mkdir -p /etc/containerd/conf.d
    bb-sync-file /etc/containerd/conf.d/metrics_settings.toml - << EOF
    [metrics]
    address = "127.0.0.1"
    grpc_histogram = true
    EOF
  nodeGroups:
    - "worker"
  weight: 31
```

## Adding a private registry with authorization

Running your own applications may require a private registry that may be protected by authorization. `containerd` allows you to configure registry settings using the `plugins."io.containerd.grpc.v1.cri".registry` parameter.

Authorization data is specified in the `auth` parameter in the docker registry auth format as a Base64 string. You can get the string with the following command:

```shell
d8 k create secret docker-registry my-secret --dry-run=client --docker-username=User --docker-password=password --docker-server=private.registry.example -o jsonpath="{ .data['\.dockerconfigjson'] }"
eyJhdXRocyI6eyJwcml2YXRlLnJlZ2lzdHJ5LmV4YW1wbGUiOnsidXNlcm5hbWUiOiJVc2VyIiwicGFzc3dvcmQiOiJwYXNzd29yZCIsImF1dGgiOiJWWE5sY2pwd1lYTnpkMjl5WkE9PSJ9fX0=
```

The NodeGroupConfiguration resource looks like this:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: containerd-additional-config.sh
spec:
  bundles:
    - '*'
  content: |

    REGISTRY_URL=private.registry.example

    mkdir -p /etc/containerd/conf.d
    bb-sync-file /etc/containerd/conf.d/additional_registry.toml - << EOF
    [plugins]
      [plugins."io.containerd.grpc.v1.cri"]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."${REGISTRY_URL}"]
          endpoint = ["https://${REGISTRY_URL}"]
        [plugins."io.containerd.grpc.v1.cri".registry.configs."${REGISTRY_URL}".auth]
          auth = "eyJhdXRocyI6eyJwcml2YXRlLnJlZ2lzdHJ5LmV4YW1wbGUiOnsidXNlcm5hbWUiOiJVc2VyIiwicGFzc3dvcmQiOiJwYXNzd29yZCIsImF1dGgiOiJWWE5sY2pwd1lYTnpkMjl5WkE9PSJ9fX0="
    EOF
  nodeGroups:
    - "*"
  weight: 31
```

## Adding a certificate for an additional registry

<span id="ca-certificate-for-an-additional-registry"></span>

A private registry may require a root certificate. Add it to the `/var/lib/containerd/certs` directory and specify it in the tls parameter of the containerd settings.

You can base such a script on the [instructions for adding a root certificate to the OS](./os/#adding-a-root-certificate). Note the differences:

1. The priority value is 31;
1. The root certificate is added to the `/var/lib/containerd/certs` directory;
1. The certificate path is added to the `plugins."io.containerd.grpc.v1.cri".registry.configs."${REGISTRY_URL}".tls` settings section.

The script uses the following bashbooster constructs:

- [bb-sync-file](http://www.bashbooster.net/#sync) to synchronize the file contents.
- [bb-tmp-file](http://www.bashbooster.net/#tmp) to create temporary files and delete them after the script runs.

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: configure-cert-containerd.sh
spec:
  bundles:
  - '*'
  nodeGroups:
  - '*'
  weight: 31
  content: |-
    REGISTRY_URL=private.registry.example
    CERT_FILE_NAME=${REGISTRY_URL}
    CERTS_FOLDER="/var/lib/containerd/certs/"
    CERT_CONTENT=$(cat <<"EOF"
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

    CONFIG_CONTENT=$(cat <<EOF
    [plugins]
      [plugins."io.containerd.grpc.v1.cri".registry.configs."${REGISTRY_URL}".tls]
        ca_file = "${CERTS_FOLDER}/${CERT_FILE_NAME}.crt"
    EOF
    )

    mkdir -p ${CERTS_FOLDER}
    mkdir -p /etc/containerd/conf.d


    CERT_TMP_FILE="$( bb-tmp-file )"
    echo -e "${CERT_CONTENT}" > "${CERT_TMP_FILE}"

    CONFIG_TMP_FILE="$( bb-tmp-file )"
    echo -e "${CONFIG_CONTENT}" > "${CONFIG_TMP_FILE}"

    # Ensure CA certificate file in the CERTS_FOLDER.
    bb-sync-file \
      "${CERTS_FOLDER}/${CERT_FILE_NAME}.crt" \
      ${CERT_TMP_FILE}

    # Ensure additional containerd configuration file.
    bb-sync-file \
      "/etc/containerd/conf.d/${REGISTRY_URL}.toml" \
      ${CONFIG_TMP_FILE}
```
