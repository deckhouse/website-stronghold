---
title: "Plugins in Standalone"
weight: 20
description: "Using external Stronghold plugins in a Standalone installation."
---

In a Standalone installation, external plugins are delivered to the server by the operator. After that, the plugin must be registered in the plugin catalog and only then enabled as a `secret`, `auth`, or `database` engine.

## How external plugins work

External plugins are separate applications that Stronghold starts and communicates with over RPC.

This means:

- the plugin process does not share memory with the main Stronghold process;
- the plugin only receives explicitly provided interfaces and arguments;
- a plugin crash should not crash the whole Stronghold server.

## Prerequisites

To use external plugins in a Standalone configuration, you must:

- set `api_addr` in the Stronghold configuration;
- configure `plugin_directory`;
- place plugin binaries in that directory;
- register the plugin in the plugin catalog.

## Configure `plugin_directory`

Stronghold uses the `plugin_directory` setting in the configuration file to define the directory from which external plugins may be started.

Example:

```hcl
api_addr = "https://stronghold.example.com:8200"

plugin_directory = "/opt/stronghold/plugins"
```

Important details:

- `plugin_directory` must not be a symbolic link;
- if `plugin_directory` is not set, external plugins cannot be used;
- the plugin binary must physically exist in that directory.

## Register an external plugin

Before enabling an external plugin, register it in the plugin catalog so Stronghold can verify its authenticity and integrity.

Example:

```shell
PLUGIN_SHA=$(sha256sum /opt/stronghold/plugins/passthrough-plugin | awk '{print $1}')

stronghold plugin register \
  -sha256="${PLUGIN_SHA}" \
  secret \
  passthrough-plugin
```

If the plugin uses versioning, you can specify the version explicitly:

```shell
stronghold plugin register \
  -sha256="${PLUGIN_SHA}" \
  -version="v1.0.0" \
  secret \
  passthrough-plugin
```

Stronghold supports registering multiple versions of the same plugin.

## Enable an external plugin

After registration, the plugin can be mounted at the desired `mount path`:

```shell
stronghold secrets enable -path=my-secrets passthrough-plugin
```

Example result:

```text
Success! Enabled the passthrough-plugin secrets engine at: my-secrets/
```

After that, the plugin appears in the list of enabled engines.

## Disable a plugin

Disabling an external plugin works the same way as for built-in engines:

```shell
stronghold secrets disable my-secrets
```

If you want to remove the external plugin from the catalog completely, deregister it after disabling:

```shell
stronghold plugin deregister secret passthrough-plugin
```

## Plugin versions

A plugin can optionally report its own semantic version. If it does, Stronghold can use that information during registration.

If you specify a version manually:

- it should match semantic versioning;
- it is recommended to use the `v` prefix, for example `v1.0.0`;
- the same plugin type and name can coexist in multiple versions in the catalog.

If no version is specified when mounting, Stronghold chooses the plugin with the following precedence:

1. A plugin registered without a version.
2. The plugin with the most recent semantic version.
3. The built-in plugin.

## File permissions and `mlock`

If Stronghold uses `mlock`, external plugin binaries may also need the capability to use `mlock`.

In such cases, you may need:

```shell
sudo setcap cap_ipc_lock=+ep /opt/stronghold/plugins/<plugin-binary>
```

Stronghold can also validate ownership and permissions of the plugin directory and plugin files when file permissions checks are enabled.

## Practical recommendations

- Keep external plugins in a dedicated directory with restricted access.
- Verify SHA256 before registration.
- Do not use symbolic links for `plugin_directory`.
- For production, validate plugin compatibility with the installed Stronghold version in advance.

## See also

- [Plugins in DKP](./dkp/)
- [Standalone configuration](../standalone/configuration/)
