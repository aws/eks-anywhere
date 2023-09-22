---
title: "anywhere generate tinkerbelltemplateconfig"
linkTitle: "anywhere generate tinkerbelltemplateconfig"
---

## anywhere generate tinkerbelltemplateconfig

Generate TinkerbellTemplateConfig objects

### Synopsis

Generate TinkerbellTemplateConfig objects for your cluster specification.

The TinkerbellTemplateConfig is part of an EKS Anywhere bare metal cluster 
specification. When no template config is specified on TinkerbellMachineConfig
objects, EKS Anywhere generates the template config internally. The template 
config defines the actions for provisioning a bare metal host such as streaming 
an OS image to disk. Actions vary based on the OS - see the EKS Anywhere 
documentation for more details on the individual actions.

The template config include it in your bare metal cluster specification and
reference it in the TinkerbellMachineConfig object using the .spec.templateRef
field.


```
anywhere generate tinkerbelltemplateconfig [flags]
```

### Options

```
      --bundles-override string          A path to a custom bundles manifest
  -f, --filename string                  Path that contains a cluster configuration
  -h, --help                             help for tinkerbelltemplateconfig
      --tinkerbell-bootstrap-ip string   The IP used to expose the Tinkerbell stack from the bootstrap cluster
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere generate](../anywhere_generate/)	 - Generate resources

