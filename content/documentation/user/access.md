---
title: "Configuring Project Access"
linkTitle: "Configuring Project Access"
description: "Obtaining access to Deckhouse Stronghold for CLI usage"
weight: 10
---

To work with Deckhouse Stronghold from the command line (CLI), you must obtain access parameters from the administrator, configure your working environment, and perform authentication. The example below covers configuration via OIDC using the `d8` utility.

## What to obtain from the administrator

Before you begin, ensure the administrator has provided you with:

- the Stronghold server address;
- the authentication method;
- confirmation that your account has the necessary access permissions;
- additional connection parameters if required by your installation.

If your organization uses a corporate identity provider, obtaining access to the required group or role is typically sufficient.

## Preparing CLI access

To configure access to your project from the command line in Deckhouse Stronghold, follow these steps.

1. Install the [`d8` utility](/products/kubernetes-platform/documentation/v1/cli/d8).

{{< alert level="info" >}}
The `d8` utility is used to work with Stronghold via the CLI in scenarios where Stronghold is integrated with the Deckhouse ecosystem.
{{< /alert >}}

1. Set your Stronghold server address as an environment variable:

```bash
export STRONGHOLD_ADDR=https://stronghold.domain.my
```

Replace `https://stronghold.domain.my` with the address you received from the administrator.

1. Log in to Stronghold using the command:

```bash
d8 stronghold login -path=oidc_deckhouse -method=oidc -no-print
```

This example uses:

- `oidc` as the authentication method;
- `oidc_deckhouse` as the authentication path.

If your installation uses different authentication parameters, use the values provided by the administrator.

1. After a successful login, use the following command format to work with Stronghold objects:

```bash
d8 stronghold <command>
```

For example, to check the service status or work with secrets.

## How to verify that access is configured correctly

After logging in, run a simple command that does not modify data, for example:

```bash
d8 stronghold status
```

If the command runs successfully, it means that:

- the Stronghold address is specified correctly;
- authentication was successful;
- CLI access is working.

## If login fails

If you cannot access Stronghold, check:

- whether `STRONGHOLD_ADDR` is set correctly;
- whether the Stronghold address is reachable from your network;
- whether you are using the correct authentication method and path;
- whether your account has the necessary permissions;
- whether your authentication session has expired.

If the problem persists, contact your Stronghold administrator.
