package tinkerbell_test

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/networkutils/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestAssertMachineConfigsValid_ValidSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	g.Expect(tinkerbell.AssertMachineConfigsValid(clusterSpec)).To(gomega.Succeed())
}

func TestAssertMachineConfigsValid_InvalidFails(t *testing.T) {
	// Invalidate the namespace check.
	for name, mutate := range map[string]func(*tinkerbell.ClusterSpec){
		"MissingName": func(clusterSpec *tinkerbell.ClusterSpec) {
			clusterSpec.ControlPlaneMachineConfig().Name = ""
		},
		"MissingHardwareSelector": func(clusterSpec *tinkerbell.ClusterSpec) {
			clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector = map[string]string{}
		},
		"MultipleKeyValuePairsInHardwareSelector": func(clusterSpec *tinkerbell.ClusterSpec) {
			clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector = map[string]string{
				"foo": "bar",
				"baz": "qux",
			}
		},
		"MissingUsers": func(clusterSpec *tinkerbell.ClusterSpec) {
			clusterSpec.ControlPlaneMachineConfig().Spec.Users = []eksav1alpha1.UserConfiguration{}
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			spec := NewDefaultValidClusterSpecBuilder().Build()
			mutate(spec)
			g.Expect(tinkerbell.AssertMachineConfigsValid(spec)).ToNot(gomega.Succeed())
		})
	}
}

func TestAssertDatacenterConfigValid_ValidSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	g.Expect(tinkerbell.AssertDatacenterConfigValid(clusterSpec)).To(gomega.Succeed())
}

func TestAssertDatacenterConfigValid_InvalidFails(t *testing.T) {
	g := gomega.NewWithT(t)

	for name, mutate := range map[string]func(*tinkerbell.ClusterSpec){
		"NoObjectName": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.ObjectMeta.Name = ""
		},
		"NoTinkerbellIP": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.TinkerbellIP = ""
		},
		"TinkerbellIPInvalid": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.TinkerbellIP = "invalid"
		},
	} {
		t.Run(name, func(t *testing.T) {
			cluster := NewDefaultValidClusterSpecBuilder().Build()
			mutate(cluster)
			g.Expect(tinkerbell.AssertDatacenterConfigValid(cluster)).ToNot(gomega.Succeed())
		})
	}
}

func TestAssertDatacenterConfigValidEmptyOSImageURL(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.DatacenterConfig.Spec.OSImageURL = "test"
	g.Expect(tinkerbell.AssertDatacenterConfigValid(clusterSpec)).To(gomega.MatchError("parsing osImageOverride: parse \"test\": invalid URI for request"))
}

func TestAssertDatacenterConfigValidEmptyHookImagesURLPath(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.DatacenterConfig.Spec.HookImagesURLPath = "test"
	g.Expect(tinkerbell.AssertDatacenterConfigValid(clusterSpec)).To(gomega.MatchError("parsing hookOverride: parse \"test\": invalid URI for request"))
}

func TestAssertMachineConfigNamespaceMatchesDatacenterConfig_Same(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	err := tinkerbell.AssertMachineConfigNamespaceMatchesDatacenterConfig(clusterSpec)
	g.Expect(err).To(gomega.Succeed())
}

func TestAssertMachineConfigNamespaceMatchesDatacenterConfig_Different(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()

	// Invalidate the namespace check.
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Namespace = "foo-bar"

	err := tinkerbell.AssertMachineConfigNamespaceMatchesDatacenterConfig(clusterSpec)
	g.Expect(err).ToNot(gomega.Succeed())
}

func TestAssertMachineConfigOSImageURL_Error(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	clusterSpec.DatacenterConfig.Spec.OSImageURL = "test-url"
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Spec.OSImageURL = "test-url"
	err := tinkerbell.AssertOSImageURL(clusterSpec)
	g.Expect(err).ToNot(gomega.Succeed())
}

func TestAssertMachineConfigOSImageURLNotSpecified_Error(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	clusterSpec.DatacenterConfig.Spec.OSImageURL = ""
	// set OsImageURL at machineConfig level but not for all machine configs
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Spec.OSImageURL = "test-url"
	err := tinkerbell.AssertOSImageURL(clusterSpec)
	g.Expect(err).ToNot(gomega.Succeed())
}

