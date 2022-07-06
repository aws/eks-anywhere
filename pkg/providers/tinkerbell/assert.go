package tinkerbell

import (
	"errors"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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

// MinimumHardwareForCreate asserts that catalogue has sufficient hardware to
// support the ClusterSpec during a create workflow.
//
// It does not protect against intersections or subsets so consumers should ensure a 1-2-1
// mapping between catalogue hardware and selectors.
func MinimumHardwareForCreate(catalogue *hardware.Catalogue) ClusterSpecAssertion {
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

// MinimumHardwareForUpgrade ensures there is sufficient hardware in catalogue to perform a
// Kubernetes upgrade operation. Operators cannot invoke a Kubernetes version upgrade with a
// scaling upgrade.
func MinimumHardwareForUpgrade(current *ClusterSpec, catalogue *hardware.Catalogue) ClusterSpecAssertion {
	return func(desired *ClusterSpec) error {
		isVersionUpgrade := isVersionUpgrade(current, desired)
		isScaleUpgrade := isScaleUpgrade(current, desired)

		if isVersionUpgrade && isScaleUpgrade {
			return errors.New("cannot upgrade kubernetes version and scale up/down simultaneously")
		}

		if err := ensureHardwareSelectorsSpecified(desired); err != nil {
			return err
		}

		if isVersionUpgrade {
			return validateHardwareForVersionUpgrade(current, desired, catalogue)
		}

		if isScaleUpgrade {
			return validateHardwareForScaleUpgrade(current, desired, catalogue)
		}

		// Noop as default indicating there isn't a change requiring more/less hardware.
		return nil
	}
}

// isScaleUpgrade returns true if there is a change in the clusters size on a per node group
// basis. For example, adjustment in control plane counts or individual worker node group counts.
//
// Changes to external etcd are unsupported.
func isScaleUpgrade(currentSpec, desiredSpec *ClusterSpec) bool {
	currentNodeGroups := buildNodeGroupMap(currentSpec.WorkerNodeGroupConfigurations())
	return hasWorkerNodeGroupDiffs(currentNodeGroups, desiredSpec.WorkerNodeGroupConfigurations()) ||
		hasControlPlaneDiff(currentSpec, desiredSpec)
}

// buildNodeGroupMap builds a map of node group name to node group config.
func buildNodeGroupMap(s []v1alpha1.WorkerNodeGroupConfiguration) map[string]v1alpha1.WorkerNodeGroupConfiguration {
	groups := map[string]v1alpha1.WorkerNodeGroupConfiguration{}
	for _, nodeGroup := range s {
		groups[nodeGroup.Name] = nodeGroup
	}
	return groups
}

// hasWorkerNodeGroupDiffs returns true if there is an increase between current and desired
// worker node group counts.
func hasWorkerNodeGroupDiffs(
	current map[string]v1alpha1.WorkerNodeGroupConfiguration,
	desired []v1alpha1.WorkerNodeGroupConfiguration,
) bool {
	// Ensure the group exists. If the group does exist check if there's a count diff.
	// If the group doesn't already exist its a new group which counts as a diff.
	for _, desiredGroup := range desired {
		current, exists := current[desiredGroup.Name]
		if !exists || desiredGroup.Count > current.Count {
			return true
		}
	}

	return false
}

func hasControlPlaneDiff(current, desired *ClusterSpec) bool {
	return current.ControlPlaneConfiguration().Count != desired.ControlPlaneConfiguration().Count
}

// isVersionUpgrade returns true if there is a difference in the current and desired Kubernetes
// versions. It does not check if the version was incremeneted.
func isVersionUpgrade(current, desired *ClusterSpec) bool {
	return current.Cluster.Spec.KubernetesVersion != desired.Cluster.Spec.KubernetesVersion
}

func validateHardwareForVersionUpgrade(current, desired *ClusterSpec, catalogue *hardware.Catalogue) error {
	requirements := minimumHardwareRequirements{}

	selector := current.ControlPlaneMachineConfig().Spec.HardwareSelector
	err := requirements.Add(selector, 1)
	if err != nil {
		return err
	}

	for _, nodeGroup := range current.WorkerNodeGroupConfigurations() {
		selector := current.WorkerNodeGroupMachineConfig(nodeGroup).Spec.HardwareSelector
		if err := requirements.Add(selector, 1); err != nil {
			return err
		}
	}

	if current.HasExternalEtcd() {
		selector := current.ExternalEtcdMachineConfig().Spec.HardwareSelector
		if err := requirements.Add(selector, 1); err != nil {
			return err
		}
	}

	return validateMinimumHardwareRequirements(requirements, catalogue)
}

func validateHardwareForScaleUpgrade(current, desired *ClusterSpec, catalogue *hardware.Catalogue) error {
	requirements := minimumHardwareRequirements{}

	controlPlaneDiff := desired.ControlPlaneConfiguration().Count - current.ControlPlaneConfiguration().Count
	if controlPlaneDiff > 0 {
		selector := desired.ControlPlaneMachineConfig().Spec.HardwareSelector
		err := requirements.Add(selector, controlPlaneDiff)
		if err != nil {
			return err
		}
	}

	currentWorkerNodeGroups := buildNodeGroupMap(current.WorkerNodeGroupConfigurations())

	for _, desiredNodeGroup := range desired.WorkerNodeGroupConfigurations() {
		currentNodeGroup, ok := currentWorkerNodeGroups[desiredNodeGroup.Name]

		// If its a new node group so we need the full hardware set.
		if !ok {
			selector := desired.WorkerNodeGroupMachineConfig(desiredNodeGroup).Spec.HardwareSelector
			if err := requirements.Add(selector, desiredNodeGroup.Count); err != nil {
				return err
			}
			continue
		}

		// If its an existing node group we need to add the difference in count.
		diff := desiredNodeGroup.Count - currentNodeGroup.Count
		if diff > 0 {
			selector := desired.WorkerNodeGroupMachineConfig(desiredNodeGroup).Spec.HardwareSelector
			if err := requirements.Add(selector, diff); err != nil {
				return err
			}
		}
	}

	return validateMinimumHardwareRequirements(requirements, catalogue)
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
