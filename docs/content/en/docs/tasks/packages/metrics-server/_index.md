---
title: "Metrics Server"
linkTitle: "Add Metrics Server"
weight: 13
date: 2022-10-20
description: >
  Install/upgrade/uninstall Metrics Server
---

{{< content "../prereq.md" >}}


## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package metrics-server --cluster clusterName > metrics-server.yaml
   ```

1. Add the desired configuration to `metrics-server.yaml`

   Please see [complete configuration options]({{< relref "../../../reference/packagespec/metrics-server" >}}) for all configuration options and their default values.

    Example package file configuring a cluster autoscaler package to run on a management cluster.
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: metrics-server
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: metrics-server
      targetNamespace: <namespace-to-install-component>
      config: |-
        args:
          - "--kubelet-insecure-tls"
    ```


1. Install Metrics Server

   ```bash
   eksctl anywhere create packages -f metrics-server.yaml
   ```

1. Validate the installation

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```

   Example command output
   ```
   NAME                   PACKAGE              AGE   STATE        CURRENTVERSION                                                     TARGETVERSION                                                               DETAIL
   metrics-server         metrics-server       8h    installed    0.6.1-eks-1-23-6-b4c2524fabb3dd4c5f9b9070a418d740d3e1a8a2          0.6.1-eks-1-23-6-b4c2524fabb3dd4c5f9b9070a418d740d3e1a8a2 (latest)
   ```

## Update
To update package configuration, update metrics-server.yaml file, and run the following command:
```bash
eksctl anywhere apply package -f metrics-server.yaml
```

## Upgrade

Metrics Server will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall Metrics Server, simply delete the package

```bash
eksctl anywhere delete package --cluster <cluster-name> metrics-server
```
