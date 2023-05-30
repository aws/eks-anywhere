---
title: Create local cluster
weight: 20
---

## EKS Anywhere docker provider deployments

EKS Anywhere supports a Docker provider for *development and testing use cases only.* 
This allows you to try EKS Anywhere on your local system before deploying to a supported provider to create either:

* A single, standalone cluster or
* Multiple management/workload clusters on the same provider, as described in [Cluster topologies]({{< relref "../../concepts/cluster-topologies" >}}).
The management/workload topology is recommended for production clusters and can be tried out here using both `eksctl` and `GitOps` tools.

## Create a standalone cluster

### Prerequisite Checklist

To install the EKS Anywhere binaries and see system requirements please follow the [installation guide]({{< relref "../install" >}}).

### Steps

<!-- this content needs to be indented so the numbers are automatically incremented -->

1. Generate a cluster config
   ```bash
   CLUSTER_NAME=mgmt
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider docker > $CLUSTER_NAME.yaml
   ```

   The command above creates a file named mgmt.yaml (if you used the provided example) with the contents below in the path where it is executed.
   The configuration specification is divided into two sections:

   * Cluster
   * DockerDatacenterConfig

   ```yaml
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: Cluster
   metadata:
      name: mgmt
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
         name: mgmt
      externalEtcdConfiguration:
         count: 1
      kubernetesVersion: "1.27"
      managementCluster:
         name: mgmt
      workerNodeGroupConfigurations:
         - count: 1
            name: md-0
   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: DockerDatacenterConfig
   metadata:
      name: mgmt
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
   **NOTE**: The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. Due to this there might be some warnings in the CLI if proper authentication is not set up.

1. Create Cluster:

   For a regular cluster create (with internet access), type the following:

      ```bash
      eksctl anywhere create cluster \
         # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
         -f $CLUSTER_NAME.yaml
      ```
   For an airgapped cluster create, follow [Preparation for airgapped deployments]({{< relref "../install/#prepare-for-airgapped-deployments-optional" >}}) instructions, then type the following:

      ```bash
      eksctl anywhere create cluster 
         # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
         -f $CLUSTER_NAME.yaml \
         --bundles-override ./eks-anywhere-downloads/bundle-release.yaml
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
      **NOTE**: to install curated packages during cluster creation, use `--install-packages packages.yaml` flag  
   
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

## Create management/workload clusters
To try the recommended EKS Anywhere [topology,]({{< relref "../../concepts/cluster-topologies" >}}) you can create a management cluster and one or more workload clusters on the same Docker provider.

### Prerequisite Checklist

To install the EKS Anywhere binaries and see system requirements please follow the [installation guide]({{< relref "../install" >}}).

### Create a management cluster

<!-- this content needs to be indented so the numbers are automatically incremented -->

1. Generate a management cluster config (named `mgmt` for this example):
   ```bash
   CLUSTER_NAME=mgmt
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider docker > eksa-mgmt-cluster.yaml
   ```

1. Modify the management cluster config (`eksa-mgmt-cluster.yaml`) you could use the same one described earlier or modify it to use GitOps, as shown below:

   ```bash
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: Cluster
   metadata:
     name: mgmt
     namespace: default
   spec:
     bundlesRef:
       apiVersion: anywhere.eks.amazonaws.com/v1alpha1
       name: bundles-1
       namespace: eksa-system
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
       name: mgmt
     externalEtcdConfiguration:
       count: 1
     gitOpsRef:
       kind: FluxConfig
       name: mgmt
     kubernetesVersion: "1.27"
     managementCluster:
       name: mgmt
     workerNodeGroupConfigurations:
     - count: 1
       name: md-1
    
   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: DockerDatacenterConfig
   metadata:
     name: mgmt
     namespace: default
   spec: {}

   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: FluxConfig
   metadata:
     name: mgmt
     namespace: default
   spec:
     branch: main
     clusterConfigPath: clusters/mgmt
     github:
       owner: <your github account, such as example for https://github.com/example>
       personal: true
       repository: <your github repo, such as test for https://github.com/example/test>
     systemNamespace: flux-system
   
   ---
   ```

