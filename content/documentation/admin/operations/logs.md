---
title: "Server logs"
description: "Stronghold server log levels and format, where to find logs in Linux and DKP, and how to change the log level without a restart."
weight: 20
---

Server logs describe the operation of the Stronghold process itself: startup, unsealing, Raft leader elections, plugin and connection errors. They do not replace [audit logs](../../audit/overview/), which record client requests.

## Log levels

The level is set with the `log_level` parameter in the configuration file or the `VAULT_LOG_LEVEL` environment variable (see [Configuration](../../../install/standalone/configuration/)). Supported values in increasing verbosity:

| Level | When to use |
| --- | --- |
| `error` | Errors only |
| `warn` | Errors and warnings |
| `info` | Default, recommended for production |
| `debug` | Troubleshooting, short-term |
| `trace` | Detailed troubleshooting, short-term. Produces a large volume of records |

In DP, `log_level` is not set in the module configuration, so the default level (`info`) is used. To raise it temporarily, use the [`/sys/loggers`](#via-the-api-without-a-restart) endpoint.

{{< alert level="warning" >}}
Do not keep `debug` and `trace` enabled for long: they increase log volume and disk load.
{{< /alert >}}

## Log format

The `log_format` parameter accepts `standard` (default) and `json`. Use `json` to ship logs to a centralized logging system:

```hcl
log_level  = "info"
log_format = "json"
```

A JSON record example:

```json
{"@level":"info","@message":"attempting to join possible raft leader node","@module":"core","@timestamp":"2025-10-20T10:54:02.578963Z","leader_addr":"https://stronghold-0.stronghold-internal:8300"}
```

In DP, `log_format = "json"` is hard-coded in the module configuration.

To write logs to a file, set `log_file` and the rotation parameters `log_rotate_duration`, `log_rotate_bytes`, `log_rotate_max_files`.

## Viewing logs

{{< tabs name="stronghold_logs_view" >}}
{{% tab name="Stronghold in Linux" %}}

When running as the `stronghold.service` systemd unit, logs go to journald:

```shell
journalctl -u stronghold.service -f
```

Logs for a specific period:

```shell
journalctl -u stronghold.service --since "1 hour ago"
```

{{% /tab %}}
{{% tab name="Stronghold in DKP" %}}

Stronghold runs in the `d8-stronghold` namespace. List the pods:

```shell
d8 k -n d8-stronghold get pods -o wide
```

The `stronghold-N` pod has the `stronghold` container, the `kube-rbac-proxy` sidecar (metrics) and the `plugin-fetcher` init container (downloads plugins). Specify the main container with `-c`:

```shell
d8 k -n d8-stronghold logs stronghold-0 -c stronghold -f
```

Logs of the previous container instance after a restart:

```shell
d8 k -n d8-stronghold logs stronghold-0 -c stronghold --previous
```

To get logs of the active node (for example, when diagnosing automatic snapshot alerts), use the `stronghold-active` service:

```shell
d8 k -n d8-stronghold logs svc/stronghold-active -c stronghold
```

{{% /tab %}}
{{< /tabs >}}

## Changing the log level

### Via the configuration file

1. Change the `log_level` value in the configuration file.
1. Send `SIGHUP` to the process:

   ```shell
   systemctl reload stronghold
   ```

   The command uses `ExecReload=/bin/kill -HUP $MAINPID` from the systemd unit. On `SIGHUP`, the log level is updated and CLI flags and environment variables are ignored. Not all subsystems (for example, plugins) support changing the level dynamically.

### Via the API without a restart

The [`/sys/loggers`](../../../reference/api/system/#post-sysloggers) endpoint changes the log level on a running node. The request applies to the node it is sent to.

Raise the level for all subsystems:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request POST \
  --data '{"level": "debug"}' \
  "${STRONGHOLD_ADDR}/v1/sys/loggers"
```

Raise the level for a single subsystem (for example, `core`):

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request POST \
  --data '{"level": "trace"}' \
  "${STRONGHOLD_ADDR}/v1/sys/loggers/core"
```

Revert to the level from the configuration:

```shell
curl \
  --header "X-Vault-Token: ${STRONGHOLD_TOKEN}" \
  --request DELETE \
  "${STRONGHOLD_ADDR}/v1/sys/loggers"
```

Subsystem names match the logger names in log lines (for example, `core`, `expiration`, `raft`, `audit`). `GET /v1/sys/loggers` returns the current list for a node.

### Streaming logs

The `d8 stronghold monitor` command connects to the [`/sys/monitor`](../../../reference/api/system/#get-sysmonitor) endpoint and streams node logs in real time at the given level without changing the configuration:

```shell
d8 stronghold monitor -log-level=debug
```
