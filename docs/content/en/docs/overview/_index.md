---
title: "Overview"
linkTitle: "Overview"
weight: 10
description: >
  Provides an overview of EKS Anywhere
---

EKS Anywhere uses the `eksctl` executable to create a Kubernetes cluster on Bare Metal or vSphere environments.
You can run cluster create and delete commands from an Ubuntu or Mac administrative machine.

To create a cluster, you need to create a specification file that includes information about your EKS Anywhere cluster.
After preparing the environment, running the `eksctl anywhere create cluster` command from your admin machine deploys the workload cluster to that environment.

For a detailed description of how EKS Anywhere creates clusters, see [Cluster creation workflow]({{< relref "../concepts/clusterworkflow/" >}}).

Next steps:
* [Getting Started]({{< relref "../getting-started/" >}})
