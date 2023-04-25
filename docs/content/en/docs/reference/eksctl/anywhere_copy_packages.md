---
title: "anywhere copy packages"
linkTitle: "anywhere copy packages"
---

## anywhere copy packages

Copy curated package images and charts from a source to a destination

### Synopsis

Copy all the EKS Anywhere curated package images and helm charts from a source to a destination.

```
anywhere copy packages <destination-registry> [flags]
```

### Options

```
      --aws-region string   Region to copy images from
  -b, --bundle string       EKS-A bundle file to read artifact dependencies from
      --dry-run             Dry run copy to print images that would be copied
      --dst-cert string     TLS certificate for destination registry
  -h, --help                help for packages
      --insecure            Skip TLS verification while copying images and charts
      --src-cert string     TLS certificate for source registry
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere copy](../anywhere_copy/)	 - Copy resources

