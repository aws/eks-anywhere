package tinkerbell

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

// ClusterSpec represents a cluster configuration for Tinkerbell provisioning.
type ClusterSpec struct {
	*cluster.Spec

	// DatacenterConfig configured in the cluster configuration YAML.
	DatacenterConfig *v1alpha1.TinkerbellDatacenterConfig

	// MachineConfigs configured in the cluster configuration YAML whether they're used or not.
	MachineConfigs map[string]*v1alpha1.TinkerbellMachineConfig
}

// NewClusterSpec creates a ClusterSpec instance.
func NewClusterSpec(
	clusterSpec *cluster.Spec,
	machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig,
	datacenterConfig *v1alpha1.TinkerbellDatacenterConfig,
) *ClusterSpec {
	return &ClusterSpec{
		Spec:             clusterSpec,
		DatacenterConfig: datacenterConfig,
		MachineConfigs:   machineConfigs,
	}
}

// ControlPlaneMachineConfig retrieves the TinkerbellMachineConfig referenced by the cluster
// control plane machine reference.
func (s *ClusterSpec) ControlPlaneMachineConfig() *v1alpha1.TinkerbellMachineConfig {
	return s.MachineConfigs[s.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
}

// ControlPlaneConfiguration retrieves the control plane configuration from s.
func (s *ClusterSpec) ControlPlaneConfiguration() *v1alpha1.ControlPlaneConfiguration {
	return &s.Cluster.Spec.ControlPlaneConfiguration
}

// HasExternalEtcd returns true if there is an external etcd configuration.
func (s *ClusterSpec) HasExternalEtcd() bool {
	return s.Spec.Cluster.Spec.ExternalEtcdConfiguration != nil
}

// ExternalEtcdConfiguration returns the etcd configuration. The configuration may be nil. Consumers
// should check if external etcd configuration is present using HasExternalEtcd().
func (s *ClusterSpec) ExternalEtcdConfiguration() *v1alpha1.ExternalEtcdConfiguration {
	return s.Cluster.Spec.ExternalEtcdConfiguration
}

// ExternalEtcdMachineConfig retrieves the TinkerbellMachineConfig referenced by the cluster etcd machine
// reference.
func (s *ClusterSpec) ExternalEtcdMachineConfig() *v1alpha1.TinkerbellMachineConfig {
	if !s.HasExternalEtcd() {
		return nil
	}

	return s.MachineConfigs[s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
}

// WorkerNodeGroupConfigurations retrieves all worker node group configurations in s.
func (s *ClusterSpec) WorkerNodeGroupConfigurations() []v1alpha1.WorkerNodeGroupConfiguration {
	return s.Cluster.Spec.WorkerNodeGroupConfigurations
}

// WorkerNodeGroupMachineConfig retrieves the machine group associated with conf.
func (s *ClusterSpec) WorkerNodeGroupMachineConfig(conf v1alpha1.WorkerNodeGroupConfiguration) *v1alpha1.TinkerbellMachineConfig {
	return s.MachineConfigs[conf.MachineGroupRef.Name]
}

// ClusterSpecAssertion makes an assertion on spec.
type ClusterSpecAssertion func(spec *ClusterSpec) error

// ClusterSpecValidator is composed of a set of ClusterSpecAssertions to be run on a ClusterSpec
// instance.
type ClusterSpecValidator []ClusterSpecAssertion

// Register registers assertions with v.
func (v *ClusterSpecValidator) Register(assertions ...ClusterSpecAssertion) {
	*v = append(*v, assertions...)
}

// Validate validates spec with all assertions registered on v.
func (v *ClusterSpecValidator) Validate(spec *ClusterSpec) error {
	for _, a := range *v {
		if err := a(spec); err != nil {
			return err
		}
	}
	return nil
}

// NewClusterSpecValidator creates a ClusterSpecValidator instance with a set of default assertions.
// Any assertions passed will be registered in addition to the default assertions.
func NewClusterSpecValidator(assertions ...ClusterSpecAssertion) *ClusterSpecValidator {
	var v ClusterSpecValidator
	// Register mandatory assertions. If an assertion becomes optional dependent on context move it
	// to a New* func and register it dynamically. See assert.go for examples.
	v.Register(
		AssertK8SVersionNot120,
		AssertDatacenterConfigValid,
		AssertControlPlaneMachineRefExists,
		AssertEtcdMachineRefExists,
		AssertWorkerNodeGroupMachineRefsExists,
		AssertMachineConfigsValid,
		AssertMachineConfigNamespaceMatchesDatacenterConfig,
		AssertOsFamilyValid,
		AssertOSImageURL,
		AssertTinkerbellIPAndControlPlaneIPNotSame,
		AssertHookRetrievableWithoutProxy,
		AssertUpgradeRolloutStrategyValid,
		AssertAutoScalerDisabledForInPlace,
	)
	v.Register(assertions...)
	return &v
}
