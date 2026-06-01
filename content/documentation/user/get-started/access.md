---
title: "Configuring project access"
linkTitle: "Configuring project access"
description: "Getting access to Deckhouse Stronghold for working via the CLI"
weight: 10
---

To work with Deckhouse Stronghold from the command line (CLI), you need to obtain the access parameters from the administrator, configure the working environment, and authenticate via OIDC using the `d8` utility.

## What to get from the administrator

Before you begin, make sure the administrator has provided the following:

- the Stronghold server address;
- the authentication method;
- confirmation that your account has the required permissions;
- additional connection parameters, if they are required in your installation.

If your organization uses a corporate identity provider, it is usually enough to get access to the required group or role.

## Prepare CLI access

To configure access to your project from the command line, follow these steps:

1. Install the `d8` utility.

   Install [the `d8` utility](/products/kubernetes-platform/documentation/v1/cli/d8).

1. Specify the Stronghold address.

   Set the Stronghold server address in an environment variable:

   ```shell
   export STRONGHOLD_ADDR=https://stronghold.domain.my
   ```

   Replace `https://stronghold.domain.my` with the address you received from the administrator.

1. Log in.

   Log in to Stronghold using the following command:

   ```shell
   d8 stronghold login -path=oidc_deckhouse -method=oidc -no-print
   ```

   This example uses the `oidc` authentication method and the `oidc_deckhouse` path. If your installation uses different parameters, specify the values provided by the administrator.

1. Use `d8 stronghold` commands.

   After logging in, use commands in the following format:

   ```shell
   d8 stronghold <command>
   ```

   The next step is described in the [First login](../first-login/) section.

## How to verify that access is configured correctly

After logging in, run a simple command that does not modify data:

```shell
d8 stronghold status
```

If the command runs successfully, it means that:

- the Stronghold address is specified correctly;
- authentication completed successfully;
- CLI access works.

## If login fails

If you cannot access Stronghold, check the following:

- whether `STRONGHOLD_ADDR` is specified correctly;
- whether the Stronghold address is reachable from your network;
- whether you are using the correct authentication method and path;
- whether your account has the required permissions.

If the problem persists, contact your Stronghold administrator.
