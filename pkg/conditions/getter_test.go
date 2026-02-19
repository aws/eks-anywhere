package conditions

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestGet(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "Ready", Status: corev1.ConditionTrue},
				{Type: "ControlPlaneReady", Status: corev1.ConditionFalse, Reason: "Scaling"},
			},
		},
	}

	t.Run("existing condition", func(t *testing.T) {
		c := Get(cluster, "Ready")
		if c == nil {
			t.Fatal("expected condition, got nil")
		}
		if c.Status != corev1.ConditionTrue {
			t.Errorf("expected True, got %s", c.Status)
		}
	})

	t.Run("non-existing condition", func(t *testing.T) {
		c := Get(cluster, "NonExistent")
		if c != nil {
			t.Errorf("expected nil, got %v", c)
		}
	})

	t.Run("nil conditions", func(t *testing.T) {
		empty := &anywherev1.Cluster{}
		c := Get(empty, "Ready")
		if c != nil {
			t.Errorf("expected nil, got %v", c)
		}
	})
}

func TestHas(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "Ready", Status: corev1.ConditionTrue},
			},
		},
	}

	if !Has(cluster, "Ready") {
		t.Error("expected Has to return true for existing condition")
	}
	if Has(cluster, "NonExistent") {
		t.Error("expected Has to return false for non-existing condition")
	}
}

func TestIsTrue(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "Ready", Status: corev1.ConditionTrue},
				{Type: "Failing", Status: corev1.ConditionFalse},
			},
		},
	}

	if !IsTrue(cluster, "Ready") {
		t.Error("expected IsTrue to return true")
	}
	if IsTrue(cluster, "Failing") {
		t.Error("expected IsTrue to return false for False condition")
	}
	if IsTrue(cluster, "NonExistent") {
		t.Error("expected IsTrue to return false for non-existing condition")
	}
}

func TestIsFalse(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "Ready", Status: corev1.ConditionTrue},
				{Type: "Failing", Status: corev1.ConditionFalse},
			},
		},
	}

	if !IsFalse(cluster, "Failing") {
		t.Error("expected IsFalse to return true")
	}
	if IsFalse(cluster, "Ready") {
		t.Error("expected IsFalse to return false for True condition")
	}
	if IsFalse(cluster, "NonExistent") {
		t.Error("expected IsFalse to return false for non-existing condition")
	}
}

func TestIsUnknown(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "Ready", Status: corev1.ConditionTrue},
				{Type: "Unknown", Status: corev1.ConditionUnknown},
			},
		},
	}

	if !IsUnknown(cluster, "Unknown") {
		t.Error("expected IsUnknown to return true for Unknown condition")
	}
	if IsUnknown(cluster, "Ready") {
		t.Error("expected IsUnknown to return false for True condition")
	}
	if !IsUnknown(cluster, "NonExistent") {
		t.Error("expected IsUnknown to return true for non-existing condition")
	}
}

func TestGetReason(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "Ready", Status: corev1.ConditionFalse, Reason: "NotReady"},
			},
		},
	}

	if r := GetReason(cluster, "Ready"); r != "NotReady" {
		t.Errorf("expected 'NotReady', got '%s'", r)
	}
	if r := GetReason(cluster, "NonExistent"); r != "" {
		t.Errorf("expected empty string, got '%s'", r)
	}
}

func TestGetMessage(t *testing.T) {
	cluster := &anywherev1.Cluster{
		Status: anywherev1.ClusterStatus{
			Conditions: anywherev1.Conditions{
				{Type: "Ready", Status: corev1.ConditionFalse, Message: "still starting"},
			},
		},
	}

	if m := GetMessage(cluster, "Ready"); m != "still starting" {
		t.Errorf("expected 'still starting', got '%s'", m)
	}
	if m := GetMessage(cluster, "NonExistent"); m != "" {
		t.Errorf("expected empty string, got '%s'", m)
	}
}

func TestSummary(t *testing.T) {
	t.Run("all true", func(t *testing.T) {
		cluster := &anywherev1.Cluster{
			Status: anywherev1.ClusterStatus{
				Conditions: anywherev1.Conditions{
					{Type: "ControlPlaneReady", Status: corev1.ConditionTrue},
					{Type: "WorkersReady", Status: corev1.ConditionTrue},
				},
			},
		}
		s := summary(cluster, WithConditions("ControlPlaneReady", "WorkersReady"))
		if s == nil {
			t.Fatal("expected summary condition, got nil")
		}
		if s.Status != corev1.ConditionTrue {
			t.Errorf("expected True, got %s", s.Status)
		}
	})

	t.Run("one false", func(t *testing.T) {
		cluster := &anywherev1.Cluster{
			Status: anywherev1.ClusterStatus{
				Conditions: anywherev1.Conditions{
					{Type: "ControlPlaneReady", Status: corev1.ConditionTrue},
					{Type: "WorkersReady", Status: corev1.ConditionFalse, Reason: "ScalingUp", Severity: anywherev1.ConditionSeverityInfo, Message: "scaling"},
				},
			},
		}
		s := summary(cluster, WithConditions("ControlPlaneReady", "WorkersReady"))
		if s == nil {
			t.Fatal("expected summary condition, got nil")
		}
		if s.Status != corev1.ConditionFalse {
			t.Errorf("expected False, got %s", s.Status)
		}
	})

	t.Run("empty conditions", func(t *testing.T) {
		cluster := &anywherev1.Cluster{}
		s := summary(cluster)
		if s != nil {
			t.Errorf("expected nil, got %v", s)
		}
	})

	t.Run("only Ready condition excluded", func(t *testing.T) {
		cluster := &anywherev1.Cluster{
			Status: anywherev1.ClusterStatus{
				Conditions: anywherev1.Conditions{
					{Type: "Ready", Status: corev1.ConditionTrue},
				},
			},
		}
		s := summary(cluster)
		if s != nil {
			t.Errorf("expected nil (Ready excluded from summary), got %v", s)
		}
	})

	t.Run("with step counter", func(t *testing.T) {
		cluster := &anywherev1.Cluster{
			ObjectMeta: metav1.ObjectMeta{Name: "test"},
			Status: anywherev1.ClusterStatus{
				Conditions: anywherev1.Conditions{
					{Type: "A", Status: corev1.ConditionTrue},
					{Type: "B", Status: corev1.ConditionFalse, Reason: "NotReady", Severity: anywherev1.ConditionSeverityInfo},
				},
			},
		}
		s := summary(cluster, WithConditions("A", "B"), WithStepCounter())
		if s == nil {
			t.Fatal("expected summary, got nil")
		}
		if s.Message != "1 of 2 completed" {
			t.Errorf("expected '1 of 2 completed', got '%s'", s.Message)
		}
	})
}
