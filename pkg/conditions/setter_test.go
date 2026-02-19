package conditions

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestTrueCondition(t *testing.T) {
	c := TrueCondition("Ready")
	if c.Type != "Ready" {
		t.Errorf("expected type Ready, got %s", c.Type)
	}
	if c.Status != corev1.ConditionTrue {
		t.Errorf("expected True, got %s", c.Status)
	}
}

func TestFalseCondition(t *testing.T) {
	c := FalseCondition("Ready", "NotReady", anywherev1.ConditionSeverityError, "node %s down", "n1")
	if c.Type != "Ready" {
		t.Errorf("expected type Ready, got %s", c.Type)
	}
	if c.Status != corev1.ConditionFalse {
		t.Errorf("expected False, got %s", c.Status)
	}
	if c.Reason != "NotReady" {
		t.Errorf("expected reason NotReady, got %s", c.Reason)
	}
	if c.Severity != anywherev1.ConditionSeverityError {
		t.Errorf("expected severity Error, got %s", c.Severity)
	}
	if c.Message != "node n1 down" {
		t.Errorf("expected 'node n1 down', got '%s'", c.Message)
	}
}

func TestUnknownCondition(t *testing.T) {
	c := UnknownCondition("Ready", "Pending", "waiting for %d nodes", 3)
	if c.Status != corev1.ConditionUnknown {
		t.Errorf("expected Unknown, got %s", c.Status)
	}
	if c.Message != "waiting for 3 nodes" {
		t.Errorf("expected 'waiting for 3 nodes', got '%s'", c.Message)
	}
}

func TestSet(t *testing.T) {
	t.Run("set new condition", func(t *testing.T) {
		cluster := &anywherev1.Cluster{}
		Set(cluster, TrueCondition("Ready"))
		if len(cluster.Status.Conditions) != 1 {
			t.Fatalf("expected 1 condition, got %d", len(cluster.Status.Conditions))
		}
		if cluster.Status.Conditions[0].Type != "Ready" {
			t.Errorf("expected Ready, got %s", cluster.Status.Conditions[0].Type)
		}
	})

	t.Run("update existing condition same state", func(t *testing.T) {
		cluster := &anywherev1.Cluster{
			Status: anywherev1.ClusterStatus{
				Conditions: anywherev1.Conditions{
					{Type: "Ready", Status: corev1.ConditionTrue},
				},
			},
		}
		origTime := cluster.Status.Conditions[0].LastTransitionTime
		Set(cluster, TrueCondition("Ready"))
		if cluster.Status.Conditions[0].LastTransitionTime != origTime {
			t.Error("LastTransitionTime should not change when state is the same")
		}
	})

	t.Run("update existing condition different state", func(t *testing.T) {
		cluster := &anywherev1.Cluster{
			Status: anywherev1.ClusterStatus{
				Conditions: anywherev1.Conditions{
					{Type: "Ready", Status: corev1.ConditionTrue},
				},
			},
		}
		Set(cluster, FalseCondition("Ready", "Failed", anywherev1.ConditionSeverityError, ""))
		if cluster.Status.Conditions[0].Status != corev1.ConditionFalse {
			t.Errorf("expected False, got %s", cluster.Status.Conditions[0].Status)
		}
		if cluster.Status.Conditions[0].LastTransitionTime.IsZero() {
			t.Error("expected LastTransitionTime to be set")
		}
	})

	t.Run("nil setter", func(t *testing.T) {
		Set(nil, TrueCondition("Ready")) // should not panic
	})

	t.Run("nil condition", func(t *testing.T) {
		cluster := &anywherev1.Cluster{}
		Set(cluster, nil) // should not panic
	})

	t.Run("conditions sorted with Ready first", func(t *testing.T) {
		cluster := &anywherev1.Cluster{}
		Set(cluster, TrueCondition("Workers"))
		Set(cluster, TrueCondition("Ready"))
		Set(cluster, TrueCondition("ControlPlane"))
		if cluster.Status.Conditions[0].Type != "Ready" {
			t.Errorf("expected Ready first, got %s", cluster.Status.Conditions[0].Type)
		}
	})
}

func TestMarkTrue(t *testing.T) {
	cluster := &anywherev1.Cluster{}
	MarkTrue(cluster, "Ready")
	c := Get(cluster, "Ready")
	if c == nil || c.Status != corev1.ConditionTrue {
		t.Error("expected condition to be True")
	}
}

