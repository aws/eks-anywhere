package tinkerbell_test

import (
	"errors"
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

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/networkutils/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestAssertMachineConfigsValid_ValidSucceds(t *testing.T) {
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
