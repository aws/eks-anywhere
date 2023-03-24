---
title: "anywhere generate support-bundle"
linkTitle: "anywhere generate support-bundle"
---

## anywhere generate support-bundle

Generate a support bundle

### Synopsis

This command is used to create a support bundle to troubleshoot a cluster

```
anywhere generate support-bundle -f my-cluster.yaml [flags]
```

### Options

```
      --bundle-config string   Bundle Config file to use when generating support bundle
  -f, --filename string        Filename that contains EKS-A cluster configuration
  -h, --help                   help for support-bundle
      --since string           Collect pod logs in the latest duration like 5s, 2m, or 3h.
      --since-time string      Collect pod logs after a specific datetime(RFC3339) like 2021-06-28T15:04:05Z
  -w, --w-config string        Kubeconfig file to use when creating support bundle for a workload cluster
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere generate](../anywhere_generate/)	 - Generate resources

