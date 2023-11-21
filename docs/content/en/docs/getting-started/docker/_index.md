---
title: Create Docker Cluster (dev only)
linkTitle: Install on Docker (dev only)
aliases:
    /docs/getting-started/local-environment/
weight: 90
description: >
  Create an EKS Anywhere cluster with Docker on your local machine, laptop, or cloud instance
---

## EKS Anywhere docker provider deployments

EKS Anywhere supports a Docker provider for *development and testing use cases only.* 
This allows you to try EKS Anywhere on your local machine or laptop before deploying to other infrastructure such as vSphere or bare metal.


### Prerequisites

* Mac OS 10.15+ or Ubuntu 20.04.2+ LTS
* [Docker 20.x.x](https://docs.docker.com/engine/install/)
* [`curl`](https://everything.curl.dev/get)
* [`yq`](https://github.com/mikefarah/yq/#install)
* Internet access
* 4 CPU cores
* 16GB memory
* 30GB free disk space
* If you are using Ubuntu, use the Docker CE installation instructions to install Docker and not the Snap installation, as described here.
* For EKS Anywhere v0.15 and earlier, if you are using Ubuntu 21.10 or 22.04, you will need to switch from cgroups v2 to cgroups v1. For details, see [Troubleshooting Guide.]({{< relref "../../troubleshooting/troubleshooting.md#for-eks-anywhere-v015-and-earlier-cgroups-v2-is-not-supported-in-ubuntu-2110-and-2204" >}})
* EKS Anywhere works with x86 and amd64 architectures, From v0.18.1 it also works with Apple Silicon or Arm based processors.

### Install EKS Anywhere CLI tools
To get started with EKS Anywhere, you must first install the `eksctl` CLI and the `eksctl anywhere` plugin.
This is the primary interface for EKS Anywhere and what you will use to create a local Docker cluster. The EKS Anywhere plugin requires eksctl version 0.66.0 or newer.

#### Homebrew

Note if you already have `eksctl` installed, you can install the `eksctl anywhere` plugin manually following the instructions in the following section.
This package also installs `kubectl` and `aws-iam-authenticator`.

```bash
brew install aws/tap/eks-anywhere
```

#### Manual

Install `eksctl`

```bash
curl "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" \
    --silent --location \
    | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin/
```

Install the `eksctl-anywhere` plugin

```bash
RELEASE_VERSION=$(curl https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml --silent --location | yq ".spec.latestVersion")
EKS_ANYWHERE_TARBALL_URL=$(curl https://anywhere-assets.eks.amazonaws.com/releases/eks-a/manifest.yaml --silent --location | yq ".spec.releases[] | select(.version==\"$RELEASE_VERSION\").eksABinary.$(uname -s | tr A-Z a-z).uri")
curl $EKS_ANYWHERE_TARBALL_URL \
    --silent --location \
    | tar xz ./eksctl-anywhere
sudo mv ./eksctl-anywhere /usr/local/bin/
```

Install `kubectl`. See the Kubernetes [documentation](https://kubernetes.io/docs/tasks/tools/) for more information.

```bash
export OS="$(uname -s | tr A-Z a-z)" ARCH=$(test "$(uname -m)" = 'x86_64' && echo 'amd64' || echo 'arm64')
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/${OS}/${ARCH}/kubectl"
sudo mv ./kubectl /usr/local/bin
sudo chmod +x /usr/local/bin/kubectl
```

### Create a local Docker cluster


<!-- this content needs to be indented so the numbers are automatically incremented -->

1. Generate a cluster config. The cluster config will contain the settings for your local Docker cluster. The eksctl anywhere generate command populates a cluster config with EKS Anywhere defaults and best practices.

   ```bash
   CLUSTER_NAME=mgmt
   eksctl anywhere generate clusterconfig $CLUSTER_NAME \
      --provider docker > $CLUSTER_NAME.yaml
   ```

   The command above creates a file named eksa-cluster.yaml with the contents below in the path where it is executed.
   The configuration specification is divided into two sections: Cluster and DockerDatacenterConfig.
   These are the minimum configuration settings you must provide to create a Docker cluster. You can optionally configure OIDC, etcd, proxy, and GitOps as described [here.]({{< relref "../optional/" >}})

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
      kubernetesVersion: "1.28"
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

1. Create Docker Cluster. Note the following command may take several minutes to complete. You can run the command with -v 6 to increase logging verbosity to see the progress of the command. 

      ```bash
      eksctl anywhere create cluster -f $CLUSTER_NAME.yaml
      ```

     Expand for sample output:

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
      ...
      ```
      **NOTE**: to install curated packages during cluster creation, use `--install-packages packages.yaml` flag  
   
1. Access Docker cluster

   Once the cluster is created you can use it with the generated `kubeconfig` in the local directory.
   If you used the same naming conventions as the example above, you will find a `eksa-cluster/eksa-cluster-eks-a-cluster.kubeconfig` in the directory where you ran the commands.

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

1. The following command will deploy a test application:
   
   ```bash
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```
   To interact with the deployed application, review the steps in the [Deploy test workload page]({{< relref "../../workloadmgmt/test-app" >}}).

## Next steps:
* See the [Cluster management]({{< relref "../../clustermgmt" >}}) section for more information on common operational tasks like scaling and deleting the cluster.

* See the [Package management]({{< relref "../../packages" >}}) section for more information on post-creation curated packages installation.
