---
title: "Stronghold plugins"
linkTitle: "Introduction"
weight: 10
description: "Overview of the Stronghold plugin system and the differences between Standalone and DKP."
---

Stronghold supports three plugin types:

- `auth` for authentication methods;
- `secret` for secrets engines;
- `database` for database plugins.

Plugins can be:

- **built-in**, shipped with Stronghold and available without additional preparation;
- **external**, delivered separately and requiring operator actions before use.

External plugins run as separate processes and communicate with Stronghold over RPC. This isolates plugin execution from the main Stronghold process so a plugin failure should not crash the server itself.

## Important considerations

- A plugin is uniquely identified by its type, name, and optionally version.
- The same plugin can be mounted at multiple `mount path`s.
- For external plugins, Stronghold validates the plugin binary with SHA256.
- A plugin must be registered in the plugin catalog before it can be enabled.

## Standalone vs DKP

The loading flow for external plugins depends on how Stronghold is installed:

- in **Standalone**, the operator brings the plugin binary to the server, places it in `plugin_directory`, registers it, and then enables it;
- in **Deckhouse Kubernetes Platform**, the operator declares plugins in `ModuleConfig`, the platform downloads and validates them, places them into the Stronghold container, and the plugin is then registered and enabled in Stronghold.

## Section pages

- [Plugins in Standalone](./standalone/) for plugin directory setup, registration, and mounting on a Linux server.
- [Plugins in DKP](./dkp/) for plugin delivery through `ModuleConfig` and later registration in Stronghold.
