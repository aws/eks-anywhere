---
title: "Frequently Asked Questions"
linkTitle: "FAQ"
aliases:
    /docs/reference/faq/
weight: 15
description: >
  Frequently asked questions about EKS Anywhere
---

## General

#### Where can I deploy EKS Anywhere?
EKS Anywhere is designed to run on customer-managed infrastructure in customer-managed environments. EKS Anywhere supports different types of infrastructure including VMware vSphere, bare metal, Nutanix, AWS Snowball Edge, and Apache CloudStack. 

#### Can I run EKS Anywhere in the cloud?
No, EKS Anywhere is not supported to run in AWS or other clouds, and EC2 instances cannot be used as the infrastructure for EKS Anywhere clusters. This includes EC2 instances running in AWS Regions, Local Zones, and Outposts.

#### What operating systems can I use with EKS Anywhere?
EKS Anywhere provides Bottlerocket, a Linux-based container-native operating system built by AWS, as the default node operating system for clusters on VMware vSphere. You can alternatively use Ubuntu and Red Hat Enterprise Linux (RHEL) as the node operating system. You can only use a single node operating system per cluster. Bottlerocket is the only operating system distributed and fully supported by AWS. If you are using the other operating systems, you must build the operating system images and configure EKS Anywhere to use the images you built when installing or updating clusters. AWS will assist with troubleshooting and configuration guidance for Ubuntu and RHEL as part of EKS Anywhere Enterprise Subscriptions. For official support for Ubuntu and RHEL operating systems, you must purchase support through their respective vendors.

#### Does EKS Anywhere require a connection to AWS?
EKS Anywhere can run connected to an AWS Region or disconnected from an AWS Region, including in air-gapped environments. If you run EKS Anywhere connected to an AWS Region, you can view your clusters in the Amazon EKS console with the EKS Connector and can optionally use AWS IAM for cluster authentication, AWS IAM Roles for Service Accounts (IRSA), cert-manager with Amazon Certificate Manager, the AWS Distro for OpenTelemetry (ADOT) collector with Amazon Managed Prometheus, and FluentBit with Amazon CloudWatch Logs.

#### What are the differences between EKS Anywhere and EKS Hybrid Nodes?

