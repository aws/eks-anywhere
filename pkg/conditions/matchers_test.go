package conditions

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestHaveSameStateOf(t *testing.T) {
	expected := &anywherev1.Condition{
		Type:     "Ready",
		Status:   corev1.ConditionTrue,
		Reason:   "AllGood",
		Severity: anywherev1.ConditionSeverityNone,
		Message:  "everything is fine",
	}
	matcher := HaveSameStateOf(expected)

	t.Run("match same state", func(t *testing.T) {
		actual := &anywherev1.Condition{
			Type:     "Ready",
			Status:   corev1.ConditionTrue,
			Reason:   "AllGood",
			Severity: anywherev1.ConditionSeverityNone,
			Message:  "everything is fine",
		}
		ok, err := matcher.Match(actual)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Error("expected match")
		}
	})

	t.Run("no match different state", func(t *testing.T) {
		actual := &anywherev1.Condition{
			Type:   "Ready",
			Status: corev1.ConditionFalse,
		}
		ok, err := matcher.Match(actual)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Error("expected no match")
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		_, err := matcher.Match("not a condition")
		if err == nil {
			t.Error("expected error for wrong type")
		}
	})

	t.Run("failure message", func(t *testing.T) {
		msg := matcher.FailureMessage(&anywherev1.Condition{})
		if msg == "" {
			t.Error("expected non-empty failure message")
		}
	})

	t.Run("negated failure message", func(t *testing.T) {
		msg := matcher.NegatedFailureMessage(&anywherev1.Condition{})
		if msg == "" {
			t.Error("expected non-empty negated failure message")
		}
	})
}
