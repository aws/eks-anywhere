package v1alpha1

import "testing"

func TestEqual(t *testing.T) {
	tests := []struct {
		testName  string
		aiOld     *AWSIamConfig
		aiNew     *AWSIamConfig
		wantEqual bool
	}{
		{
			testName: "region changed",
			aiOld: &AWSIamConfig{
				Spec: AWSIamConfigSpec{
					AWSRegion: "oldRegion",
				},
			},
			aiNew: &AWSIamConfig{
				Spec: AWSIamConfigSpec{
					AWSRegion: "newRegion",
				},
			},
			wantEqual: false,
		},
		{
			testName: "partition changed",
			aiOld: &AWSIamConfig{
				Spec: AWSIamConfigSpec{
					Partition: "oldPartition",
				},
			},
			aiNew: &AWSIamConfig{
				Spec: AWSIamConfigSpec{
					Partition: "newPartition",
				},
			},
			wantEqual: false,
		},
		{
			testName: "backendMode changed",
			aiOld: &AWSIamConfig{
				Spec: AWSIamConfigSpec{
					BackendMode: []string{"mode1", "mode2"},
				},
			},
			aiNew: &AWSIamConfig{
				Spec: AWSIamConfigSpec{
					BackendMode: []string{"mode1"},
				},
			},
			wantEqual: false,
		},
		{
			testName: "equal success",
			aiOld: &AWSIamConfig{
				Spec: AWSIamConfigSpec{},
			},
			aiNew: &AWSIamConfig{
				Spec: AWSIamConfigSpec{},
			},
			wantEqual: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if ok := tt.aiNew.Spec.Equal(&tt.aiOld.Spec); ok != tt.wantEqual {
				t.Fatalf("Equal() gotEqual = %t\nwantEqual %t", ok, tt.wantEqual)
			}
		})
	}
}
