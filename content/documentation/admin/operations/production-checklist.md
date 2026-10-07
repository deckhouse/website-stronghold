---
title: "Production checklist"
description: "A checklist of hardening and reliability measures for a Stronghold cluster before going to production."
weight: 40
---

Use this checklist before putting Stronghold into production and during periodic configuration reviews. Items that rely on Stronghold EE features are marked.

## Network and TLS

- [ ] TLS is enabled on all listeners. The `tls_disable` parameter is not used (see [Configuration](../../../install/standalone/configuration/#listener)). In DP, all network listeners use TLS; only the metrics listener `127.0.0.1:8400` inside the Pod is plain HTTP, and it is not reachable from outside.
- [ ] The minimum TLS version is `tls12` or `tls13` (the `tls_min_version` parameter).
- [ ] Raft inter-node communication uses per-node certificates and a shared CA.
- [ ] Certificate expiry monitoring and a replacement procedure are in place (see [TLS certificates](../tls-certificates/)).
- [ ] Access to the API port (`8200`) is limited to client networks, and access to the cluster port (`8201`) to Stronghold nodes only.
- [ ] In DP, the Pod ports are: `8200` (API), `8201` (cluster port), `8300`/`8301` (internal API and cluster port between Pods), `8500` (mTLS listener, only with the `Ingress` and `GatewayAPI` inlets) and `9889` (metrics via `kube-rbac-proxy`). Port `8201` is exposed outside the cluster only for replication (`publishCluster`).
- [ ] Unauthenticated access to `sys/metrics` and `sys/pprof` is either disabled or limited to a trusted network.
- [ ] When a load balancer or proxy is used, `x_forwarded_for_authorized_addrs` and a `sys/health` check are configured.

## Access and authentication

- [ ] The root token obtained during initialization is revoked after the initial setup. For emergency operations, a root token is generated again with `d8 stronghold operator generate-root` (see [Key management](../key-management/#generating-a-root-token)). This applies to standalone installations and DP `Manual` mode. In DP `Automatic` mode, the `stronghold-keys` secret contains `unsealKey` and `rootToken`, and the `stronghold-automatic` configurator uses the root token, so do not revoke it. Instead, as recommended for threat TM-01 in the Stronghold threat model, take the keys out of the secret and store them in a protected location outside the cluster, and/or rotate the values stored in it, minimize RBAC access to the secret, and enable Kubernetes audit.
- [ ] Administrators log in through an external provider ([OIDC](../../../user/auth/oidc/) or [LDAP](../../../user/auth/ldap/)) rather than [userpass](../../../user/auth/userpass/).
- [ ] [Multi-factor authentication](../../../user/auth/mfa/totp/) is enabled for administrative roles.
- [ ] Applications use machine auth methods ([Kubernetes](../../../user/auth/kubernetes/), [AppRole](../../../user/auth/approle/), [JWT](../../../user/auth/jwt/)) without long-lived static tokens.
- [ ] Policies follow the least-privilege principle; the use of `*` and `sudo` is limited (see [Policies](../../../concepts/policy/)).
- [ ] Policies are stored in version control and applied automatically.
- [ ] User lockout on password guessing is configured (the [`user_lockout`](../../../install/standalone/configuration/#user_lockout) block).

## Audit

- [ ] At least two audit devices are enabled (Stronghold EE, see [Audit in Stronghold](../../audit/overview/)).
- [ ] Audit logs are shipped to a centralized system with restricted access.
- [ ] Alerts on audit device write failures are configured (see [Monitoring](../monitoring/)).
- [ ] Log rotation and free space monitoring are configured for the `file` device.

## TTLs and limits

- [ ] Global `default_lease_ttl` and `max_lease_ttl` are reduced from the default (`768h`) to values matching your security policy.
- [ ] TTLs are configured at the mount and auth role level.
- [ ] Rate limit [quotas](../quotas/) are configured for auth methods and heavily used paths.

## Seal and keys

- [ ] Auto unseal via [HSM](../../kms-hsm/hsm/) or [KMS](../../kms-hsm/yandexcloudkms/) is used, or a manual Shamir unseal procedure is defined.
- [ ] Unseal or recovery key shares are distributed among different trusted people and encrypted with their PGP keys.
- [ ] No single person holds the threshold number of shares.
- [ ] A key custody procedure (safe, envelopes, issue log) and a holder replacement procedure are defined.
- [ ] Automatic encryption key rotation is configured (`sys/rotate/config`, see [Key management](../key-management/#rotating-the-encryption-key)).

## Backups

- [ ] Regular Raft snapshots are configured ([manual](../../backups/save/) or [automated](../../backups/automated-snapshots/) in Stronghold EE).
- [ ] Snapshots are stored outside the cluster in storage with restricted access.
- [ ] Test [restores from a snapshot](../../backups/restore/) are performed regularly in an isolated environment.
- [ ] The key shares required to unseal a restored cluster are available separately from the snapshots.

## Operating system and nodes

- [ ] A [supported OS](../../../about/requirements/#supported-os) is used, and security updates are installed regularly.
- [ ] Stronghold runs as a dedicated unprivileged user (`stronghold`); data directories have `0700` permissions. In DP, the Pod runs as the non-root user with UID 64535 with a read-only root file system and all capabilities dropped (`drop: ALL`).
- [ ] Swap is disabled or encrypted (standalone; in DP, `disable_mlock = true` is set by the module). With integrated Raft storage, `disable_mlock = true` with swap disabled is recommended (see [Configuration](../../../install/standalone/configuration/)).
- [ ] Core dumps are disabled for the Stronghold process, for example with `LimitCORE=0` in the systemd unit (standalone).
- [ ] Nodes run no unrelated services; SSH access is restricted and logged.
- [ ] Time is synchronized via NTP on all nodes.
- [ ] Raft data is stored on a local SSD (see [Sizing](../sizing/)).

## High availability

- [ ] The Raft cluster has 3 or 5 nodes placed in different failure domains.

{{< alert level="info" >}}
**In DP**: the number of replicas equals the number of master nodes plus arbiter nodes, rounded up to an odd number (+1 if even). The `stronghold` PodDisruptionBudget has `minAvailable` equal to `replicas/2 + 1` (rounded down). Destructive storage migrations require at least 3 replicas. See also [Sizing](../sizing/#in-dp).
{{< /alert >}}

- [ ] `retry_join` is configured on all nodes.
- [ ] The [lost quorum recovery](../../../install/standalone/raft-lost-quorum-recovery/) procedure has been tested.
- [ ] [DR replication](../../replication/disaster-recovery/) is configured for disaster recovery (Stronghold EE).

## Monitoring

- [ ] Metrics are collected from all nodes, and the alerts from [Monitoring](../monitoring/) are configured.
- [ ] The load balancer and external monitoring use `sys/health`.
- [ ] [Server logs](../logs/) are shipped to a centralized system.
- [ ] The [upgrade](../../../install/standalone/update/) procedure is documented and tested.
