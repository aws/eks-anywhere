package internal

import "github.com/aws/eks-anywhere/pkg/constants"

// CAPIDeployments is a map where key = namespace and value is a capi deployment.
var CAPIDeployments = map[string][]string{
	"capi-kubeadm-bootstrap-system":     {"capi-kubeadm-bootstrap-controller-manager"},
	"capi-kubeadm-control-plane-system": {"capi-kubeadm-control-plane-controller-manager"},
	"capi-system":                       {"capi-controller-manager"},
	"cert-manager":                      {"cert-manager", "cert-manager-cainjector", "cert-manager-webhook"},
}

var ExternalEtcdDeployments = map[string][]string{
	"etcdadm-controller-system":         {"etcdadm-controller-controller-manager"},
	"etcdadm-bootstrap-provider-system": {"etcdadm-bootstrap-provider-controller-manager"},
}

var EksaDeployments = map[string][]string{
	constants.EksaSystemNamespace: {constants.EksaControllerManagerDeployment},
}
