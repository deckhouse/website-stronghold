---
title: "API Proxy"
linkTitle: "API Proxy"
description: "Using Stronghold Agent as a local proxy for the Stronghold API."
weight: 80
---

API Proxy lets Stronghold Agent work as a local proxy for the Stronghold API. In this mode, the application calls the local Agent endpoint, and the Agent adds the authentication token and proxies the request to the Stronghold server.

This page describes how API Proxy works, when to use it, and how to configure Stronghold Agent for this mode.

## When to use API Proxy

Use API Proxy in the following cases:

- The application can already work with the Stronghold API.
- The authentication token must be passed centrally.
- The number of direct connections to the Stronghold server must be reduced.
- The application needs a local HTTP(S) endpoint for accessing Stronghold.

If the application needs secrets in files, use the `template` mode. If the application needs secrets in environment variables, use `env_template` together with `exec`.

## How API Proxy works

Stronghold Agent usually performs the following actions:

1. Authenticates to Stronghold.
1. Starts a local listener.
1. Accepts application requests on a local address.
1. Adds the authentication token to the outgoing request.
1. Proxies the request to the Stronghold server.
1. Returns the response to the application.

This mode lets the application use a local endpoint instead of connecting to Stronghold directly.

## Basic configuration

The following example shows a minimal API Proxy configuration:

```hcl
stronghold {
  address = "https://stronghold.example.com:8200"
}

auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
    }
  }
}

api_proxy {
  use_auto_auth_token = true
}

listener "tcp" {
  address = "127.0.0.1:8200"
  tls_disable = true
}
```

In this example, the Agent:

- authenticates through AppRole;
- starts a local listener on `127.0.0.1:8200`;
- uses the token obtained through Auto-Auth;
- proxies application requests to the Stronghold server.

## How the application uses the local endpoint

After the Agent starts, the application can call the local API Proxy, for example:

```shell
curl http://127.0.0.1:8200/v1/secret/data/myapp
```

In this case, the application sends the request to the local Agent address instead of directly to Stronghold.

## Configuration parameters

The following blocks and parameters are most often used for API Proxy:

| Block or parameter | Description |
| --- | --- |
| `api_proxy.use_auto_auth_token` | Uses the token obtained through Auto-Auth |
| `listener "tcp"` | Configures the local listener |
| `listener.address` | Sets the address and port of the local endpoint |
| `listener.tls_disable` | Disables TLS on the local listener |

## Limitations

Consider the following specifics of API Proxy mode:

- The application must be able to work with the Stronghold API.
- Local listener security depends on the selected address and TLS settings.
- If the listener binds only to `127.0.0.1`, access is limited to the local machine.
- API Proxy is not intended for passing secrets through environment variables or files.
- Do not combine `api_proxy` with `env_template` in the same configuration file.

{{< alert level="warning" >}}
If you disable TLS on the local listener, use an address that is available only locally, such as `127.0.0.1`.
{{< /alert >}}

## Verification

After configuring API Proxy, perform the following steps:

1. Start Stronghold Agent.
1. Make sure that the local listener is available.
1. Run a test request to the local endpoint.
1. Check that the request is successfully proxied to Stronghold.

Example verification:

```shell
systemctl start stronghold-agent
journalctl -u stronghold-agent -n 50
curl http://127.0.0.1:8200/v1/secret/data/myapp
```
