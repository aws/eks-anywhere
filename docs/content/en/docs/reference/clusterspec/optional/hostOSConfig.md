---
title: "Host OS Configuration"
linkTitle: "Host OS Configuration"
weight: 90
description: >
  EKS Anywhere cluster yaml specification for host OS configuration
---

## Host OS Configuration
You can configure certain host OS settings through EKS Anywhere.

{{% alert title="Note" color="primary" %}}
Currently, these settings are only supported for vSphere and Tinkerbell providers.<br>
Additionally, settings under `bottlerocketConfiguration` are only supported for `osFamily: bottlerocket`
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
    bottlerocketConfiguration:
      kubernetes:
        allowedUnsafeSysctls:
          - "net.core.somaxconn"
          - "net.ipv4.ip_local_port_range"
        clusterDNSIPs:
          - 10.96.0.10
        maxPods: 100
```

## Host OS Configuration Spec Details
### `hostOSConfiguration`
Top level object used for host OS configurations.

  * #### `ntpConfiguration`
    Key used for configuring NTP servers on your EKS Anywhere cluster nodes.

    * ##### `servers`
      Servers is a list of NTP servers that should be configured on EKS Anywhere cluster nodes.

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
