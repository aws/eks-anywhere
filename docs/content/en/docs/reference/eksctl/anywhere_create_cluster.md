---
title: "anywhere create cluster"
linkTitle: "anywhere create cluster"
---

## anywhere create cluster

Create workload cluster

### Synopsis

This command is used to create workload clusters

```
anywhere create cluster -f <cluster-config-file> [flags]
```

### Options

```
      --bundles-override string          Override default Bundles manifest (not recommended)
  -f, --filename string                  Filename that contains EKS-A cluster configuration
      --force-cleanup                    Force deletion of previously created bootstrap cluster
  -z, --hardware-csv string              Path to a CSV file containing hardware data.
  -h, --help                             help for cluster
      --install-packages string          Location of curated packages configuration files to install to the cluster
      --kubeconfig string                Management cluster kubeconfig file
      --no-timeouts                      Disable timeout for all wait operations
      --skip-ip-check                    Skip check for whether cluster control plane ip is in use
      --tinkerbell-bootstrap-ip string   Override the local tinkerbell IP in the bootstrap cluster
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere create](../anywhere_create/)	 - Create resources

