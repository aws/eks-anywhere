package tinkerbell

import (
	"errors"
	"fmt"
	"strings"

	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func validateDatacenterConfig(config *v1alpha1.TinkerbellDatacenterConfig) error {
	if err := validateObjectMeta(config.ObjectMeta); err != nil {
		return fmt.Errorf("TinkerbellDatacenterConfig: %v", err)
	}

	if config.Spec.TinkerbellIP == "" {
		return errors.New("TinkerbellDatacenterConfig: missing spec.tinkerbellIP field")
	}

	if err := networkutils.ValidateIP(config.Spec.TinkerbellIP); err != nil {
		return fmt.Errorf("TinkerbellDatacenterConfig: invalid tinkerbell ip: %v", err)
	}

	return nil
}

func validateMachineConfig(config *v1alpha1.TinkerbellMachineConfig) error {
	if err := validateObjectMeta(config.ObjectMeta); err != nil {
		return fmt.Errorf("TinkerbellMachineConfig: %v", err)
	}

	if len(config.Spec.HardwareSelector) == 0 {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.hardwareSelector: %v", config.Name)
	}

	if len(config.Spec.HardwareSelector) != 1 {
		return fmt.Errorf(
			"TinkerbellMachineConfig: spec.hardwareSelector must contain only 1 key-value pair: %v",
			config.Name,
		)
	}

	if config.Spec.OSFamily == "" {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.osFamily: %v", config.Name)
	}

	if config.Spec.OSFamily != v1alpha1.Ubuntu && config.Spec.OSFamily != v1alpha1.Bottlerocket && config.Spec.OSFamily != v1alpha1.RedHat {
		return fmt.Errorf(
			"TinkerbellMachineConfig: unsupported spec.osFamily (%v); Please use one of the following: %s, %s, %s",
			config.Spec.OSFamily,
			v1alpha1.Ubuntu,
			v1alpha1.RedHat,
			v1alpha1.Bottlerocket,
		)
	}

	return nil
}

func validateOsFamily(spec *ClusterSpec) error {
	controlPlaneRef := spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef
	controlPlaneOsFamily := spec.MachineConfigs[controlPlaneRef.Name].OSFamily()

	if spec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineRef := spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef
		if spec.MachineConfigs[etcdMachineRef.Name].OSFamily() != controlPlaneOsFamily {
			return fmt.Errorf("etcd osFamily cannot be different from control plane osFamily")
		}
	}

	for _, group := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		groupRef := group.MachineGroupRef
		if spec.MachineConfigs[groupRef.Name].OSFamily() != controlPlaneOsFamily {
			return fmt.Errorf("worker node group osFamily cannot be different from control plane osFamily")
		}
	}

	if controlPlaneOsFamily != v1alpha1.Bottlerocket && spec.DatacenterConfig.Spec.OSImageURL == "" {
		return fmt.Errorf("please use bottlerocket as osFamily for auto-importing or provide a valid osImageURL")
	}
	return nil
}

func validateObjectMeta(meta metav1.ObjectMeta) error {
	if meta.Name == "" {
		return errors.New("missing name")
	}

	return nil
}

func validateMachineRefExists(
	ref *v1alpha1.Ref,
	machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig,
) error {
	if _, ok := machineConfigs[ref.Name]; !ok {
		return fmt.Errorf("missing machine config ref: kind=%v; name=%v", ref.Kind, ref.Name)
	}
	return nil
}

func validateMachineConfigNamespacesMatchDatacenterConfig(
	datacenterConfig *v1alpha1.TinkerbellDatacenterConfig,
	machineConfigs map[string]*v1alpha1.TinkerbellMachineConfig,
) error {
	for _, machineConfig := range machineConfigs {
		if machineConfig.Namespace != datacenterConfig.Namespace {
			return fmt.Errorf(
				"TinkerbellMachineConfig's namespace must match TinkerbellDatacenterConfig's namespace: %v",
				machineConfig.Name,
			)
		}
	}
	return nil
}

func validateIPUnused(client networkutils.NetClient, ip string) error {
	if networkutils.IsIPInUse(client, ip) {
		return fmt.Errorf("ip in use: %v", ip)
	}
	return nil
}

func validatePortsAvailable(client networkutils.NetClient, host string) error {
	unavailablePorts := getPortsUnavailable(client, host)

	if len(unavailablePorts) != 0 {
		return fmt.Errorf("localhost ports [%v] are already in use, please ensure these ports are available", strings.Join(unavailablePorts, ", "))
	}
	return nil
}

