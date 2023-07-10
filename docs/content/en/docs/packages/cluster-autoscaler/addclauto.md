---
title: "Cluster Autoscaler"
linkTitle: "Add Cluster Autoscaler"
weight: 13
date: 2022-10-20
description: >
  Install/upgrade/uninstall Cluster Autoscaler
---

If you have not already done so, make sure your cluster meets the [package prerequisites.]({{< relref "../prereq" >}})
Be sure to refer to the [troubleshooting guide]({{< relref "../troubleshoot" >}}) in the event of a problem.

  {{% alert title="Important" color="warning" %}}
   * Starting at `eksctl anywhere` version `v0.12.0`, packages on workload clusters are remotely managed by the management cluster.
   * While following this guide to install packages on a workload cluster, please make sure the `kubeconfig` is pointing to the management cluster that was used to create the workload cluster. The only exception is the `kubectl create namespace` command below, which should be run with `kubeconfig` pointing to the workload cluster.
   {{% /alert %}}

## Choose a Deployment Approach

Each Cluster Autoscaler instance can target one cluster for autoscaling.

There are three ways to deploy a Cluster Autoscaler instance:

1. [RECOMMENDED] Cluster Autoscaler deployed in the management cluster to autoscale the management cluster itself
1. [RECOMMENDED] Cluster Autoscaler deployed in the management cluster to autoscale a remote workload cluster
1. Cluster Autoscaler deployed in the workload cluster to autoscale the workload cluster itself

To read more about the tradeoffs of these different approaches, see [Autoscaling configuration]({{< relref "../../getting-started/optional/autoscaling/" >}}).

## Install Cluster Autoscaler in management cluster [RECOMMENDED]

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Ensure you have configured at least one WorkerNodeGroup in your cluster to support autoscaling as outlined [Autoscaling configuration]({{< relref "../../getting-started/optional/autoscaling/" >}})

    Cluster autoscaler only works on node groups with an autoscalingConfiguration set:

    **Note**: Here, the `<cluster-name>` value represents the name of the management or workload cluster you would like to autoscale.*

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
    See [Autoscaling configuration]({{< relref "../../getting-started/optional/autoscaling/" >}}) for details.

1. Generate the package configuration
   ```bash
   eksctl anywhere generate package cluster-autoscaler --cluster <cluster-name> > cluster-autoscaler.yaml
   ```

1. Add the desired configuration to `cluster-autoscaler.yaml`

   Please see [complete configuration options]({{< relref "../cluster-autoscaler" >}}) for all configuration options and their default values.

    Example package file configuring a cluster autoscaler package to run in the management cluster.

    *Note: Here, the `<cluster-name>` value represents the name of the management or workload cluster you would like to autoscale.*

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

1. Install Cluster Autoscaler

   ```bash
   eksctl anywhere create packages -f cluster-autoscaler.yaml
   ```

1. Validate the installation

   ```bash
   eksctl anywhere get packages --cluster <cluster-name>
   ```

   Example command output
   ```
   NAMESPACE                  NAME                          PACKAGE              AGE   STATE       CURRENTVERSION                                               TARGETVERSION                                                         DETAIL
   eksa-packages-mgmt-v-vmc   cluster-autoscaler            cluster-autoscaler   18h   installed   9.21.0-1.21-147e2a701f6ab625452fe311d5c94a167270f365         9.21.0-1.21-147e2a701f6ab625452fe311d5c94a167270f365 (latest)
   ```

   To verify that autoscaling works, apply this deployment:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/aws/eks-anywhere/d8575bbd2a85a6c6bbcb1a54868cf7790df56a63/test/framework/testdata/hpa_busybox.yaml
   kubectl scale deployment hpa-busybox-test --replicas 100
   ```
   You must continue scaling pods until the deployment has pods in a pending state.
   This is when cluster autoscaler will begin to autoscale your machine deployment.
   This process may take a few minutes.

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

## Install Cluster Autoscaler in workload cluster

A few extra steps are required to install cluster autoscaler in a workload cluster instead of the management cluster.

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
