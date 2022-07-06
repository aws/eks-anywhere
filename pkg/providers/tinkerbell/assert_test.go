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
	clusterSpec := DefaultClusterSpec()
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
			spec := DefaultClusterSpec()
			mutate(spec)
			g.Expect(tinkerbell.AssertMachineConfigsValid(spec)).ToNot(gomega.Succeed())
		})
	}
}

func TestAssertDatacenterConfigValid_ValidSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := DefaultClusterSpec()
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
			cluster := DefaultClusterSpec()
			mutate(cluster)
			g.Expect(tinkerbell.AssertDatacenterConfigValid(cluster)).ToNot(gomega.Succeed())
		})
	}
}

func TestAssertMachineConfigNamespaceMatchesDatacenterConfig_Same(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := DefaultClusterSpecBuilder()
	clusterSpec := builder.Build()
	err := tinkerbell.AssertMachineConfigNamespaceMatchesDatacenterConfig(clusterSpec)
	g.Expect(err).To(gomega.Succeed())
}

func TestAssertMachineConfigNamespaceMatchesDatacenterConfig_Different(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := DefaultClusterSpecBuilder()
	clusterSpec := builder.Build()

	// Invalidate the namespace check.
	clusterSpec.MachineConfigs[builder.ControlPlaneMachineName].Namespace = "foo-bar"

	err := tinkerbell.AssertMachineConfigNamespaceMatchesDatacenterConfig(clusterSpec)
	g.Expect(err).ToNot(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_Exists(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := DefaultClusterSpec()
	g.Expect(tinkerbell.AssertEtcdMachineRefExists(clusterSpec)).To(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_Missing(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := DefaultClusterSpecBuilder()
	clusterSpec := builder.Build()
	delete(clusterSpec.MachineConfigs, builder.ExternalEtcdMachineName)
	g.Expect(tinkerbell.AssertEtcdMachineRefExists(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertWorkerNodeGroupMachineRefsExists_Exists(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := DefaultClusterSpec()
	g.Expect(tinkerbell.AssertWorkerNodeGroupMachineRefsExists(clusterSpec)).To(gomega.Succeed())
}

func TestAssertWorkerNodeGroupMachineRefsExists_Missing(t *testing.T) {
	g := gomega.NewWithT(t)
	builder := DefaultClusterSpecBuilder()
	clusterSpec := builder.Build()
	delete(clusterSpec.MachineConfigs, builder.WorkerNodeGroupMachineName)
	g.Expect(tinkerbell.AssertWorkerNodeGroupMachineRefsExists(clusterSpec)).ToNot(gomega.Succeed())
}

func TestAssertEtcdMachineRefExists_ExternalEtcdUnspecified(t *testing.T) {
	g := gomega.NewWithT(t)
	clusterSpec := DefaultClusterSpec()
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

	clusterSpec := DefaultClusterSpec()

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

	clusterSpec := DefaultClusterSpec()

	assertion := tinkerbell.NewIPNotInUseAssertion(netClient)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareForCreate_SufficientSucceeds(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := DefaultClusterSpec()

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

	assertion := tinkerbell.MinimumHardwareForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestMinimumHardwareForCreate_SufficientSucceedsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := DefaultClusterSpec()
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

	assertion := tinkerbell.MinimumHardwareForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).To(gomega.Succeed())
}

func TestMinimumHardwareForCreate_NoControlPlaneSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := DefaultClusterSpec()
	clusterSpec.ControlPlaneMachineConfig().Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}
	clusterSpec.WorkerNodeGroupConfigurations()[0].Count = 0
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.MinimumHardwareForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareForCreate_NoExternalEtcdSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := DefaultClusterSpec()
	clusterSpec.ControlPlaneConfiguration().Count = 0
	clusterSpec.WorkerNodeGroupConfigurations()[0].Count = 0
	clusterSpec.ExternalEtcdMachineConfig().Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.MinimumHardwareForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareForCreate_NoWorkerNodeGroupSelectorMatchesAnything(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := DefaultClusterSpec()
	clusterSpec.ControlPlaneConfiguration().Count = 0
	nodeGroup := clusterSpec.WorkerNodeGroupMachineConfig(clusterSpec.WorkerNodeGroupConfigurations()[0])
	nodeGroup.Spec.HardwareSelector = eksav1alpha1.HardwareSelector{}
	clusterSpec.ExternalEtcdConfiguration().Count = 0

	catalogue := hardware.NewCatalogue()

	// Add something to match the control plane selector.
	g.Expect(catalogue.InsertHardware(&v1alpha1.Hardware{})).To(gomega.Succeed())

	assertion := tinkerbell.MinimumHardwareForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareForCreate_InsufficientFails(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := DefaultClusterSpec()

	assertion := tinkerbell.MinimumHardwareForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestMinimumHardwareForCreate_InsufficientFailsWithoutExternalEtcd(t *testing.T) {
	g := gomega.NewWithT(t)

	catalogue := hardware.NewCatalogue()
	clusterSpec := DefaultClusterSpec()
	clusterSpec.Spec.Cluster.Spec.ExternalEtcdConfiguration = nil

	assertion := tinkerbell.MinimumHardwareForCreate(catalogue)
	g.Expect(assertion(clusterSpec)).ToNot(gomega.Succeed())
}

func TestHardwareSatisfiesOnlyOneSelectorAssertion_MeetsOnlyOneSelector(t *testing.T) {
	g := gomega.NewWithT(t)

	clusterSpec := DefaultClusterSpec()
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

	clusterSpec := DefaultClusterSpec()

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

	clusterSpec := DefaultClusterSpec()

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

func TestMinimumHardwareForUpgrade(t *testing.T) {
	for name, d := range map[string]struct {
		specs     func() (current, desired *tinkerbell.ClusterSpec)
		catalogue func(current, desired *tinkerbell.ClusterSpec) *hardware.Catalogue
		result    gomega.OmegaMatcher
	}{
		"KubernetesVersion": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				current, desired = DefaultClusterSpec(), DefaultClusterSpec()
				desired.Cluster.Spec.KubernetesVersion = "1.22"
				current.Cluster.Spec.ExternalEtcdConfiguration = nil
				return current, desired
			},
			catalogue: versionUpgradeCatalogue,
			result:    gomega.Succeed(),
		},
		"ScaleControlPlane": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				current, desired = DefaultClusterSpec(), DefaultClusterSpec()
				desired.Cluster.Spec.ControlPlaneConfiguration.Count = 2
				return current, desired
			},
			catalogue: scaleUpgradeCatalogue,
			result:    gomega.Succeed(),
		},
		"ScaleChangeWorkerNodeGroup": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				current, desired = DefaultClusterSpec(), DefaultClusterSpec()
				desired.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 2
				return current, desired
			},
			catalogue: scaleUpgradeCatalogue,
			result:    gomega.Succeed(),
		},
		"ScaleDeleteWorkerNodeGroup": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				current, desired = DefaultClusterSpec(), DefaultClusterSpec()
				desired.Cluster.Spec.WorkerNodeGroupConfigurations = desired.Cluster.Spec.WorkerNodeGroupConfigurations[:1]
				return current, desired
			},
			catalogue: scaleUpgradeCatalogue,
			result:    gomega.Succeed(),
		},
		"ScaleAddWorkerNodeGroup": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				current, desired = DefaultClusterSpec(), DefaultClusterSpec()

				firstGroup := desired.Cluster.Spec.WorkerNodeGroupConfigurations[0]
				desired.Cluster.Spec.WorkerNodeGroupConfigurations = append(
					desired.Cluster.Spec.WorkerNodeGroupConfigurations,
					eksav1alpha1.WorkerNodeGroupConfiguration{
						MachineGroupRef: firstGroup.MachineGroupRef,
						Count:           1,
					},
				)
				return current, desired
			},
			catalogue: scaleUpgradeCatalogue,
			result:    gomega.Succeed(),
		},
		"ScaleMultiple": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				current, desired = DefaultClusterSpec(), DefaultClusterSpec()
				desired.Cluster.Spec.ControlPlaneConfiguration.Count = 2
				desired.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 2
				desired.Cluster.Spec.WorkerNodeGroupConfigurations[1].Count = 2
				return current, desired
			},
			catalogue: scaleUpgradeCatalogue,
			result:    gomega.Succeed(),
		},
		"NoAction": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				return DefaultClusterSpec(), DefaultClusterSpec()
			},
			catalogue: emptyCatalogue,
			result:    gomega.Succeed(),
		},
		"VersionAndScaledControlPlane": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				current, desired = DefaultClusterSpec(), DefaultClusterSpec()
				desired.Cluster.Spec.KubernetesVersion = "1.22"
				desired.Cluster.Spec.ControlPlaneConfiguration.Count = 2
				return current, desired
			},
			catalogue: emptyCatalogue,
			result:    gomega.HaveOccurred(),
		},
		"VersionAndScaledWorkerNodeGroup": {
			specs: func() (current, desired *tinkerbell.ClusterSpec) {
				current, desired = DefaultClusterSpec(), DefaultClusterSpec()
				desired.Cluster.Spec.KubernetesVersion = "1.22"
				desired.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = 2
				return current, desired
			},
			catalogue: emptyCatalogue,
			result:    gomega.HaveOccurred(),
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			current, desired := d.specs()
			catalogue := d.catalogue(current, desired)
			assertion := tinkerbell.MinimumHardwareForUpgrade(current, catalogue)
			g.Expect(assertion(desired)).To(d.result)
		})
	}
}

