package types_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/types"
)

func TestHasAnyLabel(t *testing.T) {
	tests := []struct {
		testName    string
		labels      map[string]string
		wantLabels  []string
		hasAnyLabel bool
	}{
		{
			testName:    "empty labels",
			labels:      map[string]string{},
			wantLabels:  []string{"label_1"},
			hasAnyLabel: false,
		},
		{
			testName:    "nil labels and want labels",
			labels:      nil,
			wantLabels:  nil,
			hasAnyLabel: false,
		},
		{
			testName: "empty want labels",
			labels: map[string]string{
				"label_1": "val_1",
			},
			wantLabels:  []string{},
			hasAnyLabel: false,
		},
		{
			testName: "nil want labels",
			labels: map[string]string{
				"label_1": "val_1",
			},
			wantLabels:  nil,
			hasAnyLabel: false,
		},
		{
			testName: "labels present",
			labels: map[string]string{
				"label_1": "val_1",
				"label_2": "val_2",
			},
			wantLabels:  []string{"label_1"},
			hasAnyLabel: true,
		},
		{
			testName: "any label present",
			labels: map[string]string{
				"label_1": "val_1",
				"label_2": "val_2",
			},
			wantLabels:  []string{"label_1", "label_3"},
			hasAnyLabel: true,
		},
		{
			testName: "labels not present",
			labels: map[string]string{
				"label_1": "val_1",
				"label_2": "val_2",
			},
			wantLabels:  []string{"label_3"},
			hasAnyLabel: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			m := &types.Machine{}
			m.Metadata.Labels = tt.labels
			if got := m.HasAnyLabel(tt.wantLabels); got != tt.hasAnyLabel {
				t.Errorf("machine.HasAnyLabel() = %v, want %v", got, tt.hasAnyLabel)
			}
		})
	}
}

func TestWithClusterReady(t *testing.T) {
	tt := NewWithT(t)
	tests := []struct {
		testName string
		status   types.ClusterStatus
		expected bool
	}{
		{
			testName: "no conditions",
			status:   types.ClusterStatus{},
			expected: false,
		},
		{
			testName: "no Ready type",
			status: types.ClusterStatus{Conditions: []types.Condition{{
				Type:   "Runnung",
				Status: "True",
			}}},
			expected: false,
		},
		{
			testName: "Ready is False",
			status: types.ClusterStatus{Conditions: []types.Condition{{
				Type:   "Ready",
				Status: "False",
			}}},
			expected: false,
		},
		{
			testName: "Ready is True",
			status: types.ClusterStatus{Conditions: []types.Condition{{
				Type:   "Ready",
				Status: "True",
			}}},
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			checker := types.WithClusterReady()
			tt.Expect(checker(test.status)).To(Equal(test.expected))
		})
	}
}
