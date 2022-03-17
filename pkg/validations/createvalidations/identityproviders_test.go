package createvalidations_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
)

var oidcResourceType = fmt.Sprintf("oidcconfigs.%s", v1alpha1.GroupVersion.Group)

func TestValidateIdendityProviderForWorkloadClusters(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "SuccessNoIdentityProvider",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_identity_provider_response.json",
		},
		{
			name:               "FailureIdentityProviderNameExists",
			wantErr:            errors.New("the following identityProviders already exists [oidc-config-test]"),
			getClusterResponse: "testdata/identity_provider_name_exists.json",
		},
	}

	defaultOIDC := &v1alpha1.OIDCConfig{
		Spec: v1alpha1.OIDCConfigSpec{
			ClientId:     "client-id",
			GroupsClaim:  "groups-claim",
			GroupsPrefix: "groups-prefix",
			IssuerUrl:    "issuer-url",
			RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{{
				Claim: "claim",
				Value: "value",
			}},
			UsernameClaim:  "username-claim",
			UsernamePrefix: "username-prefix",
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = testclustername
		s.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{
			{
				Kind: v1alpha1.OIDCConfigKind,
				Name: "oidc-config-test",
			},
		}
		s.Cluster.SetManagedBy("management-cluster")
		s.OIDCConfig = defaultOIDC
	})
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(
				ctx, []string{
					"get", oidcResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=oidc-config-test",
				}).Return(*bytes.NewBufferString(fileContent), nil)

			err := createvalidations.ValidateIdentityProviderNameIsUnique(ctx, k, cluster, clusterSpec)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateIdentityProviderForSelfManagedCluster(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "Skip Validate GitOpsConfig name",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_identity_provider_response.json",
		},
	}

	defaultOIDC := &v1alpha1.OIDCConfig{
		Spec: v1alpha1.OIDCConfigSpec{
			ClientId:     "client-id",
			GroupsClaim:  "groups-claim",
			GroupsPrefix: "groups-prefix",
			IssuerUrl:    "issuer-url",
			RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{{
				Claim: "claim",
				Value: "value",
			}},
			UsernameClaim:  "username-claim",
			UsernamePrefix: "username-prefix",
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = testclustername
		s.Cluster.Spec.IdentityProviderRefs = []v1alpha1.Ref{
			{
				Kind: v1alpha1.OIDCConfigKind,
				Name: "oidc-config-test",
			},
		}
		s.OIDCConfig = defaultOIDC

		s.Cluster.SetSelfManaged()
	})
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			e.EXPECT().Execute(
				ctx, []string{
					"get", oidcResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=oidc-config-test",
				}).Times(0)

			err := createvalidations.ValidateIdentityProviderNameIsUnique(ctx, k, cluster, clusterSpec)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
