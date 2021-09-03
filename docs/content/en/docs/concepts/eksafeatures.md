---
title: "Compare EKS-A and EKS"
linkTitle: "Compare EKS-A and EKS"
weight: 4
date: 2017-01-05
description: >
  Comparing Amazon EKS-A features to Amazon EKS
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

| Feature                       | Amazon EKS Anywhere                  | Amazon EKS                      |
|-------------------------------|--------------------------------------|---------------------------------|
| Control plane management      | Managed by customer                  | Managed by AWS                  |
| Control plane location        | Customer's datacenter                | AWS cloud                       |
| Supported Operating Systems   | Ubuntu: 20.04, BottleRocket          | Linux x86, ARM, and Windows Server operating system distributions |
| Cluster Updates               | CLI (Flux supported rolling update for data plane, Manual update for control plane) | Managed in-place update process for control plane and data plane |
| Networking and Security       | Cilium CNI                           | Amazon VPC CNI, Other compatible 3rd-party CNI plugins |
| Console                       | Connect to EKS console using EKS Connector (Public Preview) | EKS Console |
| GitOps                        | Flux controller                      | Flux controller                 |
| Deployment Types              | EKS-A clusters on VMware vSphere     | Amazon EC2, AWS Fargate         |
| Hardware (Server, Network Equipment, Storage, etc.) | Managed by customer     | Managed by AWS          |
| Load Balancer                 | [3rd-party solutions](https://aws.amazon.com/eks/eks-anywhere/partners/)  | Elastic Load Balancing including Application Load Balancer (ALB), Network Load Balancer (NLB), and Classic Load Balancer |
| Support                       | AWS Support (EKS-A Subscription required) | AWS Support   |
| Pricing                       | Free to download, Annual subscription for AWS Support  | Hourly pricing per cluster |
| Serverless                    | Not supported                         | Amazon EKS on AWS Fargate      |
| Logging and monitoring        | [3rd-party solutions](https://aws.amazon.com/eks/eks-anywhere/partners/)            | CloudWatch, CloudTrail, [3rd-party solutions](https://aws.amazon.com/eks/partners/) |
| Service mesh                  | [3rd-party solutions](https://aws.amazon.com/eks/eks-anywhere/partners/)   | AWS App Mesh, [3rd-party solutions](https://aws.amazon.com/eks/partners/) |
| Infrastructure-as-Code        | [3rd-party solutions](https://aws.amazon.com/eks/eks-anywhere/partners/)   | AWS CloudFormation, [3rd-party solutions](https://aws.amazon.com/eks/partners/) |
| Command Line Interface (CLI)  | eksctl (OSS command line tool)        | eksctl (OSS command line tool) |