func getPortsUnavailable(client networkutils.NetClient, host string) []string {
	ports := []string{"80", "42113", "50061"}
	var unavailablePorts []string
	for _, port := range ports {
		if networkutils.IsPortInUse(client, host, port) {
			unavailablePorts = append(unavailablePorts, port)
		}
	}
	return unavailablePorts
}

// minimumHardwareRequirement defines the minimum requirement for a hardware selector.
type minimumHardwareRequirement struct {
	// MinCount is the minimum number of hardware required to satisfy the requirement
	MinCount int
	// Selector defines what labels should be present on Hardware to consider it eligable for
	// this requirement.
	Selector v1alpha1.HardwareSelector
	// count is used internally by validation to sum the actual available hardware.
	count int
}

// minimumHardwareRequirements is a collection of minimumHardwareRequirement instances.
// It stores requirements in a map where the key is derived from selectors. This ensures selectors
// specifying the same key-value pairs are combined.
type minimumHardwareRequirements map[string]*minimumHardwareRequirement

// Add a minimumHardwareRequirement to r.
func (r *minimumHardwareRequirements) Add(selector v1alpha1.HardwareSelector, min int) error {
	name, err := selector.ToString()
	if err != nil {
		return err
	}

	(*r)[name] = &minimumHardwareRequirement{
		MinCount: min,
		Selector: selector,
	}

	return nil
}

// validateminimumHardwareRequirements validates all requirements can be satisfied using hardware
// registered with catalogue.
func validateMinimumHardwareRequirements(requirements minimumHardwareRequirements, catalogue *hardware.Catalogue) error {
	// Count all hardware that meets the selector requirements for each requirement.
	// This does not consider whether or not a piece of hardware is selectable by multiple
	// selectors. That requires a different validation ideally run before this one.
	for _, h := range catalogue.AllHardware() {
		for _, r := range requirements {
			if hardware.LabelsMatchSelector(r.Selector, h.Labels) {
				r.count++
			}
		}
	}

	// Validate counts of hardware meet the minimum required count.
	for name, r := range requirements {
		if r.count < r.MinCount {
			return fmt.Errorf(
				"minimum hardware count not met for selector '%v': have %v, require %v",
				name,
				r.count,
				r.MinCount,
			)
		}
	}

	return nil
}

// validateHardwareSatifiesOnlyOneSelector ensures hardware in allHardware meets one and only one
// selector in selectors. Selectors uses the selectorSet construct to ensure we don't
// operate on duplicate selectors given a selector can be re-used among groups as they may reference
// the same TinkerbellMachineConfig.
func validateHardwareSatisfiesOnlyOneSelector(allHardware []*tinkv1alpha1.Hardware, selectors selectorSet) error {
	for _, h := range allHardware {
		if matches := getMatchingHardwareSelectors(h, selectors); len(matches) > 1 {
			slctrStrs, err := getHardwareSelectorsAsStrings(matches)
			if err != nil {
				return err
			}

			return fmt.Errorf(
				"hardware must only satisfy 1 selector: hardware name '%v'; selectors '%v'",
				h.Name,
				strings.Join(slctrStrs, ", "),
			)
		}
	}

	return nil
}

// selectorSet defines a set of selectors. Selectors should be added using the Add method to ensure
// deterministic key generation. The construct is useful to avoid treating selectors that are the
// same as different.
type selectorSet map[string]v1alpha1.HardwareSelector

// Add adds selector to ss.
func (ss *selectorSet) Add(selector v1alpha1.HardwareSelector) error {
	slctrStr, err := selector.ToString()
	if err != nil {
		return err
	}

	(*ss)[slctrStr] = selector

	return nil
}

func getMatchingHardwareSelectors(
	hw *tinkv1alpha1.Hardware,
	selectors selectorSet,
) []v1alpha1.HardwareSelector {
	var satisfies []v1alpha1.HardwareSelector
	for _, selector := range selectors {
		if hardware.LabelsMatchSelector(selector, hw.Labels) {
			satisfies = append(satisfies, selector)
		}
	}
	return satisfies
}

func getHardwareSelectorsAsStrings(selectors []v1alpha1.HardwareSelector) ([]string, error) {
	var slctrs []string
	for _, selector := range selectors {
		s, err := selector.ToString()
		if err != nil {
			return nil, err
		}
		slctrs = append(slctrs, s)
	}
	return slctrs, nil
}
