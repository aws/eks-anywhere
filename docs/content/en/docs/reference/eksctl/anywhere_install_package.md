---
title: "anywhere install package"
linkTitle: "anywhere install package"
---

## anywhere install package

Install package

### Synopsis

This command is used to Install a curated package. Use list to discover curated packages

```
anywhere install package [flags] package
```

### Options

```
      --bundles-override string   Override default Bundles manifest (not recommended)
      --cluster string            Target cluster for installation.
  -h, --help                      help for package
      --kube-version string       Kubernetes Version of the cluster to be used. Format <major>.<minor>
      --kubeconfig string         Path to an optional kubeconfig file to use.
  -n, --package-name string       Custom name of the curated package to install
      --registry string           Used to specify an alternative registry for discovery
      --set stringArray           Provide custom configurations for curated packages. Format key:value
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere install](../anywhere_install/)	 - Install resources to the cluster

