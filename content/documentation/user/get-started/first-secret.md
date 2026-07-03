---
title: "Creating a test secret"
linkTitle: "Creating a test secret"
description: "Creating a test secret in Deckhouse Stronghold"
weight: 30
---

In this section, you can learn how to create a test secret in Deckhouse Stronghold, verify its contents, and change its value if needed.

This example uses:

- [`kv` secrets engine](../../secrets-engines/kv/overview/)
- `secret/my-first-secret` test path
- `username` and `password` test keys

If your installation uses a different mount path, replace `secret` with the path provided by the administrator.

{{< alert level="warning" >}}
Use only test values in this example. Do not store real passwords, tokens, or keys in test secrets.
{{< /alert >}}

To create a test secret, follow these steps:

1. Create a secret by writing test values to the selected path:

   

1. Verify the saved values:

   

1. Edit the secret value by rewriting it to the same path:

   

{{< alert level="info" >}}
After verification, replace the test path and values with the parameters used in your project.
{{< /alert >}}
