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

// HasExternalEtcd returns true if there is an external etcd configuration.
func (s *ClusterSpec) HasExternalEtcd() bool {
	return s.Spec.Cluster.Spec.ExternalEtcdConfiguration != nil
}

// EtcdMachineConfig retrieves the TinkerbellMachineConfig referenced by the cluster etcd machine
// reference.
func (s *ClusterSpec) EtcdMachineConfig() *v1alpha1.TinkerbellMachineConfig {
	if !s.HasExternalEtcd() {
		return nil
	}

	return s.MachineConfigs[s.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
}

// FirstWorkerMachineConfig retrieves the TinkerbellMachineConfig referenced by the first node
// group machine reference.
func (s *ClusterSpec) FirstWorkerMachineConfig() *v1alpha1.TinkerbellMachineConfig {
	return s.MachineConfigs[s.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
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

// NewClusterSpecValidator creates a ClusterSpecValidator instance with a set of default asseritons.
// Any assertions passed will be registered in addition to the default assertions.
func NewClusterSpecValidator(assertions ...ClusterSpecAssertion) *ClusterSpecValidator {
	var v ClusterSpecValidator
	// Register mandatory assertions. If an assertion becomes optional dependent on context move it
	// to a New* func and register it dynamically. See assert.go for examples.
	v.Register(
		AssertDatacenterConfigValid,
		AssertControlPlaneMachineRefExists,
		AssertEtcdMachineRefExists,
		AssertWorkerNodeGroupMachineRefsExists,
		AssertMachineConfigsValid,
		AssertMachineConfigNamespaceMatchesDatacenterConfig,
	)
	v.Register(assertions...)
	return &v
}
