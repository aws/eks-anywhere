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
      --bundles-override string             A path to a custom bundles manifest
      --control-plane-wait-timeout string   Override the default control plane wait timeout (default "1h0m0s")
      --external-etcd-wait-timeout string   Override the default external etcd wait timeout (default "1h0m0s")
  -f, --filename string                     Path that contains a cluster configuration
  -z, --hardware-csv string                 Path to a CSV file containing hardware data.
  -h, --help                                help for cluster
      --install-packages string             Location of curated packages configuration files to install to the cluster
      --kubeconfig string                   Management cluster kubeconfig file
      --no-timeouts                         Disable timeout for all wait operations
      --node-startup-timeout string         (DEPRECATED) Override the default node startup timeout (Defaults to 20m for Tinkerbell clusters) (default "10m0s")
      --per-machine-wait-timeout string     Override the default machine wait timeout per machine (default "10m0s")
      --skip-ip-check                       Skip check for whether cluster control plane ip is in use
      --skip-validations stringArray        Bypass create validations by name. Valid arguments you can pass are --skip-validations=vsphere-user-privilege
      --tinkerbell-bootstrap-ip string      The IP used to expose the Tinkerbell stack from the bootstrap cluster
      --unhealthy-machine-timeout string    (DEPRECATED) Override the default unhealthy machine timeout (default "5m0s")
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere create](../anywhere_create/)	 - Create resources

