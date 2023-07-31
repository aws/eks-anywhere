---
title: "Overview"
linkTitle: "Overview"
weight: 10
aliases:
    /docs/tasks/cluster-overview/
date: 2017-01-05
description: >
  Overview of tools and interfaces for managing EKS Anywhere clusters
---

The content in this page will describe the tools and interfaces available to an administrator after an EKS Anywhere cluster is up and running.
It will also describe which administrative actions done:

* Directly in Kubernetes itself (such as adding nodes with `kubectl`)
* Through the EKS Anywhere API (such as deleting a cluster with `eksctl`).
* Through tools which interface with the Kubernetes API (such as [managing a cluster with `terraform`]({{< relref "./cluster-terraform" >}}))

Note that direct changes to OVAs before nodes are deployed is not yet supported.
However, we are working on a solution for that issue.
