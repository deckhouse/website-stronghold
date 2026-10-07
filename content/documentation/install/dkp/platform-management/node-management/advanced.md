---
title: "Advanced settings"
description: "Advanced node management scenarios in Deckhouse Kubernetes Platform: master node recovery, changing CRI, NVIDIA GPU support, and bulk addition of static nodes."
weight: 60
---

## Recovering a master node if kubelet cannot load control plane components

This situation may occur if, in a cluster with a single master node, the control plane component images
were deleted on it (for example, the `/var/lib/containerd` directory was deleted). In this case, after a restart, kubelet cannot pull the `control plane` component images because the master node does not have authorization parameters for `registry.deckhouse.ru`.

Below are the instructions for recovering the master node.

### containerd

To restore the master node, run the following command in any working cluster managed by Deckhouse:

```shell
d8 k -n d8-system get secrets deckhouse-registry -o json |
jq -r '.data.".dockerconfigjson"' | base64 -d |
jq -r '.auths."registry.deckhouse.ru".auth'
```

Copy the command output and assign it to the AUTH variable on the damaged master node.
Next, pull the `control-plane` component images on the damaged master node:

```shell
for image in $(grep "image:" /etc/kubernetes/manifests/* | awk '{print $3}'); do
  crictl pull --auth $AUTH $image
done
```

After pulling the images, restart kubelet.

## Changing CRI for a NodeGroup

