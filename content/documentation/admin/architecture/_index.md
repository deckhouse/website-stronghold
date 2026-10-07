---
title: "Architecture"
description: "Stronghold components, the request path, DKP and standalone deployment layouts, network ports, and the threat model."
weight: 65
---

This section describes what Stronghold consists of and how it is deployed:

- [Architecture overview](./overview/): server components, the request path from a client to storage, the seal, and the cryptographic barrier.
- [Deployment in DP](./deployment-dkp/): how the `stronghold` module is placed in a Deckhouse Platform cluster.
- [Standalone deployment](./deployment-standalone/): an HA cluster of several servers with Raft storage.
- [Network ports](./ports/): which connections to allow between clients, nodes, and external systems.
- [Threat model](./threat-model/): trust boundaries, what the barrier, seal, HSM, and seal wrap protect against, and what they do not.
- [Functional specifications](./functional-specifications/): the main Stronghold features.

Node layers of base Stronghold and Stronghold EE, the HA cluster, performance standby, and cross-cluster replication are described in ["Architecture: Stronghold and Stronghold EE"](../replication/architecture/).
