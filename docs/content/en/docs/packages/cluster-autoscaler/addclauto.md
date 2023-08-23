---
title: "Cluster Autoscaler"
linkTitle: "Add Cluster Autoscaler"
weight: 13
date: 2023-08-16
description: >
  Install/upgrade/uninstall Cluster Autoscaler
---

If you have not already done so, make sure your EKS Anywhere cluster meets the [package prerequisites.]({{< relref "../prereq" >}}) 

Refer to the [troubleshooting guide]({{< relref "../troubleshoot" >}}) in the event of a problem.

## Enable Cluster Autoscaling

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Ensure you have configured at least one worker node group in your cluster specification to enable autoscaling as outlined in [Autoscaling configuration.]({{< relref "../../getting-started/optional/autoscaling/" >}}) Cluster Autoscaler only works on node groups with an `autoscalingConfiguration` set:

    ```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
      name: <cluster-name>
    spec:
      ...
      workerNodeGroupConfigurations:
        - autoscalingConfiguration:
            minCount: 1
            maxCount: 5
          machineGroupRef:
            kind: VSphereMachineConfig
            name: <worker-machine-config-name>
          count: 1
          name: md-0
    ```

1. Generate the package configuration.
   ```bash
   eksctl anywhere generate package cluster-autoscaler --cluster <cluster-name> > cluster-autoscaler.yaml
   ```

1. Add the desired configuration to `cluster-autoscaler.yaml`. See [configuration options]({{< relref "../cluster-autoscaler" >}}) for all configuration options and their default values. See below for an example package file configuring a Cluster Autoscaler package.

    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: cluster-autoscaler-<cluster-name>
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: cluster-autoscaler
      targetNamespace: default
      config: |-
          cloudProvider: "clusterapi"
          autoDiscovery:
            clusterName: "<cluster-name>"
    ```

1. Install Cluster Autoscaler

   ```bash
   eksctl anywhere create packages -f cluster-autoscaler.yaml
   ```

1. Validate the installation

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```
   ```stdout
   NAMESPACE                  NAME                          PACKAGE              AGE   STATE       CURRENTVERSION                                               TARGETVERSION                                                         DETAIL
   eksa-packages-mgmt-v-vmc   cluster-autoscaler            cluster-autoscaler   18h   installed   9.21.0-1.21-147e2a701f6ab625452fe311d5c94a167270f365         9.21.0-1.21-147e2a701f6ab625452fe311d5c94a167270f365 (latest)
   ```

   To verify that autoscaling works, apply the deployment below. You must continue scaling pods until the deployment has pods in a pending state.
   This is when Cluster Autoscaler will begin to autoscale your machine deployment.
   This process may take a few minutes.
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/aws/eks-anywhere/d8575bbd2a85a6c6bbcb1a54868cf7790df56a63/test/framework/testdata/hpa_busybox.yaml
   kubectl scale deployment hpa-busybox-test --replicas 100
   ```

## Update
To update package configuration, update the `cluster-autoscaler.yaml` file and run the following command:
```bash
eksctl anywhere apply package -f cluster-autoscaler.yaml
```

## Update Worker Node Group Autoscaling Configuration
It is possible to change the autoscaling configuration of a worker node group by updating the `autoscalingConfiguration` in your cluster specification and running a cluster upgrade.

## Upgrade

The Cluster Autoscaler can be upgraded by PackageController's `activeBundle` field to a newer version.
The curated packages bundle contains the SHAs of the images and helm charts associated with a particular package. When a new version is activated, the Package Controller will reconcile all active packages to their newest versions as defined in the bundle.
The Curated Packages Controller automatically polls the bundle repository for new bundle resources.
The curated packages controller automatically polls for the latest bundle, but requires the activeBundle field on the PackageController resource to be updated before a new bundle will take effect and upgrade the resources.


## Uninstall

To uninstall Cluster Autoscaler, delete the package

```bash
eksctl anywhere delete package --cluster <cluster-name> cluster-autoscaler
```
