---
title: "Your first secret"
linkTitle: "Your first secret"
description: "The first practical scenario for working with a secret in Deckhouse Stronghold"
weight: 30
---

After you sign in to Deckhouse Stronghold successfully, you can proceed to the first practical operation — writing and reading a secret.
This example uses the `KV` secrets engine.

{{< alert level="info" >}}
The `KV` secrets engine can operate in two modes: `KV v1` — without versioning, and `KV v2` — with versioning and additional features. This section uses a basic scenario. For detailed information about the differences between `KV v1` and `KV v2`, refer to the dedicated documentation section [KV secrets engines](../secrets-engines/kv/overview/).
{{< /alert >}}

## Prerequisites

Before running the examples, make sure that:

- you are already signed in to Stronghold.
- you have read and write permissions for the selected path.
- the administrator has granted access to the `KV` storage.

If you have not signed in to Stronghold yet, complete the steps in the [First sign-in](../first-login) section first.

1. Select a path for the secret

    Secrets in the `KV` storage are written to a path.
    For the first example, use the following test path:

    ```text
    secret/my-first-secret
    ```

    If your installation uses a different mount path, use the value agreed on with the administrator.

1. Write the secret

    Save a test secret by using the following command:

    ```shell
    d8 stronghold kv put secret/my-first-secret username=demo password=secret123
    ```

    If the operation completes successfully, Stronghold confirms that the secret has been written.

1. Read the secret

    Read the saved secret:

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

    If you need to change the secret value, write it again to the same path:

    ```shell
    d8 stronghold kv put secret/my-first-secret username=demo password=new-secret
    ```

    Then read the secret again:

    ```shell
    d8 stronghold kv get secret/my-first-secret
    ```

1. Delete the test secret

    After you finish verification, delete the test secret:

    ```shell
    d8 stronghold kv delete secret/my-first-secret
    ```

{{< alert level="info" >}}
Deletion behavior depends on the mode in use: `KV v1` or `KV v2`. In `KV v2`, deletion can be soft and use versioning. For details, refer to the dedicated sections about `KV`.
{{< /alert >}}

## What happens when you work with a secret

In the basic scenario, Stronghold lets you:

- write values to the selected path.
- read saved data.
- update values.
- delete the secret.

This is usually how practical work with the `KV` secrets engine begins.

## If the command does not work

If an error occurs when writing or reading a secret, check the following:

- whether you signed in successfully.
- whether the token is valid.
- whether you have permissions for the selected path.
- whether the mount path is specified correctly.
- whether the corresponding secrets engine is enabled.

If the problem persists, contact the Stronghold administrator.
