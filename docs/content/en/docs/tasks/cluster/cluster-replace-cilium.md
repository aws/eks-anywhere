# Replace EKS Anywhere Cilium with a custom CNI

## Prerequisites

* EKS Anywhere v0.15+.
* A recent version of [Cilium CLI](https://github.com/cilium/cilium-cli).

## Creating new clusters

If an operator intends to uninstall EKS Anywhere Cilium from a new cluster they can enable the `skipUpgrade`
option when creating the cluster. 
Any future upgrades to the newly created cluster will not have EKS Anywhere Cilium upgraded.

1. Generate a cluster configuration according to the [Getting Started]({{< ref "/docs/getting-started" >}}) section.

2. Modify the `Cluster` object's `spec.clusterNetwork.cniConfig.cilium.skipUpgrade` field to equal `true`.

    ```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
        name: eks-anywhere
    spec:
      clusterNetwork:
        cniConfig:
          cilium:
            skipUpgrade: true
      ...
    ```

3. Create the cluster according to the [Getting Started]({{< ref "/docs/getting-started" >}}) section.

4. Uninstall EKS Anywhere Cilium:

    ```bash
    cilium uninstall
    ```

5. Install the custom CNI.

## Add a custom CNI to an existing cluster with eksctl

1. Modify the existing `Cluster` object's `spec.clusterNetwork.cniConfig.cilium.skipUpgrade` field to equal `true`.

    ```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
        name: eks-anywhere
    spec:
      clusterNetwork:
        cniConfig:
          cilium:
            skipUpgrade: true
      ...
    ```

3. [Upgrade the EKS Anywhere cluster]({{< ref "/docs/tasks/cluster/cluster-upgrades" >}}).

4. Uninstall EKS Anywhere Cilium:

    ```bash
    cilium uninstall
    ```

5. Install the custom CNI.

## Modifying existing clusters using EKS Anywhere Lifecycle Controller

{{% alert title="Warning" color="warning" %}}
Clusters created using the Full Lifecycle Controller prior to v0.15 that have removed the EKS Anywhere Cilium CNI must manually populate their `cluster.anywhere.eks.amazonaws.com` object with the following annotation to ensure EKS Anywhere does not attempt to re-install EKS Anywhere Cilium.

```
anywhere.eks.amazonaws.com/eksa-cilium: ""
```
{{% /alert %}}

1. Modify the existing `Cluster` object's `spec.clusterNetwork.cniConfig.cilium.skipUpgrade` field to equal `true`.

    ```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
        name: eks-anywhere
    spec:
      clusterNetwork:
        cniConfig:
          cilium:
            skipUpgrade: true
      ...
    ```

3. Apply the cluster configuration to the cluster.

    ```bash
    kubectl apply -f <cluster config path>
    ```

5. Await the cluster to reconcile.

4. Uninstall EKS Anywhere Cilium:

    ```bash
    cilium uninstall
    ```

5. Install the custom CNI.