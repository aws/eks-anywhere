---
title: "Compare EKS Anywhere and Amazon EKS"
linkTitle: "Compare EKS Anywhere"
weight: 60
date: 2017-01-05
description: >
  Comparing EKS Anywhere features to Amazon EKS
---

EKS Anywhere provides an installable software package for creating and operating Kubernetes clusters on-premises and automation tooling for cluster lifecycle operations. EKS Anywhere is a fit for isolated and air-gapped environments, and for users who prefer to manage their own Kubernetes clusters on-premises.

Amazon Elastic Kubernetes Service (Amazon EKS) is a managed Kubernetes service where the Kubernetes control plane is managed by AWS, and users can choose to run applications on EKS Auto Mode, AWS Fargate, EC2 instances in AWS Regions, AWS Local Zones, or AWS Outposts, or on customer-managed infrastructure with EKS Hybrid Nodes. If you have on-premises or edge environments with reliable connectivity to an AWS Region, consider using [EKS Hybrid Nodes](https://docs.aws.amazon.com/eks/latest/userguide/hybrid-nodes-overview.html) or [EKS on Outposts](https://docs.aws.amazon.com/eks/latest/userguide/eks-outposts.html) to benefit from the AWS-managed EKS control plane and consistent experience with EKS in AWS Cloud.

EKS Anywhere and Amazon EKS are certified Kubernetes conformant, so existing applications that run on upstream Kubernetes are compatible with EKS Anywhere and Amazon EKS.

### Comparing EKS Anywhere to EKS on Outposts

Like EKS Anywhere, EKS on Outposts enables customers to run Kubernetes workloads on-premises and at the edge.

The main differences are:

* EKS on Outposts requires a reliable connection to AWS Regions. EKS Anywhere can run in isolated or air-gapped environments.
* With EKS on Outposts, AWS provides AWS-managed infrastructure and AWS services such as Amazon EC2 for compute, Amazon VPC for networking, and Amazon EBS for storage. EKS Anywhere runs on customer-managed infrastructure and interfaces with customer-provided compute (vSphere, bare metal, Nutanix, etc.), networking, and storage.
* With EKS on Outposts, the Kubernetes control plane is managed by AWS. With EKS Anywhere, customers are responsible for managing the Kubernetes cluster lifecycle with EKS Anywhere automation tooling.
* Customers can use EKS on Outposts with the same console, APIs, and tools they use to run EKS clusters in AWS Cloud. With EKS Anywhere, customers can use the eksctl CLI or Kubernetes API-compatible tooling to manage their clusters.

For more information, see the [EKS on Outposts documentation.](https://docs.aws.amazon.com/eks/latest/userguide/eks-outposts.html)

### Comparing EKS Anywhere to EKS Hybrid Nodes

Like EKS Anywhere, EKS Hybrid Nodes enables customers to run Kubernetes workloads on-premises and at the edge. Both EKS Anywhere and EKS Hybrid Nodes run on customer-managed infrastructure (VMware vSphere, bare metal, Nutanix, etc.).

The main differences are:

* EKS Hybrid Nodes requires a reliable connection to AWS Regions. EKS Anywhere can run in isolated or air-gapped environments.
* With EKS Hybrid Nodes, the Kubernetes control plane is managed by AWS. With EKS Anywhere, customers are responsible for managing the Kubernetes cluster lifecycle with EKS Anywhere automation tooling.
* Customers can use EKS Hybrid Nodes with the same console, APIs, and tools they use to run EKS clusters in AWS Cloud. With EKS Anywhere, customers can use the eksctl CLI or Kubernetes API-compatible tooling to manage their clusters.

For more information, see the [EKS Hybrid Nodes documentation.](https://docs.aws.amazon.com/eks/latest/userguide/hybrid-nodes-overview.html)