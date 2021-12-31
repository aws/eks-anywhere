# Conformance testing Amazon EKS Anywhere

## Setup EKS Anywhere Cluster

Setup EKS Anywhere cluster according to the [EKS Anywhere documentation](https://anywhere.eks.amazonaws.com/).

Create an [EKS Anywhere production cluster](https://anywhere.eks.amazonaws.com/docs/getting-started/production-environment/) to reproduce the EKS Anywhere Conformance e2e results.


## Requirements
Create a Kubernetes cluster on a target workload environment with EKS Anywhere run on an administrative machine.

### Target Workload Environment

The target workload environment will need:

* A vSphere 7+ environment running vCenter
* Capacity to deploy 6-10VMs
* DHCP service running in vSphere environment in the primary VM network for your workload cluster
* One network in vSphere to use for the cluster. This network must have inbound access into vCenter
* A OVA imported into vSphere and converted into template for the workload VMs
* User credentials to [create vms and attach networks, etc](https://anywhere.eks.amazonaws.com/docs/reference/vsphere/user-permissions/)

Each VM will require:

* 2 vCPU
* 8GB RAM
* 25GB Disk

### Administrative Machine

The administrative machine will need:

* Docker 20.x.x
* Mac OS (10.15) / Ubuntu (20.04.2 LTS)
* 4 CPU cores
* 16GB memory
* 30GB free disk space

#### Kubectl

On the administrative machine, install and configure the Kubernetes command-line tool
[kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

#### Docker

The method to [install Docker](https://docs.docker.com/get-docker/) depends on your operating system and architecture.
If you are using Ubuntu use the [Docker CE](https://docs.docker.com/engine/install/ubuntu/) installation instructions to install Docker and not the Snap installation.

#### EKS Anywhere

Install [EKS Anywhere](https://anywhere.eks.amazonaws.com/docs/getting-started/install/) on your administrative machine.

#### Sonobuoy

Download a binary release of [sonobuoy](https://github.com/vmware-tanzu/sonobuoy/releases/).

If you are on a Mac, you many need to open the Security & Privacy and approve sonobuoy for
execution.

```shell
if [[ "$(uname)" == "Darwin" ]]
then
  SONOBUOY=https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.50.0/sonobuoy_0.50.0_darwin_amd64.tar.gz
else
  SONOBUOY=https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.50.0/sonobuoy_0.50.0_linux_386.tar.gz
fi
wget -qO- ${SONOBUOY} |tar -xz sonobuoy
chmod 755 sonobuoy
```

## Create EKS Anywhere Cluster

1. Generate a cluster configuration:

   ```shell
   CLUSTER_NAME=prod
   eksctl anywhere generate clusterconfig $CLUSTER_NAME --provider vsphere >cluster.yaml
   ```

1. Populate cluster configuration. For example:

   ```yaml
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: Cluster
   metadata:
     name: prod
   spec:
     clusterNetwork:
       cni: cilium
       pods:
         cidrBlocks:
         - 192.168.0.0/16
       services:
         cidrBlocks:
         - 10.96.0.0/12
     controlPlaneConfiguration:
       count: 2
       endpoint:
         host: ""
       machineGroupRef:
         kind: VSphereMachineConfig
         name: prod-cp
     datacenterRef:
       kind: VSphereDatacenterConfig
       name: prod
     externalEtcdConfiguration:
       count: 3
       machineGroupRef:
         kind: VSphereMachineConfig
         name: prod-etcd
     kubernetesVersion: "{{eksa_k8s_version}}"
     workerNodeGroupConfigurations:
     - count: 2
       machineGroupRef:
         kind: VSphereMachineConfig
         name: prod
   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: VSphereDatacenterConfig
   metadata:
     name: prod
   spec:
     datacenter: SDDC-Datacenter
     insecure: false
     network: /SDDC-Datacenter/network/sddc-cgw-network-1
     server: vcenter.sddc-12-345-678-9.vmwarevmc.com
     thumbprint: ""
   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: VSphereMachineConfig
   metadata:
     name: prod-cp
   spec:
     datastore: /SDDC-Datacenter/datastore/WorkloadDatastore
     diskGiB: 25
     folder: /SDDC-Datacenter/vm/capv/prod
     memoryMiB: 8192
     numCPUs: 2
     osFamily: bottlerocket
     resourcePool: '*/Resources/Compute-ResourcePool'
     users:
     - name: ec2-user
       sshAuthorizedKeys:
       - "ssh-rsa AAAA..."
   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: VSphereMachineConfig
   metadata:
     name: prod
   spec:
     datastore: /SDDC-Datacenter/datastore/WorkloadDatastore
     diskGiB: 25
     folder: /SDDC-Datacenter/vm/capv/prod
     memoryMiB: 8192
     numCPUs: 2
     osFamily: bottlerocket
     resourcePool: '*/Resources/Compute-ResourcePool'
     users:
     - name: ec2-user
       sshAuthorizedKeys:
       - "ssh-rsa AAAA..."
   ---
   apiVersion: anywhere.eks.amazonaws.com/v1alpha1
   kind: VSphereMachineConfig
   metadata:
     name: prod-etcd
   spec:
     datastore: /SDDC-Datacenter/datastore/WorkloadDatastore
     diskGiB: 25
     folder: /SDDC-Datacenter/vm/capv/prod
     memoryMiB: 8192
     numCPUs: 2
     osFamily: bottlerocket
     resourcePool: '*/Resources/Compute-ResourcePool'
     users:
     - name: ec2-user
       sshAuthorizedKeys:
       - "ssh-rsa AAAA..."
   ```
   
1. Set credential environment variables

   ```shell
   export EKSA_VSPHERE_USERNAME='billy'
   export EKSA_VSPHERE_PASSWORD='t0p$ecret'
   ```
   
1. Create a cluster

   ```shell
   eksctl anywhere create cluster -f cluster.yaml -v 4
   ```


## Run Sonobuoy e2e
```
./sonobuoy run --mode=certified-conformance --wait --kube-conformance-image k8s.gcr.io/conformance:{{conformance_version}}
results=$(./sonobuoy retrieve)
mkdir ./results
tar xzf $results -C ./results
./sonobuoy e2e ${results}
mv results/plugins/e2e/results/global/* .
```

## Cleanup
```shell
eksctl anywhere delete cluster prod -v 4
rm -rf cluster.yaml prod *tar.gz results
```
