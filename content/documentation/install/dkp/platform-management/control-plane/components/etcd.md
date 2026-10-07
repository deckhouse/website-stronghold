---
title: "Etcd operations"
description: "Backing up and restoring etcd in Deckhouse Kubernetes Platform, restoring individual objects, listing etcd members, and rebuilding the etcd cluster."
weight: 20
---

## etcd backup

### Automatic backup

Deckhouse creates the `kube-system/d8-etcd-backup-*` CronJob, which runs at 00:00 UTC+0. The etcd data backup is saved to the `/var/lib/etcd/etcd-backup.tar.gz` archive on all master nodes.

### Manual backup using Deckhouse CLI

In Deckhouse v1.65 and later clusters, you can create an etcd data backup with a single `d8 backup etcd` command:

```bash
d8 backup etcd --kubeconfig $KUBECONFIG ./etcd.db
```

<!-- TODO what is in the etcd.db file? The etcdctl option explains what file is created, but this one does not.
TODO where should this be run, on every master node or not? -->

### Manual backup using etcdctl

{{< alert level="warning" >}}
Not recommended for Deckhouse 1.65 and later.
{{< /alert >}}

In Deckhouse v1.64 and earlier clusters, run the following script on any master node as the `root` user:

```bash
#!/usr/bin/env bash
set -e

pod=etcd-`hostname`
d8 k -n kube-system exec "$pod" -- /usr/bin/etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt --cert /etc/kubernetes/pki/etcd/ca.crt --key /etc/kubernetes/pki/etcd/ca.key --endpoints https://127.0.0.1:2379/ snapshot save /var/lib/etcd/${pod##*/}.snapshot && \
mv /var/lib/etcd/"${pod##*/}.snapshot" etcd-backup.snapshot && \
cp -r /etc/kubernetes/ ./ && \
tar -cvzf kube-backup.tar.gz ./etcd-backup.snapshot ./kubernetes/
rm -r ./kubernetes ./etcd-backup.snapshot
```

The `kube-backup.tar.gz` file containing a snapshot of the etcd database of one of the etcd cluster nodes will be created in the current directory.
You can restore the etcd cluster state from this snapshot.

It is also recommended to back up the `/etc/kubernetes` directory, which contains:

- manifests and configuration of the [control plane](https://kubernetes.io/docs/concepts/overview/components/#control-plane-components) components;
- [Kubernetes cluster PKI](https://kubernetes.io/docs/setup/best-practices/certificates/).

We recommend storing etcd cluster snapshot backups and the `/etc/kubernetes/` directory backup in encrypted form outside the Deckhouse cluster.
To do this, you can use third-party file backup tools, such as [Restic](https://restic.net/), [Borg](https://borgbackup.readthedocs.io/en/stable/), [Duplicity](https://duplicity.gitlab.io/), etc.

## Full cluster state recovery from an etcd backup

The following steps describe how to restore the cluster to a previous state from a backup in case of complete data loss.

### Restoring a cluster with a single master node

To correctly restore a cluster with a single master node, follow these steps:

1. Download the [etcdutl](https://github.com/etcd-io/etcd/releases) utility to the server (preferably, its version should match the etcd version in the cluster).

   ```shell
   wget "https://github.com/etcd-io/etcd/releases/download/v3.6.1/etcd-v3.6.1-linux-amd64.tar.gz"
   tar -xzvf etcd-v3.6.1-linux-amd64.tar.gz && mv etcd-v3.6.1-linux-amd64/etcdutl /usr/local/bin/etcdutl
   ```

   You can check the etcd version in the cluster by running the following command:

   ```shell
   d8 k -n kube-system exec -ti etcd-$(hostname) -- etcdutl version
   ```

1. Stop etcd.

   etcd runs as a static pod, so it is enough to move the manifest file:

   ```shell
   mv /etc/kubernetes/manifests/etcd.yaml ~/etcd.yaml
   ```

1. Save the current etcd data.

   ```shell
   cp -r /var/lib/etcd/member/ /var/lib/deckhouse-etcd-backup
   ```

1. Clean up the etcd directory.

   ```shell
   rm -rf /var/lib/etcd
   ```

1. Put the etcd backup into the `~/etcd-backup.snapshot` file.

1. Restore the etcd database.

   ```shell
   ETCDCTL_API=3 etcdutl snapshot restore ~/etcd-backup.snapshot  --data-dir=/var/lib/etcd
   ```

1. Start etcd.

   ```shell
   mv ~/etcd.yaml /etc/kubernetes/manifests/etcd.yaml
   ```

### Restoring a multi-master cluster

To correctly restore a multi-master cluster, follow these steps:

1. Enable High Availability (HA) mode. This is necessary to keep at least one Prometheus replica and its PVC, since HA is disabled by default in a cluster with a single master node.

1. Switch the cluster to single-master mode:

   - In a cloud cluster, follow the [instructions for reducing the number of master nodes in a cloud cluster](/modules/control-plane-manager/faq.html#how-to-reduce-the-number-of-master-nodes-in-a-cloud-cluster).
   - In a static cluster, remove the `control-plane` role from the extra master nodes following the [instructions for removing the master role while keeping the node](/modules/control-plane-manager/faq.html#how-to-dismiss-the-master-role-while-keeping-the-node), and then delete them from the cluster.
   - In a static cluster with HA mode based on two master nodes and an arbiter node, delete the arbiter node and the extra master nodes.
   - In a cloud cluster with HA mode based on two master nodes and an arbiter node, follow the [instructions for reducing the number of master nodes in a cloud cluster](/modules/control-plane-manager/faq.html#how-to-reduce-the-number-of-master-nodes-in-a-cloud-cluster) to delete the extra master nodes and the arbiter node.

1. Restore etcd from the backup on the only remaining master node. Follow the [instructions for a cluster with a single master node](#restoring-a-cluster-with-a-single-master-node).

1. Once etcd is restored, delete information about the master nodes removed in the first step from the cluster using the following command (specify the node name):

   ```shell
   d8 k delete node <MASTER_NODE_NAME>
   ```

   > **Warning.** If the `d8 k` or `kubectl` commands are unavailable on the node, check the `/etc/kubernetes/kubernetes-api-proxy/nginx.conf` configuration file. It must specify only your current API server. If the configuration contains lines with IP addresses of the old master nodes, delete them. Fix the configuration on all other nodes in the same way.

1. Restart the master node. Make sure the other nodes have switched to the `Ready` state.

1. Wait for the tasks in the Deckhouse queue to complete:

   ```shell
   d8 system queue main
   ```

1. Switch the cluster back to multi-master mode. Use the corresponding instructions for [cloud clusters](/modules/control-plane-manager/faq.html#how-to-add-master-nodes-to-a-cloud-cluster-single-master-to-multi-master) and for [static clusters](/modules/control-plane-manager/faq.html#how-to-add-a-master-node-to-a-static-or-hybrid-cluster).

After these steps, the cluster will be successfully restored in the multi-master configuration.

## Restoring a Kubernetes object from an etcd backup

A brief scenario for restoring individual objects from an etcd backup:

1. Get a data backup.
1. Start a temporary etcd instance.
1. Fill it with data from the backup.
1. Get the descriptions of the required objects using the `etcdhelper` utility.

### Steps for restoring objects from an etcd backup

In the example:

- `etcd-snapshot.bin` — a file with an etcd data [backup](#manual-backup-using-deckhouse-cli) (snapshot);
- `infra-production` — the namespace in which objects need to be restored.

1. Start a pod with a temporary etcd instance.

Preferably, the version of the etcd instance being started should match the etcd version from which the backup was created. For simplicity, the instance is started in the cluster rather than locally, since the cluster is guaranteed to have the etcd image.

- Prepare the `etcd.pod.yaml` file with the pod manifest:

    ```shell
    cat <<EOF >etcd.pod.yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: etcdrestore
      namespace: default
    spec:
      containers:
      - command:
        - /bin/sh
        - -c
        - "sleep 96h"
        image: IMAGE
        imagePullPolicy: IfNotPresent
        name: etcd
        volumeMounts:
        - name: etcddir
          mountPath: /default.etcd
      volumes:
      - name: etcddir
        emptyDir: {}
    EOF
    ```

- Set the actual etcd image name:

  ```shell
  IMG=`d8 k -n kube-system get pod -l component=etcd -o jsonpath="{.items[0].spec.    containers[*].image}"`
  sed -i -e "s#IMAGE#$IMG#" etcd.pod.yaml
  ```

- Create the pod:

  ```shell
  d8 k create -f etcd.pod.yaml
  ```

Copy `etcdhelper` and the etcd snapshot to the pod container.

You can build `etcdhelper` from the [source code](https://github.com/openshift/origin/tree/master/tools/etcdhelper) or copy it from a ready-made image (for example, from the [`etcdhelper` image on Docker Hub](https://hub.docker.com/r/webner/etcdhelper/tags)).

Example:

```shell
d8 k cp etcd-snapshot.bin default/etcdrestore:/tmp/etcd-snapshot.bin
d8 k cp etcdhelper default/etcdrestore:/usr/bin/etcdhelper
```

In the container, set execute permissions for `etcdhelper`, restore data from the backup, and start etcd.

Example:

```console
~ # d8 k -n default exec -it etcdrestore -- sh
/ # chmod +x /usr/bin/etcdhelper
/ # etcdctl snapshot restore /tmp/etcd-snapshot.bin
/ # etcd &
```

Get the descriptions of the required cluster objects by filtering them with `grep`.

Example:

```console
~ # d8 k -n default exec -it etcdrestore -- sh
/ # mkdir /tmp/restored_yaml
/ # cd /tmp/restored_yaml
/tmp/restored_yaml # for o in `etcdhelper -endpoint 127.0.0.1:2379 ls /registry/ | grep infra-production` ; do etcdhelper -endpoint 127.0.0.1:2379 get $o > `echo $o | sed -e "s#/registry/##g;s#/#_#g"`.yaml ; done
```

Replacing characters with `sed` in the example allows you to save object descriptions to files named after the etcd registry structure. For example: `/registry/deployments/infra-production/supercronic.yaml` → `deployments_infra-production_supercronic.yaml`.

1. Copy the resulting object descriptions from the pod to the master node with the command:

   ```shell
   d8 k cp default/etcdrestore:/tmp/restored_yaml restored_yaml
   ```

1. Remove the creation time, UID, status, and other runtime data from the resulting object descriptions, and then restore the objects with the command:

   ```shell
   d8 k create -f restored_yaml/deployments_infra-production_supercronic.yaml
   ```

1. You can delete the pod with the temporary etcd instance with the command:

   ```shell
   d8 k delete -f etcd.pod.yaml
   ```

## How to get the list of etcd cluster members

Use the `etcdctl member list` command.

Example:

```shell
for pod in $(d8 k -n kube-system get pod -l component=etcd,tier=control-plane -o name); do
  d8 k -n kube-system exec "$pod" -- etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt \
  --cert /etc/kubernetes/pki/etcd/ca.crt --key /etc/kubernetes/pki/etcd/ca.key \
  --endpoints https://127.0.0.1:2379/ member list -w table
  if [ $? -eq 0 ]; then
    break
  fi
done
```

**Warning.** The last parameter in the output table shows whether the etcd cluster member is in the [learner](https://etcd.io/docs/v3.5/learning/design-learner/) state rather than the leader state.

## How to get the list of etcd cluster members (option 2)

To get information about etcd cluster members as a table, use the `etcdctl endpoint status` command. For the leader, the `IS LEADER` column will show `true`.

Example:

```shell
for pod in $(d8 k -n kube-system get pod -l component=etcd,tier=control-plane -o name); do
  d8 k -n kube-system exec "$pod" -- etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt \
  --cert /etc/kubernetes/pki/etcd/ca.crt --key /etc/kubernetes/pki/etcd/ca.key \
  --endpoints https://127.0.0.1:2379/ endpoint status --cluster -w table check etcd cluster status
  if [ $? -eq 0 ]; then
    break
  fi
done
```

## Rebuilding the etcd cluster

Rebuilding may be required if the etcd cluster has fallen apart, or when migrating from a multi-master cluster to a cluster with a single master node.

1. Select the node from which the etcd cluster recovery will start. When migrating to a cluster with a single master node, this is the node on which etcd should remain.
1. Stop etcd on all other nodes. To do this, delete the `/etc/kubernetes/manifests/etcd.yaml` file.
1. On the remaining node, add the `--force-new-cluster` argument to the `spec.containers.command` field in the `/etc/kubernetes/manifests/etcd.yaml` manifest.
1. After the cluster has started successfully, remove the `--force-new-cluster` parameter.

{{< alert level="danger" >}}
This operation is destructive: it completely destroys consensus and starts the etcd cluster from the state saved on the selected node. Any pending writes will be lost.
{{< /alert >}}

## Fixing an endless restart loop

This option may be needed if starting with the `--force-new-cluster` argument does not restore etcd. This can happen after an unsuccessful converge of master nodes, when a new master node was created with an old etcd disk, changed its local network address, and other master nodes are missing. Use this method if the etcd container is in an endless restart loop and its log contains the error: `panic: unexpected removal of unknown remote peer`.

1. Install the [etcdutl](https://github.com/etcd-io/etcd/releases) utility.
1. Create a new snapshot from the current local etcd database snapshot (`/var/lib/etcd/member/snap/db`):

   ```shell
   ./etcdutl snapshot restore /var/lib/etcd/member/snap/db --name <HOSTNAME> \
   --initial-cluster=<HOSTNAME>=https://<ADDRESS>:2380 --initial-advertise-peer-urls=https://<ADDRESS>:2380 \
   --skip-hash-check=true --data-dir /var/lib/etcdtest
   ```

   where:

- `<HOSTNAME>` — the master node name;
- `<ADDRESS>` — the master node address.

1. Run the commands to use the new snapshot:

   ```shell
   cp -r /var/lib/etcd /tmp/etcd-backup
   rm -rf /var/lib/etcd
   mv /var/lib/etcdtest /var/lib/etcd
   ```

1. Find the `etcd` and `kube-apiserver` containers:

   ```shell
   crictl ps -a --name "^etcd|^kube-apiserver"
   ```

1. Delete the found `etcd` and `kube-apiserver` containers:

   ```shell
   crictl rm <CONTAINER-ID>
   ```

1. Restart the master node.
