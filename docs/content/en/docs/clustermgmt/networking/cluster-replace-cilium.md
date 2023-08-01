---
title: "Replace EKS Anywhere Cilium with a custom CNI"
linkTitle: "Add custom CNI"
weight: 40
description: >
 Replace EKS Anywhere Cilium with a custom CNI
---

This page provides walkthroughs on replacing the EKS Anywhere Cilium with a custom CNI. 
For more information on CNI customization see [Use a custom CNI]({{< ref "../../getting-started/optional/cni#use-a-custom-cni"  >}}).

{{% alert title="Note" color="primary" %}}
When replacing EKS Anywhere Cilium with a custom CNI, it is your responsibility to manage the custom CNI, including version upgrades and support.
{{% /alert %}}

## Prerequisites

* EKS Anywhere v0.15+.
* [Cilium CLI](https://github.com/cilium/cilium-cli) v0.14.

## Add a custom CNI to a new cluster

If an operator intends to uninstall EKS Anywhere Cilium from a new cluster they can enable the `skipUpgrade` option when creating the cluster. 
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

3. Create the cluster according to the [Getting Started]({{< ref "/docs/getting-started" >}}) guide.

4. Pause reconciliation of the cluster. This ensures EKS Anywhere components do not attempt to remediate issues arising from a missing CNI.

  ```bash
  kubectl --kubeconfig=MANAGEMENT_KUBECONFIG -n eksa-system annotate clusters.cluster.x-k8s.io WORKLOAD_CLUSTER_NAME cluster.x-k8s.io/paused=true
  ```

5. Uninstall EKS Anywhere Cilium.

    ```bash
    cilium uninstall
    ```

6. Install a custom CNI.

7. Resume reconciliation of the cluster object.

  ```bash
  kubectl --kubeconfig=MANAGEMENT_KUBECONFIG -n eksa-system annotate clusters.cluster.x-k8s.io WORKLOAD_CLUSTER_NAME cluster.x-k8s.io/paused-
  ```

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

2. [Upgrade the EKS Anywhere cluster]({{< ref "../cluster-upgrades" >}}).

3. Pause reconciliation of the cluster. This ensures EKS Anywhere components do not attempt to remediate issues arising from a missing CNI.

  ```bash
  kubectl --kubeconfig=MANAGEMENT_KUBECONFIG -n eksa-system annotate clusters.cluster.x-k8s.io WORKLOAD_CLUSTER_NAME cluster.x-k8s.io/paused=true
  ```

4. Uninstall EKS Anywhere Cilium.

    ```bash
    cilium uninstall
    ```

5. Install a custom CNI.

6. Resume reconciliation of the cluster object.

  ```bash
  kubectl --kubeconfig=MANAGEMENT_KUBECONFIG -n eksa-system annotate clusters.cluster.x-k8s.io WORKLOAD_CLUSTER_NAME cluster.x-k8s.io/paused-
  ```

## Add a custom CNI to an existing cluster with Lifecycle Controller

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

2. Apply the cluster configuration to the cluster and _await successful object reconciliation_.

    ```bash
    kubectl apply -f <cluster config path>
    ```

3. Pause reconciliation of the cluster. This ensures EKS Anywhere components do not attempt to remediate issues arising from a missing CNI.

  ```bash
  kubectl --kubeconfig=MANAGEMENT_KUBECONFIG -n eksa-system annotate clusters.cluster.x-k8s.io WORKLOAD_CLUSTER_NAME cluster.x-k8s.io/paused=true
  ```

4. Uninstall EKS Anywhere Cilium.

  ```bash
  cilium uninstall
  ```

5. Install a custom CNI.

6. Resume reconciliation of the cluster object.

  ```bash
  kubectl --kubeconfig=MANAGEMENT_KUBECONFIG -n eksa-system annotate clusters.cluster.x-k8s.io WORKLOAD_CLUSTER_NAME cluster.x-k8s.io/paused-
  ```
