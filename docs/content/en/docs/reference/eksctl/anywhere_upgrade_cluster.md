---
title: "anywhere upgrade cluster"
linkTitle: "anywhere upgrade cluster"
---

## anywhere upgrade cluster

Upgrade workload cluster

### Synopsis

This command is used to upgrade workload clusters

```
anywhere upgrade cluster [flags]
```

### Options

```
      --bundles-override string   Override default Bundles manifest (not recommended)
  -f, --filename string           Filename that contains EKS-A cluster configuration
      --force-cleanup             Force deletion of previously created bootstrap cluster
  -z, --hardware-csv string       Path to a CSV file containing hardware data.
  -h, --help                      help for cluster
      --kubeconfig string         Management cluster kubeconfig file
      --no-timeouts               Disable timeout for all wait operations
  -w, --w-config string           Kubeconfig file to use when upgrading a workload cluster
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere upgrade](../anywhere_upgrade/)	 - Upgrade resources

