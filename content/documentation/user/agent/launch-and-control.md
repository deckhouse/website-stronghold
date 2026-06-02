---
title: "Startup and management"
linkTitle: "Startup and management"
description: "Startup, configuration validation, and management of Stronghold Agent"
weight: 50
---

This page helps you validate the Stronghold Agent configuration, perform a test run, and switch Agent to continuous operation mode.

Before starting in a production environment, first validate the configuration file. Make sure that Agent can connect to the Stronghold server, authenticate successfully, and create the required files.

## Configuration validation

Before starting in a production environment, обязательно validate that the configuration is correct.

To do this, perform a test run with automatic exit:

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

This command does the following:

1. Validates the syntax of the HCL configuration.
1. Connects to the Stronghold server.
1. Performs full authentication.
1. Creates files and templates.
1. Exits automatically.

Example of a successful result:

```text
[INFO]  agent: loaded config: path=/etc/stronghold-agent/agent.hcl
[INFO]  agent.auto_auth.approle: authentication successful
[INFO]  agent.sink.file: writing token to: /var/run/stronghold-agent/token
[INFO]  agent: exit after auth set, exiting
```

This kind of run is convenient as an initial check before running Agent in the background or as a `systemd` service.

## Starting in development mode

For debugging, Stronghold Agent can be started in foreground mode.

Use one of the following startup options:

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl
```

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -log-level=debug
```

```shell
stronghold -config=/etc/stronghold-agent/agent.hcl -exit-after-auth
```

This mode is useful if you need to:

- validate whether authentication succeeds;
- see how templates are rendered;
- make sure that Agent can write tokens and files;
- quickly find configuration errors.

## Starting Stronghold Agent as a systemd service

For continuous operation, Stronghold Agent is usually started as a `systemd` service.

Create the unit file `/etc/systemd/system/stronghold-agent.service`:

```ini
[Unit]
Description=Stronghold Agent
Documentation=https://docs.stronghold.example.com/agent
Requires=network-online.target
After=network-online.target
ConditionFileNotEmpty=/etc/stronghold-agent/agent.hcl

[Service]
Type=notify
User=stronghold-agent
Group=stronghold-agent
ExecStart=/usr/local/bin/stronghold -config=/etc/stronghold-agent/agent.hcl
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
RestartSec=5
LimitNOFILE=65536
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/run/stronghold-agent /var/log/stronghold-agent /etc/myapp
CapabilityBoundingSet=CAP_IPC_LOCK

[Install]
WantedBy=multi-user.target
```

In this example:

- `ConditionFileNotEmpty` checks that the configuration file exists and is not empty;
- `ExecStart` defines the Agent startup command;
- `ExecReload` sends `HUP` to the main process;
- `Restart=on-failure` automatically restarts Agent after a failure;
- `ProtectSystem=strict` and other parameters strengthen process isolation;
- `ReadWritePaths` defines the directories that Agent is allowed to write to.

{{< alert level="warning" >}}
List all directories that Stronghold Agent will write to in `ReadWritePaths`. These may include the directory from `template.destination`, the sink file, the Unix socket, the directory with logs, and application directories if Agent renders configuration there.
{{< /alert >}}

The example unit file above is basic. If needed, add your own paths, such as `/etc/myapp` or `/var/lib/myapp`.

## Service management

After creating the unit file, run the following commands:

```shell
sudo systemctl daemon-reload
```

```shell
sudo systemctl start stronghold-agent
```

```shell
sudo systemctl enable stronghold-agent
```

```shell
sudo systemctl status stronghold-agent
```

```shell
sudo journalctl -u stronghold-agent -f
```

```shell
sudo systemctl reload stronghold-agent
```

```shell
sudo systemctl stop stronghold-agent
```

## Practical recommendations

To ensure that Stronghold Agent starts without issues, consider the following recommendations:

- before starting it as a service, always validate it with `-exit-after-auth`;
- first make sure that Agent can authenticate and write the token;
- verify in advance that the `stronghold-agent` user has access to all required directories;
- if `ProtectSystem=strict` is used, list all writable directories in `ReadWritePaths`;
- to diagnose issues, first run Agent in foreground mode with `-log-level=debug`, and only then move it to `systemd`;
- if you use the `template` block, make sure that the destination directory is writable;
- if you use a Unix socket or a sink file, also add their paths to `ReadWritePaths`.
