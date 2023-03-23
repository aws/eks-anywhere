---
title: "anywhere download artifacts"
linkTitle: "anywhere download artifacts"
---

## anywhere download artifacts

Download EKS Anywhere artifacts/manifests to a tarball on disk

### Synopsis

This command is used to download the S3 artifacts from an EKS Anywhere bundle manifest and package them into a tarball

```
anywhere download artifacts [flags]
```

### Options

```
      --bundles-override string   Override default Bundles manifest (not recommended)
  -d, --download-dir string       Directory to download the artifacts to (default "eks-anywhere-downloads")
      --dry-run                   Print the manifest URIs without downloading them
  -f, --filename string           [Deprecated] Filename that contains EKS-A cluster configuration
  -h, --help                      help for artifacts
  -r, --retain-dir                Do not delete the download folder after creating a tarball
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere download](../anywhere_download/)	 - Download resources

