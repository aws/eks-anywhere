---
title: "Overview"
linkTitle: "Overview"
weight: 10
description: >
  Provides an overview of EKS Anywhere
---

EKS Anywhere creates a Kubernetes cluster on premises to a chosen provider.
Supported providers include Bare Metal (via Tinkerbell), CloudStack, and vSphere.
To manage that cluster, you can run cluster create and delete commands from an Ubuntu or Mac Administrative machine.

Creating a cluster involves downloading EKS Anywhere tools to an Administrative machine, then running the `eksctl anywhere create cluster` command to deploy that cluster to the provider.
A temporary bootstrap cluster runs on the Administrative machine to direct the target cluster creation.
For a detailed description, see [Cluster creation workflow]({{< relref "../getting-started/overview/" >}}).

Here’s a diagram that explains the process visually.

##### EKS Anywhere Create Cluster


![EKS Anywhere create cluster overview](/images/line-create-cluster.svg)

<br/>

Next steps:
* [Getting Started]({{< relref "../getting-started/" >}})
