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
		clusterConfig    *Cluster
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
			clusterConfig: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
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
			got, err := GetAndValidateGitOpsConfig(tt.fileName, tt.refName, tt.clusterConfig)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAndValidateGitOpsConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantGitOpsConfig) {
				t.Fatalf("GetAndValidateGitOpsConfig() = %#v, want %#v", got, tt.wantGitOpsConfig)
			}
		})
	}
}

func TestConvertGitOpsConfigToFluxConfig(t *testing.T) {
	tests := []struct {
		testName          string
		givenGitOpsConfig *GitOpsConfig
		wantFluxConfig    *FluxConfig
		clusterConfig     *Cluster
	}{
		{
			testName: "Convert GitOps Config to FluxConfig",
			givenGitOpsConfig: &GitOpsConfig{
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
							Owner:               "janedoe",
							Repository:          "flux-fleet",
							FluxSystemNamespace: "flux-system-test",
							Branch:              "test-branch",
							Personal:            false,
							ClusterConfigPath:   "test-config-path",
						},
					},
				},
			},
			wantFluxConfig: &FluxConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       FluxConfigKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-gitops",
					Namespace: "default",
				},
				Spec: FluxConfigSpec{
					SystemNamespace:   "flux-system-test",
					ClusterConfigPath: "test-config-path",
					Branch:            "test-branch",
					Github: &GithubProviderConfig{
						Owner:      "janedoe",
						Repository: "flux-fleet",
						Personal:   false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			convertedGitOps := tt.givenGitOpsConfig.ConvertToFluxConfig()
			if !reflect.DeepEqual(convertedGitOps, tt.wantFluxConfig) {
				t.Fatalf("ConvertToFluxConfig() = %#v, want %#v", convertedGitOps, tt.wantFluxConfig)
			}
		})
	}
}
