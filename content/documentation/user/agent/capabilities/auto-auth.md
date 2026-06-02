---
title: "Auto-Auth"
linkTitle: "Auto-Auth"
description: "Automatic authentication for Stronghold Agent."
weight: 40
---

Auto-Auth automates token acquisition for Stronghold Agent. The Agent authenticates to Stronghold on its own, obtains a token, and uses it for requests to the Stronghold API.

This page describes how Auto-Auth works, what tasks it solves, and which authentication methods Stronghold Agent supports.

## How Auto-Auth works

Auto-Auth performs the following tasks:

- Authenticates to Stronghold without application involvement.
- Obtains a token and passes it to internal Agent subsystems.
- Writes the token to an external sink when needed.
- Renews the token before it expires.
- Authenticates again if the token can no longer be renewed.

A typical flow is as follows:

1. Stronghold Agent starts with a configured authentication method.
1. The Agent authenticates to Stronghold.
1. The Agent obtains a token.
1. The Agent uses the token for template rendering, Process Supervisor, or API Proxy.
1. If a sink is configured, the Agent writes the token to an external file.
1. The Agent renews the token before its TTL expires.
1. If renewal is unavailable, the Agent authenticates again.

## Why use Auto-Auth

Auto-Auth is useful in the following cases:

- The application must not handle authentication directly.
- A token must be obtained automatically after the Agent starts.
- The application must keep running without manual token renewal.
- The same token must be used by several Agent modes.

## Sink

A sink is a destination where Stronghold Agent writes the obtained token.

If a sink is configured, the token can be used by an external process. For example, it can be written to `/var/run/stronghold-agent/token`.

If no sink is configured, the token is used only inside the Agent process.

Example sink configuration:

```hcl
auto_auth {
  method "approle" {
    mount_path = "auth/approle"
    config = {
      role_id_file_path = "/etc/stronghold-agent/role-id"
      secret_id_file_path = "/etc/stronghold-agent/secret-id"
    }
  }

  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
}
```

## Supported authentication methods

Stronghold Agent supports the following authentication methods:

| Method | When to use | Details |
| --- | --- | --- |
| AppRole | For applications on virtual machines and bare metal | [AppRole](../approle/) |
| Token | For simple or temporary scenarios | [Token authentication](../auth-token/) |
| JWT/OIDC | For integration with an external identity provider | [JWT/OIDC](../auth-jwt-oidc/) |
| Cloud provider methods | For environments with a native authentication mechanism | — |

## Token renewal

If Stronghold Agent obtains a token with a limited TTL, it tries to renew it in advance.

This helps:

- reduce repeated authentication attempts;
- reduce load on the Stronghold server;
- keep the application running continuously.

If a token cannot be renewed, the Agent authenticates again and obtains a new token.

## Dynamic secret renewal

Auto-Auth is related not only to token acquisition, but also to continuous Agent operation in general. If the Agent uses dynamic credentials, it can renew them before they expire.

This is especially important in the following cases:

- temporary database credentials are used;
- certificates have a limited lifetime;
- the application is long-running and must not lose access to secrets.

After a secret is renewed, the Agent can:

- render a file again;
- restart a child process;
- continue proxying requests with the current token.

## Basic configuration example

The following example shows a minimal Auto-Auth configuration with AppRole and a file sink:

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
      remove_secret_id_file_after_reading = true
    }
  }

  sink "file" {
    config = {
      path = "/var/run/stronghold-agent/token"
      mode = 0640
    }
  }
}
```

## Choosing an authentication method

Choose an authentication method based on the runtime environment and security requirements:

- Use AppRole for virtual machines and bare metal.
- Use token only for simple, test, or temporary scenarios.
- Use JWT/OIDC if you need to integrate with an external identity provider.
- Use a cloud provider mechanism if Stronghold is deployed in the corresponding cloud environment.

## Limitations

Consider the following Auto-Auth specifics:

- Auto-Auth is responsible only for obtaining and renewing a token.
- Specific requirements depend on the selected authentication method.
- Not all methods are equally suitable for production environments.
- If a token is written to an external file, restrict access to that file.

{{< alert level="info" >}}
For production environments, AppRole is usually recommended for virtual machines and bare metal, and JWT/OIDC is usually recommended for integration with an external identity provider.
{{< /alert >}}
