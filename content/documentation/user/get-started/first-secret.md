---
title: "Your first secret"
linkTitle: "Your first secret"
description: "Creating your first secret in Deckhouse Stronghold"
weight: 30
---

In Deckhouse Stronghold, you can create a test secret, verify its contents, and change its value if needed.

This example uses:

- the [`kv` secrets engine](../../secrets-engines/kv/overview/);
- the `secret/my-first-secret` test path;
- the `username` and `password` test keys.

{{< alert level="info" >}}
Use only test values in this example. Do not store real passwords, tokens, or keys in test secrets.
{{< /alert >}}

## Create a secret

1. Create a secret by writing test values to the selected path:

    ```shell
    d8 stronghold kv put secret/my-first-secret username=demo password=secret123
    ```

    If the operation completes successfully, Stronghold confirms that the secret has been saved.

1. Verify the saved values:

    ```shell
    d8 stronghold kv get secret/my-first-secret
    ```

    Example output:

    ```text
    ====== Data ======
    Key         Value
    ---         -----
    password    secret123
    username    demo
    ```

1. Write the secret to the same path with a new value:

    ```shell
    d8 stronghold kv put secret/my-first-secret username=demo password=new-secret
    ```

    Verify the changed values using the following command:

    ```shell
    d8 stronghold kv get secret/my-first-secret
    ```

{{< alert level="info" >}}
After verification, replace the test path and values with the parameters used in your project.
{{< /alert >}}
