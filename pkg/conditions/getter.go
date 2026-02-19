package conditions

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// Getter interface defines methods that an EKS-A object should implement in order to
// use the conditions package for getting conditions.
type Getter interface {
	client.Object

	// GetConditions returns the list of conditions for an EKS-A object.
	GetConditions() anywherev1.Conditions
}

// Get returns the condition with the given type, if the condition does not exist,
// it returns nil.
func Get(from Getter, t anywherev1.ConditionType) *anywherev1.Condition {
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

// Has returns true if a condition with the given type exists.
func Has(from Getter, t anywherev1.ConditionType) bool {
	return Get(from, t) != nil
}

// IsTrue is true if the condition with the given type is True, otherwise it returns false
// if the condition is not True or if the condition does not exist (is nil).
func IsTrue(from Getter, t anywherev1.ConditionType) bool {
	if c := Get(from, t); c != nil {
		return c.Status == corev1.ConditionTrue
	}
	return false
}

// IsFalse is true if the condition with the given type is False, otherwise it returns false
// if the condition is not False or if the condition does not exist (is nil).
func IsFalse(from Getter, t anywherev1.ConditionType) bool {
	if c := Get(from, t); c != nil {
		return c.Status == corev1.ConditionFalse
	}
	return false
}

// IsUnknown is true if the condition with the given type is Unknown or if the condition
// does not exist (is nil).
func IsUnknown(from Getter, t anywherev1.ConditionType) bool {
	if c := Get(from, t); c != nil {
		return c.Status == corev1.ConditionUnknown
	}
	return true
}

// GetReason returns a nil safe string of Reason for the condition with the given type.
func GetReason(from Getter, t anywherev1.ConditionType) string {
	if c := Get(from, t); c != nil {
		return c.Reason
	}
	return ""
}

// GetMessage returns a nil safe string of Message.
func GetMessage(from Getter, t anywherev1.ConditionType) string {
	if c := Get(from, t); c != nil {
		return c.Message
	}
	return ""
}

// summary returns a Ready condition with the summary of all the conditions existing
// on an object. If the object does not have other conditions, no summary condition is generated.
//
//nolint:gocyclo
func summary(from Getter, options ...MergeOption) *anywherev1.Condition {
	conditions := from.GetConditions()

	mergeOpt := &mergeOptions{}
	for _, o := range options {
		o(mergeOpt)
	}

	conditionsInScope := make([]localizedCondition, 0, len(conditions))
	for i := range conditions {
		c := conditions[i]
		if c.Type == anywherev1.ReadyCondition {
			continue
		}

		if mergeOpt.conditionTypes != nil {
			found := false
			for _, t := range mergeOpt.conditionTypes {
				if c.Type == t {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		polarity := PositivePolarity
		for _, t := range mergeOpt.negativeConditionTypes {
			if c.Type == t {
				polarity = NegativePolarity
				break
			}
		}

		conditionsInScope = append(conditionsInScope, localizedCondition{
			Condition: &c,
			Polarity:  polarity,
			Getter:    from,
		})
	}

	if mergeOpt.addStepCounterIfOnlyConditionTypes != nil {
		for _, c := range conditionsInScope {
			found := false
			for _, t := range mergeOpt.addStepCounterIfOnlyConditionTypes {
				if c.Type == t {
					found = true
					break
				}
			}
			if !found {
				mergeOpt.addStepCounter = false
				break
			}
		}
	}

	if mergeOpt.addStepCounter {
		mergeOpt.stepCounter = len(conditionsInScope)
		if mergeOpt.conditionTypes != nil {
			mergeOpt.stepCounter = len(mergeOpt.conditionTypes)
		}
		if mergeOpt.addStepCounterIfOnlyConditionTypes != nil {
			mergeOpt.stepCounter = len(mergeOpt.addStepCounterIfOnlyConditionTypes)
		}
	}

	return merge(conditionsInScope, anywherev1.ReadyCondition, mergeOpt)
}
