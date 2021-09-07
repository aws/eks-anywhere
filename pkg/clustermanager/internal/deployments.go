package internal

import "github.com/aws/eks-anywhere/pkg/types"

// map where key = namespace and value is a capi deployment
var CapiDeployments = map[string][]string{
	"capi-kubeadm-bootstrap-system":     {"capi-kubeadm-bootstrap-controller-manager"},
	"capi-kubeadm-control-plane-system": {"capi-kubeadm-control-plane-controller-manager"},
	"capi-system":                       {"capi-controller-manager"},
	"capi-webhook-system":               {"capi-controller-manager", "capi-kubeadm-bootstrap-controller-manager", "capi-kubeadm-control-plane-controller-manager"},
	"cert-manager":                      {"cert-manager", "cert-manager-cainjector", "cert-manager-webhook"},
}

var ExternalEtcdDeployments = map[string][]string{
	"etcdadm-controller-system":         {"etcdadm-controller-controller-manager"},
	"etcdadm-bootstrap-provider-system": {"etcdadm-bootstrap-provider-controller-manager"},
}

// map between file name and the capi/v deployments
var ClusterDeployments = map[string]*types.Deployment{
	"kubeadm-bootstrap-controller-manager.log":         {Name: "capi-kubeadm-bootstrap-controller-manager", Namespace: "capi-kubeadm-bootstrap-system", Container: "manager"},
	"kubeadm-control-plane-controller-manager.log":     {Name: "capi-kubeadm-control-plane-controller-manager", Namespace: "capi-kubeadm-control-plane-system", Container: "manager"},
	"capi-controller-manager.log":                      {Name: "capi-controller-manager", Namespace: "capi-system", Container: "manager"},
	"wh-capi-controller-manager.log":                   {Name: "capi-controller-manager", Namespace: "capi-webhook-system", Container: "manager"},
	"wh-capi-kubeadm-bootstrap-controller-manager.log": {Name: "capi-kubeadm-bootstrap-controller-manager", Namespace: "capi-webhook-system", Container: "manager"},
	"wh-kubeadm-control-plane-controller-manager.log":  {Name: "capi-kubeadm-control-plane-controller-manager", Namespace: "capi-webhook-system", Container: "manager"},
	"cert-manager.log":                                 {Name: "cert-manager", Namespace: "cert-manager"},
	"cert-manager-cainjector.log":                      {Name: "cert-manager-cainjector", Namespace: "cert-manager"},
	"cert-manager-webhook.log":                         {Name: "cert-manager-webhook", Namespace: "cert-manager"},
	"coredns.log":                                      {Name: "coredns", Namespace: "kube-system"},
	"local-path-provisioner.log":                       {Name: "local-path-provisioner", Namespace: "local-path-storage"},
	"capv-controller-manager.log":                      {Name: "capv-controller-manager", Namespace: "capv-system", Container: "manager"},
	"wh-capv-controller-manager.log":                   {Name: "capv-controller-manager", Namespace: "capi-webhook-system", Container: "manager"},
}

var EksaDeployments = map[string][]string{
	"eksa-system": {"eksa-controller-manager"},
}
