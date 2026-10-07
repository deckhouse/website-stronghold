---
title: "Glossary"
description: "Short definitions of core Stronghold terms: seal, barrier, tokens, leases, policies, secrets engines, identity, replication, and operations."
weight: 900
---

Terms are grouped by topic. Each term links to a page with a detailed description.

## Storage and encryption

- **Barrier** — the cryptographic layer between the Stronghold core and the storage backend. All data written to storage is encrypted by the barrier and decrypted on read. See [Architecture overview](../../admin/architecture/overview/).
- **Keyring** — the set of barrier encryption keys (AES-256-GCM). The keyring is stored encrypted and is decrypted with the root key. See [Seal/Unseal](../seal/).
- **Root key** — the key that encrypts the keyring. The root key itself is protected by the seal mechanism: Shamir shares, HSM, or KMS. See [Seal/Unseal](../seal/).
- **Seal** — the state in which Stronghold knows where its data is but cannot decrypt it. A sealed Stronghold answers only a limited set of requests, such as status requests. See [Seal/Unseal](../seal/).
- **Unseal** — the process of obtaining the root key, after which Stronghold can decrypt the keyring and serve requests. See [Seal/Unseal](../seal/).
- **Unseal keys** — the shares into which the unseal key is split with the Shamir seal. A threshold number of shares is required to unseal. See [Seal/Unseal](../seal/).
- **Recovery keys** — key shares issued instead of unseal keys with auto-unseal. They do not unseal storage but are required for privileged operations such as generating a root token. See [Seal/Unseal](../seal/).
- **Auto-unseal** — a mode in which the root key is protected by an external or built-in mechanism (HSM, KMS, `inner-cluster`) and Stronghold unseals without manual share entry. See [Seal/Unseal](../seal/).
- **Seal wrap** — additional encryption of the most sensitive internal data by the seal mechanism on top of the standard barrier. See [Double encryption](../../admin/kms-hsm/sealwrap/).
- **HSM (Hardware Security Module)** — a hardware module that stores keys and performs cryptographic operations. Stronghold uses it to protect the root key (`seal "pkcs11"`). See [HSM support](../../admin/kms-hsm/hsm/).
- **KMS (Key Management Service)** — an external key management service, such as Yandex Cloud KMS, used for auto-unseal. See [Yandex Cloud KMS](../../admin/kms-hsm/yandexcloudkms/).
- **Managed keys** — keys stored in an external HSM or KMS and used by secrets engines such as PKI without exporting the private part to Stronghold. Available in Stronghold EE. See [Managed Keys in Stronghold](../../user/managed-keys/overview/).

## Tokens and leases

- **Token** — the primary way to authenticate to Stronghold. A token carries policies, a TTL, and metadata. See [Token](../tokens/).
- **Root token** — a token with the `root` policy that is allowed to perform any operation. Use it only for initial setup and emergencies. See [Token](../tokens/).
- **Accessor** — a reference to a token that lets you look up, renew, or revoke the token without knowing its value. See [Token](../tokens/).
- **Service token** — the default token type. It is persisted in Stronghold and supports renewal, revocation, and child tokens. See [Token](../tokens/).
- **Batch token** — an encrypted token that is not persisted in Stronghold. It cannot be renewed or manually revoked and has no accessor, but it scales well. See [Token](../tokens/).
- **Orphan token** — a token without a parent. It is not revoked when the token that created it is revoked. See [Token](../tokens/).
- **Periodic token** — a token without a maximum lifetime: on every renewal, its TTL is reset to the configured period. See [Token](../tokens/).
- **Token store** — the built-in `token` auth method that creates and stores tokens. It cannot be disabled. See [Token](../../user/auth/token/).
- **Lease** — metadata Stronghold creates for a dynamic secret or token: its duration, renewability, and ID. See [Lease](../lease/).
- **TTL (Time To Live)** — the lifetime of a lease or token after which it is revoked unless renewed. See [Lease](../lease/).
- **Max TTL** — the maximum time up to which a lease or token can be renewed. It is set at the system, mount, or role level. See [Lease](../lease/).
- **Explicit max TTL** — a hard limit on token lifetime that renewal cannot exceed. See [Token](../tokens/).
- **Response wrapping** — a mechanism in which Stronghold puts a response into the cubbyhole of a single-use wrapping token and returns only that token. See [Response wrapping](../response-wrapping/).

## Authentication and policies

