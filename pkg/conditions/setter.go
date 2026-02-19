package conditions

import (
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// Setter interface defines methods that an EKS-A object should implement in order to
// use the conditions package for setting conditions.
type Setter interface {
	Getter
	SetConditions(anywherev1.Conditions)
}

// Set sets the given condition.
//
// NOTE: If a condition already exists, the LastTransitionTime is updated only if a change is detected
// in any of the following fields: Status, Reason, Severity and Message.
func Set(to Setter, condition *anywherev1.Condition) {
	if to == nil || condition == nil {
		return
	}

	conditions := to.GetConditions()
	exists := false
	for i := range conditions {
		existingCondition := conditions[i]
		if existingCondition.Type == condition.Type {
			exists = true
			if !HasSameState(&existingCondition, condition) {
				condition.LastTransitionTime = metav1.NewTime(time.Now().UTC().Truncate(time.Second))
				conditions[i] = *condition
				break
			}
			condition.LastTransitionTime = existingCondition.LastTransitionTime
			break
		}
	}

	if !exists {
		if condition.LastTransitionTime.IsZero() {
			condition.LastTransitionTime = metav1.NewTime(time.Now().UTC().Truncate(time.Second))
		}
		conditions = append(conditions, *condition)
	}

	sort.Slice(conditions, func(i, j int) bool {
		return lexicographicLess(&conditions[i], &conditions[j])
	})

	to.SetConditions(conditions)
}

// TrueCondition returns a condition with Status=True and the given type.
func TrueCondition(t anywherev1.ConditionType) *anywherev1.Condition {
	return &anywherev1.Condition{
		Type:   t,
		Status: corev1.ConditionTrue,
	}
}

// FalseCondition returns a condition with Status=False and the given type.
func FalseCondition(t anywherev1.ConditionType, reason string, severity anywherev1.ConditionSeverity, messageFormat string, messageArgs ...interface{}) *anywherev1.Condition {
	return &anywherev1.Condition{
		Type:     t,
		Status:   corev1.ConditionFalse,
		Reason:   reason,
		Severity: severity,
		Message:  fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// UnknownCondition returns a condition with Status=Unknown and the given type.
func UnknownCondition(t anywherev1.ConditionType, reason string, messageFormat string, messageArgs ...interface{}) *anywherev1.Condition {
	return &anywherev1.Condition{
		Type:    t,
		Status:  corev1.ConditionUnknown,
		Reason:  reason,
		Message: fmt.Sprintf(messageFormat, messageArgs...),
	}
}

// MarkTrue sets Status=True for the condition with the given type.
func MarkTrue(to Setter, t anywherev1.ConditionType) {
	Set(to, TrueCondition(t))
}

// MarkUnknown sets Status=Unknown for the condition with the given type.
func MarkUnknown(to Setter, t anywherev1.ConditionType, reason, messageFormat string, messageArgs ...interface{}) {
	Set(to, UnknownCondition(t, reason, messageFormat, messageArgs...))
}

// MarkFalse sets Status=False for the condition with the given type.
func MarkFalse(to Setter, t anywherev1.ConditionType, reason string, severity anywherev1.ConditionSeverity, messageFormat string, messageArgs ...interface{}) {
	Set(to, FalseCondition(t, reason, severity, messageFormat, messageArgs...))
}

// SetSummary sets a Ready condition with the summary of all the conditions existing
// on an object. If the object does not have other conditions, no summary condition is generated.
func SetSummary(to Setter, options ...MergeOption) {
	Set(to, summary(to, options...))
}

// Delete deletes the condition with the given type.
func Delete(to Setter, t anywherev1.ConditionType) {
	if to == nil {
		return
	}

	conditions := to.GetConditions()
	newConditions := make(anywherev1.Conditions, 0, len(conditions))
	for _, condition := range conditions {
		if condition.Type != t {
			newConditions = append(newConditions, condition)
		}
	}
	to.SetConditions(newConditions)
}

// lexicographicLess returns true if a condition is less than another with the Ready condition
// always going first, followed by all the other conditions sorted by Type.
func lexicographicLess(i, j *anywherev1.Condition) bool {
	if i == nil {
		return true
	}
	if j == nil {
		return false
	}
	return (i.Type == anywherev1.ReadyCondition || i.Type < j.Type) && j.Type != anywherev1.ReadyCondition
}

// HasSameState returns true if a condition has the same state of another; state is defined
// by the union of following fields: Type, Status, Reason, Severity and Message (it excludes LastTransitionTime).
func HasSameState(i, j *anywherev1.Condition) bool {
	if i == nil || j == nil {
		return i == j
	}
	return i.Type == j.Type &&
		i.Status == j.Status &&
		i.Reason == j.Reason &&
		i.Severity == j.Severity &&
		i.Message == j.Message
}
