---
title: "anywhere upgrade plan management-components"
linkTitle: "anywhere upgrade plan management-components"
---

## anywhere upgrade plan management-components

Lists the current and target versions for upgrading the management components in a management cluster

### Synopsis

Provides a list of current and target versions for upgrading the management components in a management cluster. The term _management components_ encompasses all Kubernetes controllers and their CRDs present in the management cluster that are responsible for reconciling your EKS Anywhere (EKS-A) cluster

```
anywhere upgrade plan management-components [flags]
```

### Options

```
      --bundles-override string   Override default Bundles manifest (not recommended)
  -f, --filename string           Filename that contains EKS-A cluster configuration
  -h, --help                      help for management-components
      --kubeconfig string         Management cluster kubeconfig file
  -o, --output string             Output format: text|json (default "text")
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere upgrade plan](../anywhere_upgrade_plan/)	 - Provides information for a resource upgrade

