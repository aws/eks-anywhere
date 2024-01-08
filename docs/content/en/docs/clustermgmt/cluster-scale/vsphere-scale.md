---
title: "Scale vSphere cluster"
linkTitle: "Scale vSphere cluster"
weight: 20
date: 2017-01-05
aliases:
    /docs/tasks/cluster/cluster-scale/vsphere-scale/
description: >
  How to scale your vSphere cluster
---

When you are scaling your vSphere EKS Anywhere cluster, consider the number of nodes you need for your control plane and for your data plane.
Each plane can be scaled horizontally (add more nodes) or vertically (provide nodes with more resources).
In each case you can scale the cluster manually or automatically.

See the [Kubernetes Components](https://kubernetes.io/docs/concepts/overview/components/) documentation to learn the differences between the control plane and the data plane (worker nodes).

### Manual cluster scaling

Horizontally scaling the cluster is done by increasing the number for the control plane or worker node groups under the Cluster specification.

>**_NOTE:_** If etcd is running on your control plane (the default configuration) you should scale your control plane in odd numbers (3, 5, 7...).

```
apiVersion: anywhere.eks.amazonaws.com/v1
kind: Cluster
metadata:
  name: test-cluster
spec:
  controlPlaneConfiguration:
    count: 1     # increase this number to horizontally scale your control plane
...    
  workerNodeGroupConfigurations:
  - count: 1     # increase this number to horizontally scale your data plane
```

Vertically scaling your cluster is done by updating the machine config spec for your infrastructure provider.
For a vSphere cluster an example is

```
apiVersion: anywhere.eks.amazonaws.com/v1
kind: VSphereMachineConfig
metadata:
  name: test-machine
  namespace: default
spec:
  diskGiB: 25       # increase this number to make the VM disk larger
  numCPUs: 2        # increase this number to add vCPUs to your VM
  memoryMiB: 8192   # increase this number to add memory to your VM
```

Once you have made configuration updates, you can use `eksctl`, `kubectl`, GitOps, or Terraform specified in the [upgrade cluster command]({{< relref "../cluster-upgrades/vsphere-and-cloudstack-upgrades/#upgrade-cluster-command" >}}) to apply those changes.
If you are adding or removing a node, only the terminated nodes will be affected.
If you are vertically scaling your nodes, then all nodes will be replaced one at a time.

### Autoscaling

EKS Anywhere supports autoscaling of worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/) and as a [curated package]({{< relref "../../packages/cluster-autoscaler/" >}}).

See [here](https://github.com/kubernetes/autoscaler/) and as a [curated package]({{< relref "../../getting-started/optional/autoscaling/" >}}) for details on how to configure your cluster spec to autoscale worker node groups for autoscaling.