func TestAssertMachineConfigOSImageURLSpecified_Succeed(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	clusterSpec.DatacenterConfig.Spec.OSImageURL = ""
	clusterSpec.Spec.Cluster.Spec.KubernetesVersion = "1.22"
	// set OsImageURL at machineConfig level but not for all machine configs
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Spec.OSImageURL = "test-url-122"
	clusterSpec.MachineConfigs[builder.ExternalEtcdMachineName].Spec.OSImageURL = "test-url-122"
	clusterSpec.MachineConfigs[builder.WorkerNodeGroupMachineName].Spec.OSImageURL = "test-url-122"
	err := tinkerbell.AssertOSImageURL(clusterSpec)
	g.Expect(err).To(gomega.Succeed())
}

func TestK8sVersionInDataCenterOSImageURL_Succeed(t *testing.T) {
	g := gomega.NewWithT(t)
	kube122 := eksav1alpha1.Kube122
	for name, spec := range map[string]func(*tinkerbell.ClusterSpec){
		"validate'.'specifier": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.OSImageURL = "test-url-1.22"
			c.Cluster.Spec.KubernetesVersion = kube122
		},
		"validate'-'specifier": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.OSImageURL = "test-url-1-22"
			c.Cluster.Spec.KubernetesVersion = kube122
		},
		"validate'_'specifier": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.OSImageURL = "test-url-1_22"
			c.Cluster.Spec.KubernetesVersion = kube122
		},
		"validate''specifier": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.OSImageURL = "test-url-122"
			c.Cluster.Spec.KubernetesVersion = kube122
		},
	} {
		t.Run(name, func(t *testing.T) {
			cluster := NewDefaultValidClusterSpecBuilder().Build()
			spec(cluster)
			g.Expect(tinkerbell.AssertOSImageURL(cluster)).To(gomega.Succeed())
		})
	}
}

func TestK8sVersionInDataCenterOSImageURL_Error(t *testing.T) {
	g := gomega.NewWithT(t)
	kube122 := eksav1alpha1.Kube122
	for name, spec := range map[string]func(*tinkerbell.ClusterSpec){
		"noK8sVersion": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.OSImageURL = "test-url"
			c.Cluster.Spec.KubernetesVersion = kube122
		},
		"invalidSpecifierinK8sVersion": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.OSImageURL = "test-url-1/22"
			c.Cluster.Spec.KubernetesVersion = kube122
		},
		"emptyImageURL": func(c *tinkerbell.ClusterSpec) {
			c.DatacenterConfig.Spec.OSImageURL = ""
			c.Cluster.Spec.KubernetesVersion = kube122
		},
	} {
		t.Run(name, func(t *testing.T) {
			cluster := NewDefaultValidClusterSpecBuilder().Build()
			spec(cluster)
			g.Expect(tinkerbell.AssertOSImageURL(cluster)).ToNot(gomega.Succeed())
		})
	}
}

func TestK8sVersionInMachineConfigOSImageURL_Succeed(t *testing.T) {
	g := gomega.NewWithT(t)
	kube122 := eksav1alpha1.Kube122
	kube123 := eksav1alpha1.Kube123
	for name, spec := range map[string]func(*tinkerbell.ClusterSpec){
		"validate'.'specifier": func(c *tinkerbell.ClusterSpec) {
			for _, mcRef := range c.Cluster.MachineConfigRefs() {
				c.MachineConfigs[mcRef.Name].Spec.OSImageURL = "test-url-1.22"
				c.Cluster.Spec.KubernetesVersion = kube122
			}
		},
		"validate'-'specifier": func(c *tinkerbell.ClusterSpec) {
			for _, mcRef := range c.Cluster.MachineConfigRefs() {
				c.MachineConfigs[mcRef.Name].Spec.OSImageURL = "test-url-1-22"
				c.Cluster.Spec.KubernetesVersion = kube122
			}
		},
		"validate'_'specifier": func(c *tinkerbell.ClusterSpec) {
			for _, mcRef := range c.Cluster.MachineConfigRefs() {
				c.MachineConfigs[mcRef.Name].Spec.OSImageURL = "test-url-1-22"
				c.Cluster.Spec.KubernetesVersion = kube122
			}
		},
		"validate''specifier": func(c *tinkerbell.ClusterSpec) {
			for _, mcRef := range c.Cluster.MachineConfigRefs() {
				c.MachineConfigs[mcRef.Name].Spec.OSImageURL = "test-url-122"
				c.Cluster.Spec.KubernetesVersion = kube122
			}
		},
		"validateCPWorkerDiffVersion": func(c *tinkerbell.ClusterSpec) {
			c.Cluster.Spec.KubernetesVersion = kube123
			c.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122
			c.ControlPlaneMachineConfig().Spec.OSImageURL = "test-url-123"
			c.ExternalEtcdMachineConfig().Spec.OSImageURL = "test-url"
			wngRef := c.WorkerNodeGroupConfigurations()[0].MachineGroupRef.Name
			c.MachineConfigs[wngRef].Spec.OSImageURL = "test-url-122"
		},
	} {
		t.Run(name, func(t *testing.T) {
			cluster := NewDefaultValidClusterSpecBuilder().Build()
			cluster.DatacenterConfig.Spec.OSImageURL = ""
			spec(cluster)
			g.Expect(tinkerbell.AssertOSImageURL(cluster)).To(gomega.Succeed())
		})
	}
}

