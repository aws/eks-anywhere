package cilium_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	v1alpha12 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/mocks"
	mocksprovider "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ciliumtest struct {
	*WithT
	ctx          context.Context
	client       mocks.MockClient
	h            *mocks.MockHelm
	spec         *cluster.Spec
	cilium       *cilium.Cilium
	provider     *mocksprovider.MockProvider
	ciliumValues []byte
}

func newCiliumTest(t *testing.T) *ciliumtest {
	ctrl := gomock.NewController(t)
	h := mocks.NewMockHelm(ctrl)
	client := mocks.NewMockClient(ctrl)
	return &ciliumtest{
		WithT:  NewWithT(t),
		ctx:    context.Background(),
		client: *client,
		h:      h,
		spec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.VersionsBundle.Cilium = v1alpha1.CiliumBundle{
				Cilium: v1alpha1.Image{
					URI: "public.ecr.aws/isovalent/cilium:v1.9.13-eksa.2",
				},
				Operator: v1alpha1.Image{
					URI: "public.ecr.aws/isovalent/operator-generic:v1.9.13-eksa.2",
				},
				HelmChart: v1alpha1.Image{
					Name: "cilium-chart",
					URI:  "public.ecr.aws/isovalent/cilium:1.9.13-eksa.2",
				},
			}
			s.VersionsBundle.KubeDistro.Kubernetes.Tag = "v1.21.9-eks-1-21-10"
			s.Cluster.Spec.ClusterNetwork.CNIConfig = &v1alpha12.CNIConfig{Cilium: &v1alpha12.CiliumConfig{}}
		}),
		cilium:       cilium.NewCilium(client, h),
		provider:     mocksprovider.NewMockProvider(ctrl),
		ciliumValues: []byte("manifest"),
	}
}

func TestCiliumGenerateManifestSuccess(t *testing.T) {
	tt := newCiliumTest(t)
	// templater tests already test whether templater.GenerateManifest returns expected values or not. This test ensures that cilium.GenerateManifest
	// calls the templater and does not try to load the static manifest like earlier version
	tt.h.EXPECT().Template(
		tt.ctx, gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(map[string]interface{}{}),
	).Return(tt.ciliumValues, nil)
	_, err := tt.cilium.GenerateManifest(tt.ctx, tt.spec, nil)
	tt.Expect(err).To(Not(HaveOccurred()), "GenerateManifest() should succeed")
}

func TestCiliumGenerateManifestNetworkPolicyMgmtVSphereProvider(t *testing.T) {
	tt := newCiliumTest(t)
	tt.h.EXPECT().Template(
		tt.ctx, gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(map[string]interface{}{}),
	).Return(tt.ciliumValues, nil)

	alwaysModeSpec := tt.spec
	alwaysModeSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode = "always"

	providerDeployments := map[string][]string{
		"capv-system": {"capv-controller-manager"},
	}
	tt.provider.EXPECT().GetDeployments().Return(providerDeployments)

	ciliumManifest, err := tt.cilium.GenerateManifest(tt.ctx, alwaysModeSpec, tt.provider)
	test.AssertContentToFile(t, string(ciliumManifest), "testdata/network_policy_mgmt_capv.yaml")

	tt.Expect(err).To(Not(HaveOccurred()), "GenerateManifest() should succeed")
}

func TestCiliumGenerateManifestNetworkPolicyMgmtTinkerBellProviderFluxEnabled(t *testing.T) {
	tt := newCiliumTest(t)
	tt.h.EXPECT().Template(
		tt.ctx, gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(map[string]interface{}{}),
	).Return(tt.ciliumValues, nil)

	alwaysModeSpec := tt.spec
	alwaysModeSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode = "always"
	alwaysModeSpec.Cluster.Spec.GitOpsRef = &v1alpha12.Ref{
		Kind: v1alpha12.FluxConfigKind,
		Name: "eksa-unit-test",
	}
	providerDeployments := map[string][]string{
		"capt-system": {"capt-controller-manager"},
	}
	tt.provider.EXPECT().GetDeployments().Return(providerDeployments)

	ciliumManifest, err := tt.cilium.GenerateManifest(tt.ctx, alwaysModeSpec, tt.provider)
	test.AssertContentToFile(t, string(ciliumManifest), "testdata/network_policy_mgmt_capt_flux.yaml")

	tt.Expect(err).To(Not(HaveOccurred()), "GenerateManifest() should succeed")
}

func TestCiliumGenerateManifestNetworkPolicyWorkload121(t *testing.T) {
	tt := newCiliumTest(t)
	tt.h.EXPECT().Template(
		tt.ctx, gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(map[string]interface{}{}),
	).Return(tt.ciliumValues, nil)

	alwaysModeSpec := tt.spec
	alwaysModeSpec.Cluster.Spec.ManagementCluster.Name = "mgmt"
	alwaysModeSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode = "always"
	tt.provider.EXPECT().GetDeployments().Return(nil)

	ciliumManifest, err := tt.cilium.GenerateManifest(tt.ctx, alwaysModeSpec, tt.provider)
	test.AssertContentToFile(t, string(ciliumManifest), "testdata/network_policy_workload_121.yaml")

	tt.Expect(err).To(Not(HaveOccurred()), "GenerateManifest() should succeed")
}

func TestCiliumGenerateManifestNetworkPolicyWorkload120(t *testing.T) {
	tt := newCiliumTest(t)
	tt.h.EXPECT().Template(
		tt.ctx, gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf(map[string]interface{}{}),
	).Return(tt.ciliumValues, nil)

	alwaysModeSpec := tt.spec
	alwaysModeSpec.Cluster.Spec.ManagementCluster.Name = "mgmt"
	alwaysModeSpec.VersionsBundle.KubeDistro.Kubernetes.Tag = "v1.20.9-eks-1-20-10"
	alwaysModeSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode = "always"
	tt.provider.EXPECT().GetDeployments().Return(nil)

	ciliumManifest, err := tt.cilium.GenerateManifest(tt.ctx, alwaysModeSpec, tt.provider)
	test.AssertContentToFile(t, string(ciliumManifest), "testdata/network_policy_workload_120.yaml")

	tt.Expect(err).To(Not(HaveOccurred()), "GenerateManifest() should succeed")
}
