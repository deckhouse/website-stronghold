---
title: "Overview"
description: "Stronghold Agent: a client daemon that handles authentication, token management, and secret delivery for applications without code changes."
weight: 10
---

**Stronghold Agent** is a client daemon that simplifies integrating applications with Stronghold. It provides automatic authentication, token management, and secret delivery without changing the application code.

Many applications, especially legacy systems, have no built-in support for secret management systems.
Stronghold Agent solves this problem by acting as an intermediary between the application and the Stronghold server:

- **Automatic authentication**: Agent authenticates on its own and renews tokens.
- **Secret delivery**: secrets are delivered to files or environment variables.
- **Automatic updates**: secrets are updated without restarting the application (if the application supports re-reading configuration files or environment variables).
- **Simpler integration**: no changes to the application code are required.

{{< alert level="info" >}}
Secret delivery to applications is covered in detail [in the "Deckhouse Stronghold capabilities overview" course](https://education.flant.ru/course/obzor-vozmozhnostej-deckhouse-stronghold/) (in Russian).
{{< /alert >}}

![agent](../../../../images/agent.png)

## Usage examples

Ready-made examples that use this feature:

- [Application on a virtual machine with Stronghold Agent](../../../examples/delivery/legacy-app-on-vm/)
- [TLS certificate for a web server on a VM from Stronghold PKI](../../../examples/delivery/web-server-tls/)
- [Delivering secrets to Kubernetes pods](../../../examples/delivery/kubernetes-workloads/)

See all examples in [Usage examples](../../../examples/).
