---
title: "Adding a node"
description: "Adding static nodes to a Deckhouse Kubernetes Platform cluster manually or using Cluster API Provider Static, and removing nodes from the cluster."
weight: 20
---

## Adding a static node to the cluster

<span id="adding-a-node-to-the-cluster"></span>

You can add a static node manually or using Cluster API Provider Static.

### Adding a static node manually

To add a bare-metal server to the cluster as a static node, follow these steps:

1. Use an existing NodeGroup custom resource or create a new one ([NodeGroup](/modules/node-manager/cr.html#nodegroup)). The [`nodeType`](/modules/node-manager/cr.html#nodegroup-v1-spec-nodetype) parameter of the NodeGroup custom resource for static nodes must be `Static` or `CloudStatic`.
1. Get the Base64-encoded script code for adding and configuring the node.

   Example of getting the Base64-encoded script code for adding a node to the `worker` NodeGroup:

   ```shell
   NODE_GROUP=worker
   d8 k -n d8-cloud-instance-manager get secret manual-bootstrap-for-${NODE_GROUP} -o json | jq '.data."bootstrap.sh"' -r
   ```

1. Pre-configure the new node according to the specifics of your environment:

- add the required mount points to the `/etc/fstab` file (NFS, Ceph, etc.);
- install the required packages (for example, `ceph-common`);
- configure network connectivity between the new node and the other cluster nodes.

Connect to the new node via SSH and run the following command, inserting the Base64 string obtained in step 2:

   ```shell
   echo <BASE64-SCRIPT-CODE> | base64 -d | bash
   ```

### Adding a static node using Cluster API Provider Static

Example of adding a static node to the cluster using [Cluster API Provider Static (CAPS)](../node-group/#configuring-a-node-via-caps):

**Allocate a server with an installed OS**, configure network connectivity, etc., and, if necessary, install OS-specific packages and add the mount points that the node will need.

* Create a user (`caps` in the example) that can run `sudo` by running the following commands on the server:

    ```shell
    useradd -m -s /bin/bash caps 
    usermod -aG sudo caps
    ```

* Allow the user to run commands via `sudo` without a password. To do this, add the following line to the `sudo` configuration on the server (by editing the `/etc/sudoers` file, running `sudo visudo`, or in another way):

    ```text
    caps ALL=(ALL) NOPASSWD: ALL
    ```

* Generate an SSH key pair with an empty passphrase on the server:

    ```shell
    ssh-keygen -t rsa -f caps-id -C "" -N ""
    ```

  The public and private keys of the `caps` user will be saved to the `caps-id.pub` and `caps-id` files in the current directory on the server.

* Add the generated public key to the `/home/caps/.ssh/authorized_keys` file of the `caps` user by running the following commands in the key directory on the server:

    ```shell
    mkdir -p /home/caps/.ssh 
    cat caps-id.pub >> /home/caps/.ssh/authorized_keys 
    chmod 700 /home/caps/.ssh 
    chmod 600 /home/caps/.ssh/authorized_keys
    chown -R caps:caps /home/caps/
    ```

  * In Astra Linux operating systems, when using the Parsec mandatory integrity control module, configure the maximum integrity level for the `caps` user:

```shell
pdpl-user -i 63 caps
```

**Create an `SSHCredentials` resource in the cluster.**

* To access the server being added, the CAPS component needs the private key of the `caps` service user. The key in Base64 format is added to the SSHCredentials resource.

In the user key directory on the server, run the following command to get the private key in Base64 format:

```shell
base64 -w0 caps-id
```

* On any computer configured to manage the cluster, create an environment variable with the Base64-encoded private key obtained in the previous step (add a space at the beginning of the command so that the key is not saved in the command history):

```shell
CAPS_PRIVATE_KEY_BASE64=<PRIVATE_KEY_IN_BASE64>
```

* Create an SSHCredentials resource with the service user name and its private key:

```shell
d8 k create -f - <<EOF
apiVersion: deckhouse.io/v1alpha2
kind: SSHCredentials
metadata:
  name: static-0-access
spec:
  user: caps
  privateSSHKey: "${CAPS_PRIVATE_KEY_BASE64}"
EOF
```

**Create a StaticInstance resource in the cluster**:

The StaticInstance resource defines the IP address of the static node server and the credentials for accessing the server:

```shell
d8 k create -f - <<EOF
apiVersion: deckhouse.io/v1alpha2
kind: StaticInstance
metadata:
  name: static-0
spec:
  # Specify the IP address of the static node server.
  address: "<SERVER-IP>"
  credentialsRef:
    kind: SSHCredentials
    name: static-0-access
EOF
```

**Create a NodeGroup resource in the cluster**:

```shell
d8 k create -f - <<EOF
apiVersion: deckhouse.io/v1
kind: NodeGroup
metadata:
  name: worker
spec:
  nodeType: Static
  staticInstances:
    count: 1
EOF
```

**Wait for the Ready state**:

The NodeGroup status should show 1 node in the Ready column:

```shell
d8 k get ng worker
NAME     TYPE     READY   NODES   UPTODATE   INSTANCES   DESIRED   MIN   MAX   STANDBY   STATUS   AGE    SYNCED
worker   Static   1       1       1                                                                 15m   True
```

### Adding a static node using Cluster API Provider Static and label selector filters

<span id="caps-with-label-selector"></span>

To connect different StaticInstances to different NodeGroups, you can use a label selector specified in the NodeGroup and in the StaticInstance metadata.

As an example, consider distributing 3 static nodes across 2 NodeGroups: 1 node is added to the worker group and 2 nodes to the front group.

1. Prepare the required resources (3 servers) and create SSHCredentials resources for them, similarly to steps 1 and 2 of the [previous example](#adding-a-static-node-using-cluster-api-provider-static).

1. Create two NodeGroup resources in the cluster:

   Specify a labelSelector so that only servers matching it are connected to the NodeGroup.

   ```shell
   d8 k create -f - <<EOF
   apiVersion: deckhouse.io/v1
   kind: NodeGroup
   metadata:
     name: front
   spec:
     nodeType: Static
     staticInstances:
       count: 2
       labelSelector:
         matchLabels:
           role: front
   ---
   apiVersion: deckhouse.io/v1
   kind: NodeGroup
   metadata:
     name: worker
   spec:
     nodeType: Static
     staticInstances:
       count: 1
       labelSelector:
         matchLabels:
           role: worker
   EOF
   ```

1. Create StaticInstance resources in the cluster.

   Specify the actual IP addresses of the servers and set the role label in the metadata:

   ```shell
   d8 k create -f - <<EOF
   apiVersion: deckhouse.io/v1alpha1
   kind: StaticInstance
   metadata:
     name: static-front-1
     labels:
       role: front
   spec:
     address: "<SERVER-FRONT-IP1>"
     credentialsRef:
       kind: SSHCredentials
       name: front-1-credentials
   ---
   apiVersion: deckhouse.io/v1alpha1
   kind: StaticInstance
   metadata:
     name: static-front-2
     labels:
       role: front
   spec:
     address: "<SERVER-FRONT-IP2>"
     credentialsRef:
       kind: SSHCredentials
       name: front-2-credentials
   ---
   apiVersion: deckhouse.io/v1alpha1
   kind: StaticInstance
   metadata:
     name: static-worker-1
     labels:
       role: worker
   spec:
     address: "<SERVER-WORKER-IP>"
     credentialsRef:
       kind: SSHCredentials
       name: worker-1-credentials
   EOF
   ```

Result:

```shell
d8 k get ng
NAME     TYPE     READY   NODES   UPTODATE   INSTANCES   DESIRED   MIN   MAX   STANDBY   STATUS   AGE    SYNCED
master   Static   1       1       1                                                               1h     True
front    Static   2       2       2                                                               1h     True
```

## How to tell if something went wrong?

If a node in a NodeGroup is not updated (the `UPTODATE` value in the output of `d8 k get nodegroup` is less than the `NODES` value), or you suspect other problems that may be related to the `node-manager` module, check the logs of the `bashible` service. The `bashible` service runs on every node managed by the `node-manager` module.

To view the `bashible` service logs, run the following command on the node:

```shell
journalctl -fu bashible
```

Example output when all required actions have been completed:

```console
May 25 04:39:16 kube-master-0 systemd[1]: Started Bashible service.
May 25 04:39:16 kube-master-0 bashible.sh[1976339]: Configuration is in sync, nothing to do.
May 25 04:39:16 kube-master-0 systemd[1]: bashible.service: Succeeded.
```

## Removing a node from the cluster

<span id='how-to-remove-a-node-from-node-manager-control'></span>

{{< alert level="info" >}}
These instructions apply both to a node configured manually (using the bootstrap script) and to a node configured using CAPS.
{{< /alert >}}

To remove a node from the cluster and clean up the server (VM), run the following command on the node:

```shell
bash /var/lib/bashible/cleanup_static_node.sh --yes-i-am-sane-and-i-understand-what-i-am-doing
```

### How to clean up a node for adding it to a cluster later?

This is only needed if you need to move a static node from one cluster to another. Keep in mind that these operations delete local storage data. If you only need to change the NodeGroup, follow [the instructions for changing the NodeGroup of a static node](#how-to-change-the-nodegroup-of-a-static-node).

{{< alert level="warning" >}}
If the node being cleaned up has LINSTOR/DRBD storage pools, follow the `sds-replicated-volume` module [instructions](/modules/sds-replicated-volume/stable/faq.html#how-to-evict-drbd-resources-from-a-node) to evict resources from the node and delete the LINSTOR/DRBD node.
{{< /alert >}}

1. Delete the node from the Kubernetes cluster:

   ```shell
   d8 k drain <node> --ignore-daemonsets --delete-emptydir-data
   d8 k delete node <node>
   ```

1. Run the cleanup script on the node:

   ```shell
   bash /var/lib/bashible/cleanup_static_node.sh --yes-i-am-sane-and-i-understand-what-i-am-doing
   ```

1. After the reboot, the node can be added to another cluster.

## FAQ

### Can a StaticInstance be deleted?

A StaticInstance in the `Pending` state can be deleted without any issues.

To delete a StaticInstance in any state other than `Pending` (`Running`, `Cleaning`, `Bootstraping`):

1. Add the `"node.deckhouse.io/allow-bootstrap": "false"` label to the StaticInstance.

   Example command for adding the label:

   ```shell
   d8 k label staticinstance d8cluster-worker node.deckhouse.io/allow-bootstrap=false
   ```

1. Wait until the StaticInstance switches to the `Pending` status.

   To check the StaticInstance status, use the command:

   ```shell
   d8 k get staticinstances
   ```

1. Delete the StaticInstance.

   Example command for deleting a StaticInstance:

   ```shell
   d8 k delete staticinstance d8cluster-worker
   ```

1. Decrease the `NodeGroup.spec.staticInstances.count` parameter value by 1.
1. Wait for the NodeGroup to reach the `Ready` state.

### How to change the IP address of a StaticInstance?

You cannot change the IP address in a StaticInstance resource. If a StaticInstance has an incorrect address, [delete the StaticInstance](#can-a-staticinstance-be-deleted) and create a new one.

### How to migrate a manually configured static node to CAPS management?

[Clean up the node](#how-to-clean-up-a-node-for-adding-it-to-a-cluster-later), then [add](#adding-a-static-node-using-cluster-api-provider-static) the node under CAPS management.

### How to change the NodeGroup of a static node?

<span id='how-to-change-the-nodegroup-of-a-static-node'></span>

If the node is managed by [CAPS](../node-group/#configuring-a-node-via-caps), you **cannot** change its NodeGroup membership. The only option is to [delete the StaticInstance](#can-a-staticinstance-be-deleted) and create a new one.

If the static node was added to the cluster [manually](#adding-a-static-node-manually), to move it to another NodeGroup, change the label with the group name and remove the label with the role:

```shell
d8 k label node --overwrite <node_name> node.deckhouse.io/group=<new_node_group_name>
d8 k label node <node_name> node-role.kubernetes.io/<old_node_group_name>-
```

Applying the changes will take some time.

### How to see what is currently running on a node while it is being created?

If you need to find out what is happening on a node (for example, it takes a long time to create or is stuck in Pending), you can check the `cloud-init` logs. To do this, follow these steps:

1. Find the node that is currently bootstrapping:

   ```shell
   d8 k get instances | grep Pending

   # dev-worker-2a6158ff-6764d-nrtbj   Pending   46s
   ```

1. Get the connection parameters for viewing the logs:

   ```shell
   d8 k get instances dev-worker-2a6158ff-6764d-nrtbj -o yaml | grep 'bootstrapStatus' -B0 -A2

   # bootstrapStatus:
   #   description: Use 'curl -N http://192.168.199.158:8000' to get bootstrap logs.
   #   logsEndpoint: 192.168.199.178:8000
   ```

1. Run the obtained command (`curl -N http://192.168.199.158:8000` in the example above) to get the `cloud-init` logs for further diagnostics.

The logs of the initial node configuration are located in `/var/log/cloud-init-output.log`.

### When is a node reboot required?

During node configuration, some configuration changes may require a reboot.

For example, a node reboot is required in Astra Linux when changing the `kernel.yama.ptrace_scope` sysctl parameter (the result of the `astra-ptrace-lock enable/disable` command).

The reboot mode is defined in the [`disruptions`](/modules/node-manager/cr.html#nodegroup-v1-spec-disruptions) parameter section of the NodeGroup resource.
