---
title: "Overview"
linkTitle: "Overview"
weight: 10
description: >
  Provides an overview of EKS Anywhere
---

EKS Anywhere uses the `eksctl` executable to create a Kubernetes cluster in your environment.
Currently it allows you to create and delete clusters in a vSphere environment.
You can run cluster create and delete commands from an Ubuntu or Mac administrative machine.

To create a cluster, you need to create a specification file that includes all of your vSphere details and information about your EKS cluster.
Running the `eksctl anywhere create cluster` command from your admin machine creates the workload cluster in vSphere.
It does this by first creating a temporary bootstrap cluster to direct the workload cluster creation.
Once the workload cluster is created, the cluster management resources are moved to your workload cluster and the local bootstrap cluster is deleted.

Once your workload cluster is created, a KUBECONFIG file is stored on your admin machine with RBAC admin permissions for the workload cluster.
You’ll be able to use that file with `kubectl` to set up and deploy workloads.
Here’s a diagram that explains the process visually.

![EKS-A create cluster overview](/images/eks-a_create_cluster.png)

The delete process is similar to the create process.

![EKS-A delete cluster overview](/images/eks-a_delete_cluster.png)


Next steps:
* [Getting Started](/docs/getting-started/)
