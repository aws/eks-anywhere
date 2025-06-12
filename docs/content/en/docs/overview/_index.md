---
title: "Overview"
linkTitle: "Overview"
weight: 10
---

### What is EKS Anywhere?
EKS Anywhere is container management software built by AWS that makes it easier to run and manage Kubernetes clusters on-premises and at the edge. EKS Anywhere is built on [EKS Distro](https://distro.eks.amazonaws.com/), which is the same reliable and secure Kubernetes distribution used by [Amazon Elastic Kubernetes Service (EKS)](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html) in AWS Cloud. EKS Anywhere simplifies Kubernetes cluster management through the automation of undifferentiated heavy lifting such as infrastructure setup and Kubernetes cluster lifecycle operations.

Unlike Amazon EKS in AWS Cloud, EKS Anywhere is a user-managed product that runs on user-managed infrastructure. You are responsible for cluster lifecycle operations and maintenance of your EKS Anywhere clusters. EKS Anywhere is open source and free to use at no cost. You can optionally purchase [EKS Anywhere Enterprise Subscriptions]({{< relref "../concepts/support-scope/">}}) to receive support for your EKS Anywhere clusters and for access to [EKS Anywhere Curated Packages]({{< relref "../concepts/packages/">}}) and [extended support for Kubernetes versions.]({{< relref "../concepts/support-versions">}}) EKS Anywhere Curated Packages are software packages that are built, tested, and supported by AWS and extend the core functionalities of Kubernetes on your EKS Anywhere clusters.

EKS Anywhere supports many different types of infrastructure including VMWare vSphere, Bare Metal, Nutanix, Apache CloudStack, and AWS Snow. You can run EKS Anywhere without a connection to AWS Cloud and in air-gapped environments, or you can optionally connect to AWS Cloud to integrate with other AWS services. You can use the [EKS Connector](https://docs.aws.amazon.com/eks/latest/userguide/eks-connector.html) to view your EKS Anywhere clusters in the Amazon EKS console, AWS IAM to authenticate to your EKS Anywhere clusters, IAM Roles for Service Accounts (IRSA) to authenticate Pods with other AWS services, and AWS Distro for OpenTelemetry to send metrics to Amazon Managed Prometheus for monitoring cluster resources.

If you have on-premises or edge environments with reliable connectivity to an AWS Region, consider using [EKS Hybrid Nodes](https://docs.aws.amazon.com/eks/latest/userguide/hybrid-nodes-overview.html) or [EKS on Outposts](https://docs.aws.amazon.com/eks/latest/userguide/eks-outposts.html) to benefit from the AWS-managed EKS control plane and consistent experience with EKS in AWS Cloud.

EKS Anywhere is built on the Kubernetes sub-project called [Cluster API](https://cluster-api.sigs.k8s.io/) (CAPI), which is focused on providing declarative APIs and tooling to simplify the provisioning, upgrading, and operating of multiple Kubernetes clusters. While EKS Anywhere simplifies and abstracts the CAPI primitives, it is useful to understand the basics of CAPI when using EKS Anywhere. 

### Why EKS Anywhere?
* Simplify and automate Kubernetes management on-premises
* Unify Kubernetes distribution and support across on-premises, edge, and cloud environments
* Run in isolated or air-gapped on-premises environments
* Adopt modern operational practices and tools on-premises
* Build on open source standards

### Common Use Cases
* Modernize on-premises applications from virtual machines to containers
* Internal development platforms to standardize how teams consume Kubernetes across the organization
* Telco 5G Radio Access Networks (RAN) and Core workloads
* Regulated services in private data centers on-premises

### What's Next?
* [Review EKS Anywhere Concepts]({{< relref "../concepts/" >}})