func versionUpgradeCatalogue(current, desired *tinkerbell.ClusterSpec) *hardware.Catalogue {
	catalogue := hardware.NewCatalogue()

	err := catalogue.InsertHardware(&v1alpha1.Hardware{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string(current.ControlPlaneMachineConfig().Spec.HardwareSelector),
		},
	})
	if err != nil {
		panic(err)
	}

	for _, group := range current.WorkerNodeGroupConfigurations() {
		selector := current.WorkerNodeGroupMachineConfig(group).Spec.HardwareSelector
		err := catalogue.InsertHardware(&v1alpha1.Hardware{
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string(selector),
			},
		})
		if err != nil {
			panic(err)
		}
	}

	return catalogue
}

func emptyCatalogue(current, desired *tinkerbell.ClusterSpec) *hardware.Catalogue {
	return hardware.NewCatalogue()
}

func scaleUpgradeCatalogue(current, desired *tinkerbell.ClusterSpec) *hardware.Catalogue {
	catalogue := hardware.NewCatalogue()

	controlPlaneDiff := desired.ControlPlaneConfiguration().Count - current.ControlPlaneConfiguration().Count
	if controlPlaneDiff > 0 {
		for i := 0; i < controlPlaneDiff; i++ {
			err := catalogue.InsertHardware(&v1alpha1.Hardware{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string(desired.ControlPlaneMachineConfig().Spec.HardwareSelector),
				},
			})
			if err != nil {
				panic(err)
			}
		}
	}

	currentGroups := buildNodeGroupMap(current.WorkerNodeGroupConfigurations())

	for _, desiredGroup := range desired.WorkerNodeGroupConfigurations() {
		// Default to the desired group count.
		diff := desiredGroup.Count

		// If there's a corresponding current worker node group diff the two.
		currentGroup, exists := currentGroups[desiredGroup.Name]
		if exists {
			diff = desiredGroup.Count - currentGroup.Count
		}

		if diff > 0 {
			selector := desired.WorkerNodeGroupMachineConfig(desiredGroup).Spec.HardwareSelector
			for i := 0; i < diff; i++ {
				err := catalogue.InsertHardware(&v1alpha1.Hardware{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string(selector),
					},
				})
				if err != nil {
					panic(err)
				}
			}
		}
	}

	return catalogue
}

// buildNodeGroupMap builds a map of node group name to node group config.
func buildNodeGroupMap(s []eksav1alpha1.WorkerNodeGroupConfiguration) map[string]eksav1alpha1.WorkerNodeGroupConfiguration {
	groups := map[string]eksav1alpha1.WorkerNodeGroupConfiguration{}
	for _, nodeGroup := range s {
		groups[nodeGroup.Name] = nodeGroup
	}
	return groups
}

func removeExternalEtcd(spec *tinkerbell.ClusterSpec) {
}
