---
title: "Scale Bare Metal cluster"
linkTitle: "Scale Bare Metal cluster"
weight: 20
date: 2017-01-05
description: >
  How to scale your Bare Metal cluster
---

Before you can scale up nodes on a Bare Metal cluster, you must have already defined the machines in a [hardware CSV file]({{< relref "../../../reference/baremetal/bare-preparation#prepare-hardware-inventory" >}}) and supplied it to Tinkerbell.

Then, to scale a worker node group:

```bash
kubectl scale machinedeployments -n eksa-system <workerNodeGroupName> --replicas <num replicas>
```
To scale control plane nodes:

```bash
kubectl scale kubeadmcontrolplane -n eksa-system <controlPlaneName> --replicas <num replicas>
```
