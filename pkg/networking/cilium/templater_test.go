package cilium_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	helmmocks "github.com/aws/eks-anywhere/pkg/helm/mocks"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/semver"
)

type templaterTest struct {
	*WithT
	ctx                     context.Context
	t                       *cilium.Templater
	hf                      *mocks.MockHelmClientFactory
	h                       *helmmocks.MockClient
	manifest                []byte
	uri, version, namespace string
	spec, currentSpec       *cluster.Spec
}

func newtemplaterTest(t *testing.T) *templaterTest {
	ctrl := gomock.NewController(t)
	hf := mocks.NewMockHelmClientFactory(ctrl)
	h := helmmocks.NewMockClient(ctrl)
	return &templaterTest{
		WithT:     NewWithT(t),
		ctx:       context.Background(),
		hf:        hf,
		h:         h,
		t:         cilium.NewTemplater(hf),
		manifest:  []byte("manifestContent"),
		uri:       "oci://public.ecr.aws/isovalent/cilium",
		version:   "1.9.11-eksa.1",
		namespace: "kube-system",
		currentSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Spec.KubernetesVersion = "1.22"
			s.VersionsBundles["1.22"] = test.VersionBundle()
			s.VersionsBundles["1.22"].Cilium.Version = "v1.9.10-eksa.1"
			s.VersionsBundles["1.22"].Cilium.Cilium.URI = "public.ecr.aws/isovalent/cilium:v1.9.10-eksa.1"
			s.VersionsBundles["1.22"].Cilium.Operator.URI = "public.ecr.aws/isovalent/operator-generic:v1.9.10-eksa.1"
			s.VersionsBundles["1.22"].Cilium.HelmChart.URI = "public.ecr.aws/isovalent/cilium:1.9.10-eksa.1"
			s.VersionsBundles["1.22"].KubeDistro.Kubernetes.Tag = "v1.22.5-eks-1-22-9"
			s.Cluster.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{}}
		}),
		spec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Spec.KubernetesVersion = "1.22"
			s.VersionsBundles["1.22"] = test.VersionBundle()
			s.VersionsBundles["1.22"].Cilium.Version = "v1.9.11-eksa.1"
			s.VersionsBundles["1.22"].Cilium.Cilium.URI = "public.ecr.aws/isovalent/cilium:v1.9.11-eksa.1"
			s.VersionsBundles["1.22"].Cilium.Operator.URI = "public.ecr.aws/isovalent/operator-generic:v1.9.11-eksa.1"
			s.VersionsBundles["1.22"].Cilium.HelmChart.URI = "public.ecr.aws/isovalent/cilium:1.9.11-eksa.1"
			s.VersionsBundles["1.22"].KubeDistro.Kubernetes.Tag = "v1.22.5-eks-1-22-9"
			s.Cluster.Spec.ClusterNetwork.CNIConfig = &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{}}
		}),
	}
}

func (t *templaterTest) expectHelmTemplateWith(wantValues gomock.Matcher, kubeVersion string) *gomock.Call {
	return t.h.EXPECT().Template(t.ctx, t.uri, t.version, t.namespace, wantValues, kubeVersion)
}

func (t *templaterTest) expectHelmClientFactoryGet(username, password string) {
	t.hf.EXPECT().Get(t.ctx, t.spec.Cluster).Return(t.h, nil)
}

func eqMap(m map[string]interface{}) gomock.Matcher {
	return &mapMatcher{m: m}
}

// mapMatcher implements a gomock matcher for maps
// transforms any map or struct into a map[string]interface{} and uses DeepEqual to compare.
type mapMatcher struct {
	m map[string]interface{}
}

func (e *mapMatcher) Matches(x interface{}) bool {
	xJson, err := json.Marshal(x)
	if err != nil {
		return false
	}
	xMap := &map[string]interface{}{}
	err = json.Unmarshal(xJson, xMap)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(e.m, *xMap)
}

func (e *mapMatcher) String() string {
	return fmt.Sprintf("matches map %v", e.m)
}

func TestTemplaterGenerateUpgradePreflightManifestSuccess(t *testing.T) {
	t.Skip("Temporarily skipping, need to modify mapMatcher")
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"enabled": false,
		},
		"preflight": map[string]interface{}{
			"enabled": true,
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/cilium",
				"tag":        "v1.9.11-eksa.1",
			},
			"tolerations": []map[string]string{
				{
					"operator": "Exists",
				},
			},
		},
		"agent": false,
	}

	tt := newtemplaterTest(t)

	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(eqMap(wantValues), "1.22").Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateUpgradePreflightManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateUpgradePreflightManifest() should return right manifest")
}

func TestTemplaterGenerateUpgradePreflightManifestError(t *testing.T) {
	tt := newtemplaterTest(t)

	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(gomock.Any(), "1.22").Return(nil, errors.New("error from helm")) // Using any because we only want to test the returned error

	_, err := tt.t.GenerateUpgradePreflightManifest(tt.ctx, tt.spec)
	tt.Expect(err).To(HaveOccurred(), "templater.GenerateUpgradePreflightManifest() should fail")
	tt.Expect(err).To(MatchError(ContainSubstring("error from helm")))
}

func TestTemplaterGenerateUpgradePreflightManifestInvalidKubeVersion(t *testing.T) {
	tt := newtemplaterTest(t)
	tt.spec.VersionsBundles["1.22"].KubeDistro.Kubernetes.Tag = "v1-invalid"
	_, err := tt.t.GenerateUpgradePreflightManifest(tt.ctx, tt.spec)
	tt.Expect(err).To(HaveOccurred(), "templater.GenerateUpgradePreflightManifest() should fail")
	tt.Expect(err).To(MatchError(ContainSubstring("invalid major version in semver")))
}

func TestTemplaterGenerateManifestSuccess(t *testing.T) {
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	tt := newtemplaterTest(t)

	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(eqMap(wantValues), "1.22").Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateManifest() should return right manifest")
}

func TestTemplaterGenerateManifestPolicyEnforcementModeSuccess(t *testing.T) {
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
		"policyEnforcementMode": "always",
	}

	tt := newtemplaterTest(t)
	tt.spec.Cluster.Spec.ManagementCluster.Name = "managed"
	tt.spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode = v1alpha1.CiliumPolicyModeAlways

	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(eqMap(wantValues), "1.22").Return(tt.manifest, nil)

	gotManifest, err := tt.t.GenerateManifest(tt.ctx, tt.spec)
	tt.Expect(err).NotTo(HaveOccurred())
	test.AssertContentToFile(t, string(gotManifest), "testdata/manifest_network_policy.yaml")
}

func TestTemplaterGenerateManifestEgressMasqueradeInterfacesSuccess(t *testing.T) {
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
		"egressMasqueradeInterfaces": "eth0",
	}

	tt := newtemplaterTest(t)
	tt.spec.Cluster.Spec.ManagementCluster.Name = "managed"
	tt.spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.EgressMasqueradeInterfaces = "eth0"

	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(eqMap(wantValues), "1.22").Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateManifest() should return right manifest")
}

func TestTemplaterGenerateManifestDirectRouteModeSuccess(t *testing.T) {
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods":    true,
		"tunnel":               "disabled",
		"autoDirectNodeRoutes": "true",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	tt := newtemplaterTest(t)
	tt.spec.Cluster.Spec.ManagementCluster.Name = "managed"
	tt.spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.RoutingMode = v1alpha1.CiliumRoutingModeDirect
	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(eqMap(wantValues), "1.22").Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateManifest() should return right manifest")
}

func TestTemplaterGenerateManifestDirectModeManualIPCIDRSuccess(t *testing.T) {
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods":     true,
		"tunnel":                "disabled",
		"ipv4NativeRoutingCIDR": "192.168.0.0/24",
		"ipv6NativeRoutingCIDR": "2001:db8::/32",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	tt := newtemplaterTest(t)
	tt.spec.Cluster.Spec.ManagementCluster.Name = "managed"
	tt.spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.RoutingMode = v1alpha1.CiliumRoutingModeDirect
	tt.spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.IPv4NativeRoutingCIDR = "192.168.0.0/24"
	tt.spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.IPv6NativeRoutingCIDR = "2001:db8::/32"
	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(eqMap(wantValues), "1.22").Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateManifest() should return right manifest")
}

func TestTemplaterGenerateManifestError(t *testing.T) {
	expectedAttempts := 2
	tt := newtemplaterTest(t)

	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(gomock.Any(), "1.22").Return(nil, errors.New("error from helm")).Times(expectedAttempts) // Using any because we only want to test the returned error

	_, err := tt.t.GenerateManifest(tt.ctx, tt.spec, cilium.WithRetrier(retrier.NewWithMaxRetries(expectedAttempts, 0)))
	tt.Expect(err).To(HaveOccurred(), "templater.GenerateManifest() should fail")
	tt.Expect(err).To(MatchError(ContainSubstring("error from helm")))
}

func TestTemplaterGenerateManifestGetError(t *testing.T) {
	tt := newtemplaterTest(t)

	tt.hf.EXPECT().Get(tt.ctx, tt.spec.Cluster).Return(nil, errors.New("error getting helm client"))

	_, err := tt.t.GenerateManifest(tt.ctx, tt.spec)
	tt.Expect(err).To(HaveOccurred(), "templater.GenerateManifest() should fail")
	tt.Expect(err).To(MatchError(ContainSubstring("error getting helm client")))
}

func TestTemplaterGenerateManifestInvalidKubeVersion(t *testing.T) {
	tt := newtemplaterTest(t)
	tt.spec.VersionsBundles["1.22"].KubeDistro.Kubernetes.Tag = "v1-invalid"
	_, err := tt.t.GenerateManifest(tt.ctx, tt.spec)
	tt.Expect(err).To(HaveOccurred(), "templater.GenerateManifest() should fail")
	tt.Expect(err).To(MatchError(ContainSubstring("invalid major version in semver")))
}

func TestTemplaterGenerateManifestUpgradeSameKubernetesVersionSuccess(t *testing.T) {
	tt := newtemplaterTest(t)

	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(eqMap(wantUpgradeValues()), "1.22").Return(tt.manifest, nil)

	vb := tt.currentSpec.RootVersionsBundle()

	oldCiliumVersion, err := semver.New(vb.Cilium.Version)
	tt.Expect(err).NotTo(HaveOccurred())

	tt.Expect(
		tt.t.GenerateManifest(tt.ctx, tt.spec,
			cilium.WithUpgradeFromVersion(*oldCiliumVersion),
		),
	).To(Equal(tt.manifest), "templater.GenerateUpgradeManifest() should return right manifest")
}

func TestTemplaterGenerateManifestUpgradeNewKubernetesVersionSuccess(t *testing.T) {
	tt := newtemplaterTest(t)

	tt.expectHelmClientFactoryGet("", "")
	tt.expectHelmTemplateWith(eqMap(wantUpgradeValues()), "1.21").Return(tt.manifest, nil)

	vb := tt.currentSpec.RootVersionsBundle()
	oldCiliumVersion, err := semver.New(vb.Cilium.Version)
	tt.Expect(err).NotTo(HaveOccurred())

	tt.Expect(
		tt.t.GenerateManifest(tt.ctx, tt.spec,
			cilium.WithKubeVersion("1.21"),
			cilium.WithUpgradeFromVersion(*oldCiliumVersion),
		),
	).To(Equal(tt.manifest), "templater.GenerateUpgradeManifest() should return right manifest")
}

func wantUpgradeValues() map[string]interface{} {
	return map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
		"upgradeCompatibility": "1.9",
	}
}

