---
title: "Platform update"
description: "How Deckhouse Kubernetes Platform and stronghold module updates arrive, checking the current version, preparing for an update, post-update verification, and rollback limitations."
weight: 10
---

In Deckhouse Platform (DP), Stronghold runs as the `stronghold` module and is updated by the DP update mechanism. New module versions arrive through [release channels](../../release-channels/) and are applied according to the configured update mode.

## How updates arrive

- Versions of the `stronghold` module are published to the `Alpha`, `Beta`, `Early Access`, `Stable`, and `Rock Solid` release channels (see [Release channels](../../release-channels/)).
- DP creates a ModuleRelease resource in the cluster for each available module version and applies it automatically, during an update window, or after manual approval.
- By default, the module release channel and update mode are determined by the `deckhouse` module settings. <!-- TODO(verify): whether the stronghold module inherits the release channel and update mode of the deckhouse module when no ModuleUpdatePolicy is set. --> To set them separately for the `stronghold` module, use a ModuleUpdatePolicy resource (see [Manual approval of stronghold module updates](#manual-approval-of-stronghold-module-updates)).

To list the module versions in the cluster:

```shell
d8 k get modulereleases | grep stronghold
```

The changes in each version are described in the [release notes](../../../../release-notes/). Before updating, check the DP version requirements: for example, Stronghold `1.19` requires DP version `1.76` or newer.

## Checking the current version

Check the Stronghold version in one of the following ways:

- Run the command:

  ```shell
  d8 stronghold status
  ```

  The version is shown in the `Version` field.

- Check the pod labels:

  ```shell
  d8 k -n d8-stronghold get pod -o yaml | grep version
  ```

- Check the startup banner in the logs:

  ```shell
  d8 k -n d8-stronghold logs stronghold-0 | head -20
  ```

  Example line with the version and edition:

  ```console
  Version: Stronghold v1.15.0+ee
  ```

To check the current release channel of the module:

```shell
d8 k get module stronghold -o jsonpath='{.properties.releaseChannel}'
```

If the command prints nothing, no release channel is set for the module separately.

## Preparing for an update

Before updating, do the following:

1. Make sure there are no alerts in the cluster and the DP queue is empty:

   ```shell
   d8 system queue list
   ```

1. Make sure the `stronghold` module is `Ready`:

   ```shell
   d8 k get modules stronghold
   ```

1. Check the state of the Stronghold cluster:

   ```shell
   d8 stronghold status
   d8 stronghold operator raft list-peers
   ```

   `Sealed` must be `false`, and all Raft nodes must be present in the list.

1. Save the unseal keys and the root token to secure storage outside the cluster:

   ```shell
   d8 k -n d8-stronghold get secret stronghold-keys -o yaml > stronghold-keys.yaml
   chmod 600 stronghold-keys.yaml
   ```

1. Create a storage snapshot and inspect it (see [Saving a snapshot](../../../../admin/backups/save/) and [Inspecting a snapshot](../../../../admin/backups/inspect/)):

   ```shell
   d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   d8 stronghold operator raft snapshot inspect stronghold-<SNAPSHOT_DATE>.snap
   ```

   Keep the snapshot and the key file outside the DP cluster.

1. Read the [release notes](../../../../release-notes/) for all versions between the current and the target one.

## Platform update

Platform updates are configured in the [`deckhouse`](/modules/deckhouse/configuration.html) ModuleConfig.

To view the current update settings, run:

```shell
d8 k get mc deckhouse -oyaml
```

Example output:

```yaml
# ...
spec:
  settings:
    releaseChannel: Stable
    update:
      windows:
        - days:
            - Mon
          from: "19:00"
          to: "20:00"
# ...
```

### Configuring the update mode

The platform supports three update modes:

- **Automatic, no update windows.** The cluster is updated as soon as a new version appears in the corresponding [release channel](../../release-channels/).
- **Automatic with update windows.** The cluster is updated in the nearest available window after a new version appears in the release channel.
- **Manual.** Applying an update requires [manual actions](../manual-update-mode/).

Example configuration snippet that enables automatic platform updates:

```yaml
update:
  mode: Auto
```

Example configuration snippet that enables automatic platform updates with update windows:

```yaml
update:
  mode: Auto
  windows:
    - from: "8:00"
      to: "15:00"
      days:
        - Tue
        - Sat
```

Example configuration snippet that enables the manual update mode:

```yaml
update:
  mode: Manual
```

### Release channels

The platform uses [five release channels](../../release-channels/) intended for different environments. Platform components can be updated automatically or with manual approval as updates are released in the channels.

For information about the versions available in the release channels, visit [https://releases.deckhouse.io/](https://releases.deckhouse.io/).

To switch to a different release channel, set the `.spec.settings.releaseChannel` parameter in the `deckhouse` module configuration.

Example `deckhouse` module configuration with the `Stable` release channel:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: deckhouse
spec:
  version: 1
  settings:
    releaseChannel: Stable
```

- When you switch to a **more stable** release channel (for example, from `Alpha` to `EarlyAccess`), Deckhouse downloads the release data (in the example, from the `EarlyAccess` channel) and compares it with the existing `DeckhouseRelease` resources in the cluster:
  - Later releases that have not been applied yet (in the `Pending` status) are deleted.
  - If later releases have already been applied (in the `Deployed` status), the release does not change. In this case, the platform stays on that release until a later release appears in the `EarlyAccess` channel.
- When you switch to a **less stable** release channel (for example, from `EarlyAccess` to `Alpha`):
  - Deckhouse downloads the release data (in the example, from the `Alpha` channel) and compares it with the existing `DeckhouseRelease` resources in the cluster.
  - Then the platform updates according to the configured update settings.

To list the platform releases, run:

```shell
d8 k get deckhouserelease
d8 k get modulereleases
```

{{% details summary="How the releaseChannel parameter is used during installation and platform operation" %}}
![How the releaseChannel parameter is used during installation and platform operation](/images/common/deckhouse-update-process.png)
{{% /details %}}

To disable the platform update mechanism, remove the `.spec.settings.releaseChannel` parameter from the `deckhouse` module configuration. In this case, the platform does not check for updates, and patch releases are not applied.

{{< alert level="danger" >}}
Disabling automatic updates is strongly discouraged. It blocks patch releases, which may contain fixes for critical vulnerabilities and bugs.
{{< /alert >}}

### Applying updates immediately

To apply an update immediately, set the `release.deckhouse.io/apply-now: "true"` annotation on the corresponding [DeckhouseRelease](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#deckhouserelease) resource.

{{< alert level="info" >}}
In this case, update windows, [canary release](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#deckhouserelease-v1alpha1-spec-applyafter) settings, and the [manual update mode](../manual-update-mode/) are ignored. The update is applied as soon as the annotation is set.
{{< /alert >}}

Example command that sets the annotation to skip update windows for version `v1.56.2`:

```shell
d8 k annotate deckhousereleases v1.56.2 release.deckhouse.io/apply-now="true"
```

Example resource with the annotation set:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: DeckhouseRelease
metadata:
  annotations:
    release.deckhouse.io/apply-now: "true"
```

## Manual approval of stronghold module updates

To update the `stronghold` module only after manual approval, regardless of the platform update mode:

1. Create a ModuleUpdatePolicy resource with the manual update mode and the required release channel in the `releaseChannel` field:

   ```yaml
   apiVersion: deckhouse.io/v1alpha2
   kind: ModuleUpdatePolicy
   metadata:
     name: stronghold-update-policy
   spec:
     releaseChannel: Stable
     update:
       mode: Manual
   ```

1. Specify the created resource in the `stronghold` ModuleConfig:

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"updatePolicy":"stronghold-update-policy"}}'
   ```

1. Make sure the policy is applied:

   ```shell
   d8 k get module stronghold -o jsonpath='{.properties.updatePolicy}'
   ```

1. When a new version appears, approve the corresponding ModuleRelease. <!-- TODO(verify): the ModuleRelease approval command (modules.deckhouse.io/approved="true" annotation or the approved field) for the current DKP version. -->

To receive only patch versions automatically, switch the policy to the `AutoPatch` mode:

```shell
d8 k patch mup stronghold-update-policy --type merge --patch '{"spec":{"update":{"mode":"AutoPatch"}}}'
```

## Post-update verification

1. Check the module state:

   ```shell
   d8 k get modules stronghold
   ```

   `ENABLED` and `READY` must be `True`.

1. Check the pods:

   ```shell
   d8 k -n d8-stronghold get pods
   ```

   The `stronghold-*` pods must be `Running` with `2/2` readiness. There must be no `ImagePullBackOff`, `ErrImagePull`, or `CrashLoopBackOff` states.

1. Check the Stronghold version and state:

   ```shell
   d8 stronghold status
   d8 stronghold operator raft list-peers
   ```

   The `Version` field must show the target version, and `Sealed` must be `false`.

1. Check that you can log in and read a test secret, and that applications that get secrets from Stronghold work correctly.

## Rollback limitations

- The DP update mechanism does not roll the module back to a previous version automatically. <!-- TODO(verify): whether downgrading the stronghold module is supported and how to do it. -->
- The main way to return to the pre-update state is to restore from the snapshot created during preparation (see [Restoring from a snapshot](../../../../admin/backups/restore/)). <!-- TODO(verify): whether a snapshot created on an older version can be restored to a cluster running a newer version and vice versa. -->
- Data written after the snapshot was created is lost on restore. Plan the update so that the time between creating the snapshot and updating is minimal.
- If you plan to switch editions, do it separately from the version update (see [Switching Stronghold from EE to CSE](../../platform-management/switching-editions/ee-to-cse/)).
