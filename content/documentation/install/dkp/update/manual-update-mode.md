---
title: "Manual update mode"
description: "How to enable manual approval of Deckhouse Kubernetes Platform updates, including potentially disruptive updates."
weight: 20
---

To approve updates manually, set this mode in the configuration:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: deckhouse
spec:
  version: 1
  settings:
    releaseChannel: Stable
    update:
      mode: Manual
```

In this mode, you must approve every minor platform update (patch versions are not taken into account).

Example of approving an update to version `v1.43.2`:

```shell
d8 k patch DeckhouseRelease v1.43.2 --type=merge -p='{"approved": true}'
```

### Manual approval of potentially disruptive updates

If necessary, you can enable manual approval of potentially disruptive updates that change default values or the behavior of some modules:

```yaml
apiVersion: deckhouse.io/v1alpha1
kind: ModuleConfig
metadata:
  name: deckhouse
spec:
  version: 1
  settings:
    releaseChannel: Stable
    update:
      disruptionApprovalMode: Manual
```

In this mode, you must approve every minor potentially disruptive platform update (patch versions are not taken into account) using the `release.deckhouse.io/disruption-approved=true` annotation on the corresponding [DeckhouseRelease](/products/kubernetes-platform/documentation/v1/reference/api/cr.html#deckhouserelease) resource.

Example of approving the minor potentially disruptive platform update `v1.36.4`:

```shell
d8 k annotate DeckhouseRelease v1.36.4 release.deckhouse.io/disruption-approved=true
```
