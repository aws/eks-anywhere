---
title: "anywhere download images"
linkTitle: "anywhere download images"
---

## anywhere download images

Download all eks-a images to disk

### Synopsis

Creates a tarball containing all necessary images
to create an eks-a cluster for any of the supported
Kubernetes versions.

```
anywhere download images [flags]
```

### Options

```
      --bundles-override string   Override default Bundles manifest (not recommended)
  -h, --help                      help for images
      --include-packages          this flag no longer works, use copy packages instead (DEPRECATED: use copy packages command)
      --insecure                  Flag to indicate skipping TLS verification while downloading helm charts
  -o, --output string             Output tarball containing all downloaded images
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere download](../anywhere_download/)	 - Download resources

