---
title: "Emissary Ingress"
linkTitle: "Add Emissary Ingress"
weight: 13
date: 2022-04-12
description: >
  Install/upgrade/uninstall Emissary Ingress
---

If you have not already done so, make sure your cluster meets the [package prerequisites.]({{< relref "../prereq" >}})
Be sure to refer to the [troubleshooting guide]({{< relref "../troubleshoot" >}}) in the event of a problem.

  {{% alert title="Important" color="warning" %}}
   * Starting at `eksctl anywhere` version `v0.12.0`, packages on workload clusters are remotely managed by the management cluster.
   * While following this guide to install packages on a workload cluster, please make sure the `kubeconfig` is pointing to the management cluster that was used to create the workload cluster. The only exception is the `kubectl create namespace` command below, which should be run with `kubeconfig` pointing to the workload cluster.
   {{% /alert %}}

## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package emissary --cluster <cluster-name> > emissary.yaml
   ```

1. Add the desired configuration to `emissary.yaml`

   Please see [complete configuration options]({{< relref "../emissary" >}}) for all configuration options and their default values.

    Example package file with standard configuration.
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: emissary
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: emissary
    ```

1. Install Emissary

   ```bash
   eksctl anywhere create packages -f emissary.yaml
   ```

1. Validate the installation

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```

   Example command output
   ```
   NAMESPACE     NAME       PACKAGE    AGE     STATE       CURRENTVERSION                                   TARGETVERSION                                              DETAIL
   eksa-packages emissary   emissary   2m57s   installed   3.0.0-a507e09c2a92c83d65737835f6bac03b9b341467   3.0.0-a507e09c2a92c83d65737835f6bac03b9b341467 (latest)
   ```

## Update
To update package configuration, update emissary.yaml file, and run the following command:
```bash
eksctl anywhere apply package -f emissary.yaml
```

## Upgrade

Emissary will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall Emissary, simply delete the package

```bash
eksctl anywhere delete package --cluster <cluster-name> emissary
```
