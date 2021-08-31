---
title: "Overview"
linkTitle: "Overview"
weight: 10
description: >
  Provides an overview of EKS Anywhere
---

EKS Anywhere will use the `eksctl` executable to create a Kubernetes cluster in your environment.
Currently it allows you to create and delete clusters in a vSphere environment.
You can run cluster create and delete commands from an Ubuntu or Mac administrative machine.

To create a cluster you will need to create a specification file which includes all of your vSphere details and information about your EKS cluster.
The `eksctl anywhere create cluster` command will create a temporary bootstrap cluster on your admin machine which will be used to create a workload cluster in vSphere.
Once the workload cluster is created the cluster management resources will be moved to your workload cluster and the local bootstrap cluster will be deleted.

Once your workload cluster is created a KUBECONFIG file will be stored on your admin machine with RBAC admin permissions for the workload cluster.
You’ll be able to use that file with `kubectl` to set up and deploy workloads.
Here’s a diagram that explains the process visually.

![EKS-A create cluster overview](/images/eks-a_create_cluster.png)

The delete process is similar to the create process.

![EKS-A delete cluster overview](/images/eks-a_delete_cluster.png)


Next steps:
* [Getting Started](/docs/getting-started/)
* [Examples](/docs/examples/)

