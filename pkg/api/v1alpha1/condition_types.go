package v1alpha1

import (
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

type (
	// ConditionType is an alias for clusterv1.ConditionType.
	ConditionType = clusterv1.ConditionType
	// Condition is an alias for clusterv1.Condition.
	Condition = clusterv1.Condition
)
