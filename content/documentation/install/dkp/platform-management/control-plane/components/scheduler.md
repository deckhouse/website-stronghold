---
title: "Scheduler"
description: "How the Kubernetes scheduler works in Deckhouse Kubernetes Platform, how to extend its logic with webhook plugins, and how to speed up recovery after node loss."
weight: 30
---

## Scheduler algorithm

The Kubernetes scheduler (the `kube-scheduler` component) is responsible for distributing pods across nodes.

The scheduler decision-making algorithm is divided into 2 phases: `Filtering` and `Scoring`.

Within each phase, the scheduler runs a set of plugins that implement decision-making, for example:

- **ImageLocality** — prefers nodes that already have the container images used in the pod being started. Phase: `Scoring`.
- **TaintToleration** — implements the [taints and tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) mechanism. Phases: `Filtering, Scoring`.
- **NodePorts** — checks whether the node has the free ports required to run the pod. Phase: `Filtering`.

The full list of plugins is available in the [Kubernetes documentation](https://kubernetes.io/docs/reference/scheduling/config/#scheduling-plugins).

In the first phase (`Filtering`), filter plugins check nodes against the filter conditions (taints, nodePorts, nodeName, unschedulable, etc.).

The filtered list is sorted with zone alternation so that not all pods are placed in the same zone. Suppose that after filtering, the remaining nodes are distributed across zones as follows:

```text
Zone 1: Node 1, Node 2, Node 3, Node 4
Zone 2: Node 5, Node 6
```

In this case, they will be selected in the following order:

```text
Node 1, Node 5, Node 2, Node 6, Node 3, Node 4
```

Note that, for optimization purposes, not all nodes matching the conditions are selected, but only some of them. By default, the function for selecting the number of nodes is linear. For a cluster of ≤50 nodes, 100% of the nodes will be selected; for a cluster of 100 nodes, 50%; and for a cluster of 5000 nodes, 10%. The minimum value is 5% when there are more than 5000 nodes. For more details on limiting the number of nodes, see the Kubernetes documentation for the [KubeSchedulerConfiguration](https://kubernetes.io/docs/reference/config-api/kube-scheduler-config.v1/#kubescheduler-config-k8s-io-v1-KubeSchedulerConfiguration) resource. Deckhouse uses the default value, so take this scheduler behavior into account in very large clusters.

After the nodes matching the filter conditions are selected, the `Scoring` phase starts. The plugins of this phase analyze the list of filtered nodes and assign a score to each node. Scores from different plugins are summed. This phase evaluates available node resources, pod capacity, affinity, volume provisioning, and so on.

The result of this phase is a list of nodes with the highest score. If there is more than one node in the list, a node is selected randomly.

### Documentation

- [General scheduler description](https://kubernetes.io/docs/concepts/scheduling-eviction/kube-scheduler/);
- [Plugin system](https://kubernetes.io/docs/reference/scheduling/config/#scheduling-plugins);
- [Node filtering details](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduler-perf-tuning/);
- [Scheduler source code](https://github.com/kubernetes/kubernetes/tree/master/cmd/kube-scheduler).

## Changing and extending the scheduler logic

To change the scheduler logic, you can use the [extension plugin mechanism](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/624-scheduling-framework/README.md).

Each plugin is a webhook that meets the following requirements:

* Uses TLS.
* Is accessible via a service inside the cluster.
* Supports standard *Verbs* (filterVerb = filter, prioritizeVerb = prioritize).
* It is also assumed that all connected plugins can cache node information (`nodeCacheCapable: true`).

You can connect such a webhook extender using the [KubeSchedulerWebhookConfiguration](/modules/control-plane-manager/cr.html#kubeschedulerwebhookconfiguration) resource.

{{< alert level="danger" >}}
When using the `failurePolicy: Fail` option, a webhook error stops the scheduler, and new pods will not be able to start.
{{< /alert >}}

## Speeding up recovery after node loss

<!-- TODO this needs to be aligned with virtual machines somehow. -->

By default, if a node does not report its state for 40 seconds, it is marked as unavailable. After another 5 minutes, the pods of such a node are rescheduled to other nodes. The total application downtime is about 6 minutes.

For specific tasks, when an application cannot run in multiple instances, the `control-plane-manager` module settings provide a way to reduce the downtime:

1. Reduce the time it takes for a node to switch to the `Unreachable` state after losing connection with it by configuring the `nodeMonitorGracePeriodSeconds` parameter.
1. Reduce the timeout for rescheduling pods to another node in the `failedNodePodEvictionTimeoutSeconds` parameter.

### Example

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: control-plane-manager
spec:
  version: 1
  settings:
    nodeMonitorGracePeriodSeconds: 10
    failedNodePodEvictionTimeoutSeconds: 50
```

In this case, when connection with a node is lost, applications will be started on other nodes in about 1 minute.

{{< alert level="warning" >}}
Both parameters directly affect CPU and memory consumption on master nodes. Reduced timeouts force system components to send statuses and reconcile resource states more often.

When choosing suitable values, monitor the resource consumption graphs of master nodes. Be prepared that ensuring acceptable parameter values may require increasing the capacity allocated to master nodes.
{{< /alert >}}