func TestK8sVersionInMachineConfigOSImageURL_Error(t *testing.T) {
	g := gomega.NewWithT(t)
	kube122 := eksav1alpha1.Kube122
	kube123 := eksav1alpha1.Kube123
	for name, spec := range map[string]func(*tinkerbell.ClusterSpec){
		"validateCPVersionError": func(c *tinkerbell.ClusterSpec) {
			c.Cluster.Spec.KubernetesVersion = kube123
			c.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122
			c.ControlPlaneMachineConfig().Spec.OSImageURL = "test-url-122"
			wngRef := c.WorkerNodeGroupConfigurations()[0].MachineGroupRef.Name
			c.MachineConfigs[wngRef].Spec.OSImageURL = "test-url-122"
		},
		"validateWorkerVersionError": func(c *tinkerbell.ClusterSpec) {
			c.Cluster.Spec.KubernetesVersion = kube123
			c.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube123
			c.ControlPlaneMachineConfig().Spec.OSImageURL = "test-url-123"
			wngRef := c.WorkerNodeGroupConfigurations()[0].MachineGroupRef.Name
			c.MachineConfigs[wngRef].Spec.OSImageURL = "test-url-122"
		},
		"validateCPWorkerSameVersionError": func(c *tinkerbell.ClusterSpec) {
			c.Cluster.Spec.KubernetesVersion = kube123
			c.ControlPlaneMachineConfig().Spec.OSImageURL = "test-url-123"
			wngRef := c.WorkerNodeGroupConfigurations()[0].MachineGroupRef.Name
			c.MachineConfigs[wngRef].Spec.OSImageURL = "test-url-122"
		},
	} {
		t.Run(name, func(t *testing.T) {
			cluster := NewDefaultValidClusterSpecBuilder().Build()
			spec(cluster)
			g.Expect(tinkerbell.AssertOSImageURL(cluster)).ToNot(gomega.Succeed())
		})
	}
}

func TestK8sVersionInCPMachineConfigOSImageURL_Error(t *testing.T) {
	g := gomega.NewWithT(t)
	kube122 := eksav1alpha1.Kube122
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	clusterSpec.Spec.Cluster.Spec.KubernetesVersion = kube122
	clusterSpec.DatacenterConfig.Spec.OSImageURL = ""
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Spec.OSImageURL = "test-url"
	clusterSpec.MachineConfigs[builder.ExternalEtcdMachineName].Spec.OSImageURL = "test-url-122"
	clusterSpec.MachineConfigs[builder.WorkerNodeGroupMachineName].Spec.OSImageURL = "test-url-122"
	g.Expect(tinkerbell.AssertOSImageURL(clusterSpec)).To(gomega.MatchError(gomega.ContainSubstring("missing kube version from control plane machine config OSImageURL:")))
}

func TestK8sVersionInWorkerMachineConfigOSImageURL_Error(t *testing.T) {
	g := gomega.NewWithT(t)
	kube122 := eksav1alpha1.Kube122
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	clusterSpec.Spec.Cluster.Spec.KubernetesVersion = kube122
	clusterSpec.DatacenterConfig.Spec.OSImageURL = ""
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Spec.OSImageURL = "test-url-122"
	clusterSpec.MachineConfigs[builder.ExternalEtcdMachineName].Spec.OSImageURL = "test-url-122"
	clusterSpec.MachineConfigs[builder.WorkerNodeGroupMachineName].Spec.OSImageURL = "test-url"
	g.Expect(tinkerbell.AssertOSImageURL(clusterSpec)).To(gomega.MatchError(gomega.ContainSubstring("missing kube version from worker node group machine config OSImageURL:")))
}

