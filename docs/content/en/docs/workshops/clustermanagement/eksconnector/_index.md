---
title: "EKS Connector"
linkTitle: "EKS Connector"
weight: 5
date: 2021-11-11
description: >  
---

## Overview

The Amazon EKS Connector allows you to register and connect any conformant Kubernetes cluster to AWS and visualize it in the Amazon EKS console. Once connected, you can see your cluster's status, configuration, and workloads in the Amazon EKS console. Amazon EKS displays connected clusters in Amazon EKS console for workload visualization only and does not manage them. The Amazon EKS Connector connects the following types of Kubernetes clusters to Amazon EKS.

* On-premise Kubernetes clusters
* Self-managed clusters running on Amazon EC2
* Managed clusters from other cloud providers

## Amazon EKS Connector considerations

Consider the following when using Amazon EKS Connector:

* You must have administrative privileges to the Kubernetes cluster prior to registering the cluster to Amazon EKS.
* The Amazon EKS Connector must run on Linux 64-bit (x86) worker nodes. ARM worker nodes are not supported.
* You must have worker nodes in your Kubernetes cluster that have outbound access to the ssm. and ssmmessages. Systems Manager endpoints. For more information, see [Systems Manager](https://docs.aws.amazon.com/general/latest/gr/ssm.html) endpoints in the AWS General Reference.
* You can connect up to 10 clusters per Region.
* Only the Amazon EKS `RegisterCluster`, `ListClusters`, `DescribeCluster`, and `DeregisterCluster` APIs are supported for external Kubernetes clusters.
* Tags are not supported for connected clusters.