---
title: "Deckhouse Stronghold administrator guide"
linkTitle: "Introduction"
weight: 10
---

This section is intended for administrators of Deckhouse Stronghold.

The platform’s Administrator Guide includes the following sections:

- Running on Linux OS
  - [Running on Linux OS](./standalone/installation/) – a quick start with an example of configuring a high-availability cluster.
  - [Configuration](./standalone/configuration/) – a guide to the Standalone execution configuration parameters.
- Running on Deckhouse Kubernetes Platform
  - [Platform Installation](../install/steps/prepare/) - environment preparation, installation, and initial access setup.
  - [Platform Configuration](../platform-management/node-management/node_settings/node_group/) – cluster node management, networking, storage systems, virtualization, and access control.
  - [Platform Update](../update/update/) – configuring update modes and windows, and manual approval of updates.
  - [Platform Removal](../dkp/removing/) – the process of removing the platform.
- Audit
  - [Introduction](./audit/overview/) - what Stronghold audit logs contain, which backends are supported, and how to configure auditing safely.
  - [Audit log record schema](./audit/log_format/) - audit record structure, key objects, and protection of sensitive data.
  - [Audit log filtering](./audit/audit_filtering/) - select audit records by condition and configure a fallback device.
  - [Audit field exclusion](./audit/audit_exclusion/) - remove selected fields from audit records before they are stored.
- Backups
  - [Introduction](./backups/overview/) - overview of manual and automated backups for Stronghold integrated storage.
  - [Save a storage snapshot](./backups/save/) - create a snapshot manually through CLI or API.
  - [Inspect a snapshot](./backups/inspect/) - locally inspect snapshot contents and basic consistency before restore.
  - [Restore from a snapshot](./backups/restore/) - restore a Stronghold cluster from a saved snapshot.
  - [Automated snapshots](./backups/automated-snapshots/) - configure schedules, storage targets, and status checks for automated backups.
- KV replication
  - [Introduction](./kv-replication/overview/) - pull-based KV1/KV2 replication between Stronghold clusters. English documentation is in development.
- Namespaces
  - [Introduction](./namespaces/overview/) - isolate configuration and secrets between namespaces, manage them through CLI and API, and use Namespace API Lock.
- Cryptographic algorithms
  - [Introduction](./cryptography/overview/) - overview of TLS, storage encryption, HSM, and the algorithms available in PKI and Transit.
- Plugins
  - [Introduction](./plugins/overview/) - overview of built-in and external Stronghold plugins and the differences between Standalone and DKP.
  - [Plugins in Standalone](./plugins/standalone/) - plugin directory, registration, versioning, and mounting of external plugins on Linux servers.
  - [Plugins in DKP](./plugins/dkp/) - plugin delivery through `ModuleConfig`, registration, and enablement in Deckhouse Kubernetes Platform.
- KMS and HSM
  - [HSM support](./kms-hsm/hsm/) - PKCS11-based HSM integration for auto-unseal and root key protection; currently supported only for Standalone installations.
  - [Yandex Cloud KMS](./kms-hsm/yandexcloudkms/) - configure `seal "yandexcloudkms"` for auto-unseal and root key protection; currently supported only for Standalone installations.
  - [Double encryption](./kms-hsm/sealwrap/) - the `seal wrap` mechanism that adds an extra encryption layer for critical data.

If you have any questions, you can ask for assistance in our [Telegram channel](https://t.me/deckhouse). We will be happy to help and provide guidance.

If you are using the Enterprise edition, you can also email us at&nbsp;<a href="mailto:support@deckhouse.io">support@deckhouse.io</a> for additional support.
