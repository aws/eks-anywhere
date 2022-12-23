package types_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/types"
)

func TestAppend(t *testing.T) {
	tests := []struct {
		testName         string
		changeDiffs      *types.ChangeDiff
		componentReports []types.ComponentChangeDiff
	}{
		{
			testName: "empty changeDiff",
			componentReports: []types.ComponentChangeDiff{
				{
					ComponentName: "test",
					OldVersion:    "0.0.1",
					NewVersion:    "0.0.2",
				},
			},
			changeDiffs: &types.ChangeDiff{},
		},
		{
			testName: "non empty changeDiff",
			componentReports: []types.ComponentChangeDiff{
				{
					ComponentName: "test2",
					OldVersion:    "0.0.2",
					NewVersion:    "0.0.3",
				},
			},
			changeDiffs: &types.ChangeDiff{
				[]types.ComponentChangeDiff{
					{
						ComponentName: "test",
						OldVersion:    "0.0.1",
						NewVersion:    "0.0.2",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			changeDiffs := &types.ChangeDiff{tt.componentReports}
			prevLen := len(tt.changeDiffs.ComponentReports)
			tt.changeDiffs.Append(changeDiffs)

			if len(tt.changeDiffs.ComponentReports) != (len(tt.componentReports))+prevLen {
				t.Errorf("Component Reports were not appended")
			}
		})
	}
}
