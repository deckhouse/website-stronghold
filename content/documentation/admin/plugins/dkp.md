---
title: "Plugins in DKP"
weight: 30
description: "Using Stronghold plugins in Deckhouse Kubernetes Platform."
---

In Deckhouse Kubernetes Platform, plugin delivery differs from a Standalone installation: the operator does not copy plugin binaries to the server manually, but declares the plugin list in `ModuleConfig`.

After that, the platform:

- downloads plugins from the specified URLs;
- verifies their checksums;
- places them into the Stronghold container;
- restarts Stronghold when the plugin list changes.

After delivery, the plugin still has to be registered in Stronghold and then enabled at the required path.

## Configure the plugin list

The list of downloadable plugins is configured in `ModuleConfig`.

Example:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: stronghold
spec:
  enabled: true
  version: 1
  settings:
    plugins:
      - name: "vault-plugin-secrets-github"
        url: "https://github.com/martinbaillie/vault-plugin-secrets-github/releases/download/v2.3.2/vault-plugin-secrets-github-linux-amd64"
        sha256: "72cb1f2775ee2abf12ffb725e469d0377fe7bbb93cd7aaa6921c141eddecab87"
      - name: "vault-plugin-auth-any"
        url: "https://plugins.example.local/myplugins/vault-plugin-auth-any-v1.0.0-linux-amd64"
        sha256: "c943b505b39b53e1f4cb07f2a3455b59eac523ebf600cb04813b9ad28a848b21"
        ignoreFailure: true
        insecureSkipVerify: false
        ca: |
          -----BEGIN CERTIFICATE-----
          MIIDDTCCAfWgAwIBAgIJAOb7PcmW8W9MMA0GCSqGSIb3DQEBCwUAMBQxEjAQBgNV
          BAMTCWxvY2FsaG9zdDAeFw0yNjA1MjAwMDAwMDBaFw0yNjA2MjAwMDAwMDBaMBQx
          EjAQBgNVBAMTCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoC
          ggEBAKHh4g5i1R+3+9XdG0RFLiX1x5T2PvQ92E/78vR6+Bn09+G0P+C6143+wLn
          j96/E8rHbHr4R6L0f62/OJZh8JnZ/qRqE1N8oNc06Vh9Y7X8EzF4nZ4KgX/3y6L
          vXD251Qm7g==
          -----END CERTIFICATE-----
```

## Download parameters

- `name`: plugin binary name;
- `url`: URL used to download the plugin;
- `sha256`: SHA256 checksum of the plugin;
- `ignoreFailure`: allows Stronghold startup to continue even if this plugin could not be downloaded;
- `insecureSkipVerify`: disables TLS certificate verification for the remote server;
- `ca`: additional CA certificate used to validate TLS.

## What happens when the plugin list changes

Adding or removing plugins triggers a Stronghold restart.

If a plugin cannot be downloaded or validated:

- Stronghold startup is blocked;
- plugins with `ignoreFailure: true` are the exception;
- if checksum validation fails, the plugin is treated as not loaded and is removed.

## Air-gapped environments

In air-gapped environments where Stronghold does not have outbound internet access, you can host the plugin inside the DKP cluster itself.

One practical approach is:

1. Build or prepare an `nginx` container image that includes the plugin binary.
2. Run that container in Kubernetes.
3. Expose it through an internal Kubernetes `Service`.
4. Point `ModuleConfig` to a URL such as `http://<service>.<namespace>.svc.cluster.local/...` so the platform downloads the plugin through the internal Kubernetes service.

Example `ModuleConfig` fragment:

```yaml
spec:
  settings:
    plugins:
      - name: "vault-plugin-auth-any"
        url: "http://plugin-repo.plugins.svc.cluster.local/vault-plugin-auth-any"
        sha256: "c943b505b39b53e1f4cb07f2a3455b59eac523ebf600cb04813b9ad28a848b21"
```

This approach lets you:

- avoid opening outbound internet access;
- keep plugin binaries inside the internal environment;
- manage plugin versions centrally through an internal service.

## Register a plugin

After the plugin is delivered into the container, register it through the CLI:

```bash
PLUGIN_SHA=$(sha256sum <plugin_binary> | awk '{print $1;}')

d8 stronghold plugin register \
  -command <command_to_run_plugin_binary> \
  -sha256 "${PLUGIN_SHA}" \
  -version "<semantic_version>" \
  <plugin_type> \
  <plugin_name>
```

Example: register the secret plugin `mykv`:

```bash
d8 stronghold plugin register \
  -command mykvplugin \
  -sha256 "${PLUGIN_SHA}" \
  -version "v1.0.1" \
  secret \
  mykv
```

## Enable a plugin

After registration, enable the plugin as a `secret` or `auth` engine:

```bash
d8 stronghold <secrets|auth> enable \
  -path <mount_path> \
  <plugin_name>
```

Meaning:

- `secrets`: for `secret`-type plugins;
- `auth`: for authentication plugins;
- `-path`: mount path;
- `plugin_name`: the registered plugin name.

Example:

```bash
d8 stronghold secrets enable -path test-kv mykv
```

## Disable and remove a plugin

1. Disable all `secret` and `auth` methods that use the plugin.
2. Deregister the plugin:

```bash
d8 stronghold plugin deregister secret my-custom-plugin
```

3. Remove the plugin from `ModuleConfig`.

## Practical recommendations

- For production, prefer an internal plugin repository or trusted artifact storage.
- Always define `sha256` and verify it against the actual binary.
- Use `ignoreFailure` only for non-critical plugins.
- Remember that changing the plugin list restarts Stronghold.

## See also

- [Plugins in Standalone](./standalone/)
