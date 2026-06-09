---
title: "Configuring access and first login"
linkTitle: "Configuring project access"
description: "Getting access to Deckhouse Stronghold and logging in for the first time via the CLI or web interface"
weight: 10
---

To work with Deckhouse Stronghold, you need to obtain the access parameters from the administrator and log in for the first time. Users typically log in from the command line with the `d8` utility or through the web interface if it is available in the installation.

## What to get from the administrator

Before you begin, make sure the administrator has provided the following:

- the Stronghold server address;
- the authentication method;
- the authentication path, if it differs from the default one;
- the web interface address, if you log in through a browser;
- additional connection parameters, if they are required in your installation.

If your organization uses a corporate identity provider, it is usually enough to get access to the required group or role.

## Log in via the CLI

To configure access to your project from the command line and log in to Stronghold, follow these steps.

1. Install the `d8` utility.

   Install [the `d8` utility](/products/kubernetes-platform/documentation/v1/cli/d8).

1. Specify the Stronghold address.

   Set the Stronghold server address in an environment variable:

   ```shell
   export STRONGHOLD_ADDR=https://stronghold.domain.my
   ```

   Replace `https://stronghold.domain.my` with the address you received from the administrator.

1. Log in through OIDC.

   In a typical OIDC scenario, log in to Stronghold using the following command:

   ```shell
   d8 stronghold login -path=oidc_deckhouse -method=oidc -no-print
   ```

   This example uses [OIDC authentication](../auth/oidc/overview/) with the `oidc` method and the `oidc_deckhouse` path. If your installation uses different parameters, specify the values provided by the administrator.

   After you run the command, the CLI usually opens a browser to complete login through the configured OIDC provider. If the browser does not open automatically, copy the URL from the command output and open it manually.

1. Verify access.

   Run a command that does not modify data:

   ```shell
   d8 stronghold status
   ```

   If the command runs successfully, the Stronghold address is specified correctly, authentication has completed, and CLI access works.

After a successful login, use commands in the following format:

```shell
d8 stronghold <command>
```

## Log in through the web interface

If the Stronghold web interface is available in your installation, log in through a browser:

1. Open the Stronghold web interface address provided by the administrator.
1. Select a login method, for example OIDC.
1. If required, specify the role name or another parameter provided by the administrator.
1. Continue to the external authentication provider and complete login.

After a successful login, Stronghold opens the user interface with the permissions that correspond to your account and assigned policies.

After logging in, proceed to creating your first secret in the [Your first secret](../first-secret/) section.

## Troubleshoot login issues

If you cannot log in to Stronghold, follow these steps:

- Check the `STRONGHOLD_ADDR` environment variable.
- Make sure the Stronghold address is reachable from your network.
- Compare the authentication method and path with the values provided by the administrator.
- Check whether the external OIDC provider is reachable.
- Make sure the browser does not block the redirect URI from opening.
- Check the web interface address, if you log in through a browser.
- Ask the administrator to confirm that your account has access permissions.

If the problem persists, contact your Stronghold administrator.
