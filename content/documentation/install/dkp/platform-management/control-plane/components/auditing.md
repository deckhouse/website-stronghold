---
title: "Audit"
description: "Configuring Kubernetes API audit in Deckhouse Kubernetes Platform: basic and custom audit policies, audit log handling, and output to stdout."
weight: 40
---

## Audit

To diagnose API operations, for example, in case of unexpected behavior of control plane components, Kubernetes provides an API operation logging mode. You can configure this mode by creating [Audit Policy](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/#audit-policy) rules, and the result of the audit will be the `/var/log/kube-audit/audit.log` log file with all the operations of interest. For more details, see the [Auditing](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/) section of the Kubernetes documentation.

## Basic audit policies

Deckhouse clusters have the following basic audit policies created by default:

- logging of resource creation, deletion, and modification operations;
  <!-- TODO which resources are meant here? Needs clarification. -->
- logging of actions performed on behalf of service accounts from the system namespaces: `kube-system`, `d8-*`;
- logging of actions performed on resources in the system namespaces: `kube-system`, `d8-*`.

### Disabling basic policies

You can disable log collection based on the basic policies by setting the [`basicAuditPolicyEnabled`](/modules/control-plane-manager/configuration.html#parameters-apiserver-basicauditpolicyenabled) flag to `false`.

Example of enabling audit in kube-apiserver without the basic Deckhouse policies:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: control-plane-manager
spec:
  version: 1
  settings:
    apiserver:
      auditPolicyEnabled: true
      basicAuditPolicyEnabled: false
```

You can also use a patch:

```shell
d8 k patch mc control-plane-manager --type=strategic -p '{"settings":{"apiserver":{"auditPolicyEnabled":true, "basicAuditPolicyEnabled": false}}}'
```

## Custom audit policies

The control-plane-manager module automates the kube-apiserver configuration for adding custom audit policies. For such additional policies to work, make sure that audit is enabled in the `apiserver` parameter section and create a secret with the audit policy:

1. Enable the [`auditPolicyEnabled`](/modules/control-plane-manager/configuration.html#parameters-apiserver-auditpolicyenabled) parameter in the module settings:

   ```yaml
   apiVersion: deckhouse.io/v1alpha1
   kind: ModuleConfig
   metadata:
     name: control-plane-manager
   spec:
     version: 1
     settings:
       apiserver:
         auditPolicyEnabled: true
   ```

   You can enable it by editing the resource or by using a patch:

   ```shell
   d8 k patch mc control-plane-manager --type=strategic -p '{"settings":{"apiserver":{"auditPolicyEnabled":true}}}'
   ```

1. Create the `kube-system/audit-policy` Secret with the Base64-encoded policy YAML file:

   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: audit-policy
     namespace: kube-system
   data:
     audit-policy.yaml: <base64>
   ```

   As an example of `audit-policy.yaml`, here is a rule for logging all metadata changes:

   ```yaml
   apiVersion: audit.k8s.io/v1
   kind: Policy
   rules:
   - level: Metadata
     omitStages:
     - RequestReceived
   ```

   Examples and information about audit policy rules are available in:

   - [Official Kubernetes documentation](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/#audit-policy).
   - [Our article on Habr](https://habr.com/ru/company/flant/blog/468679/) (in Russian).
   - [Code of the generator script used in GCE](https://github.com/kubernetes/kubernetes/blob/0ef45b4fcf7697ea94b96d1a2fe1d9bffb692f3a/cluster/gce/gci/configure-helper.sh#L722-L862).

{{< alert level="danger" >}}
The current implementation does not validate the contents of additional policies.

If the policy in `audit-policy.yaml` contains unsupported options or a typo, `apiserver` will not start, which will make the control plane unavailable.
{{< /alert >}}

In this case, to recover, manually remove the `--audit-log-*` parameters from the `/etc/kubernetes/manifests/kube-apiserver.yaml` manifest and restart `apiserver` with the following command:

```bash
crictl stopp $(crictl pods --name=kube-apiserver -q)
```

After the restart, you will have enough time to delete the erroneous secret:

```bash
d8 k -n kube-system delete secret audit-policy
```

## How to work with the audit log?

It is assumed that master nodes have a log collector *(for example, [`log-shipper`](/modules/log-shipper/), promtail, filebeat)* that sends records from the file to centralized storage:

```bash
/var/log/kube-audit/audit.log
```

The log file rotation parameters are preset and cannot be changed:

- Maximum disk space used: `1000 MB`.
- Maximum retention depth: `30 days`.

Keep in mind that "maximum retention depth" does not mean "guaranteed". The log write rate depends on the additional policy settings and the number of requests to **apiserver**, so the actual retention depth may be much less than 7 days, for example, 30 minutes. Take this into account when configuring the log collector and writing audit policies.

## Sending the audit log to standard output

If a pod log collector is configured in the cluster, you can collect the audit log by sending it to standard output. To do this, set the [`apiserver.auditLog.output`](/modules/control-plane-manager/configuration.html#parameters-apiserver-auditlog-output) parameter to `Stdout` in the module settings.

Example of enabling audit with output to stdout:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: control-plane-manager
spec:
  version: 1
  settings:
    apiserver:
      auditPolicyEnabled: true
      auditLog:
        output: Stdout
```

You can also use a patch:

```shell
d8 k patch mc control-plane-manager --type=strategic -p '{"settings":{"apiserver":{"auditPolicyEnabled":true, "auditLog":{"output":"Stdout"}}}}'
```

After kube-apiserver restarts, you can see audit events in its log:

```shell
d8 k -n kube-system logs $(d8 k -n kube-system get po -l component=kube-apiserver -oname | head -n1)

{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Metadata","auditID":"38a26239-7f3e-402f-8c56-2fb57a3fe49d","stage":"ResponseComplete","requestURI": ...
```
