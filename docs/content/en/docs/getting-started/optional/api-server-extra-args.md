---
title: "API Server Extra Args"
linkTitle: "API Server Extra Args"
weight: 10
aliases:
    /docs/reference/clusterspec/optional/etcd/
description: >
  EKS Anywhere cluster yaml specification API Server Extra Args reference
---

## API Server Extra Args support (optional)

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

In order to configure a cluster with api server extra args, you need to configure your cluster by updating the cluster configuration file to include the details below. The feature flag `API_SERVER_EXTRA_ARGS_ENABLED=true` needs to be set.

This is a generic template with some example api server extra args configuration below for reference:
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
            service-account-issuer: "https://issuer-url"
            service-account-jwks-uri: "https://issuer-url/openid/v1/jwks"
```

### controlPlaneConfiguration.apiServerExtraArgs (required)
A list of [flags](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options) to configure for the api server.