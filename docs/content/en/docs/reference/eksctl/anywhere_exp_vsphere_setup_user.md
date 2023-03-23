---
title: "anywhere exp vsphere setup user"
linkTitle: "anywhere exp vsphere setup user"
---

## anywhere exp vsphere setup user

Setup vSphere user

### Synopsis

Use eksctl anywhere vsphere setup user to configure EKS Anywhere vSphere user

```
anywhere exp vsphere setup user -f <config-file> [flags]
```

### Options

```
  -f, --filename string   Filename containing vsphere setup configuration
      --force             Force flag. When set, setup user will proceed even if the group and role objects already exist. Mutually exclusive with --password flag, as it expects the user to already exist. default: false
  -h, --help              help for user
  -p, --password string   Password for creating new user
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere exp vsphere setup](../anywhere_exp_vsphere_setup/)	 - Setup vSphere objects

