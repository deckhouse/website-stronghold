---
title: "Switching Stronghold to Stronghold EE"
description: "Switching from base Stronghold to Stronghold EE in DKP by specifying a license key in ModuleConfig."
weight: 10
params:
  relatedLinks:
    - title: "Editions"
      url: ../../../../../about/editions/
    - title: "Stronghold configuration"
      url: ../../../configuration/
    - title: "Switching Stronghold from EE to CSE"
      url: ../ee-to-cse/
    - title: "Creating a snapshot"
      url: ../../../../../admin/backups/save/
---

The `stronghold` module is switched from base Stronghold to Stronghold EE by specifying a license key in the [`spec.settings.license`](/modules/stronghold/stable/configuration.html#parameters-license) parameter of the `stronghold` ModuleConfig resource. You do not need to reinstall the module or migrate data: secrets, policies, secrets engines, and configured auth methods are preserved.

{{< alert level="warning" >}}
Stronghold EE is licensed separately and is available only in commercial DP editions. Switching to Stronghold EE is not possible in DP CE. For details, see [Editions](../../../../../about/editions/).
{{< /alert >}}

{{< alert level="warning" >}}
After the license key is applied, in the `Automatic` mode, the `stronghold-*` pods are recreated one by one on the Stronghold EE image, so service availability stays the same. In the `Manual` mode, recreate the pods manually. Nevertheless, perform the switch strictly during a maintenance window.
{{< /alert >}}

## Before switching

1. Make sure the cluster uses a commercial DP edition:

   ```shell
   d8 k -n d8-system get configmap d8-deckhouse-version-info -o jsonpath='{.data.data\.json}'
   ```

   The `json` object must contain an edition other than `CE`. Example output:

   ```console
   { "channel":"Stable", "version":"v1.76.0", "edition":"EE" }
   ```

   You can also see the DP edition and version in the Deckhouse web interface on the main page of the cluster management console (`https://console.<CLUSTER_DOMAIN>`).

   {{< alert level="info" >}}
   If you need to change the DP edition, follow the [instructions for switching the DP edition](/products/kubernetes-platform/documentation/v1/admin/configuration/registry/switching-editions.html).
   {{< /alert >}}

1. Make sure the `stronghold` module is enabled and working:

   ```shell
   d8 k get module stronghold
   d8 k get moduleconfig stronghold -o yaml
   ```

   Check that:

   - `Enabled` and `Ready` are `True` for the `stronghold` module;
   - the `stronghold` ModuleConfig object exists.

1. Save the unseal keys and the root token to a secure storage:

   ```shell
   d8 k -n d8-stronghold get secret stronghold-keys -o yaml > stronghold-keys.yaml
   chmod 600 stronghold-keys.yaml
   ```

1. Create a [storage snapshot](../../../../admin/backups/save/). The example below applies only to the `Ingress` inlet (it reads the host from the `stronghold` Ingress) and the `Automatic` mode (OIDC login through Dex); for other inlets and modes, set `STRONGHOLD_ADDR` yourself and log in with another method:

   ```shell
   export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
   d8 stronghold login -method=oidc -path=oidc_deckhouse
   # Or using the root token:
   ## d8 stronghold login -method=token
   d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ```

   You can check the snapshot with the following command:

   ```shell
   ls -lh ./stronghold-*.snap
   ```

   {{< alert level="info" >}}
   Store the resulting files outside the DP cluster.
   {{< /alert >}}

## Specifying the license key

1. Get a Stronghold EE license key from the product vendor. The `license` parameter must be a string of exactly 32 characters; otherwise, the configuration is rejected with the `Invalid license format` error.

1. Add the key to the `spec.settings.license` parameter of the `stronghold` ModuleConfig resource:

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"version":1,"settings":{"license":"<STRONGHOLD_EE_LICENSE_KEY>"}}}'
   ```

   You can get the same result by editing the ModuleConfig with `d8 k edit mc stronghold`. Example of the resulting manifest:

   ```yaml
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleConfig
   metadata:
     name: stronghold
   spec:
     enabled: true
     version: 1
     settings:
       license: <STRONGHOLD_EE_LICENSE_KEY>
       management:
         mode: Automatic
         administrators:
         - type: Group
           name: admins
   ```

   {{< alert level="warning" >}}
   The license key is stored in ModuleConfig in plain text. If necessary, restrict access to the resource using the [role model](../../access-control/role-model/).
   {{< /alert >}}

1. Wait for the Stronghold pods to be recreated:

   ```shell
   d8 k -n d8-stronghold get po -w
   ```

## Verifying the switch

1. Make sure the key has been saved in the module configuration:

   ```shell
   d8 k get moduleconfig stronghold -o jsonpath='{.spec.settings.license}'
   ```

1. Check the `stronghold` pods:

   ```shell
   d8 k -n d8-stronghold get po
   ```

   Check that:

   - there are no `ImagePullBackOff`, `ErrImagePull`, or `CrashLoopBackOff` states;
   - the `stronghold-*` pods are in the `Running` state with `2/2` readiness.

1. Check the edition in the startup banner lines:

   ```shell
   d8 k -n d8-stronghold logs stronghold-0 | head -20
   ```

   The output must contain a version with the `ee` suffix. For example:

   ```console
   Version: Stronghold v1.19.0+ee
   ```

1. Make sure the storage is unsealed (as above, the example applies to the `Ingress` inlet):

   ```shell
   export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
   d8 stronghold status
   ```

   In the output, `Sealed` must be `false`.

After the switch, Stronghold EE features become available: the audit log, namespaces, replication between clusters, automated snapshots, Managed Keys, and management of roles and access policies via the web interface. For the full list of differences, see [Editions](../../../../about/editions/).

## Reverting to base Stronghold

{{< alert level="warning" >}}
After the license key is removed, Stronghold EE features stop working. Stop using them in advance; otherwise, the module may fail to start. In particular, set `enableAuditLog: false`, stop replication between clusters, and delete the namespaces you created.
{{< /alert >}}

1. Disable the audit log first: set `enableAuditLog: false` in the `stronghold` ModuleConfig. Setting `enableAuditLog: true` without a license is rejected, so do this before removing the license.

1. Remove the `spec.settings.license` parameter entirely:

   ```shell
   d8 k patch moduleconfig stronghold --type json --patch '[{"op":"remove","path":"/spec/settings/license"}]'
   ```

1. Wait for the pods to be recreated and make sure the banner shows a version without the `ee` suffix:

   ```shell
   d8 k -n d8-stronghold logs stronghold-0 | head -20
   ```