func TestK8sVersionForBRAutoImport_Succeed(t *testing.T) {
	g := gomega.NewWithT(t)
	kube123 := eksav1alpha1.Kube123
	kube122 := eksav1alpha1.Kube122
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	clusterSpec.Spec.Cluster.Spec.KubernetesVersion = kube123
	clusterSpec.Spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122
	clusterSpec.DatacenterConfig.Spec.OSImageURL = ""
	for _, mc := range clusterSpec.MachineConfigs {
		mc.Spec.OSFamily = eksav1alpha1.Bottlerocket
	}
	clusterSpec.VersionsBundles = test.VersionsBundlesMap()
	clusterSpec.VersionsBundle(kube122).EksD.Raw.Bottlerocket.URI = "br-122"
	clusterSpec.VersionsBundle(kube123).EksD.Raw.Bottlerocket.URI = "br-123"
	g.Expect(tinkerbell.AssertOSImageURL(clusterSpec)).To(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_Exists(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	g.Expect(tinkerbell.AssertEtcdMachineRefExists(clusterSpec)).To(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_Missing(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	delete(clusterSpec.MachineConfigs, builder.ExternalEtcdMachineName)
	g.Expect(tinkerbell.AssertEtcdMachineRefExists(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertWorkerNodeGroupMachineRefsExists_Exists(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	g.Expect(tinkerbell.AssertWorkerNodeGroupMachineRefsExists(clusterSpec)).To(gomega.Succeed())
}

func TestAssertK8SVersionNot120_Success(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.KubernetesVersion = eksav1alpha1.Kube123
	g.Expect(tinkerbell.AssertK8SVersionNot120(clusterSpec)).Error().ShouldNot(gomega.HaveOccurred())
}

func TestAssertK8SVersionNot120_Error(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.KubernetesVersion = eksav1alpha1.Kube120
	g.Expect(tinkerbell.AssertK8SVersionNot120(clusterSpec)).Error().Should(gomega.HaveOccurred())
}

func TestAssertWorkerNodeGroupMachineRefsExists_Missing(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()
	delete(clusterSpec.MachineConfigs, builder.WorkerNodeGroupMachineName)
	g.Expect(tinkerbell.AssertWorkerNodeGroupMachineRefsExists(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_ExternalEtcdUnspecified(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	g.Expect(tinkerbell.AssertEtcdMachineRefExists(clusterSpec)).To(gomega.Succeed())
}

func TestNewIPNotInUseAssertion_NotInUseSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)

	netClient := mocks.NewMockNetClient(ctrl)
	netClient.EXPECT().
		DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("failed to connect"))

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.NewIPNotInUseAssertion(netClient)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestNewIPNotInUseAssertion_InUseFails(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)

	server, client := net.Pipe()
	defer server.Close()

	netClient := mocks.NewMockNetClient(ctrl)
	netClient.EXPECT().
		DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(client, nil)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.NewIPNotInUseAssertion(netClient)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertTinkerbellIPNotInUse_NotInUseSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)

	netClient := mocks.NewMockNetClient(ctrl)
	netClient.EXPECT().
		DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("failed to connect"))

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.AssertTinkerbellIPNotInUse(netClient)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestAssertTinkerbellIPNotInUse_InUseFails(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)

	server, client := net.Pipe()
	defer server.Close()

	netClient := mocks.NewMockNetClient(ctrl)
	netClient.EXPECT().
		DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(client, nil)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.AssertTinkerbellIPNotInUse(netClient)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertTinkerbellIPAndControlPlaneIPNotSame_DifferentSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	g.Expect(tinkerbell.AssertTinkerbellIPAndControlPlaneIPNotSame(clusterSpec)).To(gomega.Succeed())
}

func TestAssertTinkerbellIPAndControlPlaneIPNotSame_SameFails(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.DatacenterConfig.Spec.TinkerbellIP = "1.1.1.1"

	g.Expect(tinkerbell.AssertTinkerbellIPAndControlPlaneIPNotSame(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertPortsNotInUse_Succeeds(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)

	netClient := mocks.NewMockNetClient(ctrl)
	netClient.EXPECT().
		DialTimeout("tcp", gomock.Any(), 500*time.Millisecond).
		Times(3).
		Return(nil, errors.New("failed to connect"))

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.AssertPortsNotInUse(netClient)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestAssertPortsNotInUse_Fails(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)

	server, client := net.Pipe()
	defer server.Close()

	netClient := mocks.NewMockNetClient(ctrl)
	netClient.EXPECT().
		DialTimeout("tcp", gomock.Any(), 500*time.Millisecond).
		Times(3).
		Return(client, nil)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.AssertPortsNotInUse(netClient)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertAssertHookImageURLProxyNonAirgappedURLSuccess(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.ProxyConfiguration = &eksav1alpha1.ProxyConfiguration{
		HttpProxy:  "2.3.4.5",
		HttpsProxy: "2.3.4.5",
	}

	clusterSpec.DatacenterConfig.Spec.HookImagesURLPath = "https://anywhere.eks.amazonaws.com/"
	g.Expect(tinkerbell.AssertHookRetrievableWithoutProxy(clusterSpec)).To(gomega.Succeed())
}

func TestAssertAssertHookRetrievableWithoutProxyURLNotProvided(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.ProxyConfiguration = &eksav1alpha1.ProxyConfiguration{
		HttpProxy:  "2.3.4.5",
		HttpsProxy: "2.3.4.5",
	}

	g.Expect(tinkerbell.AssertHookRetrievableWithoutProxy(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertAssertHookRetrievableWithoutProxyURLUnreachable(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.ProxyConfiguration = &eksav1alpha1.ProxyConfiguration{
		HttpProxy:  "2.3.4.5",
		HttpsProxy: "2.3.4.5",
	}

	clusterSpec.DatacenterConfig.Spec.HookImagesURLPath = "https://1.2.3.4"
	g.Expect(tinkerbell.AssertHookRetrievableWithoutProxy(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareAvailableAssertionForCreate_SufficientSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	catalogue := hardware.NewCatalogue()

	// Add something for the control plane.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	// Add something for external etcd
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: clusterSpec.ExternalEtcdMachineConfig().Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	// Add something for the worker node group.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: clusterSpec.WorkerNodeGroupMachineConfig(
				clusterSpec.WorkerNodeGroupConfigurations()[0],
			).Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	assertion := tinkerbell.MinimumHardwareAvailableAssertionForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestMinimumHardwareAvailableAssertionForCreate_SufficientSucceedsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	// Add something for worker node group.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: clusterSpec.WorkerNodeGroupMachineConfig(
				clusterSpec.WorkerNodeGroupConfigurations()[0],
			).Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	assertion := tinkerbell.MinimumHardwareAvailableAssertionForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestMinimumHardwareAvailableAssertionForCreate_NoControlPlaneSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}
	clusterSpec.WorkerNodeGroupConfigurations()[0].Count = ptr.Int(0)
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.MinimumHardwareAvailableAssertionForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareAvailableAssertionForCreate_NoExternalEtcdSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.ControlPlaneConfiguration().Count = 0
	clusterSpec.WorkerNodeGroupConfigurations()[0].Count = ptr.Int(0)
	clusterSpec.ExternalEtcdMachineConfig().Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.MinimumHardwareAvailableAssertionForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareAvailableAssertionForCreate_NoWorkerNodeGroupSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.ControlPlaneConfiguration().Count = 0
	nodeGroup := clusterSpec.WorkerNodeGroupMachineConfig(clusterSpec.WorkerNodeGroupConfigurations()[0])
	nodeGroup.Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}
	clusterSpec.ExternalEtcdConfiguration().Count = 0

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.MinimumHardwareAvailableAssertionForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareAvailableAssertionForCreate_InsufficientFails(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.MinimumHardwareAvailableAssertionForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareAvailableAssertionForCreate_InsufficientFailsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.MinimumHardwareAvailableAssertionForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestValidatableClusterControlPlaneReplicaCount(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	validatableCluster := &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}

	g.Expect(validatableCluster.ControlPlaneReplicaCount()).To(gomega.Equal(1))
}

func TestValidatableClusterWorkerNodeGroupConfigs(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	validatableCluster := &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}

	workerConfigs := validatableCluster.WorkerNodeHardwareGroups()

	g.Expect(workerConfigs[0].MachineDeploymentName).To(gomega.Equal("cluster-worker-node-group-0"))
	g.Expect(workerConfigs[0].Replicas).To(gomega.Equal(1))
}

func TestValidatableClusterClusterK8sVersion(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.KubernetesVersion = eksav1alpha1.Kube125
	validatableCluster := &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}

	g.Expect(validatableCluster.ClusterK8sVersion()).To(gomega.Equal(eksav1alpha1.Kube125))
}

func TestValidatableClusterWorkerNodeGroupK8sVersion(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	kube125 := eksav1alpha1.Kube125
	clusterSpec.WorkerNodeGroupConfigurations()[0].KubernetesVersion = &kube125
	validatableCluster := &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}
	wngK8sVersion := validatableCluster.WorkerNodeGroupK8sVersion()
	mdName := fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, clusterSpec.WorkerNodeGroupConfigurations()[0].Name)

	g.Expect(wngK8sVersion[mdName]).To(gomega.Equal(kube125))
}

func TestValidatableTinkerbellCAPIControlPlaneReplicaCount(t *testing.T) {
	g := gomega.NewWithT(t)

	validatableCAPI := validatableTinkerbellCAPI()

	g.Expect(validatableCAPI.ControlPlaneReplicaCount()).To(gomega.Equal(1))
}

func TestValidatableTinkerbellCAPIWorkerNodeGroupConfigs(t *testing.T) {
	g := gomega.NewWithT(t)

	validatableCAPI := validatableTinkerbellCAPI()

	workerConfigs := validatableCAPI.WorkerNodeHardwareGroups()

	g.Expect(workerConfigs[0].MachineDeploymentName).To(gomega.Equal("cluster-worker-node-group-0"))
	g.Expect(workerConfigs[0].Replicas).To(gomega.Equal(1))
}

func TestValidateTinkerbellCAPIClusterK8sVersion(t *testing.T) {
	g := gomega.NewWithT(t)
	validatableCAPI := validatableTinkerbellCAPI()
	validatableCAPI.KubeadmControlPlane.Spec.Version = "v1.27.5-eks-1-27-12"
	k8sVersion := validatableCAPI.ClusterK8sVersion()
	kube127 := eksav1alpha1.Kube127
	g.Expect(k8sVersion).To(gomega.Equal(kube127))
}

func TestValidateTinkerbellCAPIWorkerNodeK8sVersion(t *testing.T) {
	g := gomega.NewWithT(t)
	validatableCAPI := validatableTinkerbellCAPI()
	wngK8sVersion := validatableCAPI.WorkerNodeGroupK8sVersion()
	mdName := validatableCAPI.WorkerGroups[0].MachineDeployment.Name
	kube121 := eksav1alpha1.Kube121
	g.Expect(wngK8sVersion[mdName]).To(gomega.Equal(kube121))
}

func TestAssertionsForScaleUpDown_Success(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.AssertionsForScaleUpDown(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, true)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	g.Expect(assertion(newClusterSpec)).To(gomega.Succeed())
}

func TestAssertionsForScaleUpDown_CAPISuccess(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	tinkerbellCAPI := validatableTinkerbellCAPI()

	assertion := tinkerbell.AssertionsForScaleUpDown(catalogue, tinkerbellCAPI, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	check := &tinkerbell.ValidatableTinkerbellClusterSpec{newClusterSpec}
	t.Log(tinkerbellCAPI.WorkerNodeHardwareGroups()[0].MachineDeploymentName)
	t.Log(check.WorkerNodeHardwareGroups()[0].MachineDeploymentName)

	g.Expect(assertion(newClusterSpec)).To(gomega.Succeed())
}

func TestAssertionsForScaleUpDown_ScaleUpControlPlaneSuccess(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	_ = catalogue.InsertHardware(&v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "cp"},
	}})
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.AssertionsForScaleUpDown(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	newClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count = 2

	g.Expect(assertion(newClusterSpec)).To(gomega.Succeed())
}

func TestAssertionsForScaleUpDown_ScaleUpWorkerSuccess(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	_ = catalogue.InsertHardware(&v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "worker"},
	}})
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.AssertionsForScaleUpDown(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	newClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(2)

	g.Expect(assertion(newClusterSpec)).To(gomega.Succeed())
}

func TestAssertionsForScaleUpDown_AddWorkerSuccess(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	_ = catalogue.InsertHardware(&v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "worker"},
	}})
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	clusterSpec.Spec.Cluster.Spec.WorkerNodeGroupConfigurations = []eksav1alpha1.WorkerNodeGroupConfiguration{}

	assertion := tinkerbell.AssertionsForScaleUpDown(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	g.Expect(assertion(newClusterSpec)).To(gomega.Succeed())
}

func TestAssertionsForRollingUpgrade_CPOnly(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	clusterSpec.Cluster.Spec.KubernetesVersion = eksav1alpha1.Kube124
	catalogue := hardware.NewCatalogue()
	_ = catalogue.InsertHardware(&v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "cp"},
	}})

	kube124 := eksav1alpha1.Kube124
	clusterSpec.WorkerNodeGroupConfigurations()[0].KubernetesVersion = &kube124
	assertion := tinkerbell.ExtraHardwareAvailableAssertionForRollingUpgrade(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	newClusterSpec.WorkerNodeGroupConfigurations()[0].KubernetesVersion = &kube124
	newClusterSpec.Cluster.Spec.KubernetesVersion = eksav1alpha1.Kube125
	g.Expect(assertion(newClusterSpec)).To(gomega.Succeed())
}

