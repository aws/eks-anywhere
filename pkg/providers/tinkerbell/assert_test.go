package tinkerbell_test

import (
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
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
		Times(5).
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
	clusterSpec.WorkerNodeGroupConfigurations()[0].Count = 0
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
	clusterSpec.WorkerNodeGroupConfigurations()[0].Count = 0
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
