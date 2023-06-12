package v1alpha1

import (
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// Conditions and condition Reasons for the Cluster object.

const (
	WaitingForCAPIClusterReason = "WaitingForCAPIClusterInitialized"

	WaitingForCAPIClusterConditionReason = "WaitingForCAPIClusterCondition"

	WaitingForControlPlaneNodesReadyReason = "WaitingForControlPlaneReady"

	WorkersReadyConditon clusterv1.ConditionType = "WorkersReady"

	WaitingForWorkersReadyReason = "WaitingForWorkersReady"

	DefaultCNIConfiguredCondition clusterv1.ConditionType = "DefaultCNIConfigured"

	WaitingForDefaultCNIConfiguredReason = "WaitingForDefaultCNIConfigured"

	SkippedDefaultCNIConfigurationReason = "SkippedDefaultCNIConfigurationReason"
)
