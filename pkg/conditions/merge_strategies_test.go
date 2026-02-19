package conditions

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestWithNegativePolarityConditions(t *testing.T) {
	opts := &mergeOptions{}
	WithNegativePolarityConditions("A", "B")(opts)
	if len(opts.negativeConditionTypes) != 2 {
		t.Errorf("expected 2 negative condition types, got %d", len(opts.negativeConditionTypes))
	}
}

func TestWithStepCounterIf(t *testing.T) {
	opts := &mergeOptions{}
	WithStepCounterIf(true)(opts)
	if !opts.addStepCounter {
		t.Error("expected addStepCounter to be true")
	}
	WithStepCounterIf(false)(opts)
	if opts.addStepCounter {
		t.Error("expected addStepCounter to be false")
	}
}

func TestWithStepCounterIfOnly(t *testing.T) {
	opts := &mergeOptions{}
	WithStepCounterIfOnly("A", "B")(opts)
	if len(opts.addStepCounterIfOnlyConditionTypes) != 2 {
		t.Errorf("expected 2, got %d", len(opts.addStepCounterIfOnlyConditionTypes))
	}
}

func TestAddSourceRef(t *testing.T) {
	opts := &mergeOptions{}
	AddSourceRef()(opts)
	if !opts.addSourceRef {
		t.Error("expected addSourceRef to be true")
	}
}

func TestLocalizeReason(t *testing.T) {
	g := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "test-cluster"}}

	t.Run("adds source ref", func(t *testing.T) {
		result := localizeReason("ScalingUp", g)
		if result != "ScalingUp @ /test-cluster" {
			t.Errorf("expected 'ScalingUp @ /test-cluster', got '%s'", result)
		}
	})

	t.Run("already has @", func(t *testing.T) {
		result := localizeReason("ScalingUp @ Other/x", g)
		if result != "ScalingUp @ Other/x" {
			t.Errorf("expected unchanged reason, got '%s'", result)
		}
	})
}

func TestGetFirstConditionWithPriority(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}

	t.Run("empty top group", func(t *testing.T) {
		groups := conditionGroups{
			{status: corev1.ConditionFalse, severity: anywherev1.ConditionSeverityError, conditions: []localizedCondition{}},
		}
		c := getFirstCondition(groups, nil)
		if c != nil {
			t.Errorf("expected nil for empty group, got %v", c)
		}
	})

	t.Run("single condition", func(t *testing.T) {
		groups := conditionGroups{
			{
				status:   corev1.ConditionFalse,
				severity: anywherev1.ConditionSeverityError,
				conditions: []localizedCondition{
					{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionFalse, Reason: "Err"}, Getter: cluster},
				},
			},
		}
		c := getFirstCondition(groups, nil)
		if c == nil || c.Type != "A" {
			t.Errorf("expected condition A")
		}
	})

	t.Run("multiple conditions with priority", func(t *testing.T) {
		groups := conditionGroups{
			{
				status:   corev1.ConditionFalse,
				severity: anywherev1.ConditionSeverityError,
				conditions: []localizedCondition{
					{Condition: &anywherev1.Condition{Type: "A"}, Getter: cluster},
					{Condition: &anywherev1.Condition{Type: "B"}, Getter: cluster},
				},
			},
		}
		// With priority, B should come first
		c := getFirstCondition(groups, []anywherev1.ConditionType{"B", "A"})
		if c == nil || c.Type != "B" {
			t.Errorf("expected B (priority), got %v", c)
		}
	})

	t.Run("multiple conditions without priority uses first", func(t *testing.T) {
		groups := conditionGroups{
			{
				status:   corev1.ConditionFalse,
				severity: anywherev1.ConditionSeverityError,
				conditions: []localizedCondition{
					{Condition: &anywherev1.Condition{Type: "X"}, Getter: cluster},
					{Condition: &anywherev1.Condition{Type: "Y"}, Getter: cluster},
				},
			},
		}
		c := getFirstCondition(groups, nil)
		if c == nil || c.Type != "X" {
			t.Errorf("expected X (first), got %v", c)
		}
	})

	t.Run("nil top group", func(t *testing.T) {
		groups := conditionGroups{}
		c := getFirstCondition(groups, nil)
		if c != nil {
			t.Errorf("expected nil, got %v", c)
		}
	})
}

