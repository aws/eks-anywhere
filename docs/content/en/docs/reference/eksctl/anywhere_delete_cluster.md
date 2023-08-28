---
title: "anywhere delete cluster"
linkTitle: "anywhere delete cluster"
---

## anywhere delete cluster

Workload cluster

### Synopsis

This command is used to delete workload clusters created by eksctl anywhere

```
anywhere delete cluster (<cluster-name>|-f <config-file>) [flags]
```

### Options

```
      --bundles-override string   Override default Bundles manifest (not recommended)
  -f, --filename string           Filename that contains EKS-A cluster configuration, required if <cluster-name> is not provided
  -h, --help                      help for cluster
      --kubeconfig string         kubeconfig file pointing to a management cluster
  -w, --w-config string           Kubeconfig file to use when deleting a workload cluster
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere delete](../anywhere_delete/)	 - Delete resources

