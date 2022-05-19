package tinkerbell

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

// TODO(chrisdoherty) Add worker node group assertions

// AssertMachineConfigsValid iterates over all machine configs in calling validateMachineConfig.
func AssertMachineConfigsValid(spec *ClusterSpec) error {
	for _, config := range spec.MachineConfigs {
		if err := validateMachineConfig(config); err != nil {
			return err
		}
	}
	return nil
}

// AssertDatacenterConfigValid asserts the DatacenterConfig in spec is valid.
func AssertDatacenterConfigValid(spec *ClusterSpec) error {
	return validateDatacenterConfig(spec.DatacenterConfig)
}

// AssertMachineConfigNamespaceMatchesDatacenterConfig ensures all machine configuration instances
// are configured with the same namespace as the provider specific data center configuration
// namespace.
func AssertMachineConfigNamespaceMatchesDatacenterConfig(spec *ClusterSpec) error {
	return validateMachineConfigNamespacesMatchDatacenterConfig(spec.DatacenterConfig, spec.MachineConfigs)
}

// AssertControlPlaneMachineRefExists ensures the control plane machine ref is referencing a
// known machine config.
func AssertControlPlaneMachineRefExists(spec *ClusterSpec) error {
	controlPlaneMachineRef := spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef
	if err := validateMachineRefExists(controlPlaneMachineRef, spec.MachineConfigs); err != nil {
		return fmt.Errorf("control plane configuration machine ref: %v", err)
	}
	return nil
}

// AssertEtcdMachineRefExists ensures that, if the etcd configuration is specified, it references
// a known machine config.
func AssertEtcdMachineRefExists(spec *ClusterSpec) error {
	// Unstacked etcd is optional.
	if spec.Cluster.Spec.ExternalEtcdConfiguration == nil {
		return nil
	}

	etcdMachineRef := spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef
	if err := validateMachineRefExists(etcdMachineRef, spec.MachineConfigs); err != nil {
		return fmt.Errorf("external etcd configuration machine group ref: %v", err)
	}

	return nil
}

// AssertWorkerNodeGroupMachineRefsExists ensures all worker node group machine refs are
// referencing a known machine config.
func AssertWorkerNodeGroupMachineRefsExists(spec *ClusterSpec) error {
	for _, group := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		groupRef := group.MachineGroupRef
		if err := validateMachineRefExists(groupRef, spec.MachineConfigs); err != nil {
			return fmt.Errorf("worker node group configuration machine group ref: %v", err)
		}
	}

	return nil
}

// AssertcontrolPlaneIPNotInUse ensures the endpoint host for the control plane isn't in use.
// The check may be unreliable due to its implementation.
func NewIPNotInUseAssertion(client networkutils.NetClient) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		ip := spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
		if err := validateIPUnused(client, ip); err != nil {
			return fmt.Errorf("control plane endpoint ip in use: %v", ip)
		}
		return nil
	}
}

// NewMinHardwareAvailableAssertion creates a ClusterSpecAssertion that ensures there is sufficient
// hardware in catalogue for creating the ClusterSpec. The minimum expeced hardware is the sum sum
// of requested control plane, etcd and worker node counts. The system requires hardware >= to
// requested provisioning.
func NewMinimumHardwareAvailableAssertion(catalogue *hardware.Catalogue) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		return validateMinimumExpectedHardware(spec.Cluster.Spec, catalogue)
	}
}
