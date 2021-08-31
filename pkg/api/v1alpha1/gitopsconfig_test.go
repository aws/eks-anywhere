package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAndValidateGitOpsConfig(t *testing.T) {
	tests := []struct {
		testName         string
		fileName         string
		refName          string
		wantGitOpsConfig *GitOpsConfig
		wantErr          bool
	}{
		{
			testName:         "file doesn't exist",
			fileName:         "testdata/fake_file.yaml",
			wantGitOpsConfig: nil,
			wantErr:          true,
		},
		{
			testName:         "not parseable file",
			fileName:         "testdata/not_parseable_gitopsconfig.yaml",
			wantGitOpsConfig: nil,
			wantErr:          true,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19_gitops.yaml",
			refName:  "test-gitops",
			wantGitOpsConfig: &GitOpsConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "GitOpsConfig",
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gitops",
					Namespace: "default",
				},
				Spec: GitOpsConfigSpec{
					Flux: Flux{
						Github: Github{
							Owner:      "janedoe",
							Repository: "flux-fleet",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "refName doesn't match",
			fileName: "testdata/cluster_1_19_gitops.yaml",
			refName:  "wrongName",
			wantErr:  true,
		},
		{
			testName:         "empty owner",
			fileName:         "testdata/cluster_invalid_gitops_unset_gitowner.yaml",
			wantGitOpsConfig: nil,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetAndValidateGitOpsConfig(tt.fileName, tt.refName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAndValidateGitOpsConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantGitOpsConfig) {
				t.Fatalf("GetAndValidateGitOpsConfig() = %#v, want %#v", got, tt.wantGitOpsConfig)
			}
		})
	}
}
