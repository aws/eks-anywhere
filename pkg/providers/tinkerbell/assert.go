package tinkerbell

import (
	"errors"
	"fmt"
	"net/http"

	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

// TODO(chrisdoherty) Add worker node group assertions

// AssertMachineConfigsValid iterates over all machine configs in calling validateMachineConfig.
func AssertMachineConfigsValid(spec *ClusterSpec) error {
	for _, config := range spec.MachineConfigs {
		if err := config.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// AssertDatacenterConfigValid asserts the DatacenterConfig in spec is valid.
func AssertDatacenterConfigValid(spec *ClusterSpec) error {
	return spec.DatacenterConfig.Validate()
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

// AssertK8SVersionNot120 ensures Kubernetes version is not set to v1.20.
func AssertK8SVersionNot120(spec *ClusterSpec) error {
	if spec.Cluster.Spec.KubernetesVersion == v1alpha1.Kube120 {
		return errors.New("kubernetes version v1.20 is not supported for Bare Metal")
	}

	return nil
}

func AssertOsFamilyValid(spec *ClusterSpec) error {
	return validateOsFamily(spec)
}

// AssertUpgradeRolloutStrategyValid ensures that the upgrade rollout strategy is valid for both CP and worker node configurations.
func AssertUpgradeRolloutStrategyValid(spec *ClusterSpec) error {
	return validateUpgradeRolloutStrategy(spec)
}

// AssertAutoScalerDisabledForInPlace ensures that the autoscaler configuration is not enabled when upgrade rollout strategy is InPlace.
func AssertAutoScalerDisabledForInPlace(spec *ClusterSpec) error {
	return validateAutoScalerDisabledForInPlace(spec)
}

// AssertOSImageURL ensures that the OSImageURL value is either set at the datacenter config level or set for each machine config and not at both levels.
func AssertOSImageURL(spec *ClusterSpec) error {
	return validateOSImageURL(spec)
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

// AssertTinkerbellIPNotInUse ensures tinkerbell ip isn't in use.
func AssertTinkerbellIPNotInUse(client networkutils.NetClient) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		ip := spec.DatacenterConfig.Spec.TinkerbellIP
		if err := validateIPUnused(client, ip); err != nil {
			return fmt.Errorf("tinkerbellIP <%s> is already in use, please provide a unique IP", ip)
		}
		return nil
	}
}

// AssertTinkerbellIPAndControlPlaneIPNotSame ensures tinkerbell ip and controlplane ip are not the same.
func AssertTinkerbellIPAndControlPlaneIPNotSame(spec *ClusterSpec) error {
	tinkerbellIP := spec.DatacenterConfig.Spec.TinkerbellIP
	controlPlaneIP := spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	if tinkerbellIP == controlPlaneIP {
		return fmt.Errorf("controlPlaneConfiguration.endpoint.host and tinkerbellIP are the same (%s), please provide two unique IPs", tinkerbellIP)
	}
	return nil
}

// AssertHookRetrievableWithoutProxy ensures the executing machine can retrieve Hook
// from the host URL without a proxy configured. It does not guarantee the target node
// will be able to download Hook.
func AssertHookRetrievableWithoutProxy(spec *ClusterSpec) error {
	if spec.Cluster.Spec.ProxyConfiguration == nil {
		return nil
	}

	// return an error if hookImagesURLPath field is not specified for during Proxy configuration.
	if spec.DatacenterConfig.Spec.HookImagesURLPath == "" {
		return fmt.Errorf("locally hosted hookImagesURLPath is required to support ProxyConfiguration")
	}

	// verify hookImagesURLPath is accessible locally too
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get(spec.DatacenterConfig.Spec.HookImagesURLPath)
	if err != nil {
		return fmt.Errorf("HookImagesURLPath: %s needs to be hosted locally while specifiying Proxy configuration: %v", spec.DatacenterConfig.Spec.HookImagesURLPath, err)
	}

	defer resp.Body.Close()

	return nil
}

// AssertPortsNotInUse ensures that ports 80, 42113, and 50061 are available.
func AssertPortsNotInUse(client networkutils.NetClient) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		host := "0.0.0.0"
		if err := validatePortsAvailable(client, host); err != nil {
			return err
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
				*nodeGroup.Count,
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

// WorkerNodeHardware holds machine deployment name, replica count and hardware selector for a Tinkerbell worker node.
type WorkerNodeHardware struct {
	MachineDeploymentName string
	Replicas              int
}

// ValidatableCluster allows assertions to pull worker node and control plane information.
type ValidatableCluster interface {
	// WorkerNodeHardwareGroups retrieves a list of WorkerNodeHardwares containing MachineDeployment name,
	// replica count and hardware selector for each worker node of a ValidatableCluster.
	WorkerNodeHardwareGroups() []WorkerNodeHardware

	// ControlPlaneReplicaCount retrieves the control plane replica count of the ValidatableCluster.
	ControlPlaneReplicaCount() int

	// ClusterK8sVersion retreives the Cluster level Kubernetes version
	ClusterK8sVersion() v1alpha1.KubernetesVersion

	// WorkerGroupK8sVersion maps each worker group with its Kubernetes version.
	WorkerNodeGroupK8sVersion() map[string]v1alpha1.KubernetesVersion
}

// ValidatableTinkerbellClusterSpec wraps around the Tinkerbell ClusterSpec as a ValidatableCluster.
type ValidatableTinkerbellClusterSpec struct {
	*ClusterSpec
}

// ControlPlaneReplicaCount retrieves the ValidatableTinkerbellClusterSpec control plane replica count.
func (v *ValidatableTinkerbellClusterSpec) ControlPlaneReplicaCount() int {
	return v.Cluster.Spec.ControlPlaneConfiguration.Count
}

// WorkerNodeHardwareGroups retrieves a list of WorkerNodeHardwares for a ValidatableTinkerbellClusterSpec.
func (v *ValidatableTinkerbellClusterSpec) WorkerNodeHardwareGroups() []WorkerNodeHardware {
	workerNodeGroupConfigs := make([]WorkerNodeHardware, 0, len(v.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroup := range v.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupConfig := &WorkerNodeHardware{
			MachineDeploymentName: machineDeploymentName(v.Cluster.Name, workerNodeGroup.Name),
			Replicas:              *workerNodeGroup.Count,
		}
		workerNodeGroupConfigs = append(workerNodeGroupConfigs, *workerNodeGroupConfig)
	}
	return workerNodeGroupConfigs
}

// ClusterK8sVersion retrieves the Kubernetes version set at the cluster level.
func (v *ValidatableTinkerbellClusterSpec) ClusterK8sVersion() v1alpha1.KubernetesVersion {
	return v.Cluster.Spec.KubernetesVersion
}

// WorkerNodeGroupK8sVersion returns each worker node group with its associated Kubernetes version.
func (v *ValidatableTinkerbellClusterSpec) WorkerNodeGroupK8sVersion() map[string]v1alpha1.KubernetesVersion {
	return WorkerNodeGroupWithK8sVersion(v.ClusterSpec.Spec)
}

// ValidatableTinkerbellCAPI wraps around the Tinkerbell control plane and worker CAPI obects as a ValidatableCluster.
type ValidatableTinkerbellCAPI struct {
	KubeadmControlPlane *controlplanev1.KubeadmControlPlane
	WorkerGroups        []*clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]
}

// ControlPlaneReplicaCount retrieves the ValidatableTinkerbellCAPI control plane replica count.
func (v *ValidatableTinkerbellCAPI) ControlPlaneReplicaCount() int {
	return int(*v.KubeadmControlPlane.Spec.Replicas)
}

// WorkerNodeHardwareGroups retrieves a list of WorkerNodeHardwares for a ValidatableTinkerbellCAPI.
func (v *ValidatableTinkerbellCAPI) WorkerNodeHardwareGroups() []WorkerNodeHardware {
	workerNodeHardwareList := make([]WorkerNodeHardware, 0, len(v.WorkerGroups))
	for _, workerGroup := range v.WorkerGroups {
		workerNodeHardware := &WorkerNodeHardware{
			MachineDeploymentName: workerGroup.MachineDeployment.Name,
			Replicas:              int(*workerGroup.MachineDeployment.Spec.Replicas),
		}
		workerNodeHardwareList = append(workerNodeHardwareList, *workerNodeHardware)
	}
	return workerNodeHardwareList
}

// ClusterK8sVersion returns the Kubernetes version in major.minor format for a ValidatableTinkerbellCAPI.
func (v *ValidatableTinkerbellCAPI) ClusterK8sVersion() v1alpha1.KubernetesVersion {
	return v.toK8sVersion(v.KubeadmControlPlane.Spec.Version)
}

// WorkerNodeGroupK8sVersion returns each worker node group mapped to Kubernetes version in major.minor format for a ValidatableTinkerbellCAPI.
func (v *ValidatableTinkerbellCAPI) WorkerNodeGroupK8sVersion() map[string]v1alpha1.KubernetesVersion {
	wngK8sversion := make(map[string]v1alpha1.KubernetesVersion)
	for _, wng := range v.WorkerGroups {
		k8sVersion := v.toK8sVersion(*wng.MachineDeployment.Spec.Template.Spec.Version)
		wngK8sversion[wng.MachineDeployment.Name] = k8sVersion
	}
	return wngK8sversion
}

func (v *ValidatableTinkerbellCAPI) toK8sVersion(k8sversion string) v1alpha1.KubernetesVersion {
	kubeVersion := v1alpha1.KubernetesVersion(k8sversion[1:5])
	return kubeVersion
}

// AssertionsForScaleUpDown asserts that catalogue has sufficient hardware to
// support the scaling up/down from current ClusterSpec to desired ValidatableCluster.
// nolint:gocyclo // TODO: Reduce cyclomatic complexity https://github.com/aws/eks-anywhere-internal/issues/1186
func AssertionsForScaleUpDown(catalogue *hardware.Catalogue, current ValidatableCluster, rollingUpgrade bool) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		// Without Hardware selectors we get undesirable behavior so ensure we have them for
		// all MachineConfigs.
		if err := ensureHardwareSelectorsSpecified(spec); err != nil {
			return err
		}

		if spec.HasExternalEtcd() {
			return fmt.Errorf("scale up/down not supported for external etcd")
		}
		// Build a set of required hardware counts per machine group. minimumHardwareRequirements
		// will account for the same selector being specified on different groups.
		requirements := minimumHardwareRequirements{}

		if current.ControlPlaneReplicaCount() != spec.Cluster.Spec.ControlPlaneConfiguration.Count {
			if rollingUpgrade {
				return fmt.Errorf("cannot perform scale up or down during rolling upgrades")
			}
			if current.ControlPlaneReplicaCount() < spec.Cluster.Spec.ControlPlaneConfiguration.Count {
				err := requirements.Add(
					spec.ControlPlaneMachineConfig().Spec.HardwareSelector,
					spec.Cluster.Spec.ControlPlaneConfiguration.Count-current.ControlPlaneReplicaCount(),
				)
				if err != nil {
					return fmt.Errorf("error during scale up: %v", err)
				}
			}
		}

		workerNodeHardwareMap := make(map[string]WorkerNodeHardware)
		for _, workerNodeHardware := range current.WorkerNodeHardwareGroups() {
			workerNodeHardwareMap[workerNodeHardware.MachineDeploymentName] = workerNodeHardware
		}

		for _, nodeGroupNewSpec := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
			nodeGroupMachineDeploymentNameNewSpec := machineDeploymentName(spec.Cluster.Name, nodeGroupNewSpec.Name)
			if workerNodeGroupOldSpec, ok := workerNodeHardwareMap[nodeGroupMachineDeploymentNameNewSpec]; ok {
				if *nodeGroupNewSpec.Count != workerNodeGroupOldSpec.Replicas {
					if rollingUpgrade {
						return fmt.Errorf("cannot perform scale up or down during rolling upgrades")
					}
					if *nodeGroupNewSpec.Count > workerNodeGroupOldSpec.Replicas {
						err := requirements.Add(
							spec.WorkerNodeGroupMachineConfig(nodeGroupNewSpec).Spec.HardwareSelector,
							*nodeGroupNewSpec.Count-workerNodeGroupOldSpec.Replicas,
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
					*nodeGroupNewSpec.Count,
				)
				if err != nil {
					return fmt.Errorf("error during scale up: %v", err)
				}
			}
		}

		if err := validateMinimumHardwareRequirements(requirements, catalogue); err != nil {
			return fmt.Errorf("for scale up, %v", err)
		}
		return nil
	}
}

// ExtraHardwareAvailableAssertionForRollingUpgrade asserts that catalogue has sufficient hardware to
// support the ClusterSpec during an rolling upgrade workflow.
func ExtraHardwareAvailableAssertionForRollingUpgrade(catalogue *hardware.Catalogue, current ValidatableCluster, eksaVersionUpgrade bool) ClusterSpecAssertion {
	return func(spec *ClusterSpec) error {
		// Without Hardware selectors we get undesirable behavior so ensure we have them for
		// all MachineConfigs.
		if err := ensureHardwareSelectorsSpecified(spec); err != nil {
			return err
		}

		// Build a set of required hardware counts per machine group. minimumHardwareRequirements
		// will account for the same selector being specified on different groups.
		requirements := minimumHardwareRequirements{}

		if spec.Cluster.Spec.KubernetesVersion != current.ClusterK8sVersion() || eksaVersionUpgrade {
			if err := ensureCPHardwareAvailability(spec, current, requirements); err != nil {
				return err
			}
		}

		if err := ensureWorkerHardwareAvailability(spec, current, requirements, eksaVersionUpgrade); err != nil {
			return err
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

func ensureCPHardwareAvailability(spec *ClusterSpec, current ValidatableCluster, hwReq minimumHardwareRequirements) error {
	maxSurge := 1

	rolloutStrategy := spec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy
	if rolloutStrategy != nil && rolloutStrategy.Type == "RollingUpdate" {
		maxSurge = spec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
	}
	err := hwReq.Add(
		spec.ControlPlaneMachineConfig().Spec.HardwareSelector,
		maxSurge,
	)
	if err != nil {
		return fmt.Errorf("for rolling upgrade, %v", err)
	}
	return nil
}

func ensureWorkerHardwareAvailability(spec *ClusterSpec, current ValidatableCluster, hwReq minimumHardwareRequirements, eksaVersionUpgrade bool) error {
	currentWngK8sversion := current.WorkerNodeGroupK8sVersion()
	desiredWngK8sVersion := WorkerNodeGroupWithK8sVersion(spec.Spec)
	for _, nodeGroup := range spec.WorkerNodeGroupConfigurations() {
		maxSurge := 1
		// As rolling upgrades and scale up/down is not permitted in a single operation, its safe to access directly using the md name.
		mdName := fmt.Sprintf("%s-%s", spec.Cluster.Name, nodeGroup.Name)
		if currentWngK8sversion[mdName] != desiredWngK8sVersion[mdName] || eksaVersionUpgrade {
			if nodeGroup.UpgradeRolloutStrategy != nil && nodeGroup.UpgradeRolloutStrategy.Type == "RollingUpdate" {
				maxSurge = nodeGroup.UpgradeRolloutStrategy.RollingUpdate.MaxSurge
			}
			err := hwReq.Add(
				spec.WorkerNodeGroupMachineConfig(nodeGroup).Spec.HardwareSelector,
				maxSurge,
			)
			if err != nil {
				return fmt.Errorf("for rolling upgrade, %v", err)
			}
		}
	}
	return nil
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
