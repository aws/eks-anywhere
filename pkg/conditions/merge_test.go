package conditions

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestMergeNilReturnsNil(t *testing.T) {
	result := merge(nil, "Ready", &mergeOptions{})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMergeEmptyReturnsNil(t *testing.T) {
	result := merge([]localizedCondition{}, "Ready", &mergeOptions{})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMergeAllTrueReturnsTrue(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}
	conds := []localizedCondition{
		{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionTrue}, Polarity: PositivePolarity, Getter: cluster},
		{Condition: &anywherev1.Condition{Type: "B", Status: corev1.ConditionTrue}, Polarity: PositivePolarity, Getter: cluster},
	}
	result := merge(conds, "Ready", &mergeOptions{})
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Status != corev1.ConditionTrue {
		t.Errorf("expected True, got %s", result.Status)
	}
}

func TestMergeOneFalseReturnsFalse(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}
	conds := []localizedCondition{
		{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionTrue}, Polarity: PositivePolarity, Getter: cluster},
		{Condition: &anywherev1.Condition{Type: "B", Status: corev1.ConditionFalse, Severity: anywherev1.ConditionSeverityWarning, Reason: "Bad", Message: "msg"}, Polarity: PositivePolarity, Getter: cluster},
	}
	result := merge(conds, "Ready", &mergeOptions{})
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Status != corev1.ConditionFalse {
		t.Errorf("expected False, got %s", result.Status)
	}
	if result.Severity != anywherev1.ConditionSeverityWarning {
		t.Errorf("expected Warning, got %s", result.Severity)
	}
}

func TestMergeUnknownCondition(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}
	conds := []localizedCondition{
		{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionUnknown, Reason: "Checking"}, Polarity: PositivePolarity, Getter: cluster},
	}
	result := merge(conds, "Ready", &mergeOptions{})
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Status != corev1.ConditionUnknown {
		t.Errorf("expected Unknown, got %s", result.Status)
	}
}

func TestMergeNilConditionSkipped(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}
	conds := []localizedCondition{
		{Condition: nil, Polarity: PositivePolarity, Getter: cluster},
		{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionTrue}, Polarity: PositivePolarity, Getter: cluster},
	}
	result := merge(conds, "Ready", &mergeOptions{})
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Status != corev1.ConditionTrue {
		t.Errorf("expected True, got %s", result.Status)
	}
}

func TestMergeErrorPriorityOverWarning(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}
	conds := []localizedCondition{
		{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionFalse, Severity: anywherev1.ConditionSeverityError, Reason: "Err"}, Polarity: PositivePolarity, Getter: cluster},
		{Condition: &anywherev1.Condition{Type: "B", Status: corev1.ConditionFalse, Severity: anywherev1.ConditionSeverityWarning, Reason: "Warn"}, Polarity: PositivePolarity, Getter: cluster},
	}
	result := merge(conds, "Ready", &mergeOptions{})
	if result.Severity != anywherev1.ConditionSeverityError {
		t.Errorf("expected Error severity, got %s", result.Severity)
	}
}

func TestMergeNegativePolarity(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}
	conds := []localizedCondition{
		{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionFalse}, Polarity: NegativePolarity, Getter: cluster},
	}
	result := merge(conds, "Ready", &mergeOptions{})
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Status != corev1.ConditionTrue {
		t.Errorf("expected True (negative polarity False maps to True group), got %s", result.Status)
	}
}

func TestConditionGroupMergePriority(t *testing.T) {
	tests := []struct {
		name     string
		group    conditionGroup
		expected int
	}{
		{"false-error", conditionGroup{status: corev1.ConditionFalse, severity: anywherev1.ConditionSeverityError}, 0},
		{"false-warning", conditionGroup{status: corev1.ConditionFalse, severity: anywherev1.ConditionSeverityWarning}, 1},
		{"false-info", conditionGroup{status: corev1.ConditionFalse, severity: anywherev1.ConditionSeverityInfo}, 2},
		{"true", conditionGroup{status: corev1.ConditionTrue}, 3},
		{"unknown", conditionGroup{status: corev1.ConditionUnknown}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if p := tt.group.mergePriority(); p != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, p)
			}
		})
	}
}

func TestConditionGroupsSorting(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}
	lcs := []localizedCondition{
		{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionTrue}, Polarity: PositivePolarity, Getter: cluster},
		{Condition: &anywherev1.Condition{Type: "B", Status: corev1.ConditionFalse, Severity: anywherev1.ConditionSeverityError}, Polarity: PositivePolarity, Getter: cluster},
	}
	sorted := getConditionGroups(lcs)
	if sorted.TopGroup().severity != anywherev1.ConditionSeverityError {
		t.Errorf("expected Error group first, got severity %s", sorted.TopGroup().severity)
	}
}

func TestTrueGroup(t *testing.T) {
	groups := conditionGroups{
		{status: corev1.ConditionTrue, severity: anywherev1.ConditionSeverityNone},
		{status: corev1.ConditionFalse, severity: anywherev1.ConditionSeverityError},
	}

	tg := groups.TrueGroup()
	if tg == nil {
		t.Fatal("expected TrueGroup, got nil")
	}
	if tg.status != corev1.ConditionTrue {
		t.Errorf("expected True, got %s", tg.status)
	}

	emptyGroups := conditionGroups{}
	if emptyGroups.TrueGroup() != nil {
		t.Error("expected nil for empty groups")
	}
}
