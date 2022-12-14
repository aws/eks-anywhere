---
title: Create local cluster
weight: 20
---

## EKS Anywhere docker provider deployments

EKS Anywhere supports a Docker provider for *development and testing use cases only.* 
This allows you to try EKS Anywhere on your local system before deploying to a supported provider.

To install the EKS Anywhere binaries and see system requirements please follow the [installation guide]({{< relref "../install" >}}).

## Steps

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate a cluster config
   ```bash
   CLUSTER_NAME=dev-cluster
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider docker > $CLUSTER_NAME.yaml
   ```

   The command above creates a file named eksa-cluster.yaml with the contents below in the path where it is executed.
   The configuration specification is divided into two sections:

   * Cluster
   * DockerDatacenterConfig

   ```yaml
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: Cluster
   metadata:
   name: dev-cluster
   spec:
   clusterNetwork:
      cniConfig:
         cilium: {}
      pods:
         cidrBlocks:
         - 192.168.0.0/16
      services:
         cidrBlocks:
         - 10.96.0.0/12
   controlPlaneConfiguration:
      count: 1
   datacenterRef:
      kind: DockerDatacenterConfig
      name: dev-cluster
   externalEtcdConfiguration:
      count: 1
   kubernetesVersion: "1.21"
   managementCluster:
      name: dev-cluster
   workerNodeGroupConfigurations:
   - count: 1
      name: md-0
   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: DockerDatacenterConfig
   metadata:
      name: dev-cluster
   spec: {}
   ```

   * Apart from the base configuration, you can add additional optional configuration to enable supported features:
      * [OIDC]({{< relref "../../reference/clusterspec/optional/oidc" >}})
      * [etcd]({{< relref "../../reference/clusterspec/optional/etcd" >}})
      * [proxy]({{< relref "../../reference/clusterspec/optional/proxy" >}})
      * [gitops]({{< relref "../../reference/clusterspec/optional/gitops" >}})

1. Configure Curated Packages

   **The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. To request a free trial, talk to your Amazon representative or connect with one [here](https://aws.amazon.com/contact-us/sales-support-eks/). Cluster creation will succeed if authentication is not set up, but some warnings may be generated.  Detailed package configurations can be found [here]({{< relref "../../tasks/packages" >}}).**

   If you are going to use packages, set up authentication. These credentials should have [limited capabilities]({{< relref "../../tasks/packages/#setup-authentication-to-use-curated-packages" >}}):
   ```bash
   export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
   export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"
   export EKSA_AWS_REGION="us-west-2"
   ```

1. Create Cluster:

     **Note** The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. Due to this there might be some warnings in the CLI if proper authentication is not set up.
      ```bash
      eksctl anywhere create cluster -f $CLUSTER_NAME.yaml
      ```
     Example command output
      ```
      Performing setup and validations
      ✅ validation succeeded {"validation": "docker Provider setup is valid"}
      Creating new bootstrap cluster
      Installing cluster-api providers on bootstrap cluster
      Provider specific setup
      Creating new workload cluster
      Installing networking on workload cluster
      Installing cluster-api providers on workload cluster
      Moving cluster management from bootstrap to workload cluster
      Installing EKS-A custom components (CRD and controller) on workload cluster
      Creating EKS-A CRDs instances on workload cluster
      Installing GitOps Toolkit on workload cluster
      GitOps field not specified, bootstrap flux skipped
      Deleting bootstrap cluster
      🎉 Cluster created!
      ----------------------------------------------------------------------------------
      The Amazon EKS Anywhere Curated Packages are only available to customers with the
      Amazon EKS Anywhere Enterprise Subscription
      ----------------------------------------------------------------------------------
      Installing curated packages controller on management cluster
      secret/aws-secret created
      job.batch/eksa-auth-refresher created
      ```
      **Note** to install curated packages during cluster creation, use `--install-packages packages.yaml` flag  
   
1. Use the cluster

   Once the cluster is created you can use it with the generated `KUBECONFIG` file in your local directory

   ```bash
   export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
   kubectl get ns
   ```
   Example command output
   ```
   NAME                                STATUS   AGE
   capd-system                         Active   21m
   capi-kubeadm-bootstrap-system       Active   21m
   capi-kubeadm-control-plane-system   Active   21m
   capi-system                         Active   21m
   capi-webhook-system                 Active   21m
   cert-manager                        Active   22m
   default                             Active   23m
   eksa-packages                       Active   23m
   eksa-system                         Active   20m
   kube-node-lease                     Active   23m
   kube-public                         Active   23m
   kube-system                         Active   23m
   ```

   You can now use the cluster like you would any Kubernetes cluster.
   Deploy the test application with:

   ```bash
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```

   Verify the test application in the [deploy test application section]({{< relref "../../tasks/workload/test-app" >}}).

## Next steps:
* See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section for more information on common operational tasks like scaling and deleting the cluster.

* See the [Package management]({{< relref "../../tasks/packages" >}}) section for more information on post-creation curated packages installation.
