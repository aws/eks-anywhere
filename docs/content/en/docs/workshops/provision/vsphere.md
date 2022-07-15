---
title: "vSphere cluster"
linkTitle: "vSphere cluster"
weight: 50
date: 2021-11-11
description: >  
---

EKS Anywhere supports a vSphere provider for production grade EKS Anywhere deployments.
EKS Anywhere allows you to provision and manage Amazon EKS on your own infrastructure.

This document walks you through setting up EKS Anywhere in a way that:

* Deploys an initial cluster on your vSphere environment. That cluster can be used as a self-managed cluster (to run workloads) or a management cluster (to create and manage other clusters)
* Deploys zero or more workload clusters from the management cluster

If your initial cluster is a management cluster, it is intended to stay in place so you can use it later to modify, upgrade, and delete workload clusters.
Using a management cluster makes it faster to provision and delete workload clusters.
Also it lets you keep vSphere credentials for a set of clusters in one place: on the management cluster.
The alternative is to simply use your initial cluster to run workloads.

{{% alert title="Important" color="warning" %}}

Creating an EKS Anywhere management cluster is the recommended model.
Separating management features into a separate, persistent management cluster
provides a cleaner model for managing the lifecycle of workload clusters (to create, upgrade, and delete clusters), while workload clusters run user applications.
This approach also reduces provider permissions for workload clusters.

{{% /alert %}}

## Prerequisite Checklist

EKS Anywhere needs to be run on an administrative machine that has certain [machine
requirements]({{< relref "../../getting-started/install" >}}).
An EKS Anywhere deployment will also require the availability of certain
[resources from your VMware vSphere deployment]({{< relref "/docs/reference/vsphere/vsphere-prereq/_index.md" >}}).

## Steps

The following steps are divided into two sections:

* Create an initial cluster (used as a management or self-managed cluster)
* Create zero or more workload clusters from the management cluster

### Create an initial cluster

Follow these steps to create an EKS Anywhere cluster that can be used either as a management cluster or as a self-managed cluster (for running workloads itself).

All steps listed below should be executed on the admin machine with reachability to the vSphere environment where the EKA Anywhere clusters are created.

<!-- this content needs to be indented so the numbers are automatically incremented -->
1. Generate an initial cluster config (named `mgmt-cluster` for this example):
   ```bash
   export MGMT_CLUSTER_NAME=mgmt-cluster
   eksctl anywhere generate clusterconfig $MGMT_CLUSTER_NAME \
      --provider vsphere > $MGMT_CLUSTER_NAME.yaml
   ```

   The command above creates a config file named mgmt-cluster.yaml in the path where it is executed. Refer to [vsphere configuration]({{< relref "../../reference/clusterspec/vsphere" >}}) for information on configuring this cluster config for a vSphere provider.

   The configuration specification is divided into three sections:
   * Cluster
   * VSphereDatacenterConfig
   * VSphereMachineConfig

   Some key considerations and configuration parameters:
      * Create at least two control plane nodes, three worker nodes, and three etcd nodes for a production cluster, to provide high availability and rolling upgrades.   
      * osFamily (operating System on virtual machines) parameter in VSphereMachineConfig by default is set to bottlerocket. Permitted values: ubuntu, bottlerocket.
      * The recommended mode of deploying etcd on EKS Anywhere production clusters is unstacked (etcd members have dedicated machines and are not collocated with control plane components). More information here. The generated config file comes with external etcd enabled already. So leave this part as it is.
      * Apart from the base configuration, you can optionally add additional configuration to enable supported EKS Anywhere functionalities. 
         * [OIDC]({{< relref "/docs/reference/clusterspec/optional/oidc" >}}) 
         * [etcd]({{< relref "/docs/reference/clusterspec/optional/etcd" >}}) (comes by default with the generated config file) 
         * [proxy]({{< relref "/docs/reference/clusterspec/optional/proxy" >}}) 
         * [GitOps]({{< relref "/docs/reference/clusterspec/optional/gitops" >}}) 
         * [IAM for Pods]({{< relref "/docs/reference/clusterspec/optional/irsa" >}}) 
         * [IAM Authenticator]({{< relref "/docs/reference/clusterspec/optional/iamauth" >}}) 
         * [container registry mirror]({{< relref "/docs/reference/clusterspec/optional/registrymirror" >}})

         As of now, you have to pre-determine which features you want to enable on your cluster before cluster creation. Otherwise, to enable them post-creation will require you to delete and recreate the cluster. However, the next EKS-A release will remove such limitation.
      * To enable managing cluster resources using GitOps, you would need to enable GitOps configurations on the initial/managemet cluster. You can not enable GitOps on workload clusters as long as you have enabled it on the initial/management cluster. And if you want to manage the deployment of Kubernetes resources on a workload cluster, then you would need to bootstrap Flux against your workload cluster manually, to be able deploying Kubernetes resources to this workload cluster using GitOps 



