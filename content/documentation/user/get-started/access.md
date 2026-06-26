---
title: "Configuring access and first login"
linkTitle: "Configuring access"
description: "Getting access to Deckhouse Stronghold and logging in for the first time via the CLI or web interface"
weight: 10
---

To work with Deckhouse Stronghold, get the access parameters from the administrator. Log in from the command line with the `d8` utility or through the web interface if it is available in the installation.

Access parameters:

- the Stronghold server address;
- the authentication method and path;
- the web interface address, if you log in through a browser;
- additional connection parameters, if they are required in your installation.

## Log in via the CLI

To configure access to your project from the command line, follow these steps.

1. Install [the `d8` utility](/products/kubernetes-platform/documentation/v1/cli/d8).

1. Set the Stronghold server address in an environment variable:

   ```shell
   export STRONGHOLD_ADDR=https://stronghold.domain.my
   ```

   Replace `https://stronghold.domain.my` with the address you received from the administrator.

1. Log in through OIDC. To do this, authenticate to Stronghold using the following command:

   ```shell
   d8 stronghold login -path=oidc_deckhouse -method=oidc -no-print
   ```

   This example uses [OIDC authentication](../auth/oidc/overview/) with the `oidc` method and the `oidc_deckhouse` path. If your installation uses different parameters, specify the values provided by the administrator.

   After you run the command, the CLI usually opens a browser to complete login through the configured OIDC provider. If the browser does not open automatically, copy the URL from the command output and open it manually.

1. Verify access using the following command:

   ```shell
   d8 stronghold status
   ```

To work with Deckhouse Stronghold, use commands in the following format:

```shell
d8 stronghold <command>
```

## Log in through the web interface

If the Stronghold web interface is available in your installation, log in through a browser:

1. Open the Stronghold web interface address provided by the administrator.
1. Select a login method, for example OIDC.
1. If required, specify the role name or another parameter provided by the administrator.
1. Authenticate.

After a successful login, Stronghold opens the user interface with the corresponding access permissions.

After logging in, proceed to creating your first secret in the [Your first secret](../first-secret/) section.
