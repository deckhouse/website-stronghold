---
title: "Launch and management"
description: "How to validate the Stronghold Agent configuration, run Agent in the foreground for debugging, and run it as a systemd service."
weight: 50
---

## Launch and management

### Validating the configuration

Before running in production, **always** validate the configuration.
A trial run with automatic exit is recommended:

```bash
stronghold agent -config=/etc/stronghold-agent/agent.hcl -exit-after-auth -log-level=debug
```

This command:

1. Checks the HCL configuration syntax.
1. Connects to the Stronghold server.
1. Performs full authentication.
1. Creates files and renders templates.
1. Exits automatically (no need for Ctrl+C).

**Successful result:**

```text
[INFO]  agent: loaded config: path=/etc/stronghold-agent/agent.hcl
[INFO]  agent.auto_auth.approle: authentication successful
[INFO]  agent.sink.file: writing token to: /var/run/stronghold-agent/token
[INFO]  agent: exit after auth set, exiting
```

### Running in development mode

For debugging, you can run Agent in the foreground:

```bash
# Basic run.
stronghold agent -config=/etc/stronghold-agent/agent.hcl

# With a more verbose log level.
stronghold agent -config=/etc/stronghold-agent/agent.hcl -log-level=debug

# Exit after the first successful authentication (for testing).
stronghold agent -config=/etc/stronghold-agent/agent.hcl -exit-after-auth
```

### Running Agent as a systemd service

Create the systemd unit file `/etc/systemd/system/stronghold-agent.service`:

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
ExecStart=/usr/local/bin/stronghold agent -config=/etc/stronghold-agent/agent.hcl
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
KillSignal=SIGTERM
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
# IMPORTANT: add all directories Agent writes to (template.destination, sink file, unix socket, logs).
# The example below is basic; extend it for your configuration (for example, /etc/myapp or /var/lib/myapp):
ReadWritePaths=/var/run/stronghold-agent /var/log/stronghold-agent /etc/myapp
CapabilityBoundingSet=CAP_IPC_LOCK

[Install]
WantedBy=multi-user.target
```

Managing the service:

```bash
# Reload systemd.
sudo systemctl daemon-reload

# Start Agent.
sudo systemctl start stronghold-agent

# Enable autostart.
sudo systemctl enable stronghold-agent

# Check the status.
sudo systemctl status stronghold-agent

# View logs.
sudo journalctl -u stronghold-agent -f

# Reload the configuration (SIGHUP).
sudo systemctl reload stronghold-agent

# Stop.
sudo systemctl stop stronghold-agent
```
