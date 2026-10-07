---
title: "Shipping audit logs to SIEM"
linkTitle: "SIEM"
description: "Sending Stronghold audit logs to SIEM systems: audit devices, collection with log-shipper, Vector, and Fluent Bit, and field mapping for ELK, OpenSearch, Splunk, MaxPatrol SIEM, and KUMA."
weight: 70
---

Stronghold audit logs are the main source of security events: who accessed which path, when, from which address, and with what result. For centralized analysis, ship them to a SIEM system.

{{< alert level="info" >}}
The `audit enable` command and audit devices are available only in Stronghold EE.
{{< /alert >}}

General flow:

1. Enable one or more [audit devices](../../../admin/audit/overview/) (`file`, `syslog`, `socket`).
1. Collect entries with a log shipping agent (`log-shipper` in Deckhouse Platform, Vector, Fluent Bit, rsyslog).
1. Parse JSON entries and map fields to the SIEM schema.

## Choosing an audit device

| Device | Collection method | Notes |
| --- | --- | --- |
| `file` | An agent reads the file and sends entries to the SIEM | The most predictable option. Requires file rotation and disk space monitoring |
| `syslog` | A local syslog agent forwards entries to the SIEM | Entries can be large: use a reliable transport and duplicate with a `file` device |
| `socket` | Stronghold sends entries directly to a collector TCP, UDP, or UNIX socket | If the collector is unavailable, Stronghold requests may block; UDP may lose entries |

Enable at least two audit devices: if writing to the audit log fails, Stronghold may stop serving requests. See [Audit in Stronghold](../../../admin/audit/overview/) for details.

Example: a file for SIEM delivery and syslog as a backup channel:

```bash
d8 stronghold audit enable file file_path=/var/log/stronghold_audit.log
d8 stronghold audit enable syslog tag="stronghold" facility="AUTH"
```

## Collection in Deckhouse Platform

In DP, logs are shipped by the [`log-shipper`](/products/kubernetes-platform/documentation/v1/modules/log-shipper/) module. It reads pod logs and sends them to Elasticsearch, OpenSearch, Splunk, Logstash, Kafka, Loki, Vector, or over syslog.

For `log-shipper` to collect the audit log as pod logs, entries must be written to standard output. In DP (Stronghold EE, `management.mode: Automatic`), enable the audit log with the `enableAuditLog` parameter in the `stronghold` module's ModuleConfig:

```yaml
spec:
  settings:
    enableAuditLog: true
```

The module enables the `stronghold_audit_stdout` audit device (type `file`, `file_path=stdout`) itself. On EE, if the parameter is not set, the module disables the audit log. The parameter cannot be set when `management.mode` is `Manual`; in that case, as well as on a standalone installation, enable the device manually:

```bash
d8 stronghold audit enable -path=file-stdout file file_path=stdout
```

Example `log-shipper` configuration that sends Stronghold pod logs to Elasticsearch:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ClusterLoggingConfig
metadata:
  name: stronghold-audit
spec:
  type: KubernetesPods
  kubernetesPods:
    namespaceSelector:
      labelSelector:
        matchLabels:
          kubernetes.io/metadata.name: d8-stronghold
  destinationRefs:
    - siem-elasticsearch
---
apiVersion: deckhouse.io/v1alpha1
kind: ClusterLogDestination
metadata:
  name: siem-elasticsearch
spec:
  type: Elasticsearch
  elasticsearch:
    endpoint: https://elasticsearch.example.com:9200
    index: stronghold-audit-%F
    auth:
      strategy: Basic
      user: stronghold
      password: <base64-encoded password>
```

Pod logs contain Stronghold operational logs along with audit entries. Separate audit entries by the presence of the `type` (`request` or `response`) and `request.id` fields.

## Collection outside Kubernetes

For Stronghold on Linux, use any agent that can read a file with JSON lines. Example Vector configuration:

```yaml
sources:
  stronghold_audit:
    type: file
    include:
      - /var/log/stronghold_audit.log

transforms:
  parse:
    type: remap
    inputs: [stronghold_audit]
    source: |
      . = parse_json!(.message)

sinks:
  siem:
    type: elasticsearch
    inputs: [parse]
    endpoints: ["https://elasticsearch.example.com:9200"]
    bulk:
      index: "stronghold-audit-%Y.%m.%d"
```

Example for Fluent Bit:

```ini
[INPUT]
    Name   tail
    Path   /var/log/stronghold_audit.log
    Parser json
    Tag    stronghold.audit

[OUTPUT]
    Name   forward
    Match  stronghold.audit
    Host   siem-collector.example.com
    Port   24224
```

Configure audit file rotation (for example, with `logrotate`) and send Stronghold a `SIGHUP` signal after rotation so that it reopens the file.

## Connecting SIEM systems

| SIEM system | Recommended ingestion |
| --- | --- |
| ELK / OpenSearch | Direct index writes via `log-shipper`, Vector, or Logstash; JSON parsing on the agent side |
| Splunk | HTTP Event Collector (HEC) with a `_json` `sourcetype` |
| MaxPatrol SIEM | JSON events over syslog or from a file via a collection agent; a normalization rule for the Stronghold format is required <!-- TODO(verify): availability of a ready-made Stronghold/Vault normalization package in MaxPatrol SIEM --> |
| KUMA | Collector with a JSON normalizer, ingestion over TCP/syslog <!-- TODO(verify): availability of a ready-made Stronghold/Vault normalizer in KUMA --> |

## Field mapping

The entry structure is described in [Audit log entry schema](../../../admin/audit/log-format/). Typical mapping to SIEM fields:

| Stronghold field | SIEM meaning |
| --- | --- |
| `time` | Event time |
| `type` | Event type: `request` or `response` |
| `request.id` | Identifier linking the request and the response |
| `request.operation` | Action: `create`, `read`, `update`, `delete`, `list` |
| `request.path` | Accessed object |
| `request.mount_type`, `request.mount_point` | Type and path of the secrets engine or auth method |
| `request.namespace.path` | Stronghold namespace |
| `request.remote_address` | Source IP address |
| `auth.display_name` | Subject name |
| `auth.entity_id` | Entity identifier |
| `auth.policies`, `auth.token_policies` | Applied policies |
| `error` | Outcome: if present, an error or access denial |

Practical notes:

- Sensitive strings (tokens, secret values) are hashed with HMAC-SHA256 using the audit device salt. To compare a value with a hash, use the audit hash mechanism of the same device.
- Use `response` entries for correlation: they contain both the request data and the result.
- To reduce volume, use [filtering](../../../admin/audit/filtering/) and [field exclusion](../../../admin/audit/exclusion/) on a separate audit device that feeds the SIEM. Keep the full log on another device.
- Example events for correlation rules: repeated authentication failures (`auth/*/login` with `error`), operations on `sys/audit`, `sys/policies`, `sys/auth`, and root token usage (`auth.policies` contains `root`).