func TestGetReasonAndMessage(t *testing.T) {
	cluster := &anywherev1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "c1"}}

	t.Run("getReason with source ref", func(t *testing.T) {
		groups := conditionGroups{
			{
				status:   corev1.ConditionFalse,
				severity: anywherev1.ConditionSeverityError,
				conditions: []localizedCondition{
					{Condition: &anywherev1.Condition{Type: "A", Reason: "BadThing"}, Getter: cluster},
				},
			},
		}
		reason := getReason(groups, &mergeOptions{addSourceRef: true})
		if reason != "BadThing @ /c1" {
			t.Errorf("expected localized reason, got '%s'", reason)
		}
	})

	t.Run("getMessage with step counter", func(t *testing.T) {
		groups := conditionGroups{
			{
				status: corev1.ConditionTrue,
				conditions: []localizedCondition{
					{Condition: &anywherev1.Condition{Type: "A", Status: corev1.ConditionTrue}, Getter: cluster},
				},
			},
		}
		msg := getMessage(groups, &mergeOptions{addStepCounter: true, stepCounter: 3})
		if msg != "1 of 3 completed" {
			t.Errorf("expected '1 of 3 completed', got '%s'", msg)
		}
	})

	t.Run("getMessage without step counter", func(t *testing.T) {
		groups := conditionGroups{
			{
				status:   corev1.ConditionFalse,
				severity: anywherev1.ConditionSeverityInfo,
				conditions: []localizedCondition{
					{Condition: &anywherev1.Condition{Type: "A", Message: "hello"}, Getter: cluster},
				},
			},
		}
		msg := getMessage(groups, &mergeOptions{})
		if msg != "hello" {
			t.Errorf("expected 'hello', got '%s'", msg)
		}
	})

	t.Run("empty groups return empty strings", func(t *testing.T) {
		groups := conditionGroups{}
		if r := getReason(groups, &mergeOptions{}); r != "" {
			t.Errorf("expected empty reason, got '%s'", r)
		}
		if m := getMessage(groups, &mergeOptions{}); m != "" {
			t.Errorf("expected empty message, got '%s'", m)
		}
	})
}

func TestSummaryWithStepCounterIfOnly(t *testing.T) {
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "A", Status: corev1.ConditionTrue},
				{Type: "B", Status: corev1.ConditionFalse, Reason: "NotReady", Severity: anywherev1.ConditionSeverityInfo},
				{Type: "C", Status: corev1.ConditionTrue},
			},
		},
	}

	t.Run("step counter disabled when extra conditions present", func(t *testing.T) {
		// WithStepCounterIfOnly("A","B") means: show step counter only if conditions in scope are subset of {A,B}.
		// Since C is also present, step counter should be disabled.
		s := summary(cluster, WithConditions("A", "B", "C"), WithStepCounter(), WithStepCounterIfOnly("A", "B"))
		if s == nil {
			t.Fatal("expected summary")
		}
		// When step counter is disabled, message comes from first condition
		if s.Message == "2 of 2 completed" {
			t.Error("step counter should have been disabled")
		}
	})
}

func TestSummaryWithNegativePolarity(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "A", Status: corev1.ConditionTrue},
				{Type: "B", Status: corev1.ConditionFalse}, // negative polarity: False is good
			},
		},
	}
	s := summary(cluster, WithConditions("A", "B"), WithNegativePolarityConditions("B"))
	if s == nil {
		t.Fatal("expected summary")
	}
	if s.Status != corev1.ConditionTrue {
		t.Errorf("expected True (both effectively good), got %s", s.Status)
	}
}
