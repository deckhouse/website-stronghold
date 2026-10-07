---
title: "Updating Deckhouse Stronghold Certified Security Edition"
description: "How to update Stronghold CSE in Deckhouse Kubernetes Platform from version 1.16.0 to 1.16.25."
hidden: true
---

This guide describes how to update the "Deckhouse Stronghold Certified Security Edition" software (hereinafter, Stronghold CSE) deployed in Deckhouse Kubernetes Platform from v1.16.0 to v1.16.25.

## Minimum requirements

The update requires Deckhouse Kubernetes Platform Certified Security Edition (DKP CSE) v1.73.4 or later.

## Update procedure

To update Stronghold CSE, follow these steps:

1. Before updating, make sure there are no alerts in the cluster and no errors in the DKP CSE queues.

   To check the state of the queues, run the following command:

   ```shell
   d8 system queue list
   ```

   Example output when there are no errors and no active tasks in the queue:

   ```text
   Summary:
   - 'main' queue: empty.
   - 103 other queues (0 active, 103 empty): 0 tasks.
   - no tasks to handle.
   ```

1. Make sure you have access to the Stronghold CSE unseal keys that were saved during installation in the `d8-stronghold/stronghold-keys` Secret.

1. Upload the update provided to you to your container image registry.

   Copy the delivery files from the USB flash drive or DVD to a computer that has access to the container image registry.

   ```shell
   d8 mirror pull \
     ${PACKAGE_DIR_PATH} \
     --no-packages \
     --no-installer \
     --no-security-db \
     --no-platform \
     --include-module stronghold@=v1.16.25 \
     --gost-digest \
     --source="registry-cse.deckhouse.ru/stronghold/cse" \
     --license="${LICENSE_KEY}"
   ```

   Where:

   * `${PACKAGE_DIR_PATH}` is the directory to save the images to;
   * `${LICENSE_KEY}` is the license key for accessing the public Stronghold CSE container image registry.

   Use the `d8 tools gostsum` command to get the checksums of the image archive files and compare them with the values in the corresponding files with the `.gostsum` suffix. The checksums must match.

   Upload the data to the container image registry where Stronghold CSE is hosted by running the following command (change the path to the archive file or image directory):

   ```shell
   d8 mirror push <PATH> <REGISTRY_URL>/<REGISTRY_PATH> \
     --registry-login=<USERNAME> --registry-password=<PASSWORD>
   ```

   Where:

   * `<PATH>` is the delivery directory containing the archives with the Stronghold CSE delivery images;
   * `<REGISTRY_URL>` is the address of the container image registry in the local network;
   * `<REGISTRY_PATH>` is the path in the container image registry to which the Stronghold CSE images are uploaded. The examples below use the `/stronghold/cse` path;
   * `<USERNAME>` is the user name for authenticating in the container image registry;
   * `<PASSWORD>` is the user password for authenticating in the container image registry.

1. Wait for the pods in the `d8-stronghold` namespace to restart.

1. If Stronghold CSE is sealed after the restart, unseal it using the saved unseal keys:

   ```shell
   d8 k -n d8-stronghold exec stronghold-0 -it -- stronghold operator unseal
   ```

1. After the pods restart, check the Stronghold CSE version with the following command:

   {{< tabs name="stronghold_cmd_594" >}}
   {{% tab name="Stronghold in DKP" %}}

   ```shell
   d8 stronghold status
   ```

   {{% /tab %}}
   {{% tab name="Stronghold in Linux" %}}

   ```shell
   stronghold status
   ```

   {{% /tab %}}
   {{< /tabs >}}

   Example output:

   ```text
   Key                     Value
   ---                     -----
   Seal Type               shamir
   Initialized             true
   Sealed                  false
   Total Shares            1
   Threshold               1
   Version                 1.16.25+ee
   Build Date              2026-08-07T14:52:06Z
   Storage Type            raft
   Cluster Name            stronghold-cluster-8c677db0
   Cluster ID              849ab3e4-261c-b9ec-d857-c542dfadfe01
   HA Enabled              true
   HA Cluster              https://stronghold-0.stronghold-internal:8301
   HA Mode                 active
   Raft Committed Index    325470
   Raft Applied Index      325470
   Last WAL                122365
   ```

<!-- TODO(verify): the example output (copied from the RU source) shows `Version 1.16.25+ee` for a CSE build; confirm the expected version suffix for Stronghold CSE. -->
