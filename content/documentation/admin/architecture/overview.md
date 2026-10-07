---
title: "Architecture overview"
description: "The components of a Stronghold server, how a request travels from a client to storage, and what the cryptographic barrier protects."
weight: 10
---

Stronghold is a server with an HTTP API that stores data encrypted and returns it only to authenticated clients with matching policies. The Stronghold architecture is compatible with HashiCorp Vault: base Stronghold corresponds to Vault CE.

This page describes the components of a single server and the path a request takes through them. Stronghold EE node layers (WAL Backend, Sealwrap), the HA cluster, and cross-cluster replication are described in ["Architecture: Stronghold and Stronghold EE"](../../replication/architecture/).

## Server components

| Component | Purpose |
| --- | --- |
| HTTP API (listener) | Accepts requests from clients, the CLI, and the web UI over HTTPS. All Stronghold operations, including administration, go through the API. |
| Core | The server core: connects the other components, processes a request step by step, and manages the seal state and HA mode. |
| Token store | Stores tokens together with their policies, TTLs, and metadata. Checks the token of every request. See [Token](../../../concepts/tokens/). |
| Policy engine (ACL) | Matches the request path and operation against the token policies and allows or denies the request. See [Policies](../../../concepts/policy/). |
| Auth methods | Verify client credentials (LDAP, OIDC, JWT, Kubernetes, AppRole, userpass, and others) and issue a token. See [Authentication](../../../concepts/auth/). |
| Secrets engines | Store, generate, or process secrets: KV, PKI, Transit, SSH, databases, and others. Each engine is mounted at its own path (`mount`). |
| Router | Routes a request to an auth method or secrets engine by the request path. |
| Expiration manager | Tracks leases of tokens and dynamic secrets, renews and revokes them. See [Lease](../../../concepts/lease/). |
| Audit devices | Log every request and response (`file`, `syslog`, `socket`). See [Audit](../../audit/overview/). |
| Cryptographic barrier (Security Barrier) | Encrypts everything written to storage and decrypts everything read from it. Uses `AES-256-GCM`. |
| Storage backend | Persists the encrypted data. HA uses the integrated Raft storage. |
| Seal | Protects the root key that encrypts the barrier keyring: Shamir shares, HSM (`pkcs11`), Yandex Cloud KMS, or `inner-cluster`. See [Seal/Unseal](../../../concepts/seal/). |

## Request path

A request passes the server components in a fixed order. A simplified diagram for a secret read request:

![request_path.en.png](../../../images/request_path.en.png)

1. The client connects to the listener over TLS and passes a token in the request header.
1. The token store looks up the token and its policies. If the token does not exist, has expired, or has been revoked, the request is rejected. Login requests (`auth/<method>/login`) are made without a token: the auth method verifies the credentials and returns a new token in the response.
1. The policy engine checks whether the operation (`read`, `create`, `update`, `delete`, `list`, and others) is allowed on the request path. Everything is denied by default; an explicit `deny` takes precedence over grants.
1. Before execution, the request is logged to the audit devices. Sensitive values in the entry are hashed with HMAC-SHA256.
1. The router selects the secrets engine or auth method mounted at the path prefix.
1. The secrets engine performs the operation. If it issues a dynamic secret or a token, the expiration manager registers a lease.
1. All data the engine writes or reads passes through the barrier. Above the barrier data is in plaintext, below it data is only encrypted.
1. The storage backend receives and returns ciphertext only. In an HA cluster, Raft replicates it to the other nodes.
1. Before the response is sent to the client, it is also logged to the audit devices.

The request is written to the audit log after the token and policy checks, including for denied requests.

If a request reaches a standby node of an HA cluster, it is forwarded to the active node over the cluster port. In Stronghold EE, performance standby nodes serve reads locally. See ["Architecture: Stronghold and Stronghold EE"](../../replication/architecture/#nodes-in-a-cluster).

## Leases and the expiration manager

The expiration manager is a background core component that keeps track of leases:

- when a dynamic secret or a token is issued, it creates a lease with a TTL;
- it renews a lease at the client's request, if renewal is allowed;
- when the TTL expires, it revokes the lease, and the secrets engine deletes the issued credentials in the external system (for example, a database role);
- when a token is revoked, it revokes all leases created with that token and its child tokens.

Static KV secrets do not create leases. See [Lease](../../../concepts/lease/).

## Seal and barrier

At startup, the server is sealed: it can read storage but cannot decrypt the data. Only the status check and unsealing are available.

The key chain:

![key_chain.en.png](../../../images/key_chain.en.png)

The barrier protects:

- confidentiality of data in storage, on disk, and in backups (Raft snapshots);
- data integrity: `AES-GCM` detects tampering with the ciphertext;
- data in transit between nodes over Raft and in the replication stream: it is already encrypted by the barrier.

The barrier does not protect data above it: in the memory of an unsealed server, in responses to clients, or in the external systems themselves. These limits are covered in [Threat model](../threat-model/).

Barrier keys can be rotated: new entries are encrypted with the new key, and old entries stay readable. For algorithm details, see [Cryptography](../../cryptography/overview/).

Stronghold EE offers a second encryption layer for sensitive paths: [seal wrap](../../kms-hsm/sealwrap/).
