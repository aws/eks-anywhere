package tinkerbell

import (
	"errors"
	"fmt"

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

	if config.Spec.OSFamily == "" {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.osFamily: %v", config.Name)
	}

	if config.Spec.OSFamily != v1alpha1.Ubuntu && config.Spec.OSFamily != v1alpha1.Bottlerocket {
		return fmt.Errorf(
			"TinkerbellMachineConfig: unsupported spec.osFamily (%v); Please use one of the following: %s, %s",
			config.Spec.OSFamily,
			v1alpha1.Ubuntu,
			v1alpha1.Bottlerocket,
		)
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

// ValidateminimumHardwareRequirements validates all requirements can be satisfied using hardware
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
