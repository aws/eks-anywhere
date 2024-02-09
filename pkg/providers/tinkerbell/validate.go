package tinkerbell

import (
	"errors"
	"fmt"
	"strings"

	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func validateOsFamily(spec *ClusterSpec) error {
	controlPlaneRef := spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef
	controlPlaneOsFamily := spec.MachineConfigs[controlPlaneRef.Name].OSFamily()

	if spec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineRef := spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef
		if spec.MachineConfigs[etcdMachineRef.Name].OSFamily() != controlPlaneOsFamily {
			return errors.New("etcd osFamily cannot be different from control plane osFamily")
		}
	}

	for _, group := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		groupRef := group.MachineGroupRef
		if spec.MachineConfigs[groupRef.Name].OSFamily() != controlPlaneOsFamily {
			return errors.New("worker node group osFamily cannot be different from control plane osFamily")
		}
	}

	if controlPlaneOsFamily != v1alpha1.Bottlerocket && spec.DatacenterConfig.Spec.OSImageURL == "" && spec.ControlPlaneMachineConfig().Spec.OSImageURL == "" {
		return errors.New("please use bottlerocket as osFamily for auto-importing or provide a valid osImageURL")
	}

	return nil
}

func validateUpgradeRolloutStrategy(spec *ClusterSpec) error {
	cpUpgradeRolloutStrategyType := v1alpha1.RollingUpdateStrategyType

	if spec.ControlPlaneConfiguration().UpgradeRolloutStrategy != nil {
		cpUpgradeRolloutStrategyType = spec.ControlPlaneConfiguration().UpgradeRolloutStrategy.Type
		controlPlaneRef := spec.ControlPlaneConfiguration().MachineGroupRef
		controlPlaneOsFamily := spec.MachineConfigs[controlPlaneRef.Name].OSFamily()

		if controlPlaneOsFamily != v1alpha1.Ubuntu && cpUpgradeRolloutStrategyType == v1alpha1.InPlaceStrategyType {
			return errors.New("InPlace upgrades are only supported on the Ubuntu OS family")
		}
	}

	for _, group := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		wnUpgradeRolloutStrategyType := v1alpha1.RollingUpdateStrategyType
		groupRef := group.MachineGroupRef

		if group.UpgradeRolloutStrategy != nil {
			wnUpgradeRolloutStrategyType = group.UpgradeRolloutStrategy.Type
			if spec.MachineConfigs[groupRef.Name].OSFamily() != v1alpha1.Ubuntu && wnUpgradeRolloutStrategyType == v1alpha1.InPlaceStrategyType {
				return errors.New("InPlace upgrades are only supported on the Ubuntu OS family")
			}
		}
		if wnUpgradeRolloutStrategyType != cpUpgradeRolloutStrategyType {
			return errors.New("cannot specify different upgrade rollout strategy types for control plane and worker node group configurations")
		}
	}
	return nil
}

func validateAutoScalerDisabledForInPlace(spec *ClusterSpec) error {
	cpUpgradeRolloutStrategyType := spec.ControlPlaneConfiguration().UpgradeRolloutStrategy
	// We do not support different strategy types for Inplace between CP and worker nodes so it is okay to check only CP
	if cpUpgradeRolloutStrategyType == nil || cpUpgradeRolloutStrategyType.Type != v1alpha1.InPlaceStrategyType {
		return nil
	}

	for _, wng := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if wng.AutoScalingConfiguration != nil {
			return errors.New("austoscaler configuration not supported with InPlace upgrades")
		}
	}
	return nil
}

func validateOSImageURL(spec *ClusterSpec) error {
	dcOSImageURL := spec.DatacenterConfig.Spec.OSImageURL
	for _, mc := range spec.MachineConfigs {
		if mc.Spec.OSImageURL != "" && dcOSImageURL != "" {
			return errors.New("cannot specify OSImageURL on both TinkerbellMachineConfig's and TinkerbellDatacenterConfig")
		}
		if mc.Spec.OSImageURL == "" && dcOSImageURL == "" && mc.Spec.OSFamily != v1alpha1.Bottlerocket {
			return fmt.Errorf("missing OSImageURL on TinkerbellMachineConfig '%s'", mc.ObjectMeta.Name)
		}
	}
	return validateK8sVersionInOSImageURLs(spec)
}

