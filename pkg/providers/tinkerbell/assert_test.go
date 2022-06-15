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
	clusterSpec := NewDefaultClusterSpecBuilder().Build()
	g.Expect(tinkerbell.AssertMachineConfigsValid(clusterSpec)).To(gomega.Succeed())
}

func TestAssertMachineConfigsValid_InvalidFails(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultClusterSpecBuilder()
	clusterSpec := builder.Build()

	// Invalidate the namespace check.
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Name = ""

	g.Expect(tinkerbell.AssertMachineConfigsValid(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertDatacenterConfigValid_ValidSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultClusterSpecBuilder().Build()
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
			cluster := NewDefaultClusterSpecBuilder().Build()
			mutate(cluster)
			g.Expect(tinkerbell.AssertDatacenterConfigValid(cluster)).ToNot(gomega.Succeed())
		})
	}
}

func TestAssertMachineConfigNamespaceMatchesDatacenterConfig_Same(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultClusterSpecBuilder()
	clusterSpec := builder.Build()
	err := tinkerbell.AssertMachineConfigNamespaceMatchesDatacenterConfig(clusterSpec)
	g.Expect(err).To(gomega.Succeed())
}

func TestAssertMachineConfigNamespaceMatchesDatacenterConfig_Different(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultClusterSpecBuilder()
	clusterSpec := builder.Build()

	// Invalidate the namespace check.
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Namespace = "foo-bar"

	err := tinkerbell.AssertMachineConfigNamespaceMatchesDatacenterConfig(clusterSpec)
	g.Expect(err).ToNot(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_Exists(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultClusterSpecBuilder().Build()
	g.Expect(tinkerbell.AssertEtcdMachineRefExists(clusterSpec)).To(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_Missing(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultClusterSpecBuilder()
	clusterSpec := builder.Build()
	delete(clusterSpec.MachineConfigs, builder.ExternalEtcdMachineName)
	g.Expect(tinkerbell.AssertEtcdMachineRefExists(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertWorkerNodeGroupMachineRefsExists_Exists(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultClusterSpecBuilder().Build()
	g.Expect(tinkerbell.AssertWorkerNodeGroupMachineRefsExists(clusterSpec)).To(gomega.Succeed())
}

func TestAssertWorkerNodeGroupMachineRefsExists_Missing(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := NewDefaultClusterSpecBuilder()
	clusterSpec := builder.Build()
	delete(clusterSpec.MachineConfigs, builder.WorkerNodeGroupMachineName)
	g.Expect(tinkerbell.AssertWorkerNodeGroupMachineRefsExists(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_ExternalEtcdUnspecified(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := NewDefaultClusterSpecBuilder().Build()
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

	clusterSpec := NewDefaultClusterSpecBuilder().Build()

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

	clusterSpec := NewDefaultClusterSpecBuilder().Build()

	assertion := tinkerbell.NewIPNotInUseAssertion(netClient)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestNewCreateMinimumHardwareAvailableAssertion_SufficientSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultClusterSpecBuilder().Build()

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

	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestNewCreateMinimumHardwareAvailableAssertion_SufficientSucceedsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultClusterSpecBuilder().Build()
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

	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestNewCreateMinimumHardwareAvailableAssertion_NoControlPlaneSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultClusterSpecBuilder().Build()
	clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}
	clusterSpec.WorkerNodeGroupConfigurations()[0].Count = 0
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestNewCreateMinimumHardwareAvailableAssertion_NoExternalEtcdSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultClusterSpecBuilder().Build()
	clusterSpec.ControlPlaneConfiguration().Count = 0
	clusterSpec.WorkerNodeGroupConfigurations()[0].Count = 0
	clusterSpec.ExternalEtcdMachineConfig().Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestNewCreateMinimumHardwareAvailableAssertion_NoWorkerNodeGroupSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := NewDefaultClusterSpecBuilder().Build()
	clusterSpec.ControlPlaneConfiguration().Count = 0
	nodeGroup := clusterSpec.WorkerNodeGroupMachineConfig(clusterSpec.WorkerNodeGroupConfigurations()[0])
	nodeGroup.Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}
	clusterSpec.ExternalEtcdConfiguration().Count = 0

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestNewCreateMinimumHardwareAvailableAssertion_InsufficientFails(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultClusterSpecBuilder().Build()

	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestNewCreateMinimumHardwareAvailableAssertion_InsufficientFailsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestNewCreateMinimumHardwareAvailableAssertion_TotalCountsNotMet(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())
	clusterSpecBuilder := NewDefaultClusterSpecBuilder()
	clusterSpecBuilder.WithoutHardwareSelectors()
	clusterSpec := clusterSpecBuilder.Build()

	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestNewUpgradeMinimumHardwareAvailableAssertion_SufficientSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)

	current := NewDefaultClusterSpecBuilder().Build()
	desired := NewDefaultClusterSpecBuilder().Build()

	desired.ControlPlaneConfiguration().Count += 1
	desired.WorkerNodeGroupConfigurations()[0].Count += 1
	desired.ExternalEtcdConfiguration().Count += 1

	catalogue := hardware.NewCatalogue()

	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: desired.ControlPlaneMachineConfig().Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: desired.ExternalEtcdMachineConfig().Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: desired.WorkerNodeGroupMachineConfig(
				desired.WorkerNodeGroupConfigurations()[0],
			).Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	assertion := tinkerbell.NewUpgradeMinimumHardwareAvailableAssertion(current, catalogue)
	g.Expect(assertion(desired)).To(gomega.Succeed())
}

func TestNewUpgradeMinimumHardwareAvailableAssertion_SufficientSucceedsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	current := NewDefaultClusterSpecBuilder().Build()
	desired := NewDefaultClusterSpecBuilder().Build()
	current.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil
	desired.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	// Bump all groups by a count of 1.
	desired.ControlPlaneConfiguration().Count += 1
	desired.WorkerNodeGroupConfigurations()[0].Count += 1

	catalogue := hardware.NewCatalogue()

	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: desired.ControlPlaneMachineConfig().Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: desired.WorkerNodeGroupMachineConfig(
				desired.WorkerNodeGroupConfigurations()[0],
			).Spec.HardwareSelector,
		},
	})).To(gomega.Succeed())

	assertion := tinkerbell.NewUpgradeMinimumHardwareAvailableAssertion(current, catalogue)
	g.Expect(assertion(desired)).To(gomega.Succeed())
}

func TestNewUpgradeMinimumHardwareAvailableAssertion_InsufficientFailsForExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	current := NewDefaultClusterSpecBuilder().Build()
	desired := NewDefaultClusterSpecBuilder().Build()
	desired.ExternalEtcdConfiguration().Count += 1

	catalogue := hardware.NewCatalogue()

	assertion := tinkerbell.NewUpgradeMinimumHardwareAvailableAssertion(current, catalogue)
	g.Expect(assertion(desired)).ToNot(gomega.Succeed())
}

func TestNewUpgradeMinimumHardwareAvailableAssertion_InsufficientFailsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	current := NewDefaultClusterSpecBuilder().ExternalEtcd(false).Build()
	desired := NewDefaultClusterSpecBuilder().ExternalEtcd(false).Build()
	desired.ControlPlaneConfiguration().Count += 1

	catalogue := hardware.NewCatalogue()

	assertion := tinkerbell.NewUpgradeMinimumHardwareAvailableAssertion(current, catalogue)
	g.Expect(assertion(desired)).ToNot(gomega.Succeed())
}

// func TestNewUpgradeMinimumHardwareAvailableAssertion_TotalCountsNotMet(t *testing.T) {
// 	g := gomega.NewWithT(t)

// 	catalogue := hardware.NewCatalogue()
// 	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())
// 	clusterSpecBuilder := NewDefaultClusterSpecBuilder()
// 	clusterSpecBuilder.WithoutHardwareSelectors()
// 	clusterSpec := clusterSpecBuilder.Build()

// 	assertion := tinkerbell.NewCreateMinimumHardwareAvailableAssertion(catalogue)
// 	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
// }
