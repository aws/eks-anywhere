---
title: "Cluster Autoscaler"
linkTitle: "Add Cluster Autoscaler"
weight: 13
date: 2022-10-20
description: >
  Install/upgrade/uninstall Cluster Autoscaler
---

{{< content "../prereq.md" >}}


## Install

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate the package configuration
   ```bash
   eksctl anywhere generate package cluster-autoscaler --cluster clusterName > cluster-autoscaler.yaml
   ```

1. Add the desired configuration to `cluster-autoscaler.yaml`

   Please see [complete configuration options]({{< relref "../../../reference/packagespec/cluster-autoscaler" >}}) for all configuration options and their default values.

    Example package file configuring a cluster autoscaler package to run on a management cluster.
    ```yaml
    apiVersion: packages.eks.amazonaws.com/v1alpha1
    kind: Package
    metadata:
      name: cluster-autoscaler
      namespace: eksa-packages-<cluster-name>
    spec:
      packageName: cluster-autoscaler
      targetNamespace: <namespace-to-install-component>
      config: |-
          cloudProvider: "clusterapi"
          autoDiscovery:
            clusterName: "<cluster-name>"
    ```


2. Install Cluster Autoscaler

   ```bash
   eksctl anywhere create packages -f cluster-autoscaler.yaml
   ```

3. Validate the installation

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```

   Example command output
   ```
   NAMESPACE                  NAME                          PACKAGE              AGE   STATE       CURRENTVERSION                                               TARGETVERSION                                                         DETAIL
   eksa-packages-mgmt-v-vmc   cluster-autoscaler            cluster-autoscaler   18h   installed   9.21.0-1.21-147e2a701f6ab625452fe311d5c94a167270f365         9.21.0-1.21-147e2a701f6ab625452fe311d5c94a167270f365 (latest)
   ```

## Update
To update package configuration, update cluster-autoscaler.yaml file, and run the following command:
```bash
eksctl anywhere apply package -f cluster-autoscaler.yaml
```

## Upgrade

Cluster Autoscaler will automatically be upgraded when a new bundle is activated.

## Uninstall

To uninstall Cluster Autoscaler, simply delete the package

```bash
eksctl anywhere delete package --cluster <cluster-name> cluster-autoscaler
```

## Installing Cluster Autoscaler on workload cluster

A few extra steps are required to install cluster autoscaler on a workload cluster instead of the management cluster.

First, retrieve the management cluster's kubeconfig secret:
```yaml
kubectl -n eksa-system get secrets <management-cluster-name>-kubeconfig -o yaml > mgmt-secret.yaml
```

Update the secret's namespace to the namespace in the workload cluster that you would like to deploy the cluster autoscaler to.
Then, apply the secret to the workload cluster.
```yaml
kubectl --kubeconfig /path/to/workload/kubeconfig apply -f mgmt-secret.yaml
```

Now apply this package configuration to the management cluster:
```yaml
apiVersion: packages.eks.amazonaws.com/v1alpha1
kind: Package
metadata:
    name: workload-cluster-autoscaler
    namespace: eksa-packages-<workload-cluster-name>
spec:
    packageName: cluster-autoscaler
    targetNamespace: <workload-cluster-namespace-to-install-components>
    config: |-
        cloudProvider: "clusterapi"
        autoDiscovery:
            clusterName: "<workload-cluster-name>"
        clusterAPIMode: "incluster-kubeconfig"
        clusterAPICloudConfigPath: "/etc/kubernetes/value"
        extraVolumeSecrets:
            cluster-autoscaler-cloud-config:
                mountPath: "/etc/kubernetes"
                name: "<management-cluster-name>-kubeconfig"
```