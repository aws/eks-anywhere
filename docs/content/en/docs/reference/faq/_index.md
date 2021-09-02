---
title: "Frequently Asked Questions"
linkTitle: "FAQ"
weight: 10
description: >
  Frequently asked questions about EKS Anywhere
---

## AuthN / AuthZ 

### How do my applications running on EKS-A authenticate with AWS services using IAM credentials?

You can leverage the [IAM Role for Service Account (IRSA)](https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/) feature 
and follow the [special instructions for DIY Kubernetes](https://github.com/aws/amazon-eks-pod-identity-webhook/blob/master/SELF_HOSTED_SETUP.md)
to enable AWS authentication.
This solution includes the following steps at a high level:

1. Create a key pair for signing, and host your public key somewhere (e.g. S3).

1. Configure your EKS Anywhere Kubernetes API server so it can issue and mount projected service account tokens in pods.

1. Create an IAM role defining access to the target AWS services and annotate a service account with said IAM role.

1. Finally, configure your pods by using the service account created in the previous step and assume the IAM role.

The key pair solution above requires you to set up and maintain your key hosting (e.g. key rotation).
A solution that is currently in development will allow your IAM role credentials to be securely injected to the EKS Anywhere cluster
and assumed by the pods without any key pair.

### Does EKS Anywhere support OIDC (including Azure AD and AD FS)?

Yes, EKS-A can create clusters that support API server OIDC authentication.
This means you can federate authentication through AD FS locally or through Azure AD,
along with other IDPs that support the OIDC standard.
In order to add OIDC support to your EKS-A clusters, you need to configure your cluster by updating the configuration file before creating the cluster.
Please see the [OIDC reference]({{< relref "../../reference/clusterspec/oidc.md" >}}) for details.

### Does EKS Anywhere support LDAP?
EKS-A does not support LDAP out of the box.
However, you can look into the [Dex LDAP Connector](https://dexidp.io/docs/connectors/ldap/).

### Can I use AWS IAM for Kubernetes resource access control on EKS-A?
Yes, you can install the [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) on your EKS-A cluster to achieve this.

## Miscellaneous

### Can I connect my EKS-A cluster to EKS?

Yes, you can install EKS Connector to connect your EKS-A cluster to AWS EKS.
EKS Connector is a software agent that you can install on the EKS-A cluster that enables the cluster to communicate back to AWS.
Once connected, you can immediately see the EKS-A cluster with workload and cluster configuration information on the EKS console,
alongside your EKS clusters. 

#### How does the EKS Connector authenticate with AWS?

During start-up, the EKS Connector generates and stores an RSA key-pair as Kubernetes secrets.
It also registers with AWS using the public key and the activation details from the cluster registration configuration file.
The EKS Connector needs AWS credentials to receive commands from AWS and to send the response back.
Whenever it requires AWS credentials, it uses its private key to sign the request and invokes AWS APIs to request the credentials.

#### How does the EKS Connector authenticate  with my Kubernetes cluster?

The EKS Connector acts as a proxy and forwards the EKS console requests to the Kubernetes API server on your cluster.
In the initial release, the connector uses [impersonation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation) with its service account secrets to interact with the API server.
Therefore, you need to associate the connector’s service account with a ClusterRole,
which gives permission to impersonate AWS IAM entities.

#### How do I enable an AWS user account to view my connected cluster through the EKS console?

For each AWS user or other IAM identity, you should add cluster role binding to the Kubernetes cluster with the appropriate permission for that IAM identity.
Additionally, each of these IAM entities should be associated with the [IAM policy](SOME_MANAGED_POLICY)
to invoke the EKS Connector on the cluster.

### Can I use Amazon Controllers for Kubernetes (ACK) on EKS Anywhere?

Yes, you can leverage AWS services from your EKS Anywhere clusters on-premises through [Amazon Controllers for Kubernetes (ACK)](https://aws.amazon.com/blogs/containers/aws-controllers-for-kubernetes-ack/).


### Can I deploy EKS-A on other clouds?

EKS Anywhere can be installed on any infrastructure with the required VMware vSphere versions.
See EKS Anywhere [vSphere prerequisite]({{< relref "../vsphere" >}}) documentation.

### Do you support multi-cluster operations?

You can perform cluster life cycle and configuration management at scale through GitOps-based tools.
EKS Anywhere offers git-driven cluster management through the integrated Flux Controller.
See [Manage cluster with GitOps]({{< relref "../../tasks/cluster/cluster-flux.md" >}}) documentation for details.
