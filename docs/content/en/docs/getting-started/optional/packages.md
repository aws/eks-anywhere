---
title: "Package controller"
linkTitle: "Package controller"
weight: 55
aliases:
    /docs/reference/clusterspec/optional/packages/
description: >
  EKS Anywhere cluster yaml specification for package controller configuration
---

## Package Controller Configuration (optional)
You can specify EKS Anywhere package controller configurations. For more on the curated packages and the package controller, visit the [package management]({{< relref "../../packages" >}}) documentation.

The following cluster spec shows an example of how to configure package controller:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
  packages:
    disable: false
    controller:
      resources:
        requests:
          cpu: 100m
          memory: 50Mi
        limits:
          cpu: 750m
          memory: 450Mi


```
## Package Controller Configuration Spec Details
### __packages__ (optional)
* __Description__: Top level key; required controller configuration.
* __Type__: object

### __packages.disable__ (optional)
* __Description__: Disable the package controller.
* __Type__: bool
* __Example__: ```disable: true```

### __packages.controller__ (optional)
* __Description__: Disable the package controller.
* __Type__: object

### __packages.controller.resources__ (optional)
* __Description__: Resources for the package controller.
* __Type__: object

### __packages.controller.resources.limits__ (optional)
* __Description__: Resource limits.
* __Type__: object

### __packages.controller.resources.limits.cpu__ (optional)
* __Description__: CPU limit.
* __Type__: string

### __packages.controller.resources.limits.memory__ (optional)
* __Description__: Memory limit.
* __Type__: string

### __packages.controller.resources.requests__ (optional)
* __Description__: Requested resources.
* __Type__: object

### __packages.controller.resources.requests.cpu__ (optional)
* __Description__: Requested cpu.
* __Type__: string

### __packages.controller.resources.limits.memory__ (optional)
* __Description__: Requested memory.
* __Type__: string

### __packages.cronjob__ (optional)
* __Description__: Disable the package controller.
* __Type__: object

### __packages.cronjob.disable__ (optional)
* __Description__: Disable the cron job.
* __Type__: bool
* __Example__: ```disable: true```

### __packages.cronjob.resources__ (optional)
* __Description__: Resources for the package controller.
* __Type__: object

### __packages.cronjob.resources.limits__ (optional)
* __Description__: Resource limits.
* __Type__: object

### __packages.cronjob.resources.limits.cpu__ (optional)
* __Description__: CPU limit.
* __Type__: string

### __packages.cronjob.resources.limits.memory__ (optional)
* __Description__: Memory limit.
* __Type__: string

### __packages.cronjob.resources.requests__ (optional)
* __Description__: Requested resources.
* __Type__: object

### __packages.cronjob.resources.requests.cpu__ (optional)
* __Description__: Requested cpu.
* __Type__: string

### __packages.cronjob.resources.limits.memory__ (optional)
* __Description__: Requested memory.
* __Type__: string