1. Modify the initial cluster generated config (`mgmt-cluster.yaml`) as follows:
   You will notice that the generated config file comes with the following fields with empty values. All you need is to fill them with the values we gathered in the [prerequisites]({{< relref "/docs/reference/vsphere/vsphere-prereq/_index.md" >}}) page.

   * Cluster: controlPlaneConfiguration.endpoint.host: ""

      ```yaml
      controlPlaneConfiguration:
         count: 3
         endpoint:
            # Fill this value with the IP address you want to use for the management 
            # cluster control plane endpoint. You will also need  a separate one for the 
            # controlplane of each workload cluster you add later.
            host: "" 
      ```

   * VSphereDatacenterConfig:

      ```yaml
      datacenter: "" # Fill it with the vSphere Datacenter Name. Example: "Example Datacenter"
      insecure: false
      network: "" # Fill it with VM Network Name. Example: "/Example Datacenter/network/VLAN035"
      server: "" # Fill it with the vCenter Server Domain Name. Example: "sample.exampledomain.com"
      thumbprint: "" # Fill it with the thumprint of your vCenter server. Example: "BF:B5:D4:C5:72:E4:04:40:F7:22:99:05:12:F5:0B:0E:D7:A6:35:36"
      ```
   * VSphereMachineConfig sections:
      ```yaml
      datastore: "" # Fill in the vSphere datastore name: Example "/Example Datacenter/datastore/LabStorage"
      diskGiB: 25
      # Fill in the folder name that the VMs of the cluster will be organized under.
      # You will have a separate folder for the management cluster and each cluster you are adding.
      folder: "" # Fill in the foler name Example: /Example Datacenter/vm/EKS Anywhere/mgmt-cluster
      memoryMiB: 8192 
      numCPUs: 2
      osFamily: ubuntu # You can set it to botllerocket or ubuntu
      resourcePool: "" # Fill in the vSphere Resource pool. Example: /Example Datacenter/host/Lab/Resources
      ```
      * Remove the `users` property, and it will be genrated during the cluster creation automatically. It will set the username to `capv` if osFamily=ubuntu, and `ec2-user` if osFamily=botllerocket which is the default option. It will also generate an SSH Key pair, that you can use later to connect to your cluster VMs.
      * Add template property if you chose to import the EKS-A VM OVA template, and set it to the VM template you imported. Check the [vSphere preparation steps]({{< relref "/docs/reference/vsphere/vsphere-preparation/_index.md" >}})
      ```yaml
      template: /Example Datacenter/vm/EKS Anywhere/ubuntu-2004-kube-v1.21.2
      ```
   Refer to [vsphere configuration]({{< relref "../../reference/clusterspec/vsphere" >}}) for more information on the configuring that can be used for a vSphere provider.   

1. Set Credential Environment Variables

   Before you create the initial/management  cluster, you will need to set and export these environment variables for your vSphere user name and password. Make sure you use single quotes around the values so that your shell does not interpret the values
   ```bash
   # vCenter User Credentials
   export GOVC_URL='[vCenter Server Domain Name]'     # Example: https://sample.exampledomain.com
   export GOVC_USERNAME='[vSphere user name]'         # Example: USER1@exampledomain
   export GOVC_PASSWORD='[vSphere password]'                                     
   export GOVC_INSECURE=true
   export EKSA_VSPHERE_USERNAME='[vSphere user name]' # Example: USER1@exampledomain
   export EKSA_VSPHERE_PASSWORD='[vSphere password]'               
   ```

1. Set License Environment Variable

   If you are creating a licensed cluster, set and export the license variable (see [License cluster]({{< relref "/docs/tasks/cluster/cluster-license" >}}) if you are licensing an existing cluster):

   ```bash
   export EKSA_LICENSE='my-license-here'
   ```

1. Now you are ready to create a cluster with the basic stettings.

   {{% alert title="Important" color="warning" %}}

   If you plan to enable other compnents such as, GitOps, oidc, IAM for Pods, etc, Skip creating the cluster now and go ahead adding the configuration for those components to your generated config file first. Or you would need to receate the cluster again as mentioned above.

   {{% /alert %}}

   After you have finish adding all the configuration needed to your configuration file the `mgmt-cluster.yaml` and set your credential environment variables, you are ready to create the cluster. Run the create command with the option -v 9 to get the highest level of verbosity, in case you want to troubleshoot any issue happened during the creation of the cluster. You may need also to output it to a file, so you can look at it later.

   ```bash
   eksctl anywhere create cluster -f $MGMT_CLUSTER_NAME.yaml \
   -v 9 > $MGMT_CLUSTER_NAME-$(date "+%Y%m%d%H%M").log 2>&1
   ```
1. With the completion of the above steps, the management EKS Anywhere cluster is created on the configured vSphere environment under a sub-folder of the `EKS Anywhere` folder. You can see the cluster VMs from the vSphere console as below:

   ![Import ova wizard](/images/vms.png) 

1. Once the cluster is created a folder got created on the admin machine with the cluster name which contains the kubeconfig file and the cluster configuration file used to create the cluster, in addition to the generated SSH key pair that you can use to SSH into the VMs of the cluster. 

   ```bash
   ls mgmt-cluster/
   ```

   Output

   ```bash
   eks-a-id_rsa      mgmt-cluster-eks-a-cluster.kubeconfig
   eks-a-id_rsa.pub  mgmt-cluster-eks-a-cluster.yaml
   ```

1. Now you can use your cluster with the generated `KUBECONFIG` file:

   ```bash
   export KUBECONFIG=${PWD}/${MGMT_CLUSTER_NAME}/${MGMT_CLUSTER_NAME}-eks-a-cluster.kubeconfig
   kubectl cluster-info
   ```
   The cluster endpoint in the output of this command would be the controlPlaneConfiguration.endpoint.host provided in the mgmt-cluster.yaml config file.
   
1. Check the cluster nodes:

   To check that the cluster completed, list the machines to see the control plane, etcd, and worker nodes:

   ```bash
   kubectl get machines -A
   ```

   Example command output
   ```
   NAMESPACE   NAME                PROVIDERID        PHASE    VERSION
   eksa-system mgmt-b2xyz          vsphere:/xxxxx    Running  v1.21.2-eks-1-21-5
   eksa-system mgmt-etcd-r9b42     vsphere:/xxxxx    Running  
   eksa-system mgmt-md-8-6xr-rnr   vsphere:/xxxxx    Running  v1.21.2-eks-1-21-5
   ...
   ```

   The etcd machine doesn't show the Kubernetes version because it doesn't run the kubelet service.

1. Check the initial/management cluster's CRD:

   To ensure you are looking at the initial/management cluster, list the CRD to see that the name of its management cluster is itself:

   ```bash
   kubectl get clusters mgmt -o yaml
   ```

   Example command output
   ```
   ...
   kubernetesVersion: "1.21"
   managementCluster:
     name: mgmt
   workerNodeGroupConfigurations:
   ...
   ```

   {{% alert title="Note" color="primary" %}}
   The initial cluster is now ready to deploy workload clusters.
   However, if you just want to use it to run workloads, you can deploy pod workloads directly on the initial cluster without deploying a separate workload cluster and skip the section on running separate workload clusters.
   {{% /alert %}}

### Create separate workload clusters

Follow these steps if you want to use your initial cluster to create and manage separate workload clusters. All steps listed below should be executed on the same admin machine the management cluster created on. 

1. Generate a workload cluster config:
   ```bash
   export WORKLOAD_CLUSTER_NAME='w01-cluster'
   export MGMT_CLUSTER_NAME='mgmt-cluster'
   eksctl anywhere generate clusterconfig $WORKLOAD_CLUSTER_NAME \
      --provider vsphere > $WORKLOAD_CLUSTER_NAME.yaml
   ```
   The command above creates a file named w01-cluster.yaml with similar contents to the mgmt.cluster.yaml file that was generated for the management cluster in the previous section. It will be generated in the path where it is executed.

   Same key considerations and configuration parameters apply to workload cluster as well, that were mentioned above with the initial cluster.

1. Refer to the initial config described earlier for the required and optional settings.
   The main differences are that you must have a new cluster name and cannot use the same vSphere resources.

1. Modify the generated workload cluster config parameters same way you did in the generated configuration file of the management cluster. The only differences are with the following fields:

   * controlPlaneConfiguration.endpoint.host:
   That you will use a different IP address for the Cluster filed `controlPlaneConfiguration.endpoint.host` for each workload cluster as with the initial cluster. Notice here that you use a different IP address from this one that was used with the management cluster.

   * managementCluster.name:
   By default the value of this field is the same as the cluster name, when you generate the configuration file. But because we want this workload cluster we are adding, to managed by the management cluster, then you need to change that to the management cluster name.

      ```yaml
      managementCluster:
         name: mgmt-cluster # the name of the initial/management cluster
      ```

   * VSphereMachineConfig.folder
   It's recommended to have a separate folder path for each cluster you add for organization purposes.
      ```yaml
      folder: /Example Datacenter/vm/EKS Anywhere/w01-cluster
      ```
   Other than that all other parameters will be configured the same way.

1. Create a workload cluster

   {{% alert title="Important" color="warning" %}}

   If you plan to enable other compnents such as oidc, IAM for Pods, etc, skip creating the cluster now and go ahead adding the configuration for those components to your generated config file first. Or you would need to receate the cluster again. If GitOps have been enabled on the initial/management cluster, you would not have the option to enable GitOps on the workload cluster, as the goal of using GitOps is to centrally manage all of your clusters. 

   {{% /alert %}}

   To create a new workload cluster from your management cluster run this command, identifying:

   * The workload cluster yaml file
   * The initial cluster's credentials (this causes the workload cluster to be managed from the management cluster)


   ```bash
   eksctl anywhere create cluster \
       -f $WORKLOAD_CLUSTER_NAME.yaml \
       --kubeconfig $MGMT_CLUSTER_NAME/$MGMT_CLUSTER_NAME-eks-a-cluster.kubeconfig \
       -v 9 > $WORKLOAD_CLUSTER_NAME-$(date "+%Y%m%d%H%M").log 2>&1
   ```

   As noted earlier, adding the `--kubeconfig` option tells `eksctl` to use the management cluster identified by that kubeconfig file to create a different workload cluster.

1. With the completion of the above steps, the management EKS Anywhere cluster is created on the configured vSphere environment under a sub-folder of the `EKS Anywhere` folder. You can see the cluster VMs from the vSphere console as below:

   ![Import ova wizard](/images/workload-vms.png) 

1. Once the cluster is created a folder got created on the admin machine with the cluster name which contains the kubeconfig file and the cluster configuration file used to create the cluster, in addition to the generated SSH key pair that you can use to SSH into the VMs of the cluster. 

   ```bash
   ls w01-cluster/
   ```

   Output

   ```bash
   eks-a-id_rsa      w01-cluster-eks-a-cluster.kubeconfig
   eks-a-id_rsa.pub  w01-cluster-eks-a-cluster.yaml
   ```

1. You can list the workload clusters managed by the management cluster.
   ```bash
   export KUBECONFIG=${PWD}/${MGMT_CLUSTER_NAME}/${MGMT_CLUSTER_NAME}-eks-a-cluster.kubeconfig
   kubectl get clusters
   ```

1. Check the workload cluster:

   You can now use the workload cluster as you would any Kubernetes cluster.
   Change your credentials to point to the kubconfig file of the new workload cluster, then get the cluster info

   ```bash
   export KUBECONFIG=${PWD}/${WORKLOAD_CLUSTER_NAME}/${WORKLOAD_CLUSTER_NAME}-eks-a-cluster.kubeconfig
   kubectl cluster-info
   ```

   The cluster endpoint in the output of this command should be the controlPlaneConfiguration.endpoint.host provided in the w01-cluster.yaml config file.

1. To verify that the expected number of cluster worker nodes are up and running, use the kubectl command to show that nodes are Ready.

   ```bash
   kubectl get nodes
   ```

1. Test deploying an application with:
   ```bash
   kubectl apply -f "https://anywhere.eks.amazonaws.com/manifests/hello-eks-a.yaml"
   ```

   Verify the test application in the [deploy test application section]({{< relref "../../tasks/workload/test-app" >}}).

1. Add more workload clusters:

   To add more workload clusters, go through the same steps for creating the initial workload, copying the config file to a new name (such as `w01-cluster.yaml`), modifying resource names, and running the create cluster command again.

See the [Cluster management]({{< relref "../../tasks/cluster" >}}) section with more information on common operational tasks like scaling and deleting the cluster.


