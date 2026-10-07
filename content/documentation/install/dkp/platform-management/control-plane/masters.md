---
title: "Master nodes"
description: "Adding master nodes to a Deckhouse Kubernetes Platform cluster and removing the master role from a node while keeping it in the cluster."
weight: 50
---

## Adding a master node

{{< alert level="warning" >}}
It is important to have an odd number of master nodes to maintain quorum.
{{< /alert >}}

Adding a master node to the cluster is no different from adding a regular node. Make sure a NodeGroup with the control-plane role exists (usually it is the NodeGroup named master) and follow the instructions for [adding a node](../node-management/adding-node/#adding-a-node-to-the-cluster). All the steps required to configure the cluster control plane components on the new node will be performed automatically.

Before adding the next node, wait until all master nodes are in the `Ready` status:

```shell
d8 k get no -l node-role.kubernetes.io/control-plane=
NAME       STATUS   ROLES                  AGE    VERSION
master-0   Ready    control-plane,master   276d   v1.28.15
master-1   Ready    control-plane,master   247d   v1.28.15
master-2   Ready    control-plane,master   247d   v1.28.15
```

## Removing the master role from a node while keeping the node in the cluster

1. Make a [backup of etcd](./components/etcd/#etcd-backup) and the `/etc/kubernetes` directory.
1. Copy the resulting archive outside the cluster (for example, to your local machine).
1. Make sure there are no [alerts](/modules/prometheus/faq.html#how-to-get-information-about-alerts-in-the-cluster) in the cluster that could prevent master nodes from being updated.
1. Make sure the Deckhouse queue is empty:

   ```shell
   d8 system queue list
   ```

1. Remove the `node.deckhouse.io/group: master` and `node-role.kubernetes.io/control-plane: ""` labels from the node.
1. Make sure the node is no longer listed among the etcd cluster members:

   ```bash
   for pod in $(d8 k -n kube-system get pod -l component=etcd,tier=control-plane -o name); do
     d8 k -n kube-system exec "$pod" -- etcdctl --cacert /etc/kubernetes/pki/etcd/ca.crt \
     --cert /etc/kubernetes/pki/etcd/ca.crt --key /etc/kubernetes/pki/etcd/ca.key \
     --endpoints https://127.0.0.1:2379/ member list -w table
     if [ $? -eq 0 ]; then
       break
     fi
   done
   ```

1. Delete the control plane component settings on the node:

   ```shell
   rm -f /etc/kubernetes/manifests/{etcd,kube-apiserver,kube-scheduler,kube-controller-manager}.yaml
   rm -f /etc/kubernetes/{scheduler,controller-manager}.conf
   rm -f /etc/kubernetes/authorization-webhook-config.yaml
   rm -f /etc/kubernetes/admin.conf /root/.kube/config
   rm -rf /etc/kubernetes/deckhouse
   rm -rf /etc/kubernetes/pki/{ca.key,apiserver*,etcd/,front-proxy*,sa.*}
   rm -rf /var/lib/etcd/member/
   ```

1. Make sure the number of nodes in the `master` NodeGroup has decreased.

   If there were 3 nodes, there should now be 2:

   ```shell
   d8 k get ng master
   NAME     TYPE     READY   NODES   UPTODATE   INSTANCES   DESIRED   MIN   MAX   STANDBY   STATUS   AGE    SYNCED
   master   Static   2       2       2                                                               280d   True
   ```