> **Warning!** You can only switch from `Containerd` to `NotManaged` and back (the [cri.type](/modules/node-manager/cr.html#nodegroup-v1-spec-cri-type) parameter).

Set the [cri.type](/modules/node-manager/cr.html#nodegroup-v1-spec-cri-type) parameter to `Containerd` or `NotManaged`.

Example of a NodeGroup YAML manifest:

```yaml
apiVersion: deckhouse.io/v1
kind: NodeGroup
metadata:
  name: worker
spec:
  nodeType: Static
  cri:
    type: Containerd
```

You can also perform this operation using a patch:

* For `Containerd`:

  ```shell
  d8 k patch nodegroup <NodeGroup name> --type merge -p '{"spec":{"cri":{"type":"Containerd"}}}'
  ```

* For `NotManaged`:

  ```shell
  d8 k patch nodegroup <NodeGroup name> --type merge -p '{"spec":{"cri":{"type":"NotManaged"}}}'
  ```

> **Warning!** When changing `cri.type` for NodeGroups created using `dhctl`, you must change it both in `dhctl config edit provider-cluster-configuration` and in the NodeGroup object settings.

After a new CRI is configured for a NodeGroup, the `node-manager` module drains the nodes one by one and installs the new CRI on them. The node update
is accompanied by downtime (disruption). Depending on the `disruption` setting of the NodeGroup, the `node-manager` module either automatically allows node
updates or requires manual approval.

## Changing CRI for the whole cluster

> **Warning!** You can only switch from `Containerd` to `NotManaged` and back (the [cri.type](/modules/node-manager/cr.html#nodegroup-v1-spec-cri-type) parameter).

Use the `dhctl` utility to edit the `defaultCRI` parameter in the `cluster-configuration` config.

You can also perform this operation using a patch. Example:

* For `Containerd`:

  ```shell
  data="$(d8 k -n kube-system get secret d8-cluster-configuration -o json | jq -r '.data."cluster-configuration.yaml"' | base64 -d | sed "s/NotManaged/Containerd/" | base64 -w0)"
  d8 k -n kube-system patch secret d8-cluster-configuration -p '{"data":{"cluster-configuration.yaml":"'${data}'"}}'
  ```

* For `NotManaged`:

  ```shell
  data="$(d8 k -n kube-system get secret d8-cluster-configuration -o json | jq -r '.data."cluster-configuration.yaml"' | base64 -d | sed "s/Containerd/NotManaged/" | base64 -w0)"
  d8 k -n kube-system patch secret d8-cluster-configuration -p '{"data":{"cluster-configuration.yaml":"'${data}'"}}'
  ```

If you need to keep a NodeGroup on a different CRI, set the CRI for this NodeGroup before changing `defaultCRI`,
as described in [Changing CRI for a NodeGroup](#changing-cri-for-a-nodegroup).

> **Warning!** Changing `defaultCRI` results in changing the CRI on all nodes, including master nodes.
> If there is only one master node, this operation is dangerous and may make the cluster completely inoperable!
> The preferred option is to switch to a multi-master setup and then change the CRI type!

When changing the CRI in the cluster, perform the following additional steps for master nodes:

1. Deckhouse updates the nodes in the master NodeGroup one by one, so determine which node is currently being updated:

   ```shell
   d8 k get nodes -l node-role.kubernetes.io/control-plane="" -o json | jq '.items[] | select(.metadata.annotations."update.node.deckhouse.io/approved"=="") | .metadata.name' -r
   ```

1. Approve the disruption for the master node obtained in the previous step:

   ```shell
   d8 k annotate node <master node name> update.node.deckhouse.io/disruption-approved=
   ```

1. Wait until the updated master node switches to `Ready`. Repeat the iteration for the next master node.

## How to use containerd with Nvidia GPU support?

Create a separate NodeGroup for GPU nodes.

```yaml
apiVersion: deckhouse.io/v1
kind: NodeGroup
metadata:
  name: gpu
spec:
  chaos:
    mode: Disabled
  disruptions:
    approvalMode: Automatic
  nodeType: CloudStatic
```

Next, create a NodeGroupConfiguration for the `gpu` NodeGroup to configure containerd:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: containerd-additional-config.sh
spec:
  bundles:
  - '*'
  content: |
    # Copyright 2023 Flant JSC
    #
    # Licensed under the Apache License, Version 2.0 (the "License");
    # you may not use this file except in compliance with the License.
    # You may obtain a copy of the License at
    #
    #     http://www.apache.org/licenses/LICENSE-2.0
    #
    # Unless required by applicable law or agreed to in writing, software
    # distributed under the License is distributed on an "AS IS" BASIS,
    # WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    # See the License for the specific language governing permissions and
    # limitations under the License.

    mkdir -p /etc/containerd/conf.d
    bb-sync-file /etc/containerd/conf.d/nvidia_gpu.toml - << "EOF"
    [plugins]
      [plugins."io.containerd.grpc.v1.cri"]
        [plugins."io.containerd.grpc.v1.cri".containerd]
          default_runtime_name = "nvidia"
          [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
            [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
              [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia]
                privileged_without_host_devices = false
                runtime_engine = ""
                runtime_root = ""
                runtime_type = "io.containerd.runc.v2"
                [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.nvidia.options]
                  BinaryName = "/usr/bin/nvidia-container-runtime"
                  SystemdCgroup = false
    EOF
  nodeGroups:
  - gpu
  weight: 31
```

Next, add a NodeGroupConfiguration to install Nvidia drivers for the `gpu` NodeGroup.

### Ubuntu

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: install-cuda.sh
spec:
  bundles:
  - ubuntu-lts
  content: |
    # Copyright 2023 Flant JSC
    #
    # Licensed under the Apache License, Version 2.0 (the "License");
    # you may not use this file except in compliance with the License.
    # You may obtain a copy of the License at
    #
    #     http://www.apache.org/licenses/LICENSE-2.0
    #
    # Unless required by applicable law or agreed to in writing, software
    # distributed under the License is distributed on an "AS IS" BASIS,
    # WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    # See the License for the specific language governing permissions and
    # limitations under the License.

    if [ ! -f "/etc/apt/sources.list.d/nvidia-container-toolkit.list" ]; then
      distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
      curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | sudo apt-key add -
      curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.list | sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list
    fi
    bb-apt-install nvidia-container-toolkit nvidia-driver-535-server
    nvidia-ctk config --set nvidia-container-runtime.log-level=error --in-place
  nodeGroups:
  - gpu
  weight: 30
```

### Centos

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: NodeGroupConfiguration
metadata:
  name: install-cuda.sh
spec:
  bundles:
  - centos
  content: |
    # Copyright 2023 Flant JSC
    #
    # Licensed under the Apache License, Version 2.0 (the "License");
    # you may not use this file except in compliance with the License.
    # You may obtain a copy of the License at
    #
    #     http://www.apache.org/licenses/LICENSE-2.0
    #
    # Unless required by applicable law or agreed to in writing, software
    # distributed under the License is distributed on an "AS IS" BASIS,
    # WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    # See the License for the specific language governing permissions and
    # limitations under the License.

    if [ ! -f "/etc/yum.repos.d/nvidia-container-toolkit.repo" ]; then
      distribution=$(. /etc/os-release;echo $ID$VERSION_ID) \
      curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.repo | sudo tee /etc/yum.repos.d/nvidia-container-toolkit.repo
    fi
    bb-yum-install nvidia-container-toolkit nvidia-driver
    nvidia-ctk config --set nvidia-container-runtime.log-level=error --in-place
  nodeGroups:
  - gpu
  weight: 30
```

After that, bootstrap and reboot the node.

### How to check that everything went well?

Create a Job in the cluster:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: nvidia-cuda-test
  namespace: default
spec:
  completions: 1
  template:
    spec:
      restartPolicy: Never
      nodeSelector:
        node.deckhouse.io/group: gpu
      containers:
        - name: nvidia-cuda-test
          image: nvidia/cuda:11.6.2-base-ubuntu20.04
          imagePullPolicy: "IfNotPresent"
          command:
            - nvidia-smi
```

And check the logs:

```console
$ d8 k logs job/nvidia-cuda-test
Tue Jan 24 11:36:18 2023
+-----------------------------------------------------------------------------+
| NVIDIA-SMI 525.60.13    Driver Version: 525.60.13    CUDA Version: 12.0     |
|-------------------------------+----------------------+----------------------+
| GPU  Name        Persistence-M| Bus-Id        Disp.A | Volatile Uncorr. ECC |
| Fan  Temp  Perf  Pwr:Usage/Cap|         Memory-Usage | GPU-Util  Compute M. |
|                               |                      |               MIG M. |
|===============================+======================+======================|
|   0  Tesla T4            Off  | 00000000:8B:00.0 Off |                    0 |
| N/A   45C    P0    25W /  70W |      0MiB / 15360MiB |      0%      Default |
|                               |                      |                  N/A |
+-------------------------------+----------------------+----------------------+

+-----------------------------------------------------------------------------+
| Processes:                                                                  |
|  GPU   GI   CI        PID   Type   Process name                  GPU Memory |
|        ID   ID                                                   Usage      |
|=============================================================================|
|  No running processes found                                                 |
+-----------------------------------------------------------------------------+
```

Create a Job in the cluster:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: gpu-operator-test
  namespace: default
spec:
  completions: 1
  template:
    spec:
      restartPolicy: Never
      nodeSelector:
        node.deckhouse.io/group: gpu
      containers:
        - name: gpu-operator-test
          image: nvidia/samples:vectoradd-cuda10.2
          imagePullPolicy: "IfNotPresent"
```

And check the logs:

```shell
$ d8 k logs job/gpu-operator-test
[Vector addition of 50000 elements]
Copy input data from the host memory to the CUDA device
CUDA kernel launch with 196 blocks of 256 threads
Copy output data from the CUDA device to the host memory
Test PASSED
Done
```

## How to add multiple static nodes to the cluster manually?

Use an existing [NodeGroup](/modules/node-manager/cr.html#nodegroup) custom resource or create a new one (see an [example](/modules/node-manager/cr.html#nodegroup_v1) of a NodeGroup named `worker`).

You can automate adding nodes using any automation platform. Below is an example for Ansible.

1. Get one of the Kubernetes API server addresses. Note that the IP address must be reachable from the nodes being added to the cluster:

   ```shell
   d8 k -n default get ep kubernetes -o json | jq '.subsets[0].addresses[0].ip + ":" + (.subsets[0].ports[0].port | tostring)' -r
   ```

   Check the Kubernetes version. If the version is >= 1.25, create a `node-group` token:

   ```shell
   d8 k create token node-group --namespace d8-cloud-instance-manager --duration 1h
   ```

   Save the resulting token and add it to the `token:` field of the Ansible playbook in the following steps.

1. If the Kubernetes version is lower than 1.25, get a Kubernetes API token for the special ServiceAccount managed by Deckhouse:

   ```shell
   d8 k -n d8-cloud-instance-manager get $(d8 k -n d8-cloud-instance-manager get secret -o name | grep node-group-token) \
     -o json | jq '.data.token' -r | base64 -d && echo ""
   ```

1. Create an Ansible playbook with `vars` replaced by the values obtained in the previous steps:

   ```yaml
   - hosts: all
     become: yes
     gather_facts: no
     vars:
       kube_apiserver: <KUBE_APISERVER>
       token: <TOKEN>
     tasks:
       - name: Check if node is already bootsrapped
         stat:
           path: /var/lib/bashible
         register: bootstrapped
       - name: Get bootstrap secret
         uri:
           url: "https://{{ kube_apiserver }}/api/v1/namespaces/d8-cloud-instance-manager/secrets/manual-bootstrap-for-{{ node_group }}"
           return_content: yes
           method: GET
           status_code: 200
           body_format: json
           headers:
             Authorization: "Bearer {{ token }}"
           validate_certs: no
         register: bootstrap_secret
         when: bootstrapped.stat.exists == False
       - name: Run bootstrap.sh
         shell: "{{ bootstrap_secret.json.data['bootstrap.sh'] | b64decode }}"
         args:
           executable: /bin/bash
         ignore_errors: yes
         when: bootstrapped.stat.exists == False
       - name: wait
         wait_for_connection:
           delay: 30
         when: bootstrapped.stat.exists == False
   ```

1. Define an additional `node_group` variable. Its value must match the name of the NodeGroup the node will belong to. You can pass the variable in various ways, for example, using an inventory file:

   ```text
   [system]
   system-0
   system-1

   [system:vars]
   node_group=system

   [worker]
   worker-0
   worker-1

   [worker:vars]
   node_group=worker
   ```

1. Run the playbook using the inventory file.

## How to make werf ignore the Ready state of a node group?

[werf](https://werf.io) checks the `Ready` state of resources and, if present, waits until its value becomes `True`.

Creating (updating) a NodeGroup resource in the cluster may take a significant amount of time to deploy the required number of nodes. When such a resource is deployed to the cluster using werf (for example, as part of a CI/CD process), the deployment may fail due to exceeding the resource readiness timeout. To make werf ignore the `nodeGroup` state, add the following annotations to the `nodeGroup`:

```yaml
metadata:
  annotations:
    werf.io/fail-mode: IgnoreAndContinueDeployProcess
    werf.io/track-termination-mode: NonBlocking
```