1. Configure Curated Packages

   **The Amazon EKS Anywhere Curated Packages are only available to customers with the Amazon EKS Anywhere Enterprise Subscription. To request a free trial, talk to your Amazon representative or connect with one [here](https://aws.amazon.com/contact-us/sales-support-eks/). Cluster creation will succeed if authentication is not set up, but some warnings may be generated.  Detailed package configurations can be found [here]({{< relref "../../tasks/packages" >}}).**

   If you are going to use packages, set up authentication. These credentials should have [limited capabilities]({{< relref "../../tasks/packages/#setup-authentication-to-use-curated-packages" >}}):
   ```bash
   export EKSA_AWS_ACCESS_KEY_ID="your*access*id"
   export EKSA_AWS_SECRET_ACCESS_KEY="your*secret*key"  
   ```

1. Create cluster:

   For a regular cluster create (with internet access), type the following:

      ```bash
      eksctl anywhere create cluster \ 
         # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
         -f $CLUSTER_NAME.yaml
      ```
   For an airgapped cluster create, follow [Preparation for airgapped deployments]({{< relref "../install/#prepare-for-airgapped-deployments-optional" >}}) instructions, then type the following:

      ```bash
      eksctl anywhere create cluster \
         # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
         -f $CLUSTER_NAME.yaml \
         --bundles-override ./eks-anywhere-downloads/bundle-release.yaml
      ```

1. Once the cluster is created you can use it with the generated `KUBECONFIG` file in your local directory:

   ```bash
   export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
   ```

1. Check the initial cluster's CRD:

   To ensure you are looking at the initial cluster, list the CRD to see that the name of its management cluster is itself:

   ```bash
   kubectl get clusters mgmt -o yaml
   ```

   Example command output
   ```
   ...
   kubernetesVersion: "1.27"
   managementCluster:
     name: mgmt
   workerNodeGroupConfigurations:
   ...
   ```

### Create separate workload clusters

Follow these steps to have your management cluster create and manage separate workload clusters.

1. Generate a workload cluster config:
   ```bash
   CLUSTER_NAME=w01
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider docker > eksa-w01-cluster.yaml
   ```
   Refer to the initial config described earlier for the required and optional settings.

   >**NOTE**: Ensure workload cluster object names (`Cluster`, `DockerDatacenterConfig`, etc.) are distinct from management cluster object names. Be sure to set the `managementCluster` field to identify the name of the management cluster.

1. Create a workload cluster in one of the following ways:

   * **GitOps**: See [Manage separate workload clusters with GitOps]({{< relref "../../tasks/cluster/cluster-flux.md#manage-separate-workload-clusters-using-gitops" >}})

   * **Terraform**: See [Manage separate workload clusters with Terraform]({{< relref "../../tasks/cluster/cluster-terraform.md#manage-separate-workload-clusters-using-terraform" >}})

   * **Kubernetes CLI**: The cluster lifecycle feature lets you use `kubectl` to manage a workload cluster. For example:
      ```bash
      kubectl apply -f eksa-w01-cluster.yaml 
      ```
     
   * **eksctl CLI**: Useful for temporary cluster configurations. To create a workload cluster with `eksctl`, do one of the following.
     For a regular cluster create (with internet access), type the following:

      ```bash
      eksctl anywhere create cluster \
          -f eksa-w01-cluster.yaml  \
         # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
          --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
      ```
     For an airgapped cluster create, follow [Preparation for airgapped deployments]({{< relref "../install/#prepare-for-airgapped-deployments-optional" >}}) instructions, then type the following:

      ```bash
      eksctl create cluster \
         # --install-packages packages.yaml \ # uncomment to install curated packages at cluster creation
         -f $CLUSTER_NAME.yaml \
         --bundles-override ./eks-anywhere-downloads/bundle-release.yaml \
          --kubeconfig mgmt/mgmt-eks-a-cluster.kubeconfig
      ```
      As noted earlier, adding the `--kubeconfig` option tells `eksctl` to use the management cluster identified by that kubeconfig file to create a different workload cluster.

1. To check the workload cluster, get the workload cluster credentials and run a [test workload:]({{< relref "../../tasks/workload/test-app" >}})

   * If your workload cluster was created with `eksctl`,
      change your credentials to point to the new workload cluster (for example, `w01`), then run the test application with:

      ```bash
      export CLUSTER_NAME=w01
      export KUBECONFIG=${PWD}/${CLUSTER_NAME}/${CLUSTER_NAME}-eks-a-cluster.kubeconfig
      kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
      ```

   * If your workload cluster was created with kubectl, GitOps or Terraform, you can get credentials and run the test application as follows:
      ```bash
      kubectl get secret -n eksa-system w01-kubeconfig -o jsonpath='{.data.value}' | base64 --decode > w01.kubeconfig
      export KUBECONFIG=w01.kubeconfig
      kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
      ```
     
        **NOTE**: For Docker, you must modify the `server` field of the kubeconfig file by replacing the IP with `127.0.0.1` and the port with its value. 
        The port’s value can be found by running `docker ps` and checking the workload cluster’s load balancer.

1. Add more workload clusters:

   To add more workload clusters, go through the same steps for creating the initial workload, copying the config file to a new name (such as `eksa-w02-cluster.yaml`), modifying resource names, and running the create cluster command again.

## Next steps:
* See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section for more information on common operational tasks like scaling and deleting the cluster.

* See the [Package management]({{< relref "../../tasks/packages" >}}) section for more information on post-creation curated packages installation.
