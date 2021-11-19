export CLOUDSTACK_USERNAME="cloudadmin@cloudstack.local"
export CLOUDSTACK_PASSWORD="bogus"
export CLOUDSTACK_DATACENTER="https://192.168.1.150:8080/client/api"
export CLOUDSTACK_NETWORK="GuestNet1"
export CLOUDSTACK_TEMPLATE="ubuntu-1804-kube-v1.19.6"
export CONTROL_PLANE_ENDPOINT_IP="1.2.3.4"
export CLOUDSTACK_TLS_THUMBPRINT="ABCDEFG"
export CLOUDSTACK_SSH_AUTHORIZED_KEY="ssh-rsa"
kind get clusters | grep cloudstack
if [ $? -eq 1 ]
then
    set -e
    kind create cluster --name tlhowe
    clusterctl init --infrastructure cloudstack
fi
clusterctl config cluster default --kubernetes-version v1.19.6 >out2
