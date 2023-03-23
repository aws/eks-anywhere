---
title: "anywhere exp validate create cluster"
linkTitle: "anywhere exp validate create cluster"
---

## anywhere exp validate create cluster

Validate create cluster

### Synopsis

Use eksctl anywhere validate create cluster to validate the create cluster action

```
anywhere exp validate create cluster -f <cluster-config-file> [flags]
```

### Options

```
  -f, --filename string                  Filename that contains EKS-A cluster configuration
  -z, --hardware-csv string              Path to a CSV file containing hardware data.
  -h, --help                             help for cluster
      --tinkerbell-bootstrap-ip string   Override the local tinkerbell IP in the bootstrap cluster
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere exp validate create](../anywhere_exp_validate_create/)	 - Validate create resources

