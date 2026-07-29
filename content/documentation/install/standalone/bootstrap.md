---
title: "Bootstrap"
description: "Generate Stronghold deployment artifacts with stronghold bootstrap: systemd setup, Helm chart, and Docker image."
weight: 10
---

The `stronghold bootstrap` command group generates deployment artifacts for Stronghold. It is available starting from version `v1.19`.

Subcommands:

| Subcommand | Purpose |
| --- | --- |
| `service` | Generate a shell script that prepares a single-node Linux server (directories, config, TLS, systemd, user, binary) |
| `helm` | Write the Stronghold Helm chart archive embedded in the binary to disk |
| `docker` | Build an OCI-compatible Docker image tarball with the Stronghold binary |

```shell
stronghold bootstrap -help
```

## Prepare a Linux server

`stronghold bootstrap service` generates a shell script that prepares a single-node Stronghold environment: directories under `/opt/stronghold`, a configuration file with Raft storage, optional self-signed TLS certificates, a systemd unit, a system user, and binary installation.

{{< alert level="warning" >}}
Self-signed TLS certificates are suitable for testing only.
For production, provide your own certificates and use the `-no-tls` flag.
{{< /alert >}}

By default, the script is written to stdout so you can review it before applying:

```shell
stronghold bootstrap service | less
```

Apply the script:

```shell
stronghold bootstrap service | sudo sh
```

or run the installation directly:

```shell
sudo stronghold bootstrap service -apply
```

After the service starts, initialize and unseal Stronghold as described in [Installation](../installation/#quick-installation).

For HA cluster deployment, you can prepare the base environment on each node with `-no-tls`, then place your own certificates and configuration. See [Deploying a cluster in HA mode](../installation/#deploying-a-cluster-in-ha-mode).

### Service options

For the full list of options, see the help:

```shell
stronghold bootstrap service -help
```

Frequently used options:

| Option | Purpose |
| --- | --- |
| `-apply` | Execute the generated script instead of printing it to stdout |
| `-base-dir` | Base installation directory (default: `/opt/stronghold`) |
| `-config` | Path to the Stronghold configuration file |
| `-data-dir` | Path to the Raft data directory |
| `-tls-dir` | Path to the TLS certificate directory |
| `-bin-path` | Destination path for the Stronghold binary |
| `-source-binary` | Source path of the Stronghold binary to install (defaults to the current executable) |
| `-systemd-unit` | Path to the systemd unit file (default: `/etc/systemd/system/stronghold.service`) |
| `-user` | System user to run the service (default: `stronghold`) |
| `-group` | System group to run the service (default: `stronghold`) |
| `-storage` | Storage backend type (only `raft` is supported) |
| `-node-id` | Raft node identifier (defaults to hostname) |
| `-listener-address` | TCP listener address in `host:port` form (default: `0.0.0.0:8200`) |
| `-api-addr` | API address advertised to clients (auto-detected if unset) |
| `-cluster-addr` | Cluster address for Raft (auto-detected if unset) |
| `-no-user-create` | Do not create the system user |
| `-no-systemd` | Do not create or enable a systemd unit |
| `-no-tls` | Do not generate TLS certificates |
| `-tls-disable` | Disable TLS in the listener configuration (development only) |
| `-force` | Overwrite existing files instead of failing |

## Extract the Helm chart

`stronghold bootstrap helm` writes the Stronghold Helm chart archive embedded in the binary to disk. Use the archive with `helm install` or `helm template`.

Prerequisites: Helm 3 and access to a Kubernetes cluster.

Write the chart archive to the current directory (the default file name matches the embedded chart version):

```shell
stronghold bootstrap helm
```

or set an explicit output path:

```shell
stronghold bootstrap helm -output stronghold.tgz
```

Render manifests without installing:

```shell
helm template stronghold stronghold.tgz > stronghold.yaml
```

Install the chart:

```shell
helm install stronghold stronghold.tgz
```

### Helm options

| Option | Purpose |
| --- | --- |
| `-output` | Output archive path (defaults to the embedded chart file name) |

```shell
stronghold bootstrap helm -help
```

## Build a Docker image

`stronghold bootstrap docker` builds a minimal OCI-compatible image tarball on disk. The current Stronghold binary is embedded in the image by default. Import the tarball with `docker load`.

Build the image tarball:

```shell
stronghold bootstrap docker
```

By default the output file is named `stronghold-v<version>.tar` in the current directory, and the image tag is `stronghold:<version>`.

Load and run the image:

```shell
docker load -i stronghold-v1.19.0.tar
docker run --rm -p 8200:8200 stronghold:1.19.0
```

{{< alert level="info" >}}
The default container command is `stronghold server -dev`.
For a production-like run, pass your own server arguments and mount a configuration file.
{{< /alert >}}

For a dynamically linked binary, embed extra shared libraries or plugins with `-extra-file` in `source=destination` form:

```shell
stronghold bootstrap docker \
  -extra-file /lib64/ld-linux-x86-64.so.2=/lib64/ld-linux-x86-64.so.2 \
  -extra-file /lib/x86_64-linux-gnu/libc.so.6=/lib/x86_64-linux-gnu/libc.so.6 \
  -extra-file /opt/aktivco/rutokenecp/amd64/librtpkcs11ecp.so=/lib/librtpkcs11ecp.so
```

### Docker options

| Option | Purpose |
| --- | --- |
| `-output` | Output tarball path |
| `-tag` | Docker image tag in `repo:ref` form |
| `-source-binary` | Source path of the Stronghold binary to embed (defaults to the current executable) |
| `-ca-certs` | Path to the CA certificates bundle to embed (default: `/etc/ssl/certs/ca-certificates.crt`) |
| `-extra-file` | Extra file to embed as `source=destination` (repeatable) |

```shell
stronghold bootstrap docker -help
```
