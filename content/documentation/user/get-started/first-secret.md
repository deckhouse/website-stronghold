---
title: "Your first secret"
linkTitle: "Your first secret"
description: "Creating your first secret in Deckhouse Stronghold"
weight: 30
---

After you sign in to Deckhouse Stronghold successfully, you can create your first test secret, verify its contents, and change its value if needed.

This example uses the `KV` secrets engine. Before running the commands, make sure you have read and write permissions for the selected path and that the administrator has granted access to the `KV` storage.
If you have not signed in to Stronghold yet, complete the steps in the [Configuring project access](../access/) section first.

{{< alert level="warning" >}}
Use only test values in this example. Do not store real passwords, tokens, or keys in test secrets.
{{< /alert >}}

## Create a secret

The examples below use the test path `secret/my-first-secret` and the test keys `username` and `password`. If your installation uses a different mount path, replace `secret` with the path provided by the administrator.

1. Select a path for the secret

    Secrets in the `KV` storage are written to a path.
    For the first example, use the following test path:

    ```text
    secret/my-first-secret
    ```

    Stronghold stores a set of keys and values at this path.

1. Create the secret

    Write test values to the selected path:

    ```shell
    d8 stronghold kv put secret/my-first-secret username=demo password=secret123
    ```

    If the operation completes successfully, Stronghold confirms that the secret has been written.

1. Verify the secret

    Read the saved values:

    ```shell
    d8 stronghold kv get secret/my-first-secret
    ```

    As a result, you should see the saved values.
    Example output:

    ```text
    ====== Data ======
    Key         Value
    ---         -----
    password    secret123
    username    demo
    ```

1. Update the secret

    To change a value, write the secret to the same path again:

    ```shell
    d8 stronghold kv put secret/my-first-secret username=demo password=new-secret
    ```

    Then read the secret again and make sure the value has changed:

    ```shell
    d8 stronghold kv get secret/my-first-secret
    ```

{{< alert level="info" >}}
After verification, replace the test path and values with the parameters used in your project. Additional `KV` engine capabilities, including versioning behavior, are described in the [KV secrets engines](../../secrets-engines/kv/overview/) section.

If a command returns an access or path error, check the mount path and read/write permissions. Resolve login errors by following the [Configuring project access](../access/) guide.
{{< /alert >}}
