---
title: "anywhere copy packages"
linkTitle: "anywhere copy packages"
---

## anywhere copy packages

Copy curated package images and charts from source registries to a destination registry

### Synopsis

Copy all the EKS Anywhere curated package images and helm charts from source registries to a destination registry. Registry credentials are fetched from docker config.

```
anywhere copy packages <destination-registry> [flags]
```

### Options

```
      --dry-run                     Dry run will show what artifacts would be copied, but not actually copy them
      --dst-insecure                Skip TLS verification against the destination registry
      --dst-plain-http              Whether or not to use plain http for destination registry
  -h, --help                        help for packages
      --kube-version string         The kubernetes version of the package bundle to copy
      --src-chart-registry string   The source registry that stores helm charts (default src-image-registry)
      --src-image-registry string   The source registry that stores container images
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere copy](../anywhere_copy/)	 - Copy resources