func validateK8sVersionInOSImageURLs(spec *ClusterSpec) error {
	// If the user specifies the OSImageURL via the datacenter config then ensure all kube versions specified
	// on the cluster config are specified in the OSImageURL as the user could technically use a single image.
	//
	// When the user specifies OSImageURLs on each individual machine config (typical for modular upgrades) ensure
	// each machine config OSImageURL specifies the Kubernetes version. We don't explicitly take into consideration
	// the fact control plane, etcd and worker node groups can all reference the same machine config. If 2 components
	// specify different kube versions this will ensure both are present in the image URL (as above).
	if spec.DatacenterConfig.Spec.OSImageURL != "" {
		kvs := spec.Cluster.KubernetesVersions()
		for _, v := range kvs {
			if !containsK8sVersion(spec.DatacenterConfig.Spec.OSImageURL, string(v)) {
				return fmt.Errorf("missing kube version from OSImageURL: url=%v, version=%v",
					spec.DatacenterConfig.Spec.OSImageURL, v)
			}
		}
	} else {
		// For Bottlerocket we vend images but we still allow the user to specify them if they wish. We only want
		// to default machine config OSImageURLs if the datacenter config doesn't specify one and we default
		// to whatever is in the bundle.
		//
		// TODO: Investigate how we could refactor our logic to make this unnecessary.
		//
		// We validate elsewhere that all machine configs specify the same OSFamily so we can rely on the
		// control plane machine config only for the need to default OSImageURLs.
		if spec.ControlPlaneMachineConfig().OSFamily() == v1alpha1.Bottlerocket {
			defaultBottlerocketOSImageURLs(spec)
		}

		if !containsK8sVersion(spec.ControlPlaneMachineConfig().Spec.OSImageURL, string(spec.Cluster.Spec.KubernetesVersion)) {
			return fmt.Errorf("missing kube version from control plane machine config OSImageURL: url=%v, version=%v",
				spec.ControlPlaneMachineConfig().Spec.OSImageURL, spec.Cluster.Spec.KubernetesVersion)
		}

		for _, wng := range spec.WorkerNodeGroupConfigurations() {
			url := spec.MachineConfigs[wng.MachineGroupRef.Name].Spec.OSImageURL
			version := spec.Cluster.Spec.KubernetesVersion
			if wng.KubernetesVersion != nil && *wng.KubernetesVersion != "" {
				version = *wng.KubernetesVersion
			}

			if !containsK8sVersion(url, string(version)) {
				return fmt.Errorf("missing kube version from worker node group machine config OSImageURL: url=%v, version=%v",
					url, version)
			}
		}
	}
	return nil
}

func defaultBottlerocketOSImageURLs(spec *ClusterSpec) {
	if spec.ControlPlaneMachineConfig().Spec.OSImageURL == "" {
		spec.ControlPlaneMachineConfig().Spec.OSImageURL = spec.RootVersionsBundle().EksD.Raw.Bottlerocket.URI
	}
	for _, wng := range spec.WorkerNodeGroupConfigurations() {
		mc := spec.MachineConfigs[wng.MachineGroupRef.Name]
		version := spec.Cluster.Spec.KubernetesVersion
		if wng.KubernetesVersion != nil {
			version = *wng.KubernetesVersion
		}
		if mc.Spec.OSImageURL == "" {
			mc.Spec.OSImageURL = spec.VersionsBundle(version).EksD.Raw.Bottlerocket.URI
		}
	}
}

func containsK8sVersion(imageURL, k8sVersion string) bool {
	versionExtractor := strings.NewReplacer("-", "", ".", "", "_", "")
	osImageURL := versionExtractor.Replace(imageURL)
	kubeVersion := versionExtractor.Replace(k8sVersion)
	// we set the containsK8sVersion to false if the OS image URL does not contain the specified kubernetes version.
	// For ex if the kubernetes version is 1.23,
	// the image url should include 1.23 or 1-23, 1_23 or 123 i.e. ubuntu-1-23.gz or similar in the string.
	return strings.Contains(osImageURL, kubeVersion)
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
// it stores requirements in a map where the key is derived from selectors. This ensures selectors
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

// validateMinimumHardwareRequirements validates all requirements can be satisfied using hardware
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

// validateHardwareSatisfiesOnlyOneSelector ensures hardware in allHardware meets one and only one
// selector in selectors. selectors uses the selectorSet construct to ensure we don't
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
