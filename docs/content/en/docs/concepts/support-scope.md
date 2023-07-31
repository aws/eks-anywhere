---
title: "Support"
linkTitle: "Support"
weight: 40
aliases:
    /docs/reference/support/
    /docs/reference/support/support-scope/
description: >
  Support scope for EKS Anywhere
---

Enterprise support for Amazon EKS Anywhere is available to Amazon customers who pay for the [Amazon EKS Anywhere Enterprise subscription.](https://aws.amazon.com/eks/eks-anywhere/pricing/)
If you would like to purchase the Amazon EKS Anywhere Enterprise Subscription, contact an [AWS specialist.](https://aws.amazon.com/contact-us/sales-support-eks/) 

EKS Anywhere is an open source project and it is supported by the community.
If you have a problem, open an [issue](https://github.com/aws/eks-anywhere/issues) and someone will get back to you as soon as possible.
If you discover a potential security issue in this project, we ask that you notify AWS/Amazon Security via our [vulnerability reporting page.](http://aws.amazon.com/security/vulnerability-reporting/)
Please do not create a public GitHub issue for security problems.

## Operating system support

EKS Anywhere has some level of support for the following operating system nodes:

* **Bottlerocket**: Bottlerocket is the only fully-supported operating system for EKS Anywhere nodes.
Bottlerocket OVAs and Raw images are distributed by the EKS Anywhere project.
See the [Artifacts]({{< relref "../osmgmt/artifacts" >}}) page for details on how to download Bottlerocket images for EKS Anywhere.

* **Ubuntu**: EKS Anywhere has been tested with Ubuntu-based nodes.
Amazon will assist with troubleshooting and configuration guidance with Ubuntu-based nodes under the [EKS Anywhere Enterprise Subscription.](https://aws.amazon.com/eks/eks-anywhere/pricing/)
To build your own Ubuntu-based EKS Anywhere node image, refer to [Building node images]({{< relref "../osmgmt/artifacts/#building-node-images" >}}).
For official Ubuntu support, see the Canonical [Support](https://ubuntu.com/support) page.

* **Red Hat Enterprise Linux (RHEL)**: EKS Anywhere has been tested with RHEL-based nodes.
As with Ubuntu, Amazon will assist with troubleshooting and configuration guidance with RHEL-based nodes under the [EKS Anywhere Enterprise Subscription.](https://aws.amazon.com/eks/eks-anywhere/pricing/)
To build your own RHEL-based EKS Anywhere node image, refer to [Building node images]({{< relref "../osmgmt/artifacts/#building-node-images" >}}).
For official Red Hat support, see the [Red Hat Enterprise Linux Subscriptions](https://www.redhat.com/en/store/linux-platforms) page.

## Curated packages support
Amazon [EKS Anywhere Curated Packages]({{< relref "../packages/" >}}) are Amazon-curated software packages that extend the core functionalities of Kubernetes on your EKS Anywhere clusters.
All curated packages, including the curated OSS packages, are supported under the EKS Anywhere Enterprise Subscription.
