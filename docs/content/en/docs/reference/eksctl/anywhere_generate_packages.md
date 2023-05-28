---
title: "anywhere generate packages"
linkTitle: "anywhere generate packages"
---

## anywhere generate packages

Generate package(s) configuration

### Synopsis

Generates Kubernetes configuration files for curated packages

```
anywhere generate packages [flags] package
```

### Options

```
      --bundles-override string   Override default Bundles manifest (not recommended)
      --cluster string            Name of cluster for package generation
  -h, --help                      help for packages
      --kube-version string       Kubernetes Version of the cluster to be used. Format <major>.<minor>
      --kubeconfig string         Path to an optional kubeconfig file to use.
      --registry string           Used to specify an alternative registry for package generation
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere generate](../anywhere_generate/)	 - Generate resources