func TestAssertionsForRollingUpgrade_WorkerOnly(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	kube124 := eksav1alpha1.Kube124
	clusterSpec.Cluster.Spec.KubernetesVersion = kube124
	catalogue := hardware.NewCatalogue()
	_ = catalogue.InsertHardware(&v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "worker"},
	}})

	kube125 := eksav1alpha1.Kube125
	clusterSpec.WorkerNodeGroupConfigurations()[0].KubernetesVersion = &kube124
	assertion := tinkerbell.ExtraHardwareAvailableAssertionForRollingUpgrade(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	newClusterSpec.Cluster.Spec.KubernetesVersion = kube124
	newClusterSpec.WorkerNodeGroupConfigurations()[0].KubernetesVersion = &kube125
	g.Expect(assertion(newClusterSpec)).To(gomega.Succeed())
}

func TestAssertionsForRollingUpgrade_BothCPWorker(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	kube124 := eksav1alpha1.Kube124
	clusterSpec.Cluster.Spec.KubernetesVersion = kube124
	catalogue := hardware.NewCatalogue()
	_ = catalogue.InsertHardware(&v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "cp"},
	}})
	_ = catalogue.InsertHardware(&v1alpha1.Hardware{ObjectMeta: metav1.ObjectMeta{
		Labels: map[string]string{"type": "worker"},
	}})

	assertion := tinkerbell.ExtraHardwareAvailableAssertionForRollingUpgrade(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	kube125 := eksav1alpha1.Kube125
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	newClusterSpec.Cluster.Spec.KubernetesVersion = kube125
	g.Expect(assertion(newClusterSpec)).To(gomega.Succeed())
}

