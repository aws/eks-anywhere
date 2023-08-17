---
title: "Autoscaling configuration"
linkTitle: "Autoscaling"
weight: 45
aliases:
    /docs/reference/clusterspec/optional/autoscaling/
description: >
 EKS Anywhere cluster yaml autoscaling specification reference
---

EKS Anywhere supports autoscaling worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/). The Kubernetes Cluster Autoscaler Curated Package is an image and helm chart installed via the [Curated Packages Controller]({{< relref "../../packages/overview" >}})

The helm chart utilizes the Cluster Autoscaler [`clusterapi` mode](https://cluster-api.sigs.k8s.io/tasks/automated-machine-management/autoscaling.html) to scale resources.

Configure an EKS Anywhere worker node group to be picked up by a Cluster Autoscaler deployment by adding `autoscalingConfiguration` block to the `workerNodeGroupConfiguration`.

```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
      name: my-cluster-name
    spec:
      workerNodeGroupConfigurations:
        - name: md-0
          autoscalingConfiguration:
            minCount: 1
            maxCount: 5
          machineGroupRef:
            kind: VSphereMachineConfig
            name: worker-machine-a
        - name: md-1
          autoscalingConfiguration:
            minCount: 1
            maxCount: 3
          machineGroupRef:
            kind: VSphereMachineConfig
            name: worker-machine-b
          count: 1
```

Note that if no `count` is specified for the worker node group it will default to the `autoscalingConfiguration.minCount` value.

EKS Anywhere automatically applies the following annotations to your `MachineDeployment` objects for worker node groups with autoscaling enabled. The Cluster Autoscaler component uses these annotations to identify which node groups to autoscale. If a node group is not autoscaling as expected, check for these annotations on the `MachineDeployment` to troubleshoot.
```
cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size: <minCount>
cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size: <maxCount>
```
