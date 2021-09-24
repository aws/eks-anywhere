package types_test

import (
	"testing"

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
