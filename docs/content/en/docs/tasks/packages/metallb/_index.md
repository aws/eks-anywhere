---
title: "MetalLB"
linkTitle: "Add MetalLB"
weight: 13
date: 2022-04-12
description: >
  Install/upgrade/uninstall MetalLB
---

{{% alert title="Important" color="warning" %}}

If your cluster was created with a release of EKS Anywhere prior to v0.9.0, you may need to [install the package controller.]({{< relref ".." >}})

{{% /alert %}}

## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package metallb --source cluster > metallb.yaml
   ```

1. Add the desired configuration to `metallb.yaml`

   Please see [complete configuration options]({{< relref "../../../reference/packagespec/metallb" >}}) for all configuration options and their default values.

    Example package file with bgp configuration:
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: mylb
      namespace: eksa-packages
    spec:
      packageName: metallb
      config: |
        peers:
          - peer-address: 10.220.0.2
            peer-asn: 65000
            my-asn: 65002
        address-pools:
          - name: default
            protocol: bgp
            addresses:
              - 10.220.0.90/32
              - 10.220.0.97-10.220.0.120
    ```
    Example package file with ARP configuration:
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: mylb
      namespace: eksa-packages
    spec:
      packageName: metallb
      config: |
        address-pools:
          - name: default
            protocol: layer2
            addresses:
              - 10.220.0.90/32
              - 10.220.0.97-10.220.0.120
    ```


1. Install MetalLB

   ```bash
   eksctl anywhere create packages -f metallb.yaml
   ```

1. Validate the installation

   ```bash
   eksctl anywhere get packages
   ```

   Example command output
   ```
   NAME   PACKAGE   AGE   STATE       CURRENTVERSION                                    TARGETVERSION                                              DETAIL
   mylb   metallb   22h   installed   0.12.1-ce5b5de19014202cebd4ab4c091830a3b6dfea06   0.12.1-ce5b5de19014202cebd4ab4c091830a3b6dfea06 (latest)
   ```


## Upgrade

MetalLB will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall MetalLB, simply delete the package

```bash
eksctl anywhere delete package mylb
```
