---
title: "Cert-Manager"
linkTitle: "Add Cert-Manager"
weight: 13
aliases:
    /docs/reference/packagespec/cert-manager/
date: 2022-10-21
description: >
   Install/update/upgrade/uninstall Cert-Manager
---

If you have not already done so, make sure your cluster meets the [package prerequisites.]({{< relref "../prereq" >}})
Be sure to refer to the [troubleshooting guide]({{< relref "../troubleshoot" >}}) in the event of a problem.

  {{% alert title="Important" color="warning" %}}
   * Starting at `eksctl anywhere` version `v0.12.0`, packages on workload clusters are remotely managed by the management cluster.
   * While following this guide to install packages on a workload cluster, please make sure the `kubeconfig` is pointing to the management cluster that was used to create the workload cluster. The only exception is the `kubectl create namespace` command below, which should be run with `kubeconfig` pointing to the workload cluster.
   {{% /alert %}}

## Install on workload cluster

**NOTE: The cert-manager package can only be installed on a workload cluster**
<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package cert-manager --cluster <cluster-name> > cert-manager.yaml
   ```

1. Add the desired configuration to `cert-manager.yaml`

   Please see [complete configuration options]({{< relref "../cert-manager" >}}) for all configuration options and their default values.

   Example package file configuring a cert-manager package to run on a workload cluster.
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: my-cert-manager
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: cert-manager
      targetNamespace: <namespace-to-install-component>
    ```


1. Install Cert-Manager

   ```bash
   eksctl anywhere create packages -f cert-manager.yaml
   ```

1. Validate the installation

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```

   Example command output
   ```
   NAME                          PACKAGE              AGE   STATE       CURRENTVERSION                                               TARGETVERSION                                                         DETAIL
   my-cert-manager               cert-manager         15s   installed   1.9.1-dc0c845b5f71bea6869efccd3ca3f2dd11b5c95f               1.9.1-dc0c845b5f71bea6869efccd3ca3f2dd11b5c95f (latest)
   ```

## Update
To update package configuration, update cert-manager.yaml file, and run the following command:
```bash
eksctl anywhere apply package -f cert-manager.yaml
```

## Upgrade

Cert-Manager will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall cert-manager, simply delete the package

```bash
eksctl anywhere delete package --cluster <cluster-name> cert-manager
```
