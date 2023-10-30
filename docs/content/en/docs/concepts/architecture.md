---
title: "Architecture"
linkTitle: "Architecture"
aliases:
    /docs/concepts/cluster-topologies
weight: 10
description: >
  Explanation of standalone vs. management/workload cluster topologies
---

For trying out EKS Anywhere or for times when a single cluster is needed, it is fine to create a _standalone cluster_ and run your workloads on it.
However, if you plan to create multiple clusters for running Kubernetes workloads, we recommend you create a _management cluster_.
Then use that management cluster to manage a set of workload clusters.

This document describes those two different EKS Anywhere cluster topologies.

## What is an EKS Anywhere management cluster?
An EKS Anywhere management cluster is a long-lived, on-premises Kubernetes cluster that can create and manage a fleet of EKS Anywhere workload clusters.
The workload clusters are where you run your applications.
The management cluster can only be created and managed by the Amazon CLI `eksctl`.

The management cluster runs on your on-premises hardware and it does not require any connectivity back to AWS to function.
Customers are responsible for operating the management cluster including (but not limited to) patching, upgrading, scaling, and monitoring the cluster control plane and data plane.
 
## What’s the difference between a management cluster and a standalone cluster?
From a technical point of view, they are the same.
Regardless of which deployment topology you choose, you always start by creating a singleton, standalone cluster that’s capable of managing itself.
This shows examples of separate, standalone clusters:

![Standalone clusters self-manage and can run applications](/images/eks-a_cluster_standalone.png)

Once a standalone cluster is created, you have an option to use it as a management cluster to create separate workload cluster(s) under it, hence making this cluster a long-lived management cluster.
You can only use `eksctl` to create or delete the management cluster or a standalone cluster.
This shows examples of a management cluster that deploys and manages multiple workload clusters:

![Management clusters can create and manage multiple workload clusters](/images/eks-a_cluster_management.png)

With the management cluster in place, you have a choice of tools for creating, upgrading, and deleting workload clusters.
Check each provider to see which tools it currently supports.
Supported workload cluster creation, upgrade and deletion tools include:

* `eksctl` CLI
* Terraform
* GitOps
* `kubectl` CLI to communicate with the Kubernetes API

## What’s the difference between a management cluster and a bootstrap cluster for EKS Anywhere?

A management cluster is a long-lived entity you have to actively operate.
The _bootstrap_ cluster is a temporary, short-lived kind cluster that is created on a separate [Administrative machine]({{< relref "../getting-started/install" >}}) to facilitate the creation of an initial standalone or management cluster.

The `kind` cluster is automatically deleted by the end of the initial cluster creation.

## When should I deploy a management cluster?
If you want to run three or more EKS Anywhere clusters, we recommend that you choose a management/workload cluster deployment topology because of the advantages listed in the table below.
The EKS Anywhere Curated Packages feature recommends deploying certain packages such as the container registry package or monitoring packages on the management cluster to avoid circular dependency. 


|        | Standalone cluster topology | Management/workload cluster topology  |
|--------|-----------------------------|---------------------------------------|
| **Pros**   | Save hardware resources   | Isolation of secrets                |
|        | Reduced operational overhead of maintaining a separate management cluster | Resource isolation between different teams. Reduced noisy-neighbor effect. |
|        |                             |  Isolation between development and production workloads. |
|        |                             |  Isolation between applications and fleet management services, such as monitoring server or container registry. |
|        |                             |  Provides a central control plane and API to automate cluster lifecycles |
| **Cons** |  Shared secrets such, as SSH credentials or VMware credentials, across all teams who share the cluster. |  Consumes extra resources. |
|        |  Without a central control plane (such as a parent management cluster), it is not possible to automate cluster creation/deletion with advanced methods like GitOps or IaC. |The creation/deletion of the management cluster itself can’t be automated through GitOps or IaC. |
|        | Circular dependencies arise if the cluster has to host a monitoring server or a local container registry. | 
||||


## Which EKS Anywhere features support the management/workload cluster deployment topology today?

| Features   | Supported | 
|------------|-----------|
| Create/upgrade/delete a workload cluster on... ||
| <ul><li>VMware via CLI</li>  | Y |
| <ul><li>CloudStack via CLI</li> | Y |
| <ul><li>Bare Metal via CLI</li> | Y |
| <ul><li>Snow via CLI</li> | Y |
| <ul><li>Nutanix via CLI</li> | Y |
| <ul><li>Docker via CLI (non-production only)</li> | Y |
| Create/upgrade/delete a workload cluster on...
| <ul><li>VMware via GitOps/Terraform</li> | Y |
| <ul><li>CloudStack via GitOps/Terraform</li> | Y |
| <ul><li>Bare Metal via GitOps/Terraform</li> | Y |
| <ul><li>Snow via GitOps/Terraform</li> | Y |
| <ul><li>Nutanix via GitOps/Terraform</li> | Y |
| <ul><li>Docker via GitOps/Terraform (non-production only)</li> | Y |
| Install a curated package on the management cluster | Y ||
| Install a curated package on the workload cluster from the management cluster | Y |
