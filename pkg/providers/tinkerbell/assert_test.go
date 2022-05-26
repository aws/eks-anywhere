package tinkerbell_test

import (
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"

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
	g := gomega.NewWithT(t)
	builder := NewDefaultValidClusterSpecBuilder()
	clusterSpec := builder.Build()

	// Invalidate the namespace check.
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Name = ""

	g.Expect(tinkerbell.AssertMachineConfigsValid(clusterSpec)).ToNot(gomega.Succeed())
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

func TestNewMinimumHardwareAvailableAssertion_SufficientSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.NewMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestNewMinimumHardwareAvailableAssertion_SufficientSucceedsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.NewMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestNewMinimumHardwareAvailableAssertion_InsufficientFails(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()

	assertion := tinkerbell.NewMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestNewMinimumHardwareAvailableAssertion_InsufficientFailsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := NewDefaultValidClusterSpecBuilder().Build()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.NewMinimumHardwareAvailableAssertion(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}
