---
title: "Operating system"
linkTitle: "Operating system"
weight: 10
aliases:
    /docs/reference/clusterspec/optional/hostOSConfig/
description: >
  EKS Anywhere cluster yaml specification for host OS configuration
---

## Host OS Configuration
You can configure certain host OS settings through EKS Anywhere.

#### Provider support details
|                | vSphere | Bare Metal | Nutanix | CloudStack | Snow |
|:--------------:|:-------:|:----------:|:-------:|:----------:|:----:|
| **Supported?** |   ✓	    |     ✓      |   	     |            |      |

{{% alert title="Note" color="primary" %}}
Settings under `bottlerocketConfiguration` are only supported for `osFamily: bottlerocket`
{{% /alert %}}

The following cluster spec shows an example of how to configure host OS settings:
```yaml
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig        # Replace "VSphereMachineConfig" with "TinkerbellMachineConfig" for Tinkerbell clusters
metadata:
  name: machine-config
spec:
  ...
  hostOSConfiguration:
    ntpConfiguration:
      servers:
        - time-a.ntp.local
        - time-b.ntp.local
    certBundles:
    - name: "bundle_1"
      data: |
        -----BEGIN CERTIFICATE-----
        MIIF1DCCA...
        ...
        es6RXmsCj...
        -----END CERTIFICATE-----

        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
    bottlerocketConfiguration:
      kubernetes:
        allowedUnsafeSysctls:
          - "net.core.somaxconn"
          - "net.ipv4.ip_local_port_range"
        clusterDNSIPs:
          - 10.96.0.10
        maxPods: 100
      kernel:
        sysctlSettings:
          net.core.wmem_max: "8388608"
          net.core.rmem_max: "8388608"
          ...
      boot:
        bootKernelParameters:
          slub_debug:
          - "options,slabs"
          ...
```

## Host OS Configuration Spec Details
### `hostOSConfiguration`
Top level object used for host OS configurations.

  * #### `ntpConfiguration`
    Key used for configuring NTP servers on your EKS Anywhere cluster nodes.

    * ##### `servers`
      Servers is a list of NTP servers that should be configured on EKS Anywhere cluster nodes.
  
  * #### `certBundles`
    Key used for configuring custom trusted CA certs on your EKS Anywhere cluster nodes. Multiple cert bundles can be configured.
      
    {{% alert title="Note" color="primary" %}}
    This setting is _only valid_ for Botlerocket OS.
    {{% /alert %}}

    * ##### `name`
    Name of the cert bundle that should be configured on EKS Anywhere cluster nodes. This must be a unique name for each entry

    * ##### `data`
    Data of the cert bundle that should be configured on EKS Anywhere cluster nodes. This takes in a PEM formatted cert bundle and can contain more than one CA cert per entry.

<br>

  * #### `bottlerocketConfiguration`
    Key used for configuring Bottlerocket-specific settings on EKS Anywhere cluster nodes. These settings are _only valid_ for Bottlerocket.

    * ##### `kubernetes`
      Key used for configuring Bottlerocket Kubernetes settings.

      * ##### `allowedUnsafeSysctls`
        List of unsafe sysctls that should be enabled on the node.

      * ##### `clusterDNSIPs`
        List of IPs of DNS service(s) running in the kubernetes cluster.

      * ##### `maxPods`
        Maximum number of pods that can be scheduled on each node.

    * ##### `kernel`
      Key used for configuring Bottlerocket Kernel settings.
       
      * ##### `sysctlSettings`
        Map of kernel sysctl settings that should be enabled on the node.

    * ##### `boot`
      Key used for configuring Bottlerocket Boot settings.

      {{% alert title="Note" color="primary" %}}
      This setting has not been validated for baremetal.
      {{% /alert %}}

      * ##### `bootKernelParameters`
        Map of Boot Kernel parameters Bottlerocket should configure.
