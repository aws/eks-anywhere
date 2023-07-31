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

EKS Anywhere supports autoscaling worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/)'s `clusterapi` cloudProvider.


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

EKS Anywhere will automatically apply the following annotations to your MachineDeployment objects:
```
cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size: <minCount>
cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size: <maxCount>
```

After deploying the Kubernetes Cluster Autoscaler from upstream or as a [curated package]({{< relref "../../packages/cluster-autoscaler/" >}}).

### Cluster Autoscaler Deployment Topologies

The Kubernetes Cluster Autoscaler can only scale a single cluster per deployment.

This means that each cluster you want to scale will need its own cluster autoscaler deployment.

We support three deployment topologies:
1. [RECOMMENDED] Cluster Autoscaler deployed in the management cluster to autoscale the management cluster itself
2. [RECOMMENDED] Cluster Autoscaler deployed in the management cluster to autoscale a remote workload cluster
3. Cluster Autoscaler deployed in the workload cluster to autoscale the workload cluster itself

If your cluster architecture supports management clusters with resources to run additional workloads, you may want to consider using deployment topologies (1) and (2). Instructions for using this topology can be found on the [Cluster Autoscaler]({{< relref "../../packages/cluster-autoscaler/addclauto#install-cluster-autoscaler-in-management-cluster-recommended" >}}) page.

If your deployment topology runs small management clusters though, you may want to follow deployment topology (3) and deploy the cluster autoscaler to run in a [workload cluster]({{< relref "../../packages/cluster-autoscaler/addclauto#install-cluster-autoscaler-in-workload-cluster" >}}).
