---
title: "Autoscaling configuration"
linkTitle: "Autoscaling"
weight: 45
aliases:
    /docs/reference/clusterspec/optional/autoscaling/
description: >
 EKS Anywhere cluster yaml autoscaling configuration specification reference
---

## Cluster Autoscaling (Optional)

### Cluster Autoscaler configuration in EKS Anywhere cluster spec

EKS Anywhere supports autoscaling worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/). The Kubernetes Cluster Autoscaler Curated Package is an image and helm chart installed via the Curated Packages Controller.

The helm chart utilizes the Cluster Autoscaler binary's `clusterapi` mode to scale resources.

- Configure a worker node group to be picked up by a cluster autoscaler deployment by adding a `autoscalingConfiguration` block to the `workerNodeGroupConfiguration`:
    ```yaml
    apiVersion: anywhere.eks.amazonaws.com/v1alpha1
    kind: Cluster
    metadata:
      name: my-cluster-name
    spec:
      workerNodeGroupConfigurations:
        - autoscalingConfiguration:
            minCount: 1
            maxCount: 5
          machineGroupRef:
            kind: VSphereMachineConfig
            name: worker-machine-a
          name: md-0
        - count: 1
          autoscalingConfiguration:
            minCount: 1
            maxCount: 3
          machineGroupRef:
            kind: VSphereMachineConfig
            name: worker-machine-b
          name: md-1
    ```

Note that if no `count` is specified it will default to the `minCount` value.

EKS Anywhere will automatically apply the following annotations to your MachineDeployment objects. The autoscaler component uses these annotations to identify which node groups to autoscale. If a node group is not auto scaling as expected, checking for these annotations on the machine deployment can be a good debugging step:
```
cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size: <minCount>
cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size: <maxCount>
```
