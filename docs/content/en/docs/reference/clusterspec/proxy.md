---
title: "Proxy configuration"
linkTitle: "Proxy"
weight: 90
description: >
  EKS Anywhere cluster yaml specification proxy configuration reference
---

## Proxy support (optional)
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