func TestTemplaterGenerateNetworkPolicy(t *testing.T) {
	tests := []struct {
		name                    string
		k8sVersion              string
		selfManaged             bool
		gitopsEnabled           bool
		infraProviderNamespaces []string
		wantNetworkPolicyFile   string
	}{
		{
			name:                    "CAPV mgmt cluster",
			k8sVersion:              "v1.21.9-eks-1-21-10",
			selfManaged:             true,
			gitopsEnabled:           false,
			infraProviderNamespaces: []string{"capv-system"},
			wantNetworkPolicyFile:   "testdata/network_policy_mgmt_capv.yaml",
		},
		{
			name:                    "CAPT mgmt cluster with flux",
			k8sVersion:              "v1.21.9-eks-1-21-10",
			selfManaged:             true,
			gitopsEnabled:           true,
			infraProviderNamespaces: []string{"capt-system"},
			wantNetworkPolicyFile:   "testdata/network_policy_mgmt_capt_flux.yaml",
		},
		{
			name:                    "workload cluster 1.20",
			k8sVersion:              "v1.20.9-eks-1-20-10",
			selfManaged:             false,
			gitopsEnabled:           false,
			infraProviderNamespaces: []string{},
			wantNetworkPolicyFile:   "testdata/network_policy_workload_120.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			temp := newtemplaterTest(t)
			if !tt.selfManaged {
				temp.spec.Cluster.Spec.ManagementCluster.Name = "managed"
			}
			if tt.gitopsEnabled {
				temp.spec.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
					Kind: v1alpha1.FluxConfigKind,
					Name: "eksa-unit-test",
				}
				temp.spec.Config.GitOpsConfig = &v1alpha1.GitOpsConfig{
					Spec: v1alpha1.GitOpsConfigSpec{
						Flux: v1alpha1.Flux{Github: v1alpha1.Github{FluxSystemNamespace: "flux-system"}},
					},
				}
			}
			networkPolicy, err := temp.t.GenerateNetworkPolicyManifest(temp.spec, tt.infraProviderNamespaces)
			if err != nil {
				t.Fatalf("failed to generate network policy template: %v", err)
			}
			test.AssertContentToFile(t, string(networkPolicy), tt.wantNetworkPolicyFile)
		})
	}
}

func TestTemplaterGenerateManifestForSingleNodeCluster(t *testing.T) {
	tt := newtemplaterTest(t)
	tt.spec.Cluster.Spec.WorkerNodeGroupConfigurations = nil
	tt.spec.Cluster.Spec.ControlPlaneConfiguration.Count = 1

	tt.expectHelmClientFactoryGet("", "")
	tt.h.EXPECT().
		Template(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_, _, _, _ interface{}, values map[string]interface{}, _ interface{}) ([]byte, error) {
			tt.Expect(reflect.ValueOf(values["operator"]).MapIndex(reflect.ValueOf("replicas")).Interface().(int)).To(Equal(1))
			return tt.manifest, nil
		})

	tt.Expect(tt.t.GenerateManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateManifest() should return right manifest")
}

func TestTemplaterGenerateManifestForRegistryAuth(t *testing.T) {
	tt := newtemplaterTest(t)
	tt.spec.Cluster.Spec.RegistryMirrorConfiguration = &v1alpha1.RegistryMirrorConfiguration{
		Endpoint:     "1.2.3.4",
		Port:         "443",
		Authenticate: true,
		OCINamespaces: []v1alpha1.OCINamespace{
			{
				Registry:  "public.ecr.aws",
				Namespace: "eks-anywhere",
			},
			{
				Registry:  "783794618700.dkr.ecr.us-west-2.amazonaws.com",
				Namespace: "curated-packages",
			},
		},
	}

	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")

	tt.expectHelmClientFactoryGet("username", "password")

	tt.h.EXPECT().
		Template(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateManifest() should return right manifest")
}
