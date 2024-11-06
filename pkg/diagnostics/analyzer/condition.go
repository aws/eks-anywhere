package analyzer

import (
	corev1 "k8s.io/api/core/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func condition(cluster *anywherev1.Cluster, conditionType anywherev1.ConditionType) *anywherev1.Condition {
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}

func isTrue(condition *anywherev1.Condition) bool {
	return condition != nil && condition.Status == corev1.ConditionTrue
}
