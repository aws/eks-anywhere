package cluster_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

const (
	owner             = "janedoe"
	repository        = "flux-fleet"
	fluxNamespace     = "test-ns"
	branch            = "test-branch"
	clusterConfigPath = "test-path"
	personal          = false
)

func TestGitOpsToFluxConversionProcessing(t *testing.T) {
	tests := []struct {
		name           string
		wantConfigPath string
		wantFluxSpec   anywherev1.FluxConfigSpec
	}{
		{
			name:           "workload cluster with GitOpsConfig",
			wantConfigPath: "testdata/cluster_gitops_1_21.yaml",
			wantFluxSpec: anywherev1.FluxConfigSpec{
				SystemNamespace:   fluxNamespace,
				ClusterConfigPath: clusterConfigPath,
				Branch:            branch,
				Github: &anywherev1.GithubProviderConfig{
					Owner:      owner,
					Repository: repository,
					Personal:   personal,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			config, err := cluster.ParseConfigFromFile(tt.wantConfigPath)
			if err != nil {
				t.Fatal("cluster.ParseConfigFromFile error != nil, want nil", err)
			}
			g.Expect(config.FluxConfig.Spec).To(Equal(tt.wantFluxSpec))
		})
	}
}

func TestConfigManagerValidateGitOpsConfig(t *testing.T) {
	tests := []struct {
		testName string
		config   *cluster.Config
		wantErr  bool
	}{
		{
			testName: "valid gitopsconfig",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.ClusterKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eksa-unit-test",
						Namespace: "default",
					},
					Spec: anywherev1.ClusterSpec{
						GitOpsRef: &anywherev1.Ref{
							Name: "test1", Kind: "GitOpsConfig",
						},
					},
				},
				GitOpsConfig: &anywherev1.GitOpsConfig{
					TypeMeta: metav1.TypeMeta{
						Kind:       "GitOpsConfig",
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-gitops",
						Namespace: "default",
					},
					Spec: anywherev1.GitOpsConfigSpec{
						Flux: anywherev1.Flux{
							Github: anywherev1.Github{
								Owner:      "janedoe",
								Repository: "flux-fleet",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "different namespace",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.ClusterKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eksa-unit-test",
						Namespace: "default",
					},
					Spec: anywherev1.ClusterSpec{
						GitOpsRef: &anywherev1.Ref{
							Name: "test1", Kind: "GitOpsConfig",
						},
					},
				},
				GitOpsConfig: &anywherev1.GitOpsConfig{
					TypeMeta: metav1.TypeMeta{
						Kind:       "GitOpsConfig",
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-gitops",
						Namespace: "not-default",
					},
					Spec: anywherev1.GitOpsConfigSpec{
						Flux: anywherev1.Flux{
							Github: anywherev1.Github{
								Owner:      "janedoe",
								Repository: "flux-fleet",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			testName: "no gitops config",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					TypeMeta: metav1.TypeMeta{
						Kind:       anywherev1.ClusterKind,
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "eksa-unit-test",
					},
					Spec: anywherev1.ClusterSpec{
						GitOpsRef: &anywherev1.Ref{
							Name: "test1", Kind: "GitOpsConfig",
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			c := cluster.NewConfigManager()
			c.RegisterValidations(
				func(c *cluster.Config) error {
					if c.GitOpsConfig != nil {
						if c.Cluster.Namespace != c.GitOpsConfig.Namespace {
							return fmt.Errorf("%s and Cluster objects must have the same namespace specified", anywherev1.GitOpsConfigKind)
						}
					}
					return nil
				},
				func(c *cluster.Config) error {
					if c.GitOpsConfig == nil && c.Cluster.Spec.GitOpsRef != nil && c.Cluster.Spec.GitOpsRef.Kind == anywherev1.GitOpsConfigKind {
						return fmt.Errorf("%s/%s referenced in Cluster but not present in the cluster config", anywherev1.GitOpsConfigKind, c.Cluster.Spec.GitOpsRef.Name)
					}
					return nil
				},
			)

			err := c.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
