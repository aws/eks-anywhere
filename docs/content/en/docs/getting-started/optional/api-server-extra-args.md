---
title: "API Server Extra Args"
linkTitle: "API Server Extra Args"
weight: 60
description: >
  EKS Anywhere cluster yaml specification for Kubernetes API Server Extra Args reference
---

## API Server Extra Args support (optional)

As of EKS Anywhere version v0.20.0, you can pass additional flags to configure the Kubernetes API server in your EKS Anywhere clusters.

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

In order to configure a cluster with API Server extra args, you need to configure your cluster by updating the cluster configuration file to include the details below. The feature flag `API_SERVER_EXTRA_ARGS_ENABLED=true` needs to be set as an environment variable.

This is a generic template with some example API Server extra args configuration below for reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
    ...
    controlPlaneConfiguration:
        apiServerExtraArgs:
            ...
            disable-admission-plugins: "DefaultStorageClass,DefaultTolerationSeconds"
            enable-admission-plugins: "NamespaceAutoProvision,NamespaceExists"
```

The above example configures the `disable-admission-plugins` and `enable-admission-plugins` options of the API Server to enable additional admission plugins or disable some of the default ones. You can configure any of the API Server options using the above template.

### controlPlaneConfiguration.apiServerExtraArgs (optional)
Reference the [Kubernetes documentation](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options) for the list of flags that can be configured for the Kubernetes API server in EKS Anywhere