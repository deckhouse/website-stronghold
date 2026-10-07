---
title: "Monitoring"
description: "Enabling Stronghold telemetry, collecting metrics with Prometheus, key metrics, alerting rule examples, and health checks via sys/health."
weight: 10
---

Stronghold monitoring relies on three data sources:

- telemetry metrics exposed at the `/v1/sys/metrics` endpoint;
- the `/v1/sys/health` health check endpoint;
- [server logs](../logs/) and [audit logs](../../audit/overview/).

## Enabling telemetry

{{< alert level="info" >}}
This section and the next ones up to [Monitoring in DP](#monitoring-in-dp) describe standalone installations. In DP, the configuration is managed by the module (the `stronghold-config` ConfigMap), see [Monitoring in DP](#monitoring-in-dp).
{{< /alert >}}

Telemetry settings are defined in the `telemetry` block of the server configuration file (see [Configuration](../../../install/standalone/configuration/)). To expose metrics in the Prometheus format, set the in-memory retention time:

```hcl
telemetry {
  prometheus_retention_time = "24h"
  disable_hostname          = true
}
```

- `prometheus_retention_time` — how long metrics are retained for Prometheus output. The value `0` disables the Prometheus format. The default is `24h`.
- `disable_hostname` — do not prefix metrics with the node hostname. Enable it so that metric names match across all nodes.
- `metrics_prefix` — metric name prefix. By default, metrics have the `stronghold` prefix; the DP module sets the same value explicitly.

Restart the service after changing the configuration (standalone):

```shell
systemctl restart stronghold
```

## Accessing sys/metrics

The [`GET /sys/metrics`](../../../reference/api/system/#get-sysmetrics) endpoint returns aggregated metrics. For the Prometheus format, pass `format=prometheus`:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  "${STRONGHOLD_ADDR}/v1/sys/metrics?format=prometheus"
```

### Token-based access

Create a dedicated least-privilege policy for the metrics collector:

```hcl
path "sys/metrics" {
  capabilities = ["read"]
}
```

```shell
d8 stronghold policy write prometheus-metrics prometheus-metrics.hcl
d8 stronghold token create -policy=prometheus-metrics -orphan -period=24h
```

Use a periodic token and set up its renewal, or issue the token through an auth method (for example, [Kubernetes](../../../user/auth/kubernetes/) or [AppRole](../../../user/auth/approle/)).

### Unauthenticated access

If metrics are collected from a trusted network, allow unauthenticated access at the listener level:

```hcl
listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/opt/stronghold/tls/node-1-cert.pem"
  tls_key_file  = "/opt/stronghold/tls/node-1-key.pem"

  telemetry {
    unauthenticated_metrics_access = true
  }
}
```

{{< alert level="warning" >}}
Metrics disclose information about the storage layout (mount paths, lease and token counts). Allow unauthenticated access only when network access to the listener is restricted.
{{< /alert >}}

### Prometheus scrape job example

```yaml
scrape_configs:
  - job_name: stronghold
    metrics_path: /v1/sys/metrics
    params:
      format: ["prometheus"]
    scheme: https
    tls_config:
      ca_file: /etc/prometheus/stronghold-ca.pem
    authorization:
      credentials_file: /etc/prometheus/stronghold-token
    static_configs:
      - targets:
          - raft-node-1.demo.tld:8200
          - raft-node-2.demo.tld:8200
          - raft-node-3.demo.tld:8200
```

Scrape every cluster node rather than only the load balancer address: some metrics (Raft, runtime) are node-specific.

Every node, including standby nodes, serves its own metrics: a `sys/metrics` request is not forwarded to the active node.

## Monitoring in DP

In DP, Stronghold runs in the `d8-stronghold` namespace, and its configuration (including telemetry) is managed by the module through the `stronghold-config` ConfigMap. Metrics are collected as follows:

- the telemetry settings set `metrics_prefix = "stronghold"`, so all metrics have the `stronghold_` prefix (for example, `stronghold_core_unsealed`);
- metrics are served by a loopback listener `127.0.0.1:8400` that is not reachable from outside the Pod;
- the `kube-rbac-proxy` sidecar publishes them on port `9889` (`https-metrics`); access requires the `get` permission on `statefulsets/prometheus-metrics` (the `stronghold` resource) in the `d8-stronghold` namespace. The module grants it to the `prometheus` ServiceAccount in `d8-monitoring`.

The module creates the monitoring objects itself, no manual setup is required:

- the `stronghold` PodMonitor in the `d8-monitoring` namespace;
- a PrometheusRule with the module alerts (see below);
- three Grafana dashboards: `stronghold.json`, `replication.json` and `snapshot_auto.json`.

To check that the metrics collection object exists:

```shell
d8 k -n d8-monitoring get podmonitor stronghold
```

### Module alerts in DP

The alerts are defined in the module's PrometheusRule and use the `severity_level` label, as DP alerts do:

| Alert | Condition | `for` | `severity_level` |
| --- | --- | --- | --- |
| `D8StrongholdNoReadyPod` | No `stronghold` StatefulSet Pod is in the Ready state | 3m | 4 |
| `D8StrongholdNoActiveNodes` | `stronghold_core_active` is `0` on all nodes (or there are no metrics) | 1m | 3 |
| `D8StrongholdSealedNodesPresent` | The number of nodes exporting metrics is greater than the number of unsealed nodes | 5m | 7 |
| `D8StrongholdClusterNotHealthy` | The sum of `stronghold_autopilot_healthy` is `0` | 5m | 7 |
| `D8StrongholdQuorumInCriticalState` | `stronghold_autopilot_failure_tolerance` is `0`; created only if the cluster has more than 2 master nodes (`clusterMasterCount > 2`) | 3m | 4 |
| `D8StrongholdAbsentMetrics` | The number of ready `kube-rbac-proxy` containers is greater than the number of nodes exporting metrics | 1m | 3 |
| `D8StrongholdAutoSnapshotFailed` | `stronghold_core_snapshot_auto_failed` is `1` for an automatic snapshot configuration | 5m | 6 |
| `D8StrongholdAutoSnapshotRotationFailed` | `stronghold_autosnapshots_rotate_failed` is `1`: rotation of old snapshots failed after a successful backup | 15m | 8 |

Custom alerts in DP are defined with the [CustomPrometheusRules](/modules/prometheus/cr.html#customprometheusrules) resource: move the `spec.groups` content from the examples below into it.

<!-- TODO(verify): current links to DKP docs on metrics collection and CustomPrometheusRules. -->

## Key metrics

Metric names are given with the `stronghold_` prefix: Stronghold uses it by default in both standalone installations and DP. You can change the prefix with the `metrics_prefix` parameter of the `telemetry` block.

| Metric | Meaning | What to watch |
| --- | --- | --- |
| `stronghold_core_unsealed` | Node is unsealed (`1`) or sealed (`0`) | Any `0` value |
| `stronghold_core_active` | Node is active (`1`) | Cluster-wide sum must equal `1` |
| `stronghold_core_handle_request` | Request handling time, ms | Growth of the 99th percentile |
| `stronghold_core_handle_login_request` | Login request handling time, ms | Growth of the 99th percentile |
| `stronghold_expire_num_leases` | Number of active leases | Steady growth without decrease |
| `stronghold_token_count` | Number of tokens | Steady growth without decrease |
| `stronghold_audit_log_request_failure` | Failures writing requests to audit devices | Any non-zero increase |
| `stronghold_audit_log_response_failure` | Failures writing responses to audit devices | Any non-zero increase |
| `stronghold_raft_leader_lastContact` | Time since a standby last contacted the Raft leader, ms | Values above hundreds of milliseconds |
| `stronghold_raft_state_candidate` | Node transitions to the Raft candidate state | Frequent increases indicate unstable leader election |
| `stronghold_raft_commitTime` | Raft commit time, ms | Growth indicates slow disk or network |
| `stronghold_autopilot_healthy` | Cluster health as reported by Autopilot | Value `0` |
| `stronghold_runtime_alloc_bytes` | Memory used by the process | Approaching the memory limit |

## Alerting rule examples

A PrometheusRule example with a basic set of alerts. It is a template for standalone installations: it uses the `severity` label and standalone paths. For DP, see [Module alerts in DP](#module-alerts-in-dp). Tune the thresholds to your workload.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: stronghold
spec:
  groups:
    - name: stronghold.availability
      rules:
        - alert: StrongholdSealed
          expr: stronghold_core_unsealed == 0
          for: 1m
          labels:
            severity: critical
          annotations:
            summary: "Stronghold node {{ $labels.instance }} is sealed"
            description: "The node does not serve requests. Unseal it or check auto unseal."
        - alert: StrongholdNoActiveNode
          expr: sum(stronghold_core_active) < 1
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: "The Stronghold cluster has no active node"
            description: "Check Raft and quorum state with d8 stronghold operator raft list-peers."
        - alert: StrongholdLeaderFlapping
          expr: increase(stronghold_raft_state_candidate[15m]) > 3
          labels:
            severity: warning
          annotations:
            summary: "Frequent Raft leader elections on {{ $labels.instance }}"
            description: "Check network connectivity between nodes and disk latency."
        - alert: StrongholdAutopilotUnhealthy
          expr: stronghold_autopilot_healthy == 0
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Autopilot reports the Stronghold cluster as unhealthy"
    - name: stronghold.performance
      rules:
        - alert: StrongholdHighRequestLatency
          expr: stronghold_core_handle_request{quantile="0.99"} > 500
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "99th percentile request handling time is above 500 ms"
        - alert: StrongholdLeaseCountGrowth
          expr: delta(stronghold_expire_num_leases[1h]) > 10000
          labels:
            severity: warning
          annotations:
            summary: "Lease count grew by more than 10000 in an hour"
            description: "Check lease TTLs and clients that do not reuse tokens."
        - alert: StrongholdLeaseCountHigh
          expr: stronghold_expire_num_leases > 250000
          for: 15m
          labels:
            severity: warning
          annotations:
            summary: "Active lease count exceeds 250000"
    - name: stronghold.audit
      rules:
        - alert: StrongholdAuditFailures
          expr: increase(stronghold_audit_log_request_failure[5m]) > 0 or increase(stronghold_audit_log_response_failure[5m]) > 0
          labels:
            severity: critical
          annotations:
            summary: "Stronghold audit device write failures"
            description: "If all audit devices fail, Stronghold stops serving requests."
    - name: stronghold.certificates
      rules:
        - alert: StrongholdTLSCertificateExpiringSoon
          expr: (probe_ssl_earliest_cert_expiry{job="stronghold-tls"} - time()) / 86400 < 21
          for: 1h
          labels:
            severity: warning
          annotations:
            summary: "Stronghold TLS certificate on {{ $labels.instance }} expires in less than 21 days"
    - name: stronghold.storage
      rules:
        - alert: StrongholdRaftDiskUsageHigh
          expr: |
            (1 - node_filesystem_avail_bytes{mountpoint="/opt/stronghold/data"}
              / node_filesystem_size_bytes{mountpoint="/opt/stronghold/data"}) > 0.8
          for: 15m
          labels:
            severity: warning
          annotations:
            summary: "Raft data volume on {{ $labels.instance }} is more than 80% full"
```

Notes on the example:

- The TLS expiry alert uses the `probe_ssl_earliest_cert_expiry` metric from blackbox_exporter probing the Stronghold HTTPS address. Monitor root and intermediate certificates issued by the Stronghold [PKI](../../../user/secrets-engines/pki/) engine the same way, for example with a certificate exporter that checks files or service addresses.
- The disk usage alert uses node-exporter metrics. Specify the mount point of the directory set in the `path` parameter of the `storage "raft"` block. In DP, data is stored in `/var/lib/deckhouse/stronghold` on master nodes (inside the Pod it is mounted at `/stronghold/data`). If the `storageClass` parameter is set in the module configuration, data is stored in PVCs `data-stronghold-N` instead, and the example's mount point does not apply.
- For DP, move `spec.groups` into a CustomPrometheusRules resource.

Stronghold does not export PKI certificate expiry metrics (only the tidy metrics `secrets_pki_tidy_*` exist). Monitor certificate expiry with external tools, for example blackbox-exporter.

## Health checks via sys/health

The [`GET /sys/health`](../../../reference/api/system/#get-syshealth) endpoint requires no authentication and is suitable for load balancer and external monitoring checks:

```shell
curl -s -o /dev/null -w "%{http_code}\n" "${STRONGHOLD_ADDR}/v1/sys/health"
```

Response codes:

| Code | State |
| --- | --- |
| `200` | Initialized, unsealed, active node |
| `429` | Unsealed, standby node |
| `472` | Node is a DR replication secondary |
| `473` | Performance standby node |
| `423` | The active node is not ready yet (only with `isleaderreadyok=true`) |
| `425` | Raft Autopilot is not ready yet (only with `raftautopilotok=true`) |
| `501` | Not initialized |
| `503` | Sealed |

Response codes can be overridden with query parameters (upstream Vault behavior):

- `standbyok=true` — return `200` for standby nodes. Use it for a load balancer that distributes requests across all unsealed nodes;
- `activecode`, `standbycode`, `sealedcode`, `uninitcode` — set a custom code for the corresponding state.

Example check for a load balancer that should route traffic to the active node only:

```shell
curl -s -o /dev/null -w "%{http_code}\n" "${STRONGHOLD_ADDR}/v1/sys/health?standbycode=503"
```

To check Raft state, use:

```shell
d8 stronghold operator raft list-peers
d8 stronghold operator raft autopilot state
```
