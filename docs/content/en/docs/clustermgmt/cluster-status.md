---
title: "Cluster status"
linkTitle: "Cluster status"
weight: 90
aliases:
    /docs/tasks/cluster/cluster-status/
date: 2017-01-05
description: >
  What's in an EKS Anywhere cluster status?
---

The EKS Anywhere cluster status shows information representing the actual state of the cluster vs the desired cluster specification. This is particularly useful to track the progress of cluster management operations through the cluster lifecyle feature.

### Viewing an EKS Anywhere cluster status

First set the `CLUSTER_NAME` and `KUBECONFIG` environment variables.

```
export CLUSTER_NAME=w01
export KUBECONFIG=${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
```

To view an EKS Anywhere cluster status, use `kubectl` command show the status of the cluster object.
The output should show the `yaml` definition of the EKS Anywhere `Cluster` object which has a `status` field.

```
kubectl get clusters $CLUSTER_NAME -o yaml
...
status:
  conditions:
  - lastTransitionTime: "2023-08-15T20:35:15Z"
    status: "True"
    type: Ready
  - lastTransitionTime: "2023-08-15T20:35:15Z"
    status: "True"
    type: ControlPlaneInitialized
  - lastTransitionTime: "2023-08-15T20:35:15Z"
    status: "True"
    type: ControlPlaneReady
  - lastTransitionTime: "2023-08-15T20:35:15Z"
    status: "True"
    type: DefaultCNIConfigured
  - lastTransitionTime: "2023-08-15T20:35:15Z"
    status: "True"
    type: WorkersReady
  observedGeneration: 2
```

### Cluster status attributes

The following fields may be represented on the cluster status:

**`status.failureMessage`**

Descriptive message about a fatal problem while reconciling a cluster

**`status.failureReason`**

Machine readable value about a terminal problem while reconciling the cluster set at the same time as the `status.failureMessage`.

**`status.conditions`**

Provides a collection of condition objects that report a high-level assessment of cluster readiness.

**`status.observedGeneration`**

This is the latest generation observed, set everytime the status is updated. If this is not the same as the cluster object's `metadata.generation`, it means that the status being viewed represents an old generation of the cluster specification and is not up-to-date yet. 

### Cluster status conditions

Conditions provide a high-level status report representing an assessment of cluster readiness using a collection of conditions each of a particular type. The following condition types are supported:

  * `ControlPlaneInitialized` - reports the first control plane has been initialized and the cluster is contactable with the kubeconfig. Once this condition is marked `True`, its value never changes.

  * `ControlPlaneReady` -  reports that the condition of the current state of the control plane with respect to the desired state specified in the Cluster specification. This condition is marked `True` once the number of control plane nodes in the cluster match the expected number of control plane nodes as defined in the cluster specifications and all the control plane nodes are up to date and ready.

  * `DefaultCNIConfigured` - reports the configuration state of the default CNI specified in the cluster specifications. It will be marked as `True` once the default CNI has been successfully configured on the cluster. 
  However, if the EKS Anywhere default cilium CNI has been [configured to skip upgrades]({{< relref "../getting-started/optional/cni/#use-a-custom-cni" >}}) in the cluster specification, then this condition will be marked as `False` with the reason `SkipUpgradesForDefaultCNIConfigured`.

  * `WorkersReady` - reports that the condition of the current state of worker machines versus the desired state specified in the Cluster specification. This condition is marked `True` once the number of worker nodes in the cluster match the expected number of worker nodes as defined in the cluster specifications and all the worker nodes are up to date and ready.

  * `Ready` - reports a summary of the following conditions: `ControlPlaneInitialized`, `ControlPlaneReady`, and `WorkersReady`. It indicates an overall operational state of the EKS Anywhere cluster. It will be marked `True` once the current state of the cluster has fully reached the desired state specified in the Cluster spec.

