---
title: "Capabilities overview"
linkTitle: "Capabilities overview"
description: "Overview of Stronghold Agent capabilities."
weight: 20
---

Stronghold Agent lets applications retrieve secrets from Stronghold without integrating directly with the API. The Agent supports rendering secrets to files, passing secrets through environment variables, automatic authentication, token reuse, secret renewal, and local API proxying.

This page gives an overview of the main Stronghold Agent capabilities. Detailed instructions and examples are available on separate pages.

## What Stronghold Agent can do

Stronghold Agent helps solve the following tasks:

- retrieve secrets from Stronghold without direct API calls from the application;
- render secrets into configuration files;
- pass secrets to child processes through environment variables;
- authenticate to Stronghold automatically;
- store and reuse a token through a cache;
- renew tokens and dynamic credentials before they expire;
- proxy requests to the Stronghold API through a local endpoint.

## Main operating modes

Stronghold Agent supports the following operating modes:

| Mode | Purpose | When to use |
| --- | --- | --- |
| `template` | Renders secrets to a file | The application reads configuration from files |
| `env_template` + `exec` | Passes secrets through environment variables | The application reads configuration from environment variables |
| `api_proxy` | Proxies requests to the Stronghold API | The application can already work with the Stronghold API |

## Rendering secrets to files

The `template` mode creates configuration files populated with values from Stronghold. Use it for applications that read settings from files.

Typical use cases include the following:

- the application reads parameters from `.conf`, `.ini`, `.yaml`, or `.properties` files;
- certificates, keys, or other sensitive data must be passed through files;
- the application cannot work with the Stronghold API directly;
- dynamic credentials must be renewed regularly.

After a secret changes, the Agent renders the file again. If needed, it can run a command to reload the service.

For details, see [Templates and file rendering](../templating/).

## Passing secrets through environment variables

The `env_template` mode is used together with the `exec` block. In this mode, Stronghold Agent starts the application as a child process and passes secrets to it through environment variables.

Use this mode in the following cases:

- the application reads configuration from environment variables;
- secrets must not be written to disk;
- restarting the application on secret rotation is acceptable;
- dynamic credentials need to be handled more easily.

When a secret changes, the Agent generates new environment variable values and restarts the child process.

For details, see [Environment variables and Process Supervisor](../process-supervisior/).

## Auto-Auth

Auto-Auth automates token acquisition for Stronghold Agent. The Agent authenticates on its own, gets a token, uses it for requests to Stronghold, and renews it when needed.

If a sink is configured, the Agent can write the token to an external file. If no sink is configured, the token is used only inside the Agent process.

Stronghold Agent supports the following authentication methods:

- AppRole;
- token;
- JWT/OIDC;
- cloud provider methods.

For details, see [Auto-Auth](../auto-auth/).

## Caching and renewal

Stronghold Agent supports token caching and automatic secret renewal. This reduces load on the Stronghold server and lowers the number of repeated authentication attempts.

The Agent can:

- reuse an existing token;
- renew a token before its TTL expires;
- authenticate again if a token cannot be renewed;
- renew dynamic credentials before they expire.

These mechanisms help keep the application running without manual intervention.

## API Proxy

Stronghold Agent can work as a local proxy for the Stronghold API. In this mode, the application calls the local Agent endpoint instead of calling the Stronghold server directly.

This mode is useful in the following cases:

- the application can already work with the Stronghold API;
- the authentication token must be passed centrally;
- the number of direct connections to the Stronghold server must be reduced.

For details, see [API Proxy](../api-proxy/).

## Choosing an operating mode

Choose an operating mode based on how the application receives secrets:

- use `template` if the application needs configuration files;
- use `env_template` together with `exec` if the application needs environment variables;
- use `api_proxy` if the application is already integrated with the Stronghold API.
