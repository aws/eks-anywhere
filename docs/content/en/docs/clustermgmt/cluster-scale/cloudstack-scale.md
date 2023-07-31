---
title: "Scale CloudStack cluster"
linkTitle: "Scale CloudStack cluster"
weight: 20
date: 2017-01-05
aliases:
    /docs/tasks/cluster/cluster-scale/cloudstack-scale/
description: >
  How to scale your CloudStack cluster
---

When you are scaling your CloudStack EKS Anywhere cluster, consider the number of nodes you need for your control plane and for your data plane.
In each case you can scale the cluster manually, semi-automatically, or automatically.

See the [Kubernetes Components](https://kubernetes.io/docs/concepts/overview/components/) documentation to learn the differences between the control plane and the data plane (worker nodes).

### Manual cluster scaling

>**_NOTE:_** If etcd is running on your control plane (the default configuration) you should scale your control plane in odd numbers (3, 5, 7, and so on).

```
apiVersion: anywhere.eks.amazonaws.com/v1
kind: Cluster
metadata:
  name: test-cluster
spec:
  controlPlaneConfiguration:
    count: 1     # increase this number to horizontally scale your control plane
...    
  workerNodeGroupsConfiguration:
  - count: 1     # increase this number to horizontally scale your data plane
```

Once you have made configuration updates you can apply the changes to your cluster.
If you are adding or removing a node, only the terminated nodes will be affected.

```bash
eksctl anywhere upgrade cluster -f cluster.yaml
```

### Semi-automatic scaling

Scaling your cluster in a semi-automatic way still requires changing your cluster manifest configuration.
In a semi-automatic mode you change your cluster spec and then have automation make the cluster changes.

You can do this by storing your cluster config manifest in git and then having a CI/CD system deploy your changes.
Or you can use a GitOps controller to apply the changes.
To read more about making changes with the integrated Flux GitOps controller you can read how to [Manage a cluster with GitOps]({{< relref "../cluster-flux" >}}).

### Autoscaling

EKS Anywhere supports autoscaling of worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/) and as a [curated package]({{< relref "../../packages/cluster-autoscaler/" >}}).

See [here]({{< relref "../../getting-started/optional/autoscaling/" >}}) for details on how to configure your cluster spec to autoscale worker node groups for autoscaling.
