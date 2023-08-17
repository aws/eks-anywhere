---
title: "anywhere import images"
linkTitle: "anywhere import images"
---

## anywhere import images

Import images and charts to a registry from a tarball

### Synopsis

Import all the images and helm charts necessary for EKS Anywhere clusters into a registry.
Use this command in conjunction with download images, passing it output tarball as input to this command.

```
anywhere import images [flags]
```

### Options

```
  -b, --bundles string     Bundles file to read artifact dependencies from
  -h, --help               help for images
      --include-packages   Flag to indicate inclusion of curated packages in imported images (DEPRECATED: use copy packages command)
  -i, --input string       Input tarball containing all images and charts to import
      --insecure           Flag to indicate skipping TLS verification while pushing helm charts and bundles
  -r, --registry string    Registry where to import images and charts
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere import](../anywhere_import/)	 - Import resources

