---
title: "Node group configuration"
description: "Running custom scripts on node group nodes with the NodeGroupConfiguration resource: template variables, bashbooster commands, and script writing specifics."
weight: 30
---

## Custom settings on nodes

The [NodeGroupConfiguration](/modules/node-manager/cr.html#nodegroupconfiguration) resource is provided to automate actions on group nodes. The resource allows you to run bash scripts on nodes that can use the [bashbooster](https://github.com/deckhouse/deckhouse/tree/main/candi/bashible/bashbooster) command set, and also allows you to use the [Go Template](https://pkg.go.dev/text/template) templating engine. This is convenient for automating operations such as:

- Installing and configuring additional OS packages.

  Example:
  - [installing a kubectl plugin](./os/#installing-the-cert-manager-kubectl-plugin-on-master-nodes).

- Updating the OS kernel to a specific version.

  Examples:
  - [updating the Debian kernel](./os/#for-debian-based-distributions).
  - [updating the CentOS kernel](./os/#for-centos-based-distributions).

- Changing OS parameters.

  Examples:
  - [setting a sysctl parameter](./os/#setting-a-sysctl-parameter).
  - [adding a root certificate](./os/#adding-a-root-certificate).

- Collecting information on the node and performing other similar actions.

- Configuring containerd.

  Examples:
  - [configuring metrics](./containerd/#additional-containerd-settings).
  - [adding a private registry](./containerd/#adding-a-private-registry-with-authorization).

## NodeGroupConfiguration settings

The NodeGroupConfiguration resource allows you to set the [priority](/modules/node-manager/cr.html#nodegroupconfiguration-v1alpha1-spec-weight) of the scripts being run and to limit their execution to specific [node groups](/modules/node-manager/cr.html#nodegroupconfiguration-v1alpha1-spec-nodegroups) and [OS types](/modules/node-manager/cr.html#nodegroupconfiguration-v1alpha1-spec-bundles).

The script code is specified in the [`content`](/modules/node-manager/cr.html#nodegroupconfiguration-v1alpha1-spec-content) parameter of the resource. When the script is created on the node, the contents of the `content` parameter are processed by the [Go Template](https://pkg.go.dev/text/template) templating engine, which allows you to add an extra layer of logic when generating the script. During template processing, a context with a set of dynamic variables is available.

Variables available for use in templates:
<ul>
<li><code>.cloudProvider</code> (for node groups with nodeType <code>CloudEphemeral</code> or <code>CloudPermanent</code>) — an array of cloud provider data.
{{% details summary="Example data..." %}}
```yaml
cloudProvider:
  instanceClassKind: OpenStackInstanceClass
  machineClassKind: OpenStackMachineClass
  openstack:
    connection:
      authURL: https://cloud.provider.com/v3/
      domainName: Default
      password: p@ssw0rd
      region: region2
      tenantName: mytenantname
      username: mytenantusername
    externalNetworkNames:
    - public
    instances:
      imageName: ubuntu-22-04-cloud-amd64
      mainNetwork: kube
      securityGroups:
      - kube
      sshKeyPairName: kube
    internalNetworkNames:
    - kube
    podNetworkMode: DirectRoutingWithPortSecurityEnabled
  region: region2
  type: openstack
  zones:
  - nova
```
{{% /details %}}</li>
<li><code>.cri</code> — the CRI in use (starting with Deckhouse 1.49, only <code>Containerd</code> is used).</li>
<li><code>.kubernetesVersion</code> — the Kubernetes version in use.</li>
<li><code>.nodeUsers</code> — an array of data about node users added via the <a href="/modules/node-manager/cr.html#nodeuser">NodeUser</a> resource.
{{% details summary="Example data..." %}}
```yaml
nodeUsers:
- name: user1
  spec:
    isSudoer: true
    nodeGroups:
    - '*'
    passwordHash: PASSWORD_HASH
    sshPublicKey: SSH_PUBLIC_KEY
    uid: 1050
```
{{% /details %}}
</li>
<li><code>.nodeGroup</code> — an array of node group data.
{{% details summary="Example data..." %}}
```yaml
nodeGroup:
  cri:
    type: Containerd
  disruptions:
    approvalMode: Automatic
  kubelet:
    containerLogMaxFiles: 4
    containerLogMaxSize: 50Mi
    resourceReservation:
      mode: "Off"
  kubernetesVersion: "1.27"
  manualRolloutID: ""
  name: master
  nodeTemplate:
    labels:
      node-role.kubernetes.io/control-plane: ""
      node-role.kubernetes.io/master: ""
    taints:
    - effect: NoSchedule
      key: node-role.kubernetes.io/master
  nodeType: CloudPermanent
  updateEpoch: "1699879470"
```
{{% /details %}}</li>
</ul>
Example of using variables in a template:

```shell
{{- range .nodeUsers }}
echo 'Tuning environment for user {{ .name }}'
# Some code for tuning user environment
{{- end }}
```

Example of using bashbooster commands:

```shell
bb-event-on 'bb-package-installed' 'post-install'
post-install() {
  bb-log-info "Setting reboot flag due to kernel was updated"
  bb-flag-set reboot
}
```

You can view the progress of script execution on the node in the bashible service log using the command:

```bash
journalctl -u bashible.service
```  

The scripts themselves are located on the node in the `/var/lib/bashible/bundle_steps/` directory.

The service decides whether to rerun the scripts by comparing the single checksum of all files located at `/var/lib/bashible/configuration_checksum` with the checksum stored in the Kubernetes cluster in the `configuration-checksums` secret of the `d8-cloud-instance-manager` namespace.

You can check the checksum with the following command:

```bash
d8 k -n d8-cloud-instance-manager get secret configuration-checksums -o yaml
```  

The service compares the checksums every minute.

The checksum in the cluster changes every 4 hours, thereby rerunning the scripts on all nodes.  
To force bashible to run on a node, delete the script checksum file using the following command:

```bash
rm /var/lib/bashible/configuration_checksum
```  

### Script writing specifics

When writing scripts, consider the following specifics of how they are used in Deckhouse:

1. Scripts in Deckhouse run every 4 hours or based on external triggers. Therefore, write scripts so that they check whether their changes are needed in the system before performing any actions, rather than making changes on every run.
1. When choosing the [priority](/modules/node-manager/cr.html#nodegroupconfiguration-v1alpha1-spec-weight) of custom scripts, take into account the [built-in scripts](https://github.com/deckhouse/deckhouse/tree/main/candi/bashible/common-steps/all) that perform various actions, including installing and configuring services. For example, if a script is supposed to restart a service and the service is installed by a built-in script with priority N, the custom script priority must be at least N+1; otherwise, the custom script will fail when a new node is deployed.
