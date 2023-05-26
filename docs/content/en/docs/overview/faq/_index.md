---
title: "Frequently Asked Questions"
linkTitle: "FAQ"
aliases:
    /docs/reference/faq/
weight: 15
description: >
  Frequently asked questions about EKS Anywhere
---

## AuthN / AuthZ

### How do my applications running on EKS Anywhere authenticate with AWS services using IAM credentials?

You can now leverage the [IAM Role for Service Account (IRSA)](https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/) feature 
by following the [IRSA reference]({{< relref "../../getting-started/optional/irsa.md" >}}) guide for details.


### Does EKS Anywhere support OIDC (including Azure AD and AD FS)?

Yes, EKS Anywhere can create clusters that support API server OIDC authentication.
This means you can federate authentication through AD FS locally or through Azure AD, along with other IDPs that support the OIDC standard.
In order to add OIDC support to your EKS Anywhere clusters, you need to configure your cluster by updating the configuration file before creating the cluster.
Please see the [OIDC reference]({{< relref "../../getting-started/optional/oidc.md" >}}) for details.

### Does EKS Anywhere support LDAP?
EKS Anywhere does not support LDAP out of the box.
However, you can look into the [Dex LDAP Connector](https://dexidp.io/docs/connectors/ldap/).

### Can I use AWS IAM for Kubernetes resource access control on EKS Anywhere?
Yes, you can install the [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) on your EKS Anywhere cluster to achieve this.

## Miscellaneous

### How much does EKS Anywhere cost?

EKS Anywhere is free, open source software that you can download, install on your existing hardware, and run in your own data centers.
It includes management and CLI tooling for all supported [cluster topologies]({{< relref "../../concepts/architecture" >}}) on all supported [providers]({{< relref "../../getting-started/chooseprovider" >}}).
You are responsible for providing infrastructure where EKS Anywhere runs (e.g. VMware, bare metal), and some providers require third party hardware and software contracts.

The [EKS Anywhere Enterprise Subscription](https://aws.amazon.com/eks/eks-anywhere/pricing/) provides access to curated packages and enterprise support.
This is an optional—but recommended—cost based on how many clusters and how many years of support you need.

### Can I connect my EKS Anywhere cluster to EKS?

Yes, you can install EKS Connector to connect your EKS Anywhere cluster to AWS EKS.
EKS Connector is a software agent that you can install on the EKS Anywhere cluster that enables the cluster to communicate back to AWS.
Once connected, you can immediately see a read-only view of the EKS Anywhere cluster with workload and cluster configuration information on the EKS console, alongside your EKS clusters.

#### How does the EKS Connector authenticate with AWS?

During start-up, the EKS Connector generates and stores an RSA key-pair as Kubernetes secrets.
It also registers with AWS using the public key and the activation details from the cluster registration configuration file.
The EKS Connector needs AWS credentials to receive commands from AWS and to send the response back.
Whenever it requires AWS credentials, it uses its private key to sign the request and invokes AWS APIs to request the credentials.

#### How does the EKS Connector authenticate with my Kubernetes cluster?

The EKS Connector acts as a proxy and forwards the EKS console requests to the Kubernetes API server on your cluster.
In the initial release, the connector uses [impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation) with its service account secrets to interact with the API server.
Therefore, you need to associate the connector’s service account with a ClusterRole,
which gives permission to impersonate AWS IAM entities.

#### How do I enable an AWS user account to view my connected cluster through the EKS console?

For each AWS user or other IAM identity, you should add cluster role binding to the Kubernetes cluster with the appropriate permission for that IAM identity.
Additionally, each of these IAM entities should be associated with the IAM policy
to invoke the EKS Connector on the cluster.

### Can I use Amazon Controllers for Kubernetes (ACK) on EKS Anywhere?

Yes, you can leverage AWS services from your EKS Anywhere clusters on-premises through [Amazon Controllers for Kubernetes (ACK)](https://aws.amazon.com/blogs/containers/aws-controllers-for-kubernetes-ack/).


### Can I deploy EKS Anywhere on other clouds?

EKS Anywhere can be installed on any infrastructure with the required Bare Metal, Cloudstack, or VMware vSphere components.
See EKS Anywhere [Baremetal]({{< relref "../../getting-started/baremetal" >}}), [CloudStack]({{< relref "../../getting-started/cloudstack" >}}), or [vSphere]({{< relref "../../getting-started/vsphere" >}}) documentation.

### How is EKS Anywhere different from ECS Anywhere?

[Amazon ECS Anywhere](https://aws.amazon.com/ecs/anywhere/) is an option for [Amazon Elastic Container Service (ECS)](https://aws.amazon.com/ecs/) to run containers on your on-premises infrastructure.
The ECS Anywhere Control Plane runs in an AWS region and allows you to install the ECS agent on worker nodes that run outside of an AWS region.
Workloads that run on ECS Anywhere nodes are scheduled by ECS.
You are not responsible for running, managing, or upgrading the ECS Control Plane.

EKS Anywhere runs the Kubernetes Control Plane and worker nodes on your infrastructure.
You are responsible for managing the EKS Anywhere Control Plane and worker nodes.
There is no requirement to have an AWS account to run EKS Anywhere.

If you'd like to see how EKS Anywhere compares to EKS please see the [information here.]({{< relref "../../concepts/eksafeatures.md" >}})

### How can I manage EKS Anywhere at scale?

You can perform cluster life cycle and configuration management at scale through GitOps-based tools.
EKS Anywhere offers git-driven cluster management through the integrated Flux Controller.
See [Manage cluster with GitOps]({{< relref "../../clustermgmt/cluster-flux.md" >}}) documentation for details.

### Can I run EKS Anywhere on ESXi?

No. EKS Anywhere is only supported on providers listed on the [EKS Anywhere providers]({{< relref "../../getting-started/chooseprovider/" >}}) page.
There would need to be a change to the upstream project to support ESXi.

### Can I deploy EKS Anywhere on a single node?

Yes. Single node cluster deployment is supported for Bare Metal. See [workerNodeGroupConfigurations]({{< relref "../../getting-started/baremetal/bare-spec/#workernodegroupconfigurations">}})
