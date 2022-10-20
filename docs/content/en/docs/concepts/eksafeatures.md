---
title: "Compare EKS Anywhere and EKS"
linkTitle: "Compare EKS Anywhere"
weight: 4
date: 2017-01-05
description: >
  Comparing Amazon EKS Anywhere features to Amazon EKS
---

Amazon EKS Anywhere is a new deployment option for Amazon EKS
that enables you to easily create and operate Kubernetes clusters on-premises.
EKS Anywhere provides an installable software package for creating and operating Kubernetes clusters on-premises
and automation tooling for cluster lifecycle support.
To learn more, see [EKS Anywhere](https://aws.amazon.com/eks/eks-anywhere/).


Amazon Elastic Kubernetes Service (Amazon EKS) is a managed Kubernetes service that makes it easy for you to run Kubernetes on the AWS cloud.
Amazon EKS is certified Kubernetes conformant, so existing applications that run on upstream Kubernetes are compatible with Amazon EKS.
To learn more about Amazon EKS, see [Amazon Elastic Kubernetes Service](https://aws.amazon.com/eks/).


### Comparing Amazon EKS Anywhere to Amazon EKS

| Feature                 | Amazon EKS Anywhere | Amazon EKS                      |
|-------------------------|---------------------|---------------------------------|
| **Control plane** ||||
| K8s control plane management      | Managed by customer                  | Managed by AWS                  |
| K8s control plane location        | Customer's datacenter                 | AWS cloud                       |
| Cluster updates        | Manual CLI updates for control plane. Flux supported rolling updates for data plane | Managed in-place updates for control plane and managed rolling updates for data plane.                       |
||||
| **Compute** |||
| Compute options | CloudStack, VMware vSphere, Bare Metal servers | Amazon EC2, AWS Fargate | 
| Supported node operating systems   | Bottlerocket, Ubuntu         | Amazon Linux 2, Windows Server, Bottlerocket, Ubuntu |
| Physical hardware (servers, network equipment, storage, etc.) | Managed by customer | Managed by AWS |
| Serverless | Not supported | Amazon EKS on AWS Fargate |
||||
| **Management** | | |
| Command line interface (CLI)  | `eksctl` (OSS command line tool)        | `eksctl` (OSS command line tool) |
| Console view for Kubernetes objects | Optional EKS console connection using EKS Connector (public preview) | Native EKS console connection|
| Infrastructure-as-code        | Cluster manifest, Kubernetes controllers, [3rd-party solutions](https://aws.amazon.com/eks/eks-anywhere/partners/)            | AWS CloudFormation, [3rd-party solutions](https://aws.amazon.com/eks/partners/) |
| Logging and monitoring        | [3rd-party solutions](https://aws.amazon.com/eks/eks-anywhere/partners/)            | CloudWatch, CloudTrail, [3rd-party solutions](https://aws.amazon.com/eks/partners/) |
| GitOps                        | Flux controller | Flux controller                 |
||||
| **Functions and tooling** | | |
| Networking and Security       | Cilium CNI and network policy supported | Amazon VPC CNI supported. Calico supported for network policy. Other compatible [3rd-party CNI plugins](https://docs.aws.amazon.com/eks/latest/userguide/alternate-cni-plugins.html) available.|
| Load balancer                 | Metallb | Elastic Load Balancing including Application Load Balancer (ALB), and Network Load Balancer (NLB) |
| Service mesh                  | Community or [3rd-party solutions](https://aws.amazon.com/eks/eks-anywhere/partners/)    | AWS App Mesh, community, or [3rd-party solutions](https://aws.amazon.com/eks/partners/) |
| Community tools and Helm      | Works with compatible community tooling and helm charts.  | Works with compatible community tooling and helm charts. |
||||
| **Pricing and support** |||
| Control plane pricing                       | Free to download, paid support subscription option  | Hourly pricing per cluster |
| AWS Support                       | Additional annual subscription (per cluster) for AWS support | Basic support included. Included in paid AWS support plans (developer, business, and enterprise)  |
||||
