package tinkerbell

import (
	"errors"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
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

// AssertK8SVersionNot120 ensures Kubernetes version is not set to v1.20
func AssertK8SVersionNot120(spec *ClusterSpec) error {
	if spec.Cluster.Spec.KubernetesVersion == v1alpha1.Kube120 {
		return errors.New("kubernetes version v1.20 is not supported for Bare Metal")
	}

	return nil
}

func AssertOsFamilyValid(spec *ClusterSpec) error {
	return validateOsFamily(spec)
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

//AssertTinkerbellIPNotInUse ensures tinkerbell ip isn't in use
func AssertTinkerbellIPNotInUse(client networkutils.NetClient) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		ip := spec.DatacenterConfig.Spec.TinkerbellIP
		if networkutils.IsIPInUse(client, ip) {
			return fmt.Errorf("tinkerbellIP <%s> is already in use, please provide a unique IP", ip)
		}
		return nil
	}
}

// HardwareSatisfiesOnlyOneSelectorAssertion ensures hardware in catalogue only satisfies 1
// of the MachineConfig's HardwareSelector's from the spec.
func HardwareSatisfiesOnlyOneSelectorAssertion(catalogue *hardware.Catalogue) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		selectors, err := selectorsFromClusterSpec(spec)
		if err != nil {
			return err
		}

		return validateHardwareSatisfiesOnlyOneSelector(catalogue.AllHardware(), selectors)
	}
}

// selectorsFromClusterSpec extracts all selectors specified on MachineConfig's from spec.
func selectorsFromClusterSpec(spec *ClusterSpec) (selectorSet, error) {
	selectors := selectorSet{}

	if err := selectors.Add(spec.ControlPlaneMachineConfig().Spec.HardwareSelector); err != nil {
		return nil, err
	}

	for _, nodeGroup := range spec.WorkerNodeGroupConfigurations() {
		err := selectors.Add(spec.WorkerNodeGroupMachineConfig(nodeGroup).Spec.HardwareSelector)
		if err != nil {
			return nil, err
		}
	}

	if spec.HasExternalEtcd() {
		if err := selectors.Add(spec.ExternalEtcdMachineConfig().Spec.HardwareSelector); err != nil {
			return nil, err
		}
	}

	return selectors, nil
}

// MinimumHardwareAvailableAssertionForCreate asserts that catalogue has sufficient hardware to
// support the ClusterSpec during a create workflow.
//
// It does not protect against intersections or subsets so consumers should ensure a 1-2-1
// mapping between catalogue hardware and selectors.
func MinimumHardwareAvailableAssertionForCreate(catalogue *hardware.Catalogue) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		// Without Hardware selectors we get undesirable behavior so ensure we have them for
		// all MachineConfigs.
		if err := ensureHardwareSelectorsSpecified(spec); err != nil {
			return err
		}

		// Build a set of required hardware counts per machine group. minimumHardwareRequirements
		// will account for the same selector being specified on different groups.
		requirements := minimumHardwareRequirements{}

		err := requirements.Add(
			spec.ControlPlaneMachineConfig().Spec.HardwareSelector,
			spec.ControlPlaneConfiguration().Count,
		)
		if err != nil {
			return err
		}

		for _, nodeGroup := range spec.WorkerNodeGroupConfigurations() {
			err := requirements.Add(
				spec.WorkerNodeGroupMachineConfig(nodeGroup).Spec.HardwareSelector,
				nodeGroup.Count,
			)
			if err != nil {
				return err
			}
		}

		if spec.HasExternalEtcd() {
			err := requirements.Add(
				spec.ExternalEtcdMachineConfig().Spec.HardwareSelector,
				spec.ExternalEtcdConfiguration().Count,
			)
			if err != nil {
				return err
			}
		}

		return validateMinimumHardwareRequirements(requirements, catalogue)
	}
}

func AssertionsForScaleUpDown(catalogue *hardware.Catalogue, currentSpec *cluster.Spec, rollingUpgrade bool) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		// Without Hardware selectors we get undesirable behavior so ensure we have them for
		// all MachineConfigs.
		if err := ensureHardwareSelectorsSpecified(spec); err != nil {
			return err
		}

		// Build a set of required hardware counts per machine group. minimumHardwareRequirements
		// will account for the same selector being specified on different groups.
		requirements := minimumHardwareRequirements{}

		if currentSpec.Cluster.Spec.ControlPlaneConfiguration.Count != spec.Cluster.Spec.ControlPlaneConfiguration.Count {
			if rollingUpgrade {
				return fmt.Errorf("cannot perform scale up or down during rolling upgrades")
			}
			if currentSpec.Cluster.Spec.ControlPlaneConfiguration.Count < spec.Cluster.Spec.ControlPlaneConfiguration.Count {
				err := requirements.Add(
					spec.ControlPlaneMachineConfig().Spec.HardwareSelector,
					spec.ControlPlaneConfiguration().Count-currentSpec.Cluster.Spec.ControlPlaneConfiguration.Count,
				)
				if err != nil {
					return fmt.Errorf("error during scale up: %v", err)
				}
			}
		}

		workerNodeGroupMap := make(map[string]*v1alpha1.WorkerNodeGroupConfiguration)
		for _, workerNodeGroupConfiguration := range currentSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
			workerNodeGroupMap[workerNodeGroupConfiguration.Name] = &workerNodeGroupConfiguration
		}

		for _, nodeGroupNewSpec := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
			if workerNodeGrpOldSpec, ok := workerNodeGroupMap[nodeGroupNewSpec.Name]; ok {
				if nodeGroupNewSpec.Count != workerNodeGrpOldSpec.Count {
					if rollingUpgrade {
						return fmt.Errorf("cannot perform scale up or down during rolling upgrades")
					}
					if nodeGroupNewSpec.Count > workerNodeGrpOldSpec.Count {
						err := requirements.Add(
							spec.WorkerNodeGroupMachineConfig(nodeGroupNewSpec).Spec.HardwareSelector,
							nodeGroupNewSpec.Count-workerNodeGrpOldSpec.Count,
						)
						if err != nil {
							return fmt.Errorf("error during scale up: %v", err)
						}
					}
				}
			} else { // worker node group was newly added
				if rollingUpgrade {
					return fmt.Errorf("cannot perform scale up or down during rolling upgrades")
				}
				err := requirements.Add(
					spec.WorkerNodeGroupMachineConfig(nodeGroupNewSpec).Spec.HardwareSelector,
					nodeGroupNewSpec.Count,
				)
				if err != nil {
					return fmt.Errorf("error during scale up: %v", err)
				}
			}
		}

		if spec.HasExternalEtcd() {
			return fmt.Errorf("scale up/down not supported for external etcd")
			/* if spec.Cluster.Spec.ExternalEtcdConfiguration.Count > currentSpec.Cluster.Spec.ExternalEtcdConfiguration.Count {
				err := requirements.Add(
					spec.ExternalEtcdMachineConfig().Spec.HardwareSelector,
					spec.ExternalEtcdConfiguration().Count - currentSpec.Cluster.Spec.ExternalEtcdConfiguration.Count,
				)
				if err != nil {
					return err
				}
			} */
		}

		if err := validateMinimumHardwareRequirements(requirements, catalogue); err != nil {
			return fmt.Errorf("for scale up, %v", err)
		}
		return nil
	}
}

func ExtraHardwareAvailableAssertionForRollingUpgrade(catalogue *hardware.Catalogue, maxSurge int) ClusterSpecAssertion {
	return ExtraHardwareAvailableAssertion(catalogue, maxSurge)
}

func ExtraHardwareAvailableAssertion(catalogue *hardware.Catalogue, maxSurge int) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		// Without Hardware selectors we get undesirable behavior so ensure we have them for
		// all MachineConfigs.
		if err := ensureHardwareSelectorsSpecified(spec); err != nil {
			return err
		}

		// Build a set of required hardware counts per machine group. minimumHardwareRequirements
		// will account for the same selector being specified on different groups.
		requirements := minimumHardwareRequirements{}

		err := requirements.Add(
			spec.ControlPlaneMachineConfig().Spec.HardwareSelector,
			maxSurge,
		)
		if err != nil {
			return fmt.Errorf("for rolling upgrade, %v", err)
		}

		for _, nodeGroup := range spec.WorkerNodeGroupConfigurations() {
			err := requirements.Add(
				spec.WorkerNodeGroupMachineConfig(nodeGroup).Spec.HardwareSelector,
				maxSurge,
			)
			if err != nil {
				return fmt.Errorf("for rolling upgrade, %v", err)
			}
		}

		if spec.HasExternalEtcd() {
			return fmt.Errorf("external etcd upgrade is not supported")
		}

		if err := validateMinimumHardwareRequirements(requirements, catalogue); err != nil {
			return fmt.Errorf("for rolling upgrade, %v", err)
		}
		return nil
	}
}

// ensureHardwareSelectorsSpecified ensures each machine config present in spec has a hardware
// selector.
func ensureHardwareSelectorsSpecified(spec *ClusterSpec) error {
	if len(spec.ControlPlaneMachineConfig().Spec.HardwareSelector) == 0 {
		return missingHardwareSelectorErr{
			Name: spec.ControlPlaneMachineConfig().Name,
		}
	}

	for _, nodeGroup := range spec.WorkerNodeGroupConfigurations() {
		if len(spec.WorkerNodeGroupMachineConfig(nodeGroup).Spec.HardwareSelector) == 0 {
			return missingHardwareSelectorErr{
				Name: spec.WorkerNodeGroupMachineConfig(nodeGroup).Name,
			}
		}
	}

	if spec.HasExternalEtcd() {
		if len(spec.ExternalEtcdMachineConfig().Spec.HardwareSelector) == 0 {
			return missingHardwareSelectorErr{
				Name: spec.ExternalEtcdMachineConfig().Name,
			}
		}
	}

	return nil
}

type missingHardwareSelectorErr struct {
	Name string
}

func (e missingHardwareSelectorErr) Error() string {
	return fmt.Sprintf("missing hardware selector for %v", e.Name)
}
