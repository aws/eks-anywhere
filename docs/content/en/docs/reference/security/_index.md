---
title: "Security best practices"
linkTitle: "Security best practices"
weight: 40
description: >
  Using security best practices with your EKS-A deployments
---

**If you discover a potential
security issue in this project, we ask that you notify AWS/Amazon Security via our
[vulnerability reporting page](http://aws.amazon.com/security/vulnerability-reporting/).
Please do not create a public GitHub issue for security problems.**

This guide provides advice about best practices for EKS-A specific security concerns. 
For a more complete treatment of Kubernetes security generally
please refer to the official [Kubernetes documentation on Securing a Cluster](https://kubernetes.io/docs/tasks/administer-cluster/securing-a-cluster/) and the [Amazon EKS Best Practices Guide for Security](https://aws.github.io/aws-eks-best-practices/security/docs/index.html).

## The Shared Responsibility Model and EKS-A
AWS Cloud Services follow the [Shared Responsibility Model,](https://aws.amazon.com/compliance/shared-responsibility-model/)
where AWS is responsible for security “of” the cloud,
while the customer is responsible for security “in” the cloud. 
However, EKS-A is an open-source tool
and the distribution of responsibility differs from that of a managed cloud service like EKS.

### AWS Responsibilities
AWS is responsible for building and delivering a secure tool. 
This tool will provision an initially secure Kubernetes cluster.

AWS is responsible for vetting and securely sourcing the services
and tools packaged with EKS-A and the cluster it creates
(such as coredns, cilium, flux, capi, and govc). 

The EKS-A build and delivery infrastructure, or supply chain, is secured to the standard of any AWS service
and AWS takes responsibility for the secure and reliable delivery of a quality product 
which provisions a secure and stable Kubernetes cluster. 
When the `eksctl anywhere` plugin is executed, EKS-A components are automatically downloaded from AWS.
`eksctl` will then perform checksum verification on the components to ensure their authenticity.

AWS is responsible for the secure development and testing of the EKS-A controller and associated custom resource definitions.

AWS is responsible for the secure development and testing of the EKS-A CLI,
and ensuring it handles sensitive data and cluster resources securely.

### End user responsibilities
The end user is responsible for the entire EKS-A cluster after it has been provisioned. 
AWS provides a mechanism to upgrade the cluster in-place, but it is the responsibility of the end user to perform that upgrade using the provided tools.
End users are responsible for operating their clusters in accordance with [Kubernetes security best practices,](https://kubernetes.io/docs/tasks/administer-cluster/securing-a-cluster/) 
and for the ongoing security of the cluster after it has been provisioned. 
This includes but is not limited to:
- creation or modification of RBAC roles and bindings
- creation or modification of namespaces
- modification of the default container network interface plugin
- configuration of network ingress and load balancing
- use and configuration of container storage interfaces
- the inclusion of add-ons and other services

End users are also responsible for:

* The hardware and software which make up the infrastructure layer
(such as vSphere, ESXi, physical servers, and physical network infrastructure).

* The ongoing maintenance of the cluster nodes,
including the underlying guest operating systems. 
Additionally, while EKS-A provides a streamlined process for upgrading a cluster to a new Kubernetes version,
it is the responsibility of the user to perform the upgrade as necessary.

* Any applications which run “on” the cluster, including their secure operation,
least privilege, and use of well-known and vetted container images.

## EKS-A Security Best Practices
This section captures EKS-A specific security best practices.
Please read this section carefully and follow any guidance
to ensure the ongoing security and reliability of your EKS-A cluster.

### Critical Namespaces

EKS-A creates and uses resources in several critical namespaces. 
All of the EKS-A managed namespaces should be treated as sensitive
and access should be limited to only the most trusted users and processes. 
Allowing additional access or modifying the existing RBAC resources
could potentially allow a subject to access the namespace and the resources that it contains. 
This could lead to the exposure of secrets
or the failure of your cluster due to modification of critical resources.
Here are rules you should follow when dealing with critical namespaces:

* Avoid creating [Roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-example) in these namespaces
or providing users access to them with [ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#clusterrole-example).
For more information about creating limited roles for day-to-day administration and development,
please see the official introduction to [Role Based Access Control (RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

* **Do not modify existing [Roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-example) in these namespaces,
bind existing roles to additional [subjects](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#referring-to-subjects), or create new Roles in the namespace.**

* **Do not modify existing [ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#clusterrole-example)
or bind them to additional subjects.**

* **Avoid using the cluster-admin role,
as it grants permissions over all namespaces.**

* **No subjects except for the most trusted administrators should be permitted to perform ANY action in the critical namespaces.**

The critical namespaces include:

* `eksa-system`
* `capv-system`
* `flux-system`
* `capi-system`
* `capi-webhook-system`
* `capi-kubeadm-control-plane-system`
* `capi-kubeadm-bootstrap-system`
* `cert-manager`
* `kube-system` (as with any Kubernetes cluster,
this namespace is critical to the functioning of your cluster and should be treated with the highest level of sensitivity.)


### Secrets

EKS-A stores sensitive information, like the vSphere credentials and Github Personal Access Token,
in the cluster as native Kubernetes [secrets](https://kubernetes.io/docs/concepts/configuration/secret/).
These secret objects are namespaced, for example in the `eksa-system` and `flux-system` namespace,
and limiting access to the sensitive namespaces will ensure that these secrets will not be exposed.
Additionally, limit access to the underlying node. Access to the node could allow access to the secret content.

EKS-A does not currently support encryption-at-rest for Kubernetes secrets.
EKS-A support for [Key Management Services (KMS)](https://kubernetes.io/docs/tasks/administer-cluster/kms-provider/) is planned.


### The EKS-A `kubeconfig` file

`eksctl anywhere create cluster` creates an EKS-A-based Kubernetes cluster
and outputs a [`kubeconfig`](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file with administrative privileges to the `$PWD/$CLUSTER_NAME` directory.

By default, this `kubeconfig` file uses certificate-based authentication and contains the user certificate data for the administrative user.

**The `kubeconfig` file grants administrative privileges over your cluster to the bearer and the certificate key should be treated as you would any other private key or administrative password.**

The EKS-A-generated kubeconfig file should only be used for interacting with the cluster via `eksctl anywhere` commands,
such as `upgrade`, and for the most privileged administrative tasks.
For more information about creating limited roles for day-to-day administration and development,
please see the [official introduction to Role Based Access Control (RBAC)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

### GitOps

GitOps enabled EKS-A clusters maintain a copy of their cluster configuration in the user provided Git repository.
This configuration acts as the source of truth for the cluster.
Changes made to this configuration will be reflected in the cluster configuration.

AWS recommends that you gate any changes to this repository with mandatory pull request reviews.
Carefully review pull requests for changes which could impact the availability of the cluster
(such as scaling nodes to 0 and deleting the cluster object) or contain secrets.

### Github Personal Access Token

Treat the [GitHub PAT](https://docs.github.com/en/github/authenticating-to-github/keeping-your-account-and-data-secure/creating-a-personal-access-token) used with EKS-A as you would any highly privileged secret,
as it could potentially be used to make changes to your cluster by modifying the contents of the cluster configuration file through the [Github.com](https://github.com/) API.

* Never commit the PAT to a Git repository
* Never share the PAT via untrusted channels
* Never grant non-administrative subjects access to the `flux-system` namespace where the PAT is stored as a native Kubernetes secret.

### Executing EKS-A

Ensure that you execute `eksctl anywhere create cluster` on a trusted workstation
in order to protect the values of sensitive environment variables and the EKS-A generated kubeconfig file.

### SSH Access to Cluster Nodes and ETCD Nodes

EKS-A provides the option to configure an ssh authorized key for access to underlying nodes in a cluster, via `vsphereMachineConfig.Users.sshAuthorizedKeys`.
This grants the associated private key the ability to connect to the cluster
via `ssh` as the user `capv` with `sudo` permissions.
The associated private key should be treated as extremely sensitive,
as `sudo` access to the cluster and ETCD nodes can permit access to secret object data and potentially confer arbitrary control over the cluster.

### VMWare OVAs

Only download OVAs for cluster nodes from official sources,
and do not allow untrusted users or processes to modify the templates used by EKS-A for provisioning nodes.