func TestMarkFalse(t *testing.T) {
	cluster := &anywherev1.Cluster{}
	MarkFalse(cluster, "Ready", "NotReady", anywherev1.ConditionSeverityError, "msg")
	c := Get(cluster, "Ready")
	if c == nil {
		t.Fatal("expected condition")
	}
	if c.Status != corev1.ConditionFalse {
		t.Errorf("expected False, got %s", c.Status)
	}
	if c.Reason != "NotReady" {
		t.Errorf("expected NotReady, got %s", c.Reason)
	}
}

func TestMarkUnknown(t *testing.T) {
	cluster := &anywherev1.Cluster{}
	MarkUnknown(cluster, "Ready", "Pending", "msg")
	c := Get(cluster, "Ready")
	if c == nil {
		t.Fatal("expected condition")
	}
	if c.Status != corev1.ConditionUnknown {
		t.Errorf("expected Unknown, got %s", c.Status)
	}
}

func TestSetSummary(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "ControlPlaneReady", Status: corev1.ConditionTrue},
				{Type: "WorkersReady", Status: corev1.ConditionTrue},
			},
		},
	}
	SetSummary(cluster, WithConditions("ControlPlaneReady", "WorkersReady"))
	c := Get(cluster, "Ready")
	if c == nil {
		t.Fatal("expected Ready condition")
	}
	if c.Status != corev1.ConditionTrue {
		t.Errorf("expected True, got %s", c.Status)
	}
}

func TestDelete(t *testing.T) {
	t.Run("delete existing", func(t *testing.T) {
		cluster := &anywherev1.Cluster{
			Status: anywherev1.ClusterStatus{
				Conditions: anywherev1.Conditions{
					{Type: "Ready", Status: corev1.ConditionTrue},
					{Type: "Other", Status: corev1.ConditionTrue},
				},
			},
		}
		Delete(cluster, "Ready")
		if Has(cluster, "Ready") {
			t.Error("expected Ready condition to be deleted")
		}
		if !Has(cluster, "Other") {
			t.Error("expected Other condition to remain")
		}
	})

	t.Run("delete non-existing", func(t *testing.T) {
		cluster := &anywherev1.Cluster{
			Status: anywherev1.ClusterStatus{
				Conditions: anywherev1.Conditions{
					{Type: "Ready", Status: corev1.ConditionTrue},
				},
			},
		}
		Delete(cluster, "NonExistent") // should not panic
		if len(cluster.Status.Conditions) != 1 {
			t.Errorf("expected 1 condition, got %d", len(cluster.Status.Conditions))
		}
	})

	t.Run("nil setter", func(t *testing.T) {
		Delete(nil, "Ready") // should not panic
	})
}

func TestHasSameState(t *testing.T) {
	a := &anywherev1.Condition{Type: "Ready", Status: corev1.ConditionTrue, Reason: "R", Severity: anywherev1.ConditionSeverityNone, Message: "m"}
	b := &anywherev1.Condition{Type: "Ready", Status: corev1.ConditionTrue, Reason: "R", Severity: anywherev1.ConditionSeverityNone, Message: "m"}

	if !HasSameState(a, b) {
		t.Error("expected same state")
	}

	c := &anywherev1.Condition{Type: "Ready", Status: corev1.ConditionFalse, Reason: "R", Severity: anywherev1.ConditionSeverityNone, Message: "m"}
	if HasSameState(a, c) {
		t.Error("expected different state")
	}

	if !HasSameState(nil, nil) {
		t.Error("expected nil == nil")
	}
	if HasSameState(a, nil) {
		t.Error("expected non-nil != nil")
	}
	if HasSameState(nil, a) {
		t.Error("expected nil != non-nil")
	}
}

func TestLexicographicLess(t *testing.T) {
	ready := &anywherev1.Condition{Type: "Ready"}
	alpha := &anywherev1.Condition{Type: "Alpha"}
	beta := &anywherev1.Condition{Type: "Beta"}

	if !lexicographicLess(ready, alpha) {
		t.Error("Ready should come before Alpha")
	}
	if lexicographicLess(alpha, ready) {
		t.Error("Alpha should not come before Ready")
	}
	if !lexicographicLess(alpha, beta) {
		t.Error("Alpha should come before Beta")
	}
	if !lexicographicLess(nil, alpha) {
		t.Error("nil should come before anything")
	}
	if lexicographicLess(alpha, nil) {
		t.Error("non-nil should not come before nil")
	}
}
