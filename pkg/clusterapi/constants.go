package clusterapi

import clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

const (
	ControlPlaneReadyCondition clusterv1.ConditionType = "ControlPlaneReady"
)
