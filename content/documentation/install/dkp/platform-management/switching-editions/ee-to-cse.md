---
title: "Switching Stronghold from EE to CSE"
description: "How to upgrade the stronghold module in Deckhouse Kubernetes Platform from Stronghold EE 1.15.x to Stronghold CSE 1.16.0."
weight: 20
---

Stronghold Enterprise Edition (EE) can be upgraded to Stronghold Certified Security Edition (CSE) in one of the following ways:

- [in a standalone installation](../../../standalone/switching-editions/ee-to-cse/);
- in a DP installation.

{{< alert level="warning" >}}
Upgrading from Stronghold EE 1.15.x to Stronghold CSE 1.16.0 is supported. If you use a Stronghold EE version earlier than 1.15.x, first [update to the latest version of the branch](../../../update/update/).
{{< /alert >}}

{{< alert level="warning" >}}
The service may be temporarily unavailable during the upgrade to Stronghold CSE.
{{< /alert >}}

## Upgrading in a DP installation

{{< alert level="warning" >}}
Switching to Stronghold CSE requires DKP CSE version 1.73 or later.
{{< /alert >}}

Before starting the upgrade, do the following:

1. Check the current Stronghold version:

   ```shell
   stronghold version
   ```

1. Save the unseal keys and the root token to a secure storage. Example:

   ```shell
   d8 k -n d8-stronghold get secret stronghold-keys -o yaml > stronghold-keys.yaml
   chmod 600 stronghold-keys.yaml
   ```

1. Create a backup or a snapshot of the Stronghold cluster. Example:

   {{< tabs name="stronghold_cmd_21012" >}}
   {{% tab name="Stronghold in DKP" %}}

   ```sh
   export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
   d8 stronghold login -method=oidc -path=oidc_deckhouse
   # Or using the root token:
   ## d8 stronghold login -method=token
   d8 stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ```

   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}

   ```sh
   export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
   stronghold login -method=oidc -path=oidc_deckhouse
   # Or using the root token:
   ## stronghold login -method=token
   stronghold operator raft snapshot save stronghold-$(date +%F_%H-%M).snap
   ```

   {{% /tab %}}
   {{< /tabs >}}

   You can check the snapshot with the following command:

   ```sh
   ls -lh ./stronghold-*.snap
   ```

   {{< alert level="info" >}}
   Store the resulting files outside the DP cluster.
   {{< /alert >}}

1. Prepare the Stronghold CSE 1.16.0 package or binary on each node.

1. Make sure the `stronghold` ModuleConfig exists:

   ```shell
   d8 k get mc stronghold -o yaml
   ```

   If the ModuleConfig does not exist, create it:

   ```shell
   cat | d8 k apply -f - <<EOF
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleConfig
   metadata:
     name: stronghold
   spec:
     enabled: false
   EOF
   ```

### Disabling automatic module updates

To disable automatic module updates, do the following:

1. Determine the current release channel of the `stronghold` module:

   ```shell
   d8 k get module stronghold -o jsonpath='{.properties.releaseChannel}'
   ```

   If the command outputs nothing, use the `stable` channel in the next step.

1. Create a ModuleUpdatePolicy resource with the manual update mode, specifying the current release channel in the `releaseChannel` field, and apply it:

   ```shell
   cat <<EOF > stronghold-mup.yaml
   apiVersion: deckhouse.io/v1alpha2
   kind: ModuleUpdatePolicy
   metadata:
     name: stronghold-update-policy
   spec:
     releaseChannel: <STRONGHOLD_RELEASE_CHANNEL>
     update:
       mode: Manual
   EOF
   d8 k apply -f stronghold-mup.yaml
   d8 k get mup stronghold-update-policy
   ```

   Expected output:

   ```console
   NAME                       RELEASE CHANNEL   UPDATE MODE
   stronghold-update-policy   Stable            Manual
   ```

1. Specify the created ModuleUpdatePolicy resource in the `stronghold` ModuleConfig:

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"updatePolicy":"stronghold-update-policy"}}'
   ```

1. Check that the value has been saved:

   ```shell
   d8 k get mc stronghold -o jsonpath='{.spec.updatePolicy}'
   ```

   Expected output:

   ```text
   stronghold-update-policy
   ```

1. Make sure the changes have been applied to the `stronghold` module:

   ```shell
   d8 k get module stronghold -o jsonpath='{.properties.updatePolicy}'
   ```

   Expected output:

   ```text
   stronghold-update-policy
   ```

   {{< alert level="info" >}}
   If DKP CSE 1.67 or earlier is used, or if the module has never been started, the command output is empty. No additional actions are required in this case.
   {{< /alert >}}

### Pre-upgrade checks

1. Check the DKP version. Switching to Stronghold CSE requires DKP CSE version 1.73 or later. You can check the DKP version with the following command:

   ```shell
   d8 k -n d8-system get configmap d8-deckhouse-version-info -o yaml
   ```

   You can also check the DP version in the Deckhouse web interface on the main page of the cluster management console (`https://console.<CLUSTER_DOMAIN>`).

   {{< alert level="info" >}}
   If you need to update DP or switch its edition, use the following instructions:
    - [DP update instructions](/products/kubernetes-platform/documentation/v1/admin/configuration/update/configuration.html);
    - [instructions for switching DKP EE to DKP CSE](/products/kubernetes-platform/documentation/v1/admin/configuration/registry/switching-editions.html).
    {{< /alert >}}