func TestAssertionsForRollingUpgrade_CPError(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	kube124 := eksav1alpha1.Kube124
	clusterSpec.Cluster.Spec.KubernetesVersion = kube124
	catalogue := hardware.NewCatalogue()

	assertion := tinkerbell.ExtraHardwareAvailableAssertionForRollingUpgrade(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	newClusterSpec.WorkerNodeGroupConfigurations()[0].KubernetesVersion = &kube124
	newClusterSpec.Cluster.Spec.KubernetesVersion = eksav1alpha1.Kube125
	g.Expect(assertion(newClusterSpec)).To(gomega.MatchError(gomega.ContainSubstring("minimum hardware count not met for selector '{\"type\":\"cp\"}'")))
}

func TestAssertionsForRollingUpgrade_WorkerError(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	kube124 := eksav1alpha1.Kube124
	kube125 := eksav1alpha1.Kube125
	clusterSpec.Cluster.Spec.KubernetesVersion = kube125
	clusterSpec.WorkerNodeGroupConfigurations()[0].KubernetesVersion = &kube124
	catalogue := hardware.NewCatalogue()

	assertion := tinkerbell.ExtraHardwareAvailableAssertionForRollingUpgrade(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, false)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	newClusterSpec.WorkerNodeGroupConfigurations()[0].KubernetesVersion = &kube125
	newClusterSpec.Cluster.Spec.KubernetesVersion = kube125
	g.Expect(assertion(newClusterSpec)).To(gomega.MatchError(gomega.ContainSubstring("minimum hardware count not met for selector '{\"type\":\"worker\"}'")))
}

func TestAssertionsForScaleUpDown_ExternalEtcdErrorFails(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.AssertionsForScaleUpDown(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, true)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	g.Expect(assertion(newClusterSpec)).To(gomega.MatchError(gomega.ContainSubstring("scale up/down not supported for external etcd")))
}

func TestAssertionsForScaleUpDown_FailsScaleUpAndRollingError(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.AssertionsForScaleUpDown(catalogue, &tinkerbell.ValidatableTinkerbellClusterSpec{clusterSpec}, true)
	newClusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	newClusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	newClusterSpec.WorkerNodeGroupConfigurations()[0].Count = ptr.Int(2)
	g.Expect(assertion(newClusterSpec)).NotTo(gomega.Succeed())
}

func TestHardwareSatisfiesOnlyOneSelectorAssertion_MeetsOnlyOneSelector(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	catalogue := hardware.NewCatalogue()
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	assertion := tinkerbell.HardwareSatisfiesOnlyOneSelectorAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestHardwareSatisfiesOnlyOneSelectorAssertion_MeetsMultipleSelectorFails(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	// Ensure we have distinct labels for selectors so we can populate the same key on the
	// test hardware.
	clusterSpec.ExternalEtcdMachineConfig().Spec.HardwareSelector = map[string]string{
		"etcd": "etcd",
	}

	catalogue := hardware.NewCatalogue()
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
			Labels: mergeHardwareSelectors(
				clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector,
				clusterSpec.ExternalEtcdMachineConfig().Spec.HardwareSelector,
			),
		},
	})).To(gomega.Succeed())

	assertion := tinkerbell.HardwareSatisfiesOnlyOneSelectorAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestHardwareSatisfiesOnlyOneSelectorAssertion_NoLabelsMeetsNothing(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	catalogue := hardware.NewCatalogue()
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.HardwareSatisfiesOnlyOneSelectorAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestAssertUpgradeRolloutStrategyValid_Succeeds(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	g.Expect(tinkerbell.AssertUpgradeRolloutStrategyValid(clusterSpec)).To(gomega.Succeed())
}

func TestAssertUpgradeRolloutStrategyValid_InPlaceFails(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &eksav1alpha1.ControlPlaneUpgradeRolloutStrategy{
		Type: "InPlace",
	}

	for _, mc := range clusterSpec.MachineConfigs {
		// InPlace upgrades are only supported on the Ubuntu OS family
		mc.Spec.OSFamily = eksav1alpha1.Bottlerocket
	}

	g.Expect(tinkerbell.AssertUpgradeRolloutStrategyValid(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertUpgradeRolloutStrategyValid_UpgradeStrategyNotEqual(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &eksav1alpha1.ControlPlaneUpgradeRolloutStrategy{
		Type: "InPlace",
	}
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].UpgradeRolloutStrategy = &eksav1alpha1.WorkerNodesUpgradeRolloutStrategy{
		Type: "RollingUpdate",
	}

	g.Expect(tinkerbell.AssertUpgradeRolloutStrategyValid(clusterSpec)).ToNot(gomega.Succeed())

	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].UpgradeRolloutStrategy = nil
	g.Expect(tinkerbell.AssertUpgradeRolloutStrategyValid(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertAutoScalerDisabledForInPlace_Success(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].AutoScalingConfiguration = &eksav1alpha1.AutoScalingConfiguration{
		MinCount: 1,
		MaxCount: 3,
	}
	g.Expect(tinkerbell.AssertAutoScalerDisabledForInPlace(clusterSpec)).To(gomega.Succeed())
}

func TestAssertAutoScalerDisabledForInPlace(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &eksav1alpha1.ControlPlaneUpgradeRolloutStrategy{
		Type: "InPlace",
	}
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].AutoScalingConfiguration = &eksav1alpha1.AutoScalingConfiguration{
		MinCount: 1,
		MaxCount: 3,
	}
	g.Expect(tinkerbell.AssertAutoScalerDisabledForInPlace(clusterSpec)).To(gomega.MatchError(gomega.ContainSubstring("austoscaler configuration not supported with InPlace")))
}

// mergeHardwareSelectors merges m1 with m2. Values already in m1 will be overwritten by m2.
func mergeHardwareSelectors(m1, m2 map[string]string) map[string]string {
	for name, value := range m2 {
		m1[name] = value
	}
	return m1
}

func validatableTinkerbellCAPI() *tinkerbell.ValidatableTinkerbellCAPI {
	return &tinkerbell.ValidatableTinkerbellCAPI{
		KubeadmControlPlane: &controlplanev1.KubeadmControlPlane{
			Spec: controlplanev1.KubeadmControlPlaneSpec{
				Replicas: ptr.Int32(1),
				Version:  "1.22",
			},
		},
		WorkerGroups: workerGroups(),
	}
}

func workerGroups() []*clusterapi.WorkerGroup[*v1beta1.TinkerbellMachineTemplate] {
	return []*clusterapi.WorkerGroup[*v1beta1.TinkerbellMachineTemplate]{
		{
			MachineDeployment: machineDeployment(func(md *clusterv1.MachineDeployment) {
				md.Name = "cluster-worker-node-group-0"
			}),
			ProviderMachineTemplate: machineTemplate(),
		},
	}
}
