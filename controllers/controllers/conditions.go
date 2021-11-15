package controllers

import (
	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

type ConditionsObject interface {
	GetConditions() capi.Conditions
}

// TODO: hacky. Duplicating this from capi code bc the package we import only supports this logic for v1alpha4 and we only use v1alpha3 structs
//  there might be a way to use conversion
func IsTrue(obj ConditionsObject, condition string) bool {
	if c := get(obj, capi.ConditionType(condition)); c != nil {
		return c.Status == corev1.ConditionTrue
	}
	return false
}

func get(from ConditionsObject, t capi.ConditionType) *capi.Condition {
	conditions := from.GetConditions()
	if conditions == nil {
		return nil
	}

	for _, condition := range conditions {
		if condition.Type == t {
			return &condition
		}
	}
	return nil
}
