package cluster_test

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestConfigManagerValidateOIDCConfig(t *testing.T) {
	tests := []struct {
		testName string
		config   *cluster.Config
		wantErr  bool
	}{
		{
			testName: "valid oidc config",
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
						IdentityProviderRefs: []anywherev1.Ref{
							{
								Kind: "OIDCConfig",
								Name: "test",
							},
						},
					},
				},
				OIDCConfigs: map[string]*anywherev1.OIDCConfig{
					"test": {
						TypeMeta: metav1.TypeMeta{
							Kind:       "OIDCConfig",
							APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "eksa-unit-test",
							Namespace: "default",
						},
						Spec: anywherev1.OIDCConfigSpec{
							ClientId:     "id11",
							GroupsClaim:  "claim1",
							GroupsPrefix: "prefix-for-groups",
							IssuerUrl:    "https://mydomain.com/issuer",
							RequiredClaims: []anywherev1.OIDCConfigRequiredClaim{
								{
									Claim: "sub",
									Value: "test",
								},
							},
							UsernameClaim:  "username-claim",
							UsernamePrefix: "username-prefix",
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
						IdentityProviderRefs: []anywherev1.Ref{
							{
								Kind: "OIDCConfig",
								Name: "test",
							},
						},
					},
				},
				OIDCConfigs: map[string]*anywherev1.OIDCConfig{
					"test": {
						TypeMeta: metav1.TypeMeta{
							Kind:       "OIDCConfig",
							APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "eksa-unit-test",
							Namespace: "not-default",
						},
						Spec: anywherev1.OIDCConfigSpec{
							ClientId:     "id11",
							GroupsClaim:  "claim1",
							GroupsPrefix: "prefix-for-groups",
							IssuerUrl:    "https://mydomain.com/issuer",
							RequiredClaims: []anywherev1.OIDCConfigRequiredClaim{
								{
									Claim: "sub",
									Value: "test",
								},
							},
							UsernameClaim:  "username-claim",
							UsernamePrefix: "username-prefix",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			testName: "no oidc config",
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
						IdentityProviderRefs: []anywherev1.Ref{
							{
								Name: "test1", Kind: "OIDCConfig",
							},
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
					for _, a := range c.OIDCConfigs {
						if a.Namespace != c.Cluster.Namespace {
							return fmt.Errorf("%s and Cluster objects must have the same namespace specified", anywherev1.OIDCConfigKind)
						}
					}
					return nil
				},
				func(c *cluster.Config) error {
					for _, idr := range c.Cluster.Spec.IdentityProviderRefs {
						if idr.Kind == anywherev1.OIDCConfigKind && c.OIDCConfigs == nil {
							return fmt.Errorf("%s/%s referenced in Cluster but not present in the cluster config", anywherev1.OIDCConfigKind, idr.Name)
						}
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
