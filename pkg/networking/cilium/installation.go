package cilium

import appsv1 "k8s.io/api/apps/v1"

// Installation represents the Cilium components installed in a cluster
type Installation struct {
	DaemonSet *appsv1.DaemonSet
	Operator  *appsv1.Deployment
}

// Installed determines if all Cilium components are present
func (i Installation) Installed() bool {
	return i.DaemonSet != nil && i.Operator != nil
}
