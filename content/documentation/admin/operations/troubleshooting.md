---
title: "Troubleshooting"
description: "Diagnosing and fixing common Stronghold problems: permission and TLS errors, sealed nodes, lost Raft quorum, slow responses, audit, login, and DKP pod issues."
weight: 30
---

This page covers common Stronghold operational problems, how to diagnose them, and how to fix them. Before troubleshooting, check the overall node and cluster state:

```shell
d8 stronghold status
d8 stronghold operator raft list-peers
```

## Permission denied

Symptom: a `permission denied` error (HTTP code `403`) when accessing a path.

Diagnostics:

1. Check which policies are attached to the token:

   ```shell
   d8 stronghold token lookup
   ```

1. Check the token capabilities on a specific path:

   ```shell
   d8 stronghold token capabilities secret/data/app/config
   ```

   For another token, pass it explicitly or use its accessor via the [`/sys/capabilities-accessor`](../../../reference/api/system/#post-syscapabilities-accessor) endpoint:

   ```shell
   d8 stronghold token capabilities <TOKEN> secret/data/app/config
   ```

1. Review the policy contents:

   ```shell
   d8 stronghold policy read <POLICY_NAME>
   ```

Common causes:

- for KV v2, the policy lacks the `data/` or `metadata/` segment (for example, `secret/app/*` instead of `secret/data/app/*`), see [KV v2](../../../user/secrets-engines/kv/kv-v2/);
- the path is blocked by an explicit `deny` rule, which takes precedence;
- the request is made in a different [namespace](../../namespaces/overview/);
- the token has expired: `d8 stronghold token lookup` returns an error.

For policy syntax, see [Policies](../../../concepts/policy/).

## x509 errors

Symptoms: `x509: certificate signed by unknown authority`, `x509: certificate is valid for ..., not ...`, `x509: certificate has expired`.

| Error | Cause | Fix |
| --- | --- | --- |
| `certificate signed by unknown authority` | The client does not trust the CA that issued the server certificate | Specify the CA via `STRONGHOLD_CACERT` or the `-ca-cert` flag |
| `certificate is valid for X, not Y` | The address in `STRONGHOLD_ADDR` is not in the certificate SAN | Use a name from the SAN or reissue the certificate with the required SANs |
| `certificate has expired` | The listener certificate has expired | Replace the certificate as described in [TLS certificates](../tls-certificates/) |

Check the certificate presented by the server:

```shell
openssl s_client -connect stronghold.example.com:8200 -showcerts </dev/null \
  | openssl x509 -noout -subject -issuer -dates -ext subjectAltName
```

For Raft inter-node communication (`retry_join`), the same errors appear in standby node logs. Check `leader_ca_cert_file`, `leader_client_cert_file`, and `leader_client_key_file` (see [Installation](../../../install/standalone/installation/)).

## Node is sealed after a restart

With Shamir seal, every node starts sealed after a restart. This is expected behavior (see [Seal](../../../concepts/seal/)).

1. Check the state:

   ```shell
   d8 stronghold status
   ```

1. Unseal the node by entering the threshold number of key shares:

   ```shell
   d8 stronghold operator unseal
   ```

If auto unseal ([KMS or HSM](../../kms-hsm/hsm/)) is configured but the node stays sealed, check the [server logs](../logs/) for KMS/HSM access errors: service availability, credentials, HSM PIN or slot.

In DP, the module in `Automatic` mode initializes the cluster with one key share and a threshold of 1. The unseal key and root token are stored in the `stronghold-keys` secret (`unsealKey`, `rootToken`) in the `d8-stronghold` namespace. Nodes are unsealed by the `stronghold-automatic` Deployment (check period is 60 seconds by default). If a node stays sealed, check that the secret exists and look at the unsealer logs:

```shell
d8 k -n d8-stronghold get secret stronghold-keys
d8 k -n d8-stronghold logs deploy/stronghold-automatic
```

In `Manual` mode, there is neither the secret nor the unsealer: unseal nodes manually.

## Lost Raft quorum

Symptom: the `local node not active but active cluster node not found` error; the cluster does not elect a leader.

Quorum requires a majority of voting nodes: 2 of 3 or 3 of 5. If the majority is permanently lost, follow the recovery guide:

- [Recovering from lost quorum in Linux](../../../install/standalone/raft-lost-quorum-recovery/);
- [Recovering from lost quorum in DP](../../../install/dkp/raft-lost-quorum-recovery/).

Before recovery, make sure the unavailable nodes will not come back: starting old nodes after recovery may lead to diverging data.

## Leader election flapping

Symptoms: logs frequently show transitions to the candidate state, clients get short-lived errors, the active node keeps changing.

Diagnostics:

- check the `stronghold_raft_leader_lastContact` and `stronghold_raft_state_candidate` metrics (see [Monitoring](../monitoring/));
- check network latency and packet loss between nodes on the cluster port (`cluster_addr`, `8201` by default);
- check Raft data disk latency (`iostat -x 1`), see [Sizing](../sizing/);
- check time synchronization (NTP) on all nodes;
- check Autopilot state:

  ```shell
  d8 stronghold operator raft autopilot state
  ```

Typical causes: an overloaded or network-attached disk with high latency, CPU starvation during load spikes, nodes placed in different sites with high latency, process pauses due to memory pressure.

## Slow responses and lease explosion

Symptoms: response time grows, memory usage and storage size increase, startup and leader changes take long.

A frequent cause is uncontrolled growth of leases and tokens: clients authenticate on every request instead of reusing a token, or lease TTLs are too long.

Diagnostics:

1. Check the total lease count via the [`/sys/leases/count`](../../../reference/api/system/#get-sysleasescount) endpoint and the `stronghold_expire_num_leases` metric.
1. Find prefixes with the most leases:

   ```shell
   d8 stronghold list sys/leases/lookup/auth/approle/login
   ```

1. Check TTLs of auth methods and secrets engines:

   ```shell
   d8 stronghold read sys/auth/approle/tune
   ```

Fixes:

- reduce `default_lease_ttl` and `max_lease_ttl` for the affected mounts;
- make clients reuse and renew tokens (for example, via [Stronghold Agent](../../../user/agent/overview/));
- revoke excess leases by prefix: `d8 stronghold lease revoke -prefix <PATH>`;
- limit load with [quotas](../quotas/).

## Audit device blocks requests

If no audit device can write an event, Stronghold rejects the request or requests hang until the problem is fixed (see [Audit in Stronghold](../../audit/overview/)).

Diagnostics:

1. List audit devices:

   ```shell
   d8 stronghold audit list -detailed
   ```

1. For a `file` device, check free space and write permissions for the log directory.
1. For `syslog` and `socket` devices, check that the receiver is reachable.
1. Check the `stronghold_audit_log_request_failure` and `stronghold_audit_log_response_failure` metrics.

Fix: free up space or restore the receiver. Keep at least two audit devices so that a single failure does not stop request processing.

## Login issues

### OIDC

- **`redirect_uri` or `invalid redirect` error**: the redirect URI must match in the role `allowed_redirect_uris` parameter and in the provider settings. The CLI uses `http://localhost:8250/oidc/callback`, the web UI uses an address like `https://<STRONGHOLD_ADDR>/ui/stronghold/auth/oidc/oidc/callback`. See [OIDC](../../../user/auth/oidc/).
- **Provider token validation errors**: check `oidc_discovery_url`, `oidc_client_id`, and `oidc_discovery_ca_pem` for a private CA.
- **No permissions after login**: check `bound_claims`, `groups_claim`, and group-to-policy mapping.

### LDAP

- **Bind error**: check `binddn` and `bindpass`, and that the server is reachable from all Stronghold nodes.
- **TLS errors**: with `ldaps://` or `starttls`, specify the server CA in the `certificate` parameter; do not use `insecure_tls` in production.
- **User logs in but gets no policies**: check `groupdn`, `groupfilter`, and `groupattr`.

See [LDAP](../../../user/auth/ldap/).

## DP pod issues

Check pod state and events:

```shell
d8 k -n d8-stronghold get pods -o wide
d8 k -n d8-stronghold describe pod stronghold-0
d8 k -n d8-stronghold get events --sort-by=.lastTimestamp
```

| Symptom | What to check |
| --- | --- |
| Pod is `ContainerCreating`, no `ingress-tls` secret (`ingress-tls-customcertificate` in `CustomCertificate` mode) | In `CertManager` mode, cert-manager certificate issuance (the `stronghold` Certificate exists only with the `Ingress` inlet); in `CustomCertificate` mode, the secret in `d8-system`. See [Configuring Stronghold](../../../install/dkp/configuration/#troubleshooting) |
| Pod is in `CrashLoopBackOff` | Logs of the current and previous run: `d8 k -n d8-stronghold logs stronghold-0 --previous` |
| Pod is `Pending` | Availability of master nodes and their resources |
| Pod is running but the node is sealed | Presence of the `stronghold-keys` secret and `stronghold-automatic` logs (`Automatic` mode only) |

Module state:

```shell
d8 k get module stronghold
```

<!-- TODO(verify): module status command (d8 k get module stronghold) and status fields. -->

## Collecting diagnostic data

### Debug logs

Temporarily raise the log level without a restart via the `/sys/loggers` endpoint (see [Server logs](../logs/#changing-the-log-level)) and revert it after collecting data.

### debug bundle

The `d8 stronghold debug` command collects node state, metrics, pprof profiles, and logs for a given period into an archive:

```shell
d8 stronghold debug -duration=5m -interval=30s -output=stronghold-debug.tar.gz
```

The token needs read access to `sys/metrics`, `sys/pprof/*`, `sys/host-info`, `sys/in-flight-req`, and `sys/monitor`. Send the archive to support over a secure channel: it contains information about the cluster configuration and layout.