- **Auth method** — a component that verifies a client (by a Kubernetes token, JWT, password, and so on) and issues it a token with policies. See [Authentication methods](../../user/auth/overview/).
- **AppRole** — an auth method for applications and automation based on a `role_id` and `secret_id` pair. See [AppRole method](../../user/auth/approle/).
- **role_id / secret_id** — the AppRole role identifier (not secret, similar to a username) and the secret identifier (similar to a password) used together to log in. See [AppRole method](../../user/auth/approle/).
- **Policy** — an HCL or JSON document that defines which operations are allowed on which paths. Access is denied by default. See [Policies](../policy/).
- **Capability** — an operation allowed on a path in a policy: `create`, `read`, `update`, `patch`, `delete`, `list`, `sudo`, or `deny`. See [Policies](../policy/).
- **Path templating** — substituting entity parameters, such as `{{identity.entity.id}}`, into a policy path so that one policy gives every client access to its own secrets. See [Policies](../policy/).
- **Password policy** — a set of rules for generating passwords for dynamic credentials. See [Password policy](../password-policy/).
- **MFA (multi-factor authentication)** — an additional check (for example, TOTP) at login on top of the primary auth method. See [2FA / MFA](../../user/auth/mfa/).

## Secrets engines

- **Secrets engine** — a component that stores, generates, or encrypts data. It is mounted at a path and defines its own API. See [Secrets engines](../../user/secrets-engines/overview/).
- **Mount** — the path at which a secrets engine or auth method is enabled, for example `kv/` or `auth/ldap/`. See [Secrets engines](../../user/secrets-engines/overview/).
- **Dynamic secret** — credentials that Stronghold creates on demand in an external system (for example, a database) and revokes automatically when the lease expires. See [Databases secrets engine](../../user/secrets-engines/databases/overview/).
- **Static role** — a mapping of a Stronghold role to an existing account in an external system whose password Stronghold stores and rotates periodically. See [Databases secrets engine](../../user/secrets-engines/databases/overview/).
- **Cubbyhole** — per-token private secret storage. The data is available only to that token and is deleted with it. See [Cubbyhole secrets engine](../../user/secrets-engines/cubbyhole/).
- **KV (Key/Value)** — a secrets engine for storing arbitrary static secrets. Version 2 supports versioning. See [KV secrets engine](../../user/secrets-engines/kv/overview/).
- **Plugin** — a built-in or external module that implements a secrets engine, an auth method, or a database plugin. See [Stronghold plugins](../../admin/plugins/overview/).

## Identity and namespaces

- **Entity** — the representation of a client (a person or a service) in Identity. It combines accounts from different auth methods. See [Identity](../identity/).
- **Alias** — the link between an entity or group and an account or group in a specific auth method. See [Identity](../identity/).
- **Group** — a set of entities with shared policies. An external group is mapped to a group in an external system through an alias. See [Identity](../identity/).
- **Namespace** — an isolated area of Stronghold with its own secrets engines, auth methods, policies, and tokens. Available in Stronghold EE and Stronghold CSE. See [Namespaces](../../admin/namespaces/overview/).

## Cluster, replication, and backups

- **Raft** — Stronghold's built-in integrated storage with consensus between cluster nodes. See [Architecture overview](../../admin/architecture/overview/).
- **Active node** — the HA cluster node that handles write requests. See [Architecture: Stronghold and Stronghold EE](../../admin/replication/architecture/).
- **Standby** — a non-leading HA cluster node that is ready to become the active node if the current one fails. See [Architecture: Stronghold and Stronghold EE](../../admin/replication/architecture/).
- **Performance standby** — horizontal read scaling inside a cluster: a Stronghold EE standby node serves read requests locally and forwards writes to the active node. See [Performance standby](../../admin/replication/performance-standby/).
- **Primary / secondary** — cluster roles in cross-cluster replication: the primary is the data source, and the secondary receives the data. See [Cross-cluster replication](../../admin/replication/overview/).
- **WAL (Write-Ahead Log)** — the change log through which the primary sends data to the secondary during replication. See [Architecture: Stronghold and Stronghold EE](../../admin/replication/architecture/).
- **Performance replication (PR)** — a replication type where a secondary cluster receives the primary's shared state, serves reads locally and forwards writes to the primary. See [Performance replication](../../admin/replication/performance/).
- **DR (Disaster Recovery)** — a replication type with a hot standby cluster that can be promoted to primary in a disaster. See [Disaster recovery](../../admin/replication/disaster-recovery/).
- **Promote / demote** — the operations of promoting a secondary cluster to primary and demoting a primary to secondary. See [Disaster recovery](../../admin/replication/disaster-recovery/).
- **Snapshot** — a backup of the Raft integrated storage data from which a cluster can be restored. See [Stronghold backups](../../admin/backups/overview/).

## Audit

- **Audit device** — a component that logs all Stronghold requests and responses to a file, syslog, or socket. Sensitive values in the log are hashed. Available in Stronghold EE and Stronghold CSE. See [Audit in Stronghold](../../admin/audit/overview/).
