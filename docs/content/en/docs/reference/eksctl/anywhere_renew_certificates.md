---
title: "anywhere renew certificates"
linkTitle: "anywhere renew certificates"
aliases:
    /docs/reference/eksctl/anywhere_renew_certificates/
---

## anywhere renew certificates

Renew EKS Anywhere cluster certificates

### Synopsis

This command is used to renew certificates for EKS Anywhere clusters

For detailed documentation on this command, see [Renew certificates using eksctl anywhere](../../../clustermgmt/certificate-management/eksctl-renew-certs/).

```
anywhere renew certificates -f <config-file> [flags]
```

### Options

```
      --component string   Component to renew certificates for (control-plane or etcd)
  -f, --filename string    Path to the certificate renewal configuration file
  -h, --help              help for certificates
```

### Options inherited from parent commands

```
  -v, --verbosity int   Set the log level verbosity
```

### SEE ALSO

* [anywhere renew](../anywhere_renew/)	 - Renew resources
