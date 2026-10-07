---
title: "Removing"
description: "Removing Stronghold installed in Linux: revoking leases, taking a backup, stopping the service, and deleting data and configuration."
weight: 50
---

{{< alert level="danger" >}}
Deleting Stronghold data is irreversible. Without a storage snapshot and the unseal or recovery keys, secrets cannot be recovered.
{{< /alert >}}

## Preparation

1. Make sure Stronghold is no longer in use: applications, CI/CD, and other systems have moved to another secrets store or no longer need access.
1. Take a final storage snapshot and keep it together with the unseal or recovery keys for the period defined by your data retention policy (see [Creating a snapshot](../../../admin/backups/save/)):

   ```shell
   d8 stronghold operator raft snapshot save final.snap
   ```

1. Revoke dynamic secret leases so that Stronghold removes the credentials it issued in external systems (databases, LDAP, Kubernetes). List secrets engines:

   ```shell
   d8 stronghold secrets list
   ```

   For each engine that issues dynamic secrets, revoke leases by prefix:

   ```shell
   d8 stronghold lease revoke -prefix database/
   ```

   Check that no leases remain:

   ```shell
   d8 stronghold list sys/leases/lookup/database/creds/
   ```

   Disabling a secrets engine with `d8 stronghold secrets disable <PATH>` also revokes all its leases.

1. If Stronghold was used as a [PKI](../../../user/secrets-engines/pki/) certificate authority, issue CA certificates in another system in advance and replace trusted CAs on clients. After Stronghold is removed, CRLs and OCSP responses stop being updated.
1. Preserve audit logs according to your retention policy if they are stored on Stronghold nodes.

## Stopping and removing

Perform on each cluster node:

1. Stop and disable the service:

   ```shell
   systemctl stop stronghold
   systemctl disable stronghold
   ```

1. Delete the systemd unit and reload the systemd configuration:

   ```shell
   rm /etc/systemd/system/stronghold.service
   systemctl daemon-reload
   ```

1. Delete data, configuration, certificates, and the binary. The default paths match [Installation](../installation/); if the configuration uses other directories (the `path` parameter of the `storage` block, `log_file`, audit files), delete them too:

   ```shell
   rm -rf /opt/stronghold
   ```

   For media that stored Raft data, follow your organization's secure data destruction procedure.

1. Delete the system user:

   ```shell
   userdel stronghold
   ```

1. Remove firewall rules for ports `8200` and `8201`, DNS records, and load balancer settings created for Stronghold.

## After removal

- Destroy the unseal or recovery key shares and root tokens if the final snapshot does not need to be kept; otherwise, store them together with the snapshot.
- If auto unseal was used, delete the Stronghold keys in the HSM or KMS only after the snapshot retention period expires: without them, the snapshot cannot be unsealed.
- Revoke the credentials Stronghold used to connect to external systems (for example, secrets engine accounts in databases).
