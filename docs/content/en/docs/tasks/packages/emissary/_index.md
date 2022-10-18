---
title: "Emissary Ingress"
linkTitle: "Add Emissary Ingress"
weight: 13
date: 2022-04-12
description: >
  Install/upgrade/uninstall Emissary Ingress
---

{{< content "../prereq.md" >}}


## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package emissary --cluster clusterName > emissary.yaml
   ```

1. Add the desired configuration to `emissary.yaml`

   Please see [complete configuration options]({{< relref "../../../reference/packagespec/emissary" >}}) for all configuration options and their default values.

    Example package file with standard configuration.
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: emissary
      namespace: eksa-packages
    spec:
      packageName: emissary
    ```

2. Install Emissary

   ```bash
   eksctl anywhere create packages -f emissary.yaml
   ```

3. Validate the installation

   ```bash
   eksctl anywhere get packages
   ```

   Example command output
   ```
   NAMESPACE     NAME       PACKAGE    AGE     STATE       CURRENTVERSION                                   TARGETVERSION                                              DETAIL
   eksa-packages emissary   emissary   2m57s   installed   3.0.0-a507e09c2a92c83d65737835f6bac03b9b341467   3.0.0-a507e09c2a92c83d65737835f6bac03b9b341467 (latest)
   ```


## Upgrade

Emissary will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall Emissary, simply delete the package

```bash
eksctl anywhere delete package emissary
```