[EKS Hybrid Nodes](https://docs.aws.amazon.com/eks/latest/userguide/hybrid-nodes-overview.html) is a feature of Amazon EKS, a managed Kubernetes service, whereas EKS Anywhere is AWS-supported Kubernetes management software that you manage. EKS Hybrid Nodes is a fit for customers with on-premises environments that can be connected to the cloud, whereas EKS Anywhere is a fit for customers with isolated or air-gapped on-premises environments. 

With EKS Hybrid Nodes, AWS manages the security, availability, and scalability of the Kubernetes control plane, which is hosted in AWS Cloud, and only nodes run on your infrastructure. With EKS Anywhere, you are responsible for managing the Kubernetes clusters that run entirely on your infrastructure. EKS Hybrid Nodes uses a "bring-your-own-infrastructure" approach where you are responsible for the provisioning and management of the infrastructure used for your hybrid nodes compute with your own choice of tooling whereas EKS Anywhere integrates with Cluster API (CAPI) to provision and manage the infrastructure used as nodes in EKS Anywhere clusters.

With EKS Hybrid Nodes, there are no upfront commitments or minimum fees and you pay for the hourly use of your cluster and nodes as you use them. The EKS Hybrid Nodes fee is based on vCPU of the connected hybrid nodes. With EKS Anywhere, you can purchase EKS Anywhere Enterprise Subscriptions for a one-year or three-year term on a per cluster basis.

#### What are the differences between EKS Anywhere and ECS Anywhere?

[ECS Anywhere](https://aws.amazon.com/ecs/anywhere/) is a feature of Amazon ECS that can be used to run containers on your on-premises infrastructure. ECS Anywhere is similar to EKS Hybrid Nodes, where the ECS control plane runs in an AWS Region, managed by ECS, and you connect your on-premises hosts as instances to your ECS clusters to enable tasks to be scheduled on your on-premises ECS instances. With EKS Anywhere, you are responsible for managing the Kubernetes clusters that run entirely on your infrastructure.

## Architecture

#### What infrastructure do I need to get started with EKS Anywhere?
To get started with EKS Anywhere, you need 1 [admin machine]({{< relref "../../getting-started/install" >}}) and at least 1 VM for the control plane and 1 VM for the worker node if you are running on VMware vSphere, Nutanix, AWS Snowball Edge, or Apache CloudStack. If you are running on bare metal, you need at least 1 [admin machine]({{< relref "../../getting-started/install" >}}) and 1 physical server for the co-located control plane and worker node.

#### What infrastructure do I need to run EKS Anywhere in production?
To use EKS Anywhere in production, it is generally recommended to run separate management and workload clusters, see [EKS Anywhere Architecture]({{< relref "../../concepts/architecture" >}}) for more information. It is recommended to run both the management cluster and workload clusters in a highly available fashion with the Kubernetes control plane instances spread across multiple virtual or physical hosts. For management clusters, management components are run on worker nodes that are separate from the Kubernetes control plane machines. For workload clusters, application workloads are run on worker nodes that are separate from the Kubernetes control plane machines, unless you are running on bare metal which allows for co-locating the Kubernetes control plane and worker nodes on the same physical machines. 

If you are using VMware vSphere, Nutanix, AWS Snowball Edge, or Apache CloudStack for your infrastructure, it is recommended to run at least 3 separate virtual machines for the etcd instances of the Kubernetes control plane, which can be configured with the `externalEctdConfiguration` setting of the EKS Anywhere cluster specification. For more information, see the installation [Overview]({{< relref "../../getting-started/overview" >}}) and the requirements for using EKS Anywhere for each infrastructure provider below.

- [Requirements for VMware vSphere]({{< relref "../../getting-started/vsphere/vsphere-prereq">}})
- [Requirements for bare metal]({{< relref "../../getting-started/baremetal/bare-prereq">}})
- [Requirements for Nutanix]({{< relref "../../getting-started/nutanix/nutanix-prereq">}})
- [Requirements for Snow]({{< relref "../../getting-started/snow/snow-getstarted/#prerequisite-checklist">}})
- [Requirements for Apache CloudStack]({{< relref "../../getting-started/cloudstack/cloudstack-prereq">}})

#### What permissions does EKS Anywhere need to manage infrastructure used for the cluster?
EKS Anywhere needs permissions to create the virtual machines that are used as nodes in EKS Anywhere clusters. If you are running EKS Anywhere on bare metal, EKS Anywhere needs to be able to remotely manage your bare metal servers for network booting. You must configure these permissions before creating EKS Anywhere clusters.

- [Prepare VMware vSphere]({{< relref "../../getting-started/vsphere/vsphere-prereq">}})
- [Prepare hardware for bare metal]({{< relref "../../getting-started/baremetal/bare-preparation">}})
- [Prepare Nutanix]({{< relref "../../getting-started/nutanix/nutanix-preparation">}})
- [Prepare Snow]({{< relref "../../getting-started/snow/snow-getstarted">}})
- [Prepare Apache CloudStack]({{< relref "../../getting-started/cloudstack/cloudstack-preparation">}})

#### What components does EKS Anywhere use?
EKS Anywhere is built on the Kubernetes sub-project called [Cluster API](https://cluster-api.sigs.k8s.io/) (CAPI), which is focused on providing declarative APIs and tooling to simplify the provisioning, upgrading, and operating of multiple Kubernetes clusters. EKS Anywhere inherits many of the same architectural patterns and concepts that exist in CAPI. Reference the [CAPI documentation](https://cluster-api.sigs.k8s.io/user/concepts) to learn more about the core CAPI concepts.

EKS Anywhere has four categories of components, all based on open source software: 

- Administrative / CLI components: Responsible for lifecycle operations of management or standalone clusters, building images, and collecting support diagnostics. Admin / CLI components run on Admin machines or image building machines.
- Management components: Responsible for infrastructure and cluster lifecycle management (create, update, upgrade, scale, delete). Management components run on standalone or management clusters.
- Cluster components: Components that make up a Kubernetes cluster where applications run. Cluster components run on standalone, management, and workload clusters.
- Curated packages: Amazon-curated software packages that extend the core functionalities of Kubernetes on your EKS Anywhere clusters

For more information on EKS Anywhere components, reference [EKS Anywhere Architecture.]({{< relref "../../concepts/architecture" >}})

#### What interfaces can I use to manage EKS Anywhere clusters?

The tools available for cluster lifecycle operations (create, update, upgrade, scale, delete) vary based on the EKS Anywhere architecture you run. You must use the eksctl CLI for cluster lifecycle operations with standalone clusters and management clusters. If you are running a management / workload cluster architecture, you can use the management cluster to manage one-to-many downstream workload clusters. With the management cluster architecture, you can use the eksctl CLI, any Kubernetes API-compatible client, or Infrastructure as Code (IAC) tooling such as Terraform and GitOps to manage the lifecycle of workload clusters. For details on the differences between the architecture options, reference the Architecture page .

To perform cluster lifecycle operations for standalone, management, or workload clusters, you modify the EKS Anywhere Cluster specification, which is a Kubernetes Custom Resource for EKS Anywhere clusters. When you modify a field in an existing Cluster specification, EKS Anywhere reconciles the infrastructure and Kubernetes components until they match the new desired state you defined. 

## EKS Anywhere Enterprise Subscriptions

For more information on EKS Anywhere Enterprise Subscriptions, see the [Overview of support for EKS Anywhere.]({{< relref "../../concepts/support-scope" >}})

#### What is included in EKS Anywhere Enterprise Subscriptions?
EKS Anywhere Enterprise Subscriptions include support for EKS Anywhere clusters, access to EKS Anywhere Curated Packages, and access to extended support for Kubernetes versions. If you do not have an EKS Anywhere Enterprise Subscription, you cannot get support for EKS Anywhere clusters through AWS Support. 

#### How much do EKS Anywhere Enterprise Subscriptions cost?
For pricing information, visit the [EKS Anywhere Pricing](https://aws.amazon.com/eks/eks-anywhere/pricing/) page.

#### How can I purchase an EKS Anywhere Enterprise Subscription?
Reference the [Purchase Subscriptions]({{< relref "../../clustermgmt/support/purchase-subscription" >}}) documentation for instructions on how to purchase.

#### Is there a free trial for EKS Anywhere Enterprise Subscriptions?
Free trial access to EKS Anywhere Curated Packages is available upon request. Free trial access to EKS Anywhere Curated Packages does not include troubleshooting support for your EKS Anywhere deployments. Contact your AWS account team for more information.