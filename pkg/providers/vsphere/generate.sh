export VSPHERE_USERNAME="cloudadmin@vmc.local"
export VSPHERE_PASSWORD="bogus"
export VSPHERE_SERVER="vsphere_server"
export VSPHERE_DATACENTER="SDDC-Datacenter"
export VSPHERE_DATASTORE="/SDDC-Datacenter/datastore/WorkloadDatastore"
export VSPHERE_NETWORK="/SDDC-Datacenter/network/sddc-cgw-network-1"
export VSPHERE_RESOURCE_POOL="*/Resources"
export VSPHERE_FOLDER="/SDDC-Datacenter/vm"
export VSPHERE_TEMPLATE="/SDDC-Datacenter/vm/Templates/ubuntu-1804-kube-v1.19.6"
export CONTROL_PLANE_ENDPOINT_IP="1.2.3.4"
export VSPHERE_TLS_THUMBPRINT="ABCDEFG"
export VSPHERE_SSH_AUTHORIZED_KEY="ssh-rsa"
export EXP_CLUSTER_RESOURCE_SET="true"
export VSPHERE_STORAGE_POLICY=""
kind get clusters | grep tlhowe
if [ $? -eq 1 ]
then
    set -e
    kind create cluster --name tlhowe
    clusterctl init --infrastructure vsphere:v0.7.8
fi
clusterctl config cluster default --kubernetes-version v1.19.6 >out2
