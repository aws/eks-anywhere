---
title: "MetalLB"
linkTitle: "Add MetalLB"
weight: 13
date: 2022-04-12
description: >
  Install/upgrade/uninstall MetalLB
---

{{< content "../prereq.md" >}}


## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package metallb --cluster clusterName > metallb.yaml
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
        IPAddressPools:
          - name: default
            addresses:
              - 10.220.0.93/32
              - 10.220.0.97-10.220.0.120
        BGPAdvertisements:
          - IPAddressPools:
            - default
        BGPPeers:
          - peerAddress: 10.220.0.2
            peerASN: 65000
            myASN: 65002
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
        IPAddressPools:
          - name: default
            addresses:
              - 10.220.0.93/32
              - 10.220.0.97-10.220.0.120
        L2Advertisements:
          - IPAddressPools:
            - default
    ```

1. Create the namespace
  (If overriding `targetNamespace`, change `metallb-system` to the value of `targetNamespace`)
   ```bash
   kubectl create namespace metallb-system
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
