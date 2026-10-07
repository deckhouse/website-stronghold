---
title: "Threat model"
description: "Stronghold trust boundaries, which threats the barrier, seal, HSM, and seal wrap protect against, which threats remain out of scope, and how to reduce them."
weight: 40
---

This page describes what Stronghold trusts, what its mechanisms protect against, and which risks remain the administrator's responsibility. The threat model is compatible with the HashiCorp Vault model. The components are described in [Architecture overview](../overview/).

## Protected assets

- Secrets and keys stored and issued by secrets engines (KV, PKI, Transit, SSH, databases, and others).
- The barrier keyring and the root key.
- Unseal keys (Shamir shares), recovery keys, and root tokens.
- Client tokens and auth method credentials.
- Audit logs.

## Trust boundaries

![trust_boundaries.en.png](../../../images/trust_boundaries.en.png)

- **Clients and the network** are untrusted. Every request must pass TLS and the token and policy checks.
- **The storage backend** is untrusted: it sees only ciphertext. The same applies to Raft snapshots, Raft traffic between nodes, and the replication stream.
- **The Stronghold process** in the unsealed state is the trusted zone. Its memory holds the keyring, the root key, and plaintext data.
- **The seal** is an external trusted zone: Shamir share holders, HSM, or KMS protecting the root key.
- **The operating environment** (server, OS, hypervisor, and in DP the Kubernetes cluster and the nodes holding Stronghold data, which are master nodes by default) is trusted. Compromising it equals compromising Stronghold.

## What Stronghold mechanisms protect against

| Threat | Mechanism | Result |
| --- | --- | --- |
| Theft of a disk, data directory, or Raft snapshot | `AES-256-GCM` cryptographic barrier | Data cannot be decrypted without the unseal keys or access to the seal. |
| Tampering with data in storage | `AES-GCM` authenticated encryption | Ciphertext tampering is detected on read. |
| Eavesdropping on client, node-to-node, and cluster-to-cluster traffic | TLS on the listener and cluster port, barrier encryption | Data travels over TLS, and in Raft and replication also as ciphertext. |
| Compromise of a single unseal key holder | Shamir's Secret Sharing, `-key-threshold` | One share is not enough to unseal or generate a root token. |
| Theft of the root key from storage | Seal: Shamir, HSM (`pkcs11`), Yandex Cloud KMS | The root key is stored encrypted; the seal key never leaves the HSM or KMS. |
| Decrypting key material after the barrier keyring is compromised | [Seal wrap](../../kms-hsm/sealwrap/) (Stronghold EE) | The root key, keyring, recovery key, and selected PKI, SSH, Transit, and LDAP values are additionally encrypted through the seal. |
| A client accessing other clients' secrets | Tokens and ACL policies, default deny | A client gets only the paths and operations allowed by its policies. |
| Long-term use of leaked credentials | Leases and TTLs, token and lease revocation | Dynamic secrets and tokens expire and are revoked, including in cascade. |
| Secrets leaking through audit logs | HMAC-SHA256 hashing with the audit device salt | Sensitive values do not appear in the log in plaintext. |
| Process memory being swapped out | `mlock` | Memory pages are not written to disk unless `mlock` is disabled with `disable_mlock`. In the DP module, `disable_mlock = true`, so this protection does not apply there. |

## What the mechanisms do not protect against

| Threat | Why it is not covered | How to reduce the risk |
| --- | --- | --- |
| Compromise of a root token or a token with broad policies | The barrier and seal take no part in authorization: a token grants everything its policies allow. | In a standalone installation, do not keep the root token after initial setup; issue one with `operator generate-root` by quorum when needed. In DP in the `Automatic` mode, the root token is needed by the `stronghold-automatic` configurator, so do not revoke it; follow the measures for the DP cluster compromise threat below. Assign minimal policies, short TTLs, [MFA](../../../user/auth/mfa/), and CIDR-bound tokens. |
| Compromise of an unsealed node (root on the server, memory dump, debugger) | The keyring and plaintext data are in process memory. | Isolate Stronghold servers, restrict root access and debugging, do not disable `mlock` if swap is not encrypted. |
| Collusion or compromise of a quorum of unseal or recovery key holders | A quorum can by definition unseal the storage and generate a root token. | Distribute shares among independent people, store them separately, and rekey when holders change. |
| An HSM or KMS administrator with a copy of the storage | Whoever controls the seal key and has the data can unseal the copy. | Separate HSM/KMS and storage administrator roles, protect snapshots separately. |
| Compromise of the DP cluster (`Automatic` mode) | In the `Automatic` mode, the `stronghold-keys` secret in the `d8-stronghold` namespace contains the root token and unseal key; compromising the secret gives administrative access to the whole storage. In the `Manual` mode, there is no `stronghold-keys` secret. Data is on the nodes holding Stronghold data (master nodes by default). | Restrict RBAC to the `d8-stronghold` namespace and access to the nodes, forbid unneeded `exec` and volume reads, and enable Kubernetes audit. Then choose one of the options: take the keys and delete the `stronghold-keys` secret from the cluster; take the keys and rotate the unseal key and the root token; or use the `Manual` mode. Keep the keys outside the cluster in a protected location. See [Deployment in DP](../deployment-dkp/). |
| Spoofing a client IP through the `8500` listener (DP, `Ingress` and `GatewayAPI` inlets) | The listener on port `8500` requires a client certificate (mTLS) but trusts `X-Forwarded-For` from any address (`x_forwarded_for_authorized_addrs = "0.0.0.0/0"`). If an mTLS client is compromised or the proxy trust model is wrong, the client IP can be spoofed, which affects CIDR-bound tokens and audit. | Restrict access to port `8500` to trusted networks (NetworkPolicy, firewall) and rely on `X-Forwarded-For` only with a strictly defined proxy chain. |
| Secrets after they are issued to a client | Stronghold does not control what an application does with the received data. | Use dynamic secrets with short TTLs, [response wrapping](../../../concepts/response-wrapping/), and Stronghold Agent. |
| Loss or deletion of the seal key | Recovery keys cannot decrypt the root key. Without the seal mechanism, the cluster cannot be restored even from a backup. | Strictly control HSM and KMS key management, perform a [seal migration](../../../concepts/seal/#seal-migration) before retiring a key. |
| Denial of service | Loss of Raft quorum or a blocking audit device stops request processing. | Use at least three nodes, several audit devices, [backups](../../backups/overview/), and DR replication. |
| Actions of a legitimate administrator | An administrator with broad policies can change the configuration and policies. | Enable [audit](../../audit/overview/) right after initialization, ship logs to an external system, and use [GitOps](../../../user/secrets-engines/gitops/overview/) with a signature quorum for configuration changes (Stronghold EE). |

## Recommendations

1. Issue listener certificates from a trusted CA and verify them on clients. TLS and GOST cipher settings are described in [Cryptography](../../cryptography/overview/).
1. Allow only the required connections from the tables in [Network ports](../ports/).
1. Use auto unseal through an [HSM](../../kms-hsm/hsm/) or [Yandex Cloud KMS](../../kms-hsm/yandexcloudkms/) if manual handling of key shares must be avoided, and enable [seal wrap](../../kms-hsm/sealwrap/) for key material.
1. Grant access through [auth methods](../../../concepts/auth/) and [policies](../../../concepts/policy/) with least privilege instead of shared tokens.
1. Enable at least two [audit devices](../../audit/overview/).
1. Keep [backups](../../backups/overview/) outside the protected system and protect them like the storage itself.
