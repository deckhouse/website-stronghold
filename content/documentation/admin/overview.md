---
title: "Deckhouse Stronghold administrator guide"
linkTitle: "Overview"
weight: 10
---

This section is intended for Deckhouse Stronghold administrators.

The administrator guide includes the following sections:

- Audit
  - ["Overview"](./audit/overview/): What Stronghold audit logs contain, which backends are supported, and how to configure auditing safely.
  - ["Audit log record schema"](./audit/log-format/): Audit record structure, key objects, and protection of sensitive data.
  - ["Audit log filtering"](./audit/filtering/): Selecting audit records by condition and configure a fallback device.
  - ["Audit field exclusion"](./audit/exclusion/): Removing selected fields from audit records before they are stored.

- Backup and restore
  - ["Overview"](./backups/overview/): Overview of manual and automated backups for Stronghold integrated storage.
  - ["Creating a snapshot"](./backups/save/): Creating a snapshot manually through CLI or API.
  - ["Inspecting a snapshot"](./backups/inspect/): Inspecting snapshot contents and basic consistency locally.
  - ["Restoring from a snapshot"](./backups/restore/): Restoring a Stronghold cluster from a saved snapshot.
  - ["Automated snapshots"](./backups/automated-snapshots/): Configuring schedules, storage targets, and status checks for automated backups.

- KMS and HSM
  - ["HSM support"](./kms-hsm/hsm/): PKCS11-based HSM integration for auto-unseal and root key protection; currently supported only for Standalone installations.
  - ["Yandex Cloud KMS"](./kms-hsm/yandexcloudkms/): Configure `seal "yandexcloudkms"` for auto-unseal and root key protection; currently supported only for Standalone installations.
  - ["Double encryption"](./kms-hsm/sealwrap/): The `seal wrap` mechanism that adds an extra encryption layer for critical data.

- Replication
  - ["Overview"](./replication/overview/): Native Performance and DR replication between Stronghold clusters.
  - ["Architecture: CE and EE"](./replication/architecture/): Node layers and how their order shapes replication behavior.
  - ["Performance replication"](./replication/performance/): Read scaling, secondary setup, and path filters.
  - ["Disaster recovery"](./replication/disaster-recovery/): Hot standby, failover, and the promote ceremony.
  - ["Performance standby"](./replication/performance-standby/): Serving reads from non-active HA nodes within a cluster.
  - ["KV1/KV2 replication"](./replication/kv-replication/): Pull-based KV1/KV2 replication between Stronghold clusters. English documentation is in development.

- Namespaces
  - ["Overview"](./namespaces/overview/): Isolate configuration and secrets between namespaces, manage them through CLI and API, and use Namespace API Lock.

- Cryptographic algorithms
  - ["Overview"](./cryptography/overview/): Overview of TLS, storage encryption, HSM, and the algorithms available in PKI and Transit.

- Extensions and integrations
  - ["Overview"](./plugins/overview/): Overview of built-in and external Stronghold plugins and the differences between Standalone and DKP.
  - ["Plugins in Standalone"](./plugins/standalone/): Plugin directory, registration, versioning, and mounting of external plugins on Linux servers.
  - ["Plugins in DKP"](./plugins/dkp/): Plugin delivery through `ModuleConfig`, registration, and enablement in Deckhouse Kubernetes Platform.
