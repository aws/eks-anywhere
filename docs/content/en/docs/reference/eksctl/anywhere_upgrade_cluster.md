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
      --bundles-override string             A path to a custom bundles manifest
      --control-plane-wait-timeout string   Override the default control plane wait timeout (default "1h0m0s")
      --external-etcd-wait-timeout string   Override the default external etcd wait timeout (default "1h0m0s")
  -f, --filename string                     Path that contains a cluster configuration
  -z, --hardware-csv string                 Path to a CSV file containing hardware data.
  -h, --help                                help for cluster
      --kubeconfig string                   Management cluster kubeconfig file
      --no-timeouts                         Disable timeout for all wait operations
      --node-startup-timeout string         (DEPRECATED) Override the default node startup timeout (Defaults to 20m for Tinkerbell clusters) (default "10m0s")
      --per-machine-wait-timeout string     Override the default machine wait timeout per machine (default "10m0s")
      --skip-validations stringArray        Bypass upgrade validations by name. Valid arguments you can pass are --skip-validations=pod-disruption,vsphere-user-privilege,eksa-version-skew
      --unhealthy-machine-timeout string    (DEPRECATED) Override the default unhealthy machine timeout (default "5m0s")
  -w, --w-config string                     Kubeconfig file to use when upgrading a workload cluster
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere upgrade](../anywhere_upgrade/)	 - Upgrade resources

