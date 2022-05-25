package cluster_test

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestConfigManagerValidateFluxConfig(t *testing.T) {
	tests := []struct {
		testName string
		config   *cluster.Config
		wantErr  bool
	}{
		{
			testName: "valid flux config",
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
							Name: "test1", Kind: "FluxConfig",
						},
					},
				},
				FluxConfig: &anywherev1.FluxConfig{
					TypeMeta: metav1.TypeMeta{
						Kind:       "FluxConfig",
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-gitops",
						Namespace: "default",
					},
					Spec: anywherev1.FluxConfigSpec{
						Git: &anywherev1.GitProviderConfig{
							RepositoryUrl: "test",
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
						Namespace: "not-default",
					},
					Spec: anywherev1.ClusterSpec{
						GitOpsRef: &anywherev1.Ref{
							Name: "test1", Kind: "FluxConfig",
						},
					},
				},
				FluxConfig: &anywherev1.FluxConfig{
					TypeMeta: metav1.TypeMeta{
						Kind:       "FluxConfig",
						APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-gitops",
						Namespace: "default",
					},
					Spec: anywherev1.FluxConfigSpec{
						Git: &anywherev1.GitProviderConfig{
							RepositoryUrl: "test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			testName: "no flux config",
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
							Name: "test1", Kind: "FluxConfig",
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
					if c.FluxConfig != nil {
						if c.Cluster.Namespace != c.FluxConfig.Namespace {
							return fmt.Errorf("%s and Cluster objects must have the same namespace specified", anywherev1.FluxConfigKind)
						}
					}
					return nil
				},
				func(c *cluster.Config) error {
					if c.FluxConfig == nil && c.Cluster.Spec.GitOpsRef != nil && c.Cluster.Spec.GitOpsRef.Kind == anywherev1.FluxConfigKind {
						return fmt.Errorf("%s/%s referenced in Cluster but not present in the cluster config", anywherev1.FluxConfigKind, c.Cluster.Spec.GitOpsRef.Name)
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