1. Make sure DP is working normally, the leader node is determined, and the queue is empty:

   ```shell
   d8 k -n d8-system get deploy deckhouse
   d8 k -n d8-system get pod -l leader=true
   d8 system queue list
   ```

   Expected output:

   - The `deckhouse` Deployment is `Ready`:

     ```console
     NAME        READY   UP-TO-DATE   AVAILABLE   AGE
     deckhouse   3/3     3            3           1d
     ```

   - There is a working pod in the `Running` status:

     ```console
     NAME                        READY   STATUS    RESTARTS   AGE
     deckhouse-596d944f95-d82m8  2/2     Running   0          1d
     ```

   - The task queue is empty:

     ```console
     Summary:
     - 'main' queue: empty.
     - 121 other queues (0 active, 121 empty): 0 tasks.
     - no tasks to handle.
     ```

1. Make sure the `stronghold` module is enabled and working:

   ```shell
   d8 k get module stronghold
   d8 k get moduleconfig stronghold -o yaml
   ```

   Check that:

   - `Enabled` and `Ready` are `True` for `stronghold`;
   - the `stronghold` ModuleConfig object exists;
   - a license key is specified in `spec.settings.license`.

1. Make sure Stronghold EE version 1.15.x is used. You can do this in one of the following ways:

   - In the Deckhouse web interface: open the main page of the cluster management console (`https://console.<CLUSTER_DOMAIN>`) and check that version 1.15.x with the `ee` suffix is shown at the bottom of the page.

   - Using the pod manifest or image. To do this, run the command:

     ```shell
     d8 k -n d8-stronghold get pod -o yaml | grep version
     ```

     The pod labels must contain version 1.15.x with the `ee` suffix.

   - Using the logs. To do this, run the command:

     ```shell
     d8 k -n d8-stronghold logs stronghold-0 | head -20
     ```

     The startup banner lines must contain the version with the edition:

     ```console
     Version: Stronghold v1.15.0+ee
     ```

### Preparing the module for installation

1. Prepare Stronghold CSE in your image registry (performed on a machine that can push images to the registry). Verify the archive checksum and push the Stronghold CSE module bundle:

   ```shell
   # Verify the archive checksum:
   gost12sum module-stronghold.tar

   # Move the archive to the modules directory:
   mkdir modules
   mv module-stronghold.tar modules/.

   # Image registry host address. For example, 10.129.0.18:5000 or my-registry.com
   export REGISTRY_HOST="<REGISTRY_HOST:PORT>"

   # DKP image registry address. For example, 10.129.0.18:5000/dkp-cse/stable
   export MODULES_MODULE_REPO="${REGISTRY_HOST}/<PATH_TO_DKP_REPO>"

   # Use an account with write permissions to the image registry.
   d8 mirror push modules $MODULES_MODULE_REPO -u <USERNAME> -p <PASSWORD>
   ```

1. Create a ModuleSource for Stronghold CSE. Using your infrastructure tools, make sure the image registry is reachable from each master node. Prepare and apply the ModuleSource:

   ```shell
   # Path to the modules in the image registry.
   MODULES_MODULE_SOURCE="$MODULES_MODULE_REPO/modules"

   # Name of a user with read access to images.
   REGISTRY_USER="<REGISTRY_USER>"

   # User password.
   REGISTRY_PASSWORD="<REGISTRY_PASSWORD>"

   # CA certificate of the domain used for the image registry.
   REGISTRY_CA_CERT="<CA_CERT_FOR_REGISTRY>"

   AUTH_STRING=$(echo -n "${REGISTRY_USER}:${REGISTRY_PASSWORD}" | base64)
   DOCKER_CFG=$(echo -n '{"auths":{"'${REGISTRY_HOST}'":{"username":"'${REGISTRY_USER}'","password":"'${REGISTRY_PASSWORD}'","auth":"'${AUTH_STRING}'"}}}' | base64 -w0)

   cat <<EOF > stronghold-cse-ms.yaml
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleSource
   metadata:
     name: stronghold-cse
   spec:
     registry:
       ca: "$REGISTRY_CA_CERT"
       dockerCfg: $DOCKER_CFG
       repo: $MODULES_MODULE_SOURCE
       scheme: HTTPS
   EOF

   # Check the configuration.
   cat stronghold-cse-ms.yaml

   # Apply the resource.
   d8 k apply -f stronghold-cse-ms.yaml
   ```

1. Check the ModuleSource status:

   ```shell
   d8 k get modulesource stronghold-cse -o yaml
   ```

   Check that there are no errors in the status and that the `status.message` field is empty:

   ```shell
   ...
   status:
     message: ""
   ...
   ```

### Switching the Stronghold module from EE to CSE

Update the `stronghold` ModuleConfig to use the new ModuleSource:

1. Check that the `stronghold` ModuleConfig exists:

   ```shell
   d8 k get mc stronghold -o yaml
   ```

1. Add the `spec.source` field with the `stronghold-cse` value to the ModuleConfig:

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"source":"stronghold-cse"}}'
   d8 k get moduleconfig stronghold -o yaml
   ```

   Check that the changes have been applied:

   - the `stronghold` ModuleConfig contains `spec.source: stronghold-cse`;
   - the other fields in `spec.settings` have not changed.

1. If the module reports a license problem after switching the source, update only the license field (`CSE_LICENSE`):

   ```shell
   d8 k patch moduleconfig stronghold --type merge --patch '{"spec":{"settings":{"license":"<CSE_LICENSE>"},"version":1}}'
   ```

1. If the module was previously disabled, enable it:

   ```shell
   d8 system module enable stronghold
   ```

1. Wait until DP stabilizes and the queue is empty:

   ```shell
   d8 k -n d8-system get pods -l app=deckhouse
   d8 system queue list
   ```

   Check that:

   - the `deckhouse` pod is `Running`/`Ready`;
   - the queue is empty.

Switch the ModuleUpdatePolicy to the `AutoPatch` mode:

1. Change the update mode:

   ```shell
   d8 k patch mup stronghold-update-policy --type merge --patch '{"spec":{"update":{"mode":"AutoPatch"}}}'
   ```

1. Check that the update mode has changed:

   ```shell
   d8 k get mup stronghold-update-policy
   ```

   Expected output:

   ```console
   NAME                       RELEASE CHANNEL   UPDATE MODE
   stronghold-update-policy   Stable            AutoPatch
   ```

Verify that the switch to CSE is complete:

1. Check the module state:

   ```shell
   d8 k get modules stronghold
   ```

   Expected output:

   ```console
   NAME            STAGE                  SOURCE           PHASE   ENABLED   READY
   stronghold      General Availability   stronghold-cse   Ready   True      True
   ```

1. Check the module source:

   ```shell
   d8 k get modulesource stronghold-cse -o yaml
   d8 k get moduleconfig stronghold -o yaml
   d8 k get module stronghold -o yaml
   ```

   Check that:

   - there are no `auth`, `tls`, `x509`, or `timeout` errors;
   - there are no warning events;
   - the `stronghold` ModuleConfig contains `spec.source: stronghold-cse`;
   - the `stronghold` Module contains `properties.source: stronghold-cse`.

1. Check the `stronghold` pods:

   ```shell
   d8 k -n d8-stronghold get po
   ```

   Check that:

   - there are no `ImagePullBackOff`, `ErrImagePull`, or `CrashLoopBackOff` states;
   - the `stronghold-*` pods are in the `Running` state with `2/2` readiness.

### Alternative: installing a new cluster of the required version

If you could not pass the [pre-upgrade checks](#pre-upgrade-checks), that is, bring the current DP cluster and the `stronghold` module to the required versions (DKP CSE 1.73 or later and Stronghold EE 1.15.x), you can move the Stronghold cluster to a new DKP CSE 1.73 cluster by restoring a backup.

To do this, follow these steps:

1. [Back up the cluster data](#upgrading-in-a-dp-installation).
1. Deploy a DKP CSE 1.73 cluster.
1. Starting from the [Disabling automatic module updates](#disabling-automatic-module-updates) section, perform all the steps on the new cluster in order.
1. Restore the backup, the unseal keys, and the root token. Run the commands on the host where the `stronghold-*.snap` backup files and the `stronghold-keys.yaml` keys are located. Example of restoring a Stronghold cluster backup:

   - Restore the snapshot:

     {{< tabs name="stronghold_cmd_70751" >}}
     {{% tab name="Stronghold in DKP" %}}

     ```shell
     export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
     d8 stronghold login -method=oidc -path=oidc_deckhouse
     # Or using the root token:
     ## d8 stronghold login -method=token
     d8 stronghold operator raft snapshot restore -force stronghold-<SNAPSHOT_DATE>.snap
     ```

     {{% /tab %}}
     {{% tab name="Stronghold in Linux" %}}

     ```shell
     export STRONGHOLD_ADDR=https://$(d8 k -n d8-stronghold get ing stronghold -o json | jq -r '.spec.rules[0].host')
     stronghold login -method=oidc -path=oidc_deckhouse
     # Or using the root token:
     ## stronghold login -method=token
     stronghold operator raft snapshot restore -force stronghold-<SNAPSHOT_DATE>.snap
     ```

     {{% /tab %}}
     {{< /tabs >}}

   - Restore the unseal keys and the root token:

     ```shell
     d8 k -n d8-stronghold delete secret stronghold-keys
     d8 k -n d8-stronghold create -f stronghold-keys.yaml
     ```

   - Check the `stronghold` pods:

     ```shell
     d8 k -n d8-stronghold get po
     ```

     Check that:

     - there are no `ImagePullBackOff`, `ErrImagePull`, or `CrashLoopBackOff` states;
     - the `stronghold-*` pods are in the `Running` state with `2/2` readiness.
