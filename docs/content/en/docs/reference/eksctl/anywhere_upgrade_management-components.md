---
title: "anywhere upgrade management-components"
linkTitle: "anywhere upgrade management-components"
---

## anywhere upgrade management-components

Upgrade management components in a management cluster

### Synopsis

The term _management components_ encompasses all Kubernetes controllers and their CRDs present in the management cluster that are responsible for reconciling your EKS Anywhere (EKS-A) cluster. This command is specifically designed to facilitate the upgrade of these management components. Post this upgrade, the cluster itself can be upgraded by updating the 'eksaVersion' field in your EKS-A Cluster object.

```
anywhere upgrade management-components [flags]
```

### Options

```
      --bundles-override string   A path to a custom bundles manifest
  -f, --filename string           Path that contains a cluster configuration
  -h, --help                      help for management-components
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere upgrade](../anywhere_upgrade/)	 - Upgrade resources
* [Upgrade management components]({{< relref "../../clustermgmt/cluster-upgrades/management-components-upgrade" >}})
