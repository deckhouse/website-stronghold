---
title: "Audit in Stronghold"
linkTitle: "Overview"
weight: 10
description: "Overview of Stronghold audit logs: what gets logged, which backends are available, and how to configure auditing safely."
---

Auditing in Stronghold is used to record API requests and responses. Since every operation in Stronghold is performed through the API, audit logs let you reconstruct user, service, and administrator activity and use that data for investigations, monitoring, and compliance controls.

## Audit log contents

When an audit device is enabled, Stronghold records:

- API requests
- Responses to those requests
- Errors that occur while processing operations

As a result, the audit log becomes the primary source of information about who accessed which path, when an operation happened, what kind of operation was performed, and what the result was.

## Paths that are not audited

Some system paths bypass auditing:

- `sys/init`
- `sys/seal-status`
- `sys/seal`
- `sys/step-down`
- `sys/unseal`
- `sys/leader`
- `sys/health`
- `sys/rekey/init`
- `sys/rekey/update`
- `sys/rekey/verify`
- `sys/rekey-recovery-key/init`
- `sys/rekey-recovery-key/update`
- `sys/rekey-recovery-key/verify`
- `sys/storage/raft/bootstrap`
- `sys/storage/raft/join`
- `sys/internal/ui/feature-flags`

Depending on listener settings, the following unauthenticated paths may also bypass auditing:

- `sys/metrics`
- `sys/pprof/*`
- `sys/in-flight-req`

These exceptions should be considered when designing your overall observability and investigation model.

## Format and protection of audit data

Each audit log entry is a JSON object. The `type` field indicates the record type:

- `request`: For request entries.
- `response`: For response entries.

By default, sensitive string values in audit logs are not stored in plaintext. Instead, Stronghold hashes them using HMAC-SHA256 and a cryptographic salt specific to the audit device. This reduces the risk of leaking secrets through logs while still allowing value verification through the audit hash mechanism.

Keep in mind:

- Not all data types are masked in exactly the same way.
- Once an audit device is disabled, its hashing salt is lost, so comparing old values through that device is no longer possible.
- Raw logging of sensitive data should only be used deliberately and under tight control.

For a more detailed explanation of audit record contents and storage format, see ["Audit log record schema"](./log-format/).

## Using multiple audit devices

If multiple audit devices are enabled, Stronghold tries to write an event to all of them. This gives you several benefits:

- Redundant copies of audit records.
- The ability to send logs to different systems.
- Better visibility into delivery issues and possible log tampering.

This means the complete audit trail should be treated as the combination of logs from all configured devices.

{{< alert level="warning" >}}
It is recommended that you use at least two audit devices, or at minimum one primary device plus a well-designed fallback strategy. Audit failures can affect Stronghold's ability to serve requests.
{{< /alert >}}

## Audit device failures

Auditing is security-critical in Stronghold, so failure handling depends on the failure mode:

- If at least one audit device writes the event successfully, the request can usually complete successfully.
- If an audit device enters a blocking state and the write never completes, requests to Stronghold may hang until the problem is resolved.
- When using network-based backends, you must separately account for timeouts, socket availability, and possible UDP packet loss.

Before enabling an audit device, ensure that every cluster node can actually write to the it.

## Supported audit backends

Stronghold supports the same basic audit backends as Vault:

- `file`: For writing to a file.
- `syslog`: For sending logs to the local syslog agent.
- `socket`: For sending logs to a TCP, UDP, or UNIX socket.

### File

This is the most predictable backend for local storage and later forwarding to a centralized logging system.

Example:

```shell
d8 stronghold audit enable file file_path=/var/log/stronghold_audit.log
```

### Syslog

Useful on Unix systems that already rely on a syslog-based logging stack.

Example:

```shell
d8 stronghold audit enable syslog tag="stronghold" facility="AUTH"
```

{{< alert level="warning" >}}
Audit messages can be large, so for `syslog` it is safer to use a reliable transport configuration or enable a `file` backend alongside it.
{{< /alert >}}

### Socket

Useful for integration with external systems over TCP, UDP, or UNIX sockets.

Example:

```shell
d8 stronghold audit enable socket address=127.0.0.1:9090 socket_type=tcp
```

If you use UDP, account for the possibility of silent message loss. For production scenarios, it is best to combine `socket` with another, more reliable audit device.

## Basic operations

Enable an audit device:

```shell
d8 stronghold audit enable file file_path=/var/log/stronghold_audit.log
```

List enabled devices:

```shell
d8 stronghold audit list
```

Disable a device:

```shell
d8 stronghold audit disable file/
```

## Advanced capabilities

Stronghold also supports advanced ways to fine-tune audit logs:

- [Audit log record schema](./log-format/) explains the audit record structure, the `auth`, `request`, and `response` objects, and protection of sensitive values.
- [Audit log filtering](./filtering/) lets you send only selected records to a specific device.
- [Audit field exclusion](./exclusion/) lets you remove selected fields from records before they are written.

These capabilities are useful for separating audit streams, reducing log volume, and limiting retention of selected sensitive fields, but they should be used carefully.

{{< alert level="warning" >}}
Filtering and field exclusion are advanced features. Incorrect configuration can lead to gaps in audit logs or loss of important investigation data.
{{< /alert >}}

## Recommendations

- Enable auditing immediately after Stronghold initialization.
- Use more than one audit device in production.
- Test filter and exclusion rules in a non-production environment.
- Treat auditing as part of the security model, not just a troubleshooting tool.
