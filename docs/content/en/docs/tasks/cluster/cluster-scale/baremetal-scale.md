---
title: "Scale Bare Metal cluster"
linkTitle: "Scale Bare Metal cluster"
weight: 20
date: 2017-01-05
description: >
  How to scale your Bare Metal cluster
---

### Scaling nodes on Bare Metal clusters
Before you can scale up nodes on a Bare Metal cluster, you must ensure you have enough available hardware for the scale up operation to function.
For scale down operation, you can skip directly to the scale commands.
To check if you have enough available hardware for scale up, you can use the `kubectl` command below to check if there are hardware with the selector labels corresponding to the controlplane/worker node group and without the `ownerName` label. 

```bash
kubectl get hardware -n eksa-system --show-labels
```

For example, if you want to scale a worker node group with selector label `type=worker-group-1`, then you must have an additional hardware object in your cluster with the label `type=worker-group-1` that doesn't have the `ownerName` label. 

In the command shown below, `eksa-worker2` matches the selector label and it doesn't have the `ownerName` label. Thus, it can be used to scale up `worker-group-1` by 1.

```bash
kubectl get hardware -n eksa-system --show-labels 
NAME                STATE       LABELS
eksa-controlplane               type=controlplane,v1alpha1.tinkerbell.org/ownerName=abhnvp-control-plane-template-1656427179688-9rm5f,v1alpha1.tinkerbell.org/ownerNamespace=eksa-system
eksa-worker1                    type=worker-group-1,v1alpha1.tinkerbell.org/ownerName=abhnvp-md-0-1656427179689-9fqnx,v1alpha1.tinkerbell.org/ownerNamespace=eksa-system
eksa-worker2                    type=worker-group-1
```

If you don't have any available hardware that match this requirement in the cluster, you can [setup a new hardware CSV]({{< relref "../../../reference/baremetal/bare-preparation/#prepare-hardware-inventory" >}}) and then run the following command to push the additional hardware to your cluster

```bash
eksctl anywhere generate hardware -z <hardware.csv> | kubectl apply -f -
```

Once you verify you have the additional hardware available, you are ready to scale your cluster.

To scale a worker node group:

```bash
kubectl scale machinedeployments -n eksa-system <workerNodeGroupName> --replicas <num replicas>
```

To scale control plane nodes:

```bash
kubectl scale kubeadmcontrolplane -n eksa-system <controlPlaneName> --replicas <num replicas>
```

### Autoscaling

EKS Anywhere supports autoscaling of worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/) and as a [curated package](../../../../reference/packagespec/cluster-autoscaler/).

See [here](../../../../reference/clusterspec/optional/autoscaling/) for details on how to configure your cluster spec to autoscale worker node groups for autoscaling.
