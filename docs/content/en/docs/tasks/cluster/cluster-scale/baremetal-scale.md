---
title: "Scale Bare Metal cluster"
linkTitle: "Scale Bare Metal cluster"
weight: 20
date: 2017-01-05
description: >
  How to scale your Bare Metal cluster
---

### Scaling nodes on Bare Metal clusters
When you are horizontally scaling your bare metal EKS Anywhere cluster, consider the number of nodes you need for your control plane and for your data plane.

See the [Kubernetes Components](https://kubernetes.io/docs/concepts/overview/components/) documentation to learn the differences between the control plane and the data plane (worker nodes).

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
  workerNodeGroupsConfiguration:
  - count: 1     # increase this number to horizontally scale your data plane
```

Next, you must ensure you have enough available hardware for the scale up operation to function. Available hardware could have been fed to the cluster as extra hardware during a prior create command, or could be fed to the cluster during the scale up process by providing the hardware CSV file to the upgrade cluster command (explained in detail below).
For scale down operation, you can skip directly to the upgrade cluster command.

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

If you don't have any available hardware that match this requirement in the cluster, you can [setup a new hardware CSV]({{< relref "../../../reference/baremetal/bare-preparation/#prepare-hardware-inventory" >}}). You can feed this hardware inventory file during the upgrade cluster command as shown below.

#### Upgrade Cluster Command for Scale Up/Down

```bash
eksctl anywhere upgrade cluster -f cluster.yaml --hardware-csv <hardware.csv>
```

or 

Prior to running the upgrade cluster command shown above, you can run the following command to manually push the additional hardware to your cluster:

```bash
eksctl anywhere generate hardware -z <hardware.csv> | kubectl apply -f -
```

### Autoscaling

EKS Anywhere supports autoscaling of worker node groups using the [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/) and as a [curated package](../../../../reference/packagespec/cluster-autoscaler/).

See [here](../../../../reference/clusterspec/optional/autoscaling/) for details on how to configure your cluster spec to autoscale worker node groups for autoscaling.
