package cilium

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Installation represents the Cilium components installed in a cluster.
type Installation struct {
	DaemonSet *appsv1.DaemonSet
	Operator  *appsv1.Deployment
	ConfigMap *corev1.ConfigMap
}

// Installed determines if all Cilium components are present.
// If the ConfigMap doesn't exist we still considered Cilium is installed.
// The installation might not be complete but it can be functional.
func (i Installation) Installed() bool {
	return i.DaemonSet != nil && i.Operator != nil
}
