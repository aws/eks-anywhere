package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
