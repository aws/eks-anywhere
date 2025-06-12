---
title: "Amazon EKS Anywhere"
linkTitle: Documentation
noedit: true
weight: 10
aliases:
    /docs/workshops/
    /docs/workshops/introduction/
    /docs/workshops/introduction/benefits-and-usecases/
    /docs/workshops/introduction/faq/
    /docs/workshops/introduction/overview/
    /docs/workshops/packages/
    /docs/workshops/packages/adot/
    /docs/workshops/packages/harbor/
    /docs/workshops/packages/prometheus/
    /docs/workshops/packages/credential-provider-package/
    /docs/workshops/provision/
    /docs/workshops/provision/local_cluster/
    /docs/workshops/provision/overview/
    /docs/workshops/provision/prerequisites/
    /docs/workshops/provision/vSphere-prepeartion/
    /docs/workshops/provision/vsphere/
menu:
  main:
    weight: 20
description: >
  EKS Anywhere documentation homepage
---

EKS Anywhere is container management software built by AWS that makes it easier to run and manage Kubernetes clusters on-premises and at the edge. EKS Anywhere is built on [EKS Distro](https://distro.eks.amazonaws.com/), which is the same reliable and secure Kubernetes distribution used by [Amazon Elastic Kubernetes Service (EKS)](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html) in AWS Cloud. EKS Anywhere simplifies Kubernetes cluster management through the automation of undifferentiated heavy lifting such as infrastructure setup and Kubernetes cluster lifecycle operations.

Unlike Amazon EKS in AWS Cloud, EKS Anywhere is a user-managed product that runs on user-managed infrastructure. You are responsible for cluster lifecycle operations and maintenance of your EKS Anywhere clusters. EKS Anywhere does not have any strict dependencies on AWS regional services and is a fit for isolated or air-gapped environments.

If you have on-premises or edge environments with reliable connectivity to an AWS Region, consider using [EKS Hybrid Nodes](https://docs.aws.amazon.com/eks/latest/userguide/hybrid-nodes-overview.html) or [EKS on Outposts](https://docs.aws.amazon.com/eks/latest/userguide/eks-outposts.html) to benefit from AWS-managed EKS control planes and a consistent experience with EKS in the AWS Cloud.

The tenets of the EKS Anywhere project are:

- **Simple**: Make using a Kubernetes distribution simple and boring (reliable and secure).
- **Opinionated Modularity**: Provide opinionated defaults about the best components to include with Kubernetes, but give customers the ability to swap them out
- **Open**: Provide open source tooling backed, validated and maintained by Amazon
- **Ubiquitous**: Enable customers and partners to integrate a Kubernetes distribution in the most common tooling.
- **Stand Alone**: Provided for use anywhere without AWS dependencies
- **Better with AWS**: Enable AWS customers to easily adopt additional AWS services
