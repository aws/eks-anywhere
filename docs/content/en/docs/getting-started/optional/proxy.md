---
title: "Proxy"
linkTitle: "Proxy"
weight: 35
aliases:
    /docs/reference/clusterspec/optional/proxy/
description: >
  EKS Anywhere cluster yaml specification proxy configuration reference
---

## Proxy support (optional)

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	 ✓   |     ✓      |  ✓   |

You can configure EKS Anywhere to use a proxy to connect to the Internet. This is the
generic template with proxy configuration for your reference:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
   name: my-cluster-name
spec:
   ...
   proxyConfiguration:
      httpProxy: http-proxy-ip:port
      httpsProxy: https-proxy-ip:port
      noProxy:
      - list of no proxy endpoints
```

### Configuring Docker daemon
EKS Anywhere will proxy for you given the above configuration file.
However, to successfully use EKS Anywhere you will also need to ensure your Docker daemon is configured to use the proxy.

This generally means updating your daemon to launch with the HTTPS_PROXY, HTTP_PROXY, and NO_PROXY environment variables.

For an example of how to do this with systemd, please see Docker's documentation [here](https://docs.docker.com/config/daemon/systemd/#httphttps-proxy).

### Configuring EKS Anywhere proxy without config file
For commands using a cluster config file, EKS Anywhere will derive its proxy config from the cluster configuration file.

However, for commands that do not utilize a cluster config file, you can set the following environment variables:
```bash
export HTTPS_PROXY=https-proxy-ip:port
export HTTP_PROXY=http-proxy-ip:port
export NO_PROXY=no-proxy-domain.com,another-domain.com,localhost
```


## Proxy Configuration Spec Details
### __proxyConfiguration__ (required)
* __Description__: top level key; required to use proxy.
* __Type__: object

### __httpProxy__ (required)
* __Description__: HTTP proxy to use to connect to the internet; must be in the format IP:port
* __Type__: string
* __Example__: ```httpProxy: 192.168.0.1:3218```

### __httpsProxy__ (required)
* __Description__: HTTPS proxy to use to connect to the internet; must be in the format IP:port
* __Type__: string
* __Example__: ```httpsProxy: 192.168.0.1:3218```

### __noProxy__ (optional)
* __Description__: list of endpoints that should not be routed through the proxy; can be an IP, CIDR block, or a domain name
* __Type__: list of strings
* __Example__
```yaml
  noProxy:
   - localhost
   - 192.168.0.1
   - 192.168.0.0/16
   - .example.com
```

{{% alert title="Note" color="primary" %}}
- For Bottlerocket OS, it is required to add the local subnet CIDR range in the `noProxy` list.
- For Bare Metal provider, it is required to host hook images locally which should be accessible by admin machines as well as all the nodes without using proxy configuration. Please refer to the documentation for getting hook images [here]({{< relref "../../osmgmt/artifacts/#hookos-kernel-and-initial-ramdisk-for-bare-metal" >}}).
{{% /alert %}}
