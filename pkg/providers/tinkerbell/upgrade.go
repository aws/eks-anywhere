package tinkerbell

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	rufiov1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/rufio"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/collection"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/rufiounreleased"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/yaml"
)

func needsNewControlPlaneTemplate(oldSpec, newSpec *cluster.Spec) bool {
	// Another option is to generate MachineTemplates based on the old and new eksa spec,
	// remove the name field and compare them with DeepEqual
	// We plan to approach this way since it's more flexible to add/remove fields and test out for validation
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}

	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}

	return false
}

func needsNewWorkloadTemplate(oldSpec, newSpec *cluster.Spec, oldWorker, newWorker v1alpha1.WorkerNodeGroupConfiguration) bool {
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}

	if !v1alpha1.TaintsSliceEqual(oldWorker.Taints, newWorker.Taints) ||
		!v1alpha1.MapEqual(oldWorker.Labels, newWorker.Labels) ||
		!v1alpha1.WorkerNodeGroupConfigurationKubeVersionUnchanged(&oldWorker, &newWorker, oldSpec.Cluster, newSpec.Cluster) {
		return true
	}

	return false
}

func needsNewKubeadmConfigTemplate(newWorkerNodeGroup, oldWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration) bool {
	return !v1alpha1.TaintsSliceEqual(newWorkerNodeGroup.Taints, oldWorkerNodeGroup.Taints) || !v1alpha1.MapEqual(newWorkerNodeGroup.Labels, oldWorkerNodeGroup.Labels)
}

func (p *Provider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, currentClusterSpec *cluster.Spec) error {
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		return errExternalEtcdUnsupported
	}

	if err := p.configureSshKeys(); err != nil {
		return err
	}

	// If we've been given a CSV with additional hardware for the cluster, validate it and
	// write it to the catalogue so it can be used for further processing.
	if p.hardwareCSVIsProvided() {
		machineCatalogueWriter := hardware.NewMachineCatalogueWriter(p.catalogue)

		machines, err := hardware.NewNormalizedCSVReaderFromFile(p.hardwareCSVFile, p.BMCOptions)
		if err != nil {
			return err
		}

		machineValidator := hardware.NewDefaultMachineValidator()

		if err := hardware.TranslateAll(machines, machineCatalogueWriter, machineValidator); err != nil {
			return err
		}
	}

	// Retrieve all unprovisioned hardware from the existing cluster and populate the catalogue so
	// it can be considered for the upgrade.
	hardware, err := p.providerKubectlClient.GetUnprovisionedTinkerbellHardware(
		ctx,
		cluster.KubeconfigFile,
		constants.EksaSystemNamespace,
	)
	if err != nil {
		return fmt.Errorf("retrieving unprovisioned hardware: %v", err)
	}
	for i := range hardware {
		if err := p.catalogue.InsertHardware(&hardware[i]); err != nil {
			return err
		}
	}

	// Retrieve all provisioned hardware from the existing cluster and populate diskExtractors's
	// disksProvisionedHardware map for use during upgrade
	hardware, err = p.providerKubectlClient.GetProvisionedTinkerbellHardware(
		ctx,
		cluster.KubeconfigFile,
		constants.EksaSystemNamespace,
	)
	if err != nil {
		return fmt.Errorf("retrieving provisioned hardware: %v", err)
	}

	// Remove all the provisioned hardware from the existing cluster if repeated from the hardware csv input.
	if err := p.catalogue.RemoveHardwares(hardware); err != nil {
		return err
	}

	upgradeStrategy := currentClusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy
	// skip extra hardware validation for InPlace upgrades
	if upgradeStrategy == nil || upgradeStrategy.Type != v1alpha1.InPlaceStrategyType {
		if err := p.validateAvailableHardwareForUpgrade(ctx, currentClusterSpec, clusterSpec); err != nil {
			return err
		}
	}

	if p.clusterConfig.IsManaged() {

		// Update stack helm enviorment variable NO_PROXY value and append management cluster's Control plane Endpoint IP in case of workload cluster upgrade
		if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
			managementCluster, err := p.providerKubectlClient.GetEksaCluster(ctx, clusterSpec.ManagementCluster, clusterSpec.Cluster.Spec.ManagementCluster.Name)
			if err != nil {
				return err
			}
			p.stackInstaller.AddNoProxyIP(managementCluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
		}

		if err := p.applyHardwareUpgrade(ctx, clusterSpec.ManagementCluster); err != nil {
			return err
		}
		if p.catalogue.TotalHardware() > 0 && p.catalogue.AllHardware()[0].Spec.BMCRef != nil {
			err = p.providerKubectlClient.WaitForRufioMachines(ctx, cluster, "5m", "Contactable", constants.EksaSystemNamespace)
			if err != nil {
				return fmt.Errorf("waiting for baseboard management to be contactable: %v", err)
			}
		}
	}

	return nil
}

// SetupAndValidateUpgradeManagementComponents performs necessary setup for upgrade management components operation.
func (p *Provider) SetupAndValidateUpgradeManagementComponents(_ context.Context, _ *cluster.Spec) error {
	return nil
}

func (p *Provider) validateAvailableHardwareForUpgrade(ctx context.Context, currentSpec, newClusterSpec *cluster.Spec) (err error) {
	clusterSpecValidator := NewClusterSpecValidator(
		HardwareSatisfiesOnlyOneSelectorAssertion(p.catalogue),
	)
	eksaVersionUpgrade := currentSpec.Bundles.Spec.Number != newClusterSpec.Bundles.Spec.Number

	currentTinkerbellSpec := NewClusterSpec(currentSpec, currentSpec.TinkerbellMachineConfigs, currentSpec.TinkerbellDatacenter)
	rollingUpgrade := p.isRollingUpgrade(currentSpec, newClusterSpec)
	currentCluster := &ValidatableTinkerbellClusterSpec{currentTinkerbellSpec}
	if rollingUpgrade || eksaVersionUpgrade {
		clusterSpecValidator.Register(ExtraHardwareAvailableAssertionForRollingUpgrade(p.catalogue, currentCluster, eksaVersionUpgrade))
	}
	// ScaleUpDown should not be supported in case of either rolling upgrade or eksa version upgrade.
	clusterSpecValidator.Register(AssertionsForScaleUpDown(p.catalogue, currentCluster, rollingUpgrade || eksaVersionUpgrade))

	tinkerbellClusterSpec := NewClusterSpec(newClusterSpec, p.machineConfigs, p.datacenterConfig)

	if err := clusterSpecValidator.Validate(tinkerbellClusterSpec); err != nil {
		return err
	}

	return nil
}

// PostBootstrapDeleteForUpgrade runs any provider-specific operations after bootstrap cluster has been deleted.
func (p *Provider) PostBootstrapDeleteForUpgrade(ctx context.Context, cluster *types.Cluster) error {
	if err := p.stackInstaller.UninstallLocal(ctx); err != nil {
		return err
	}
	return nil
}

func (p *Provider) PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
	return p.applyHardwareUpgrade(ctx, cluster)
}

// ApplyHardwareToCluster adds all the hardwares to the cluster.
func (p *Provider) applyHardwareUpgrade(ctx context.Context, cluster *types.Cluster) error {
	allHardware := p.catalogue.AllHardware()
	if len(allHardware) == 0 {
		return nil
	}
	hardwareSpec, err := hardware.MarshalCatalogue(p.catalogue)
	if err != nil {
		return fmt.Errorf("failed marshalling resources for hardware spec: %v", err)
	}
	err = p.providerKubectlClient.ApplyKubeSpecFromBytesForce(ctx, cluster, hardwareSpec)
	if err != nil {
		return fmt.Errorf("applying hardware yaml: %v", err)
	}

	return nil
}

func (p *Provider) PostMoveManagementToBootstrap(ctx context.Context, bootstrapCluster *types.Cluster) error {
	// Check if the hardware in the catalogue have a BMCRef. Since we only allow either all hardware with bmc
	// or no hardware with bmc, its sufficient to check the first hardware.
	if p.catalogue.TotalHardware() > 0 && p.catalogue.AllHardware()[0].Spec.BMCRef != nil {
		// Waiting to ensure all the new and exisiting baseboardmanagement connections are valid.
		err := p.providerKubectlClient.WaitForRufioMachines(ctx, bootstrapCluster, "5m", "Contactable", constants.EksaSystemNamespace)
		if err != nil {
			return fmt.Errorf("waiting for baseboard management to be contactable: %v", err)
		}
	}

	return nil
}

func (p *Provider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// @TODO: do we need this for bare metal upgrade?

	// Use retrier so that cluster upgrade does not fail due to any intermittent failure while connecting to kube-api server

	// This is unfortunate, but ClusterResourceSet's don't support any type of reapply of the resources they manage
	// Even if we create a new ClusterResourceSet, if such resources already exist in the cluster, they won't be reapplied
	// The long term solution is to add this capability to the cluster-api controller,
	// with a new mode like "ReApplyOnChanges" or "ReApplyOnCreate" vs the current "ReApplyOnce"
	/* err := p.retrier.Retry(
		func() error {
			return p.resourceSetManager.ForceUpdate(ctx, resourceSetName(clusterSpec), constants.EksaSystemNamespace, managementCluster, workloadCluster)
		},
	)
	if err != nil {
		return fmt.Errorf("failed updating the tinkerbell provider resource set post upgrade: %v", err)
	} */
	return nil
}

// ValidateNewSpec satisfies the Provider interface.
func (p *Provider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	desiredClstrSpec := clusterSpec

	currentClstr, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, desiredClstrSpec.Cluster.Name)
	if err != nil {
		return err
	}

	currentDCName := currentClstr.Spec.DatacenterRef.Name
	currentDCCfg, err := p.providerKubectlClient.GetEksaTinkerbellDatacenterConfig(ctx, currentDCName, cluster.KubeconfigFile, currentClstr.Namespace)
	if err != nil {
		return err
	}

	currentWNGs := currentClstr.Spec.WorkerNodeGroupConfigurations
	desiredWNGs := desiredClstrSpec.Cluster.Spec.WorkerNodeGroupConfigurations

	err = p.validateMachineCfgsImmutability(ctx, cluster, currentClstr, desiredClstrSpec, currentWNGs, desiredWNGs)
	if err != nil {
		return err
	}

	desiredDCCfgSpec := p.datacenterConfig.Spec

	if desiredDCCfgSpec.TinkerbellIP != currentDCCfg.Spec.TinkerbellIP {
		return fmt.Errorf("spec.tinkerbellIP is immutable; previous = %s, new = %s",
			currentDCCfg.Spec.TinkerbellIP, desiredDCCfgSpec.TinkerbellIP)
	}

	// for any operation other than k8s version change, hookImageURL is immutable
	if currentClstr.Spec.KubernetesVersion == desiredClstrSpec.Cluster.Spec.KubernetesVersion {
		if desiredDCCfgSpec.HookImagesURLPath != currentDCCfg.Spec.HookImagesURLPath {
			return fmt.Errorf("spec.hookImagesURLPath is immutable. previoius = %s, new = %s",
				currentDCCfg.Spec.HookImagesURLPath, desiredDCCfgSpec.HookImagesURLPath)
		}
	}

	return nil
}

func (p *Provider) validateMachineCfgsImmutability(ctx context.Context, clstr *types.Cluster, currentClstr *v1alpha1.Cluster, desiredClstrSpec *cluster.Spec, currentWNGs, desiredWNGs []v1alpha1.WorkerNodeGroupConfiguration) error {
	currentCPMachineRef := currentClstr.Spec.ControlPlaneConfiguration.MachineGroupRef
	desiredCPMachineRef := desiredClstrSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef
	if !currentCPMachineRef.Equal(desiredCPMachineRef) {
		return errors.New("control plane machine config reference is immutable")
	}

	err := validateRefsUnchanged(currentWNGs, desiredWNGs)
	if err != nil {
		return err
	}

	currentMachineCfgRefsMap := p.machineConfigs

	currentWNGsSet := collection.MapSet(currentWNGs, func(v v1alpha1.WorkerNodeGroupConfiguration) string {
		return v.Name
	})

	// newWNGs contains the set of worker node group names specified in the desired spec that are new.
	newWNGs := collection.NewSet[string]()
	for _, wng := range desiredWNGs {
		if !currentWNGsSet.Contains(wng.Name) {
			newWNGs.Add(wng.MachineGroupRef.Name)
		}
	}

	for _, machineCfgRef := range desiredClstrSpec.Cluster.MachineConfigRefs() {
		machineCfg, ok := currentMachineCfgRefsMap[machineCfgRef.Name]
		if !ok {
			return fmt.Errorf("cannot find machine config %s in tinkerbell provider machine configs", machineCfgRef.Name)
		}

		// If the machine config reference is for a new worker node group don't bother with
		// immutability checks as we want users to be able to add worker node groups.
		if !newWNGs.Contains(machineCfgRef.Name) {
			if _, has := currentMachineCfgRefsMap[machineCfg.Name]; !has {
				return fmt.Errorf("cannot change machine config references")
			}
			err := p.validateMachineCfg(ctx, clstr, machineCfg, desiredClstrSpec)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func validateRefsUnchanged(current, desired []v1alpha1.WorkerNodeGroupConfiguration) error {
	lookup := collection.ToMap(desired, func(v v1alpha1.WorkerNodeGroupConfiguration) string {
		return v.Name
	})

	var errs []error

	// For every current worker node group that still exists in the desired config, ensure the
	// machine config is still the same.
	for _, curr := range current {
		desi, exists := lookup[curr.Name]
		if !exists {
			continue
		}

		if !curr.MachineGroupRef.Equal(desi.MachineGroupRef) {
			errs = append(errs, fmt.Errorf("%v: worker node group machine config reference is immutable", curr.Name))
		}
	}

	return kerrors.NewAggregate(errs)
}

func (p *Provider) UpgradeNeeded(_ context.Context, _, _ *cluster.Spec, _ *types.Cluster) (bool, error) {
	// TODO: Figure out if something is needed here
	return false, nil
}

func (p *Provider) hardwareCSVIsProvided() bool {
	return p.hardwareCSVFile != ""
}

func (p *Provider) isScaleUpDown(oldCluster *v1alpha1.Cluster, newCluster *v1alpha1.Cluster) bool {
	if oldCluster.Spec.ControlPlaneConfiguration.Count != newCluster.Spec.ControlPlaneConfiguration.Count {
		return true
	}

	workerNodeGroupMap := make(map[string]*v1alpha1.WorkerNodeGroupConfiguration)
	for i := range oldCluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMap[oldCluster.Spec.WorkerNodeGroupConfigurations[i].Name] = &oldCluster.Spec.WorkerNodeGroupConfigurations[i]
	}

	for _, nodeGroupNewSpec := range newCluster.Spec.WorkerNodeGroupConfigurations {
		if workerNodeGrpOldSpec, ok := workerNodeGroupMap[nodeGroupNewSpec.Name]; ok {
			if *nodeGroupNewSpec.Count != *workerNodeGrpOldSpec.Count {
				return true
			}
		}
	}

	return false
}

func (p *Provider) isRollingUpgrade(currentSpec, newClusterSpec *cluster.Spec) bool {
	if currentSpec.Cluster.Spec.KubernetesVersion != newClusterSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	currentWNGSwithK8sVersion := WorkerNodeGroupWithK8sVersion(currentSpec)
	desiredWNGwithK8sVersion := WorkerNodeGroupWithK8sVersion(newClusterSpec)
	for wngName, K8sVersion := range desiredWNGwithK8sVersion {
		currentWngK8sVersion, ok := currentWNGSwithK8sVersion[wngName]
		if ok && (currentWngK8sVersion != K8sVersion) {
			return true
		}
	}
	return false
}

// WorkerNodeGroupWithK8sVersion maps each worker node group configurations in s to its K8s version.
func WorkerNodeGroupWithK8sVersion(spec *cluster.Spec) map[string]v1alpha1.KubernetesVersion {
	WNGwithK8sVersion := make(map[string]v1alpha1.KubernetesVersion)
	K8sVersion := spec.Cluster.Spec.KubernetesVersion
	for _, wng := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		mdName := fmt.Sprintf("%s-%s", spec.Cluster.Name, wng.Name)
		if wng.KubernetesVersion != nil {
			K8sVersion = *wng.KubernetesVersion
		}
		WNGwithK8sVersion[mdName] = K8sVersion
	}
	return WNGwithK8sVersion
}

func (p *Provider) validateMachineCfg(ctx context.Context, cluster *types.Cluster, newConfig *v1alpha1.TinkerbellMachineConfig, clusterSpec *cluster.Spec) error {
	prevMachineConfig, err := p.providerKubectlClient.GetEksaTinkerbellMachineConfig(ctx, newConfig.Name, cluster.KubeconfigFile, clusterSpec.Cluster.Namespace)
	if err != nil {
		return err
	}

	if newConfig.Spec.OSFamily != prevMachineConfig.Spec.OSFamily {
		return fmt.Errorf("spec.osFamily is immutable. Previous value %v,   New value %v", prevMachineConfig.Spec.OSFamily, newConfig.Spec.OSFamily)
	}

	if newConfig.Spec.Users[0].SshAuthorizedKeys[0] != prevMachineConfig.Spec.Users[0].SshAuthorizedKeys[0] {
		return fmt.Errorf("spec.Users[0].SshAuthorizedKeys[0] is immutable. Previous value %s,   New value %s", prevMachineConfig.Spec.Users[0].SshAuthorizedKeys[0], newConfig.Spec.Users[0].SshAuthorizedKeys[0])
	}

	if newConfig.Spec.Users[0].Name != prevMachineConfig.Spec.Users[0].Name {
		return fmt.Errorf("spec.Users[0].Name is immutable. Previous value %s,   New value %s", prevMachineConfig.Spec.Users[0].Name, newConfig.Spec.Users[0].Name)
	}

	if !reflect.DeepEqual(newConfig.Spec.HardwareSelector, prevMachineConfig.Spec.HardwareSelector) {
		return fmt.Errorf("spec.HardwareSelector is immutable. Previous value %v,   New value %v", prevMachineConfig.Spec.HardwareSelector, newConfig.Spec.HardwareSelector)
	}

	return nil
}

// PreCoreComponentsUpgrade staisfies the Provider interface.
func (p *Provider) PreCoreComponentsUpgrade(
	ctx context.Context,
	cluster *types.Cluster,
	managementComponents *cluster.ManagementComponents,
	clusterSpec *cluster.Spec,
) error {
	// When a workload cluster the cluster object could be nil. Noop if it is.
	if cluster == nil {
		return nil
	}

	if clusterSpec == nil {
		return errors.New("cluster spec is nil")
	}

	// PreCoreComponentsUpgrade can be called for workload clusters. Ensure we only attempt to
	// upgrade the stack if we're upgrading a management cluster.
	if clusterSpec.Cluster.IsManaged() {
		return nil
	}

	// Attempt the upgrade. This should upgrade the stack in the mangement cluster by updating
	// images, installing new CRDs and possibly removing old ones.
	err := p.stackInstaller.Upgrade(
		ctx,
		managementComponents.Tinkerbell,
		p.datacenterConfig.Spec.TinkerbellIP,
		cluster.KubeconfigFile,
		p.datacenterConfig.Spec.HookImagesURLPath,
		stack.WithLoadBalancerEnabled(
			len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations) != 0 && // load balancer is handled by kube-vip in control plane nodes
				!p.datacenterConfig.Spec.SkipLoadBalancerDeployment), // configure load balancer based on datacenterConfig.Spec.SkipLoadBalancerDeployment
	)
	if err != nil {
		return fmt.Errorf("upgrading stack: %v", err)
	}

	hasBaseboardManagement, err := p.providerKubectlClient.HasCRD(
		ctx,
		rufiounreleased.BaseboardManagementResourceName,
		cluster.KubeconfigFile,
	)
	if err != nil {
		return fmt.Errorf("upgrading rufio crds: %v", err)
	}

	// We introduced the Rufio dependency prior to its initial release. Between its introduction
	// and its official release breaking changes occured to the CRDs. We're using the presence
	// of the obsolete BaseboardManagement CRD to determine if there's an old Rufio installed.
	// If there is, we need to convert all obsolete BaseboardManagement CRs to Machine CRs (the
	// CRD that superseeds BaseboardManagement).
	if hasBaseboardManagement {
		if err := p.handleRufioUnreleasedCRDs(ctx, cluster); err != nil {
			return fmt.Errorf("upgrading rufio crds: %v", err)
		}

		// Remove the unreleased Rufio CRDs from the cluster; this will also remove any residual
		// resources.
		err = p.providerKubectlClient.DeleteCRD(
			ctx,
			rufiounreleased.BaseboardManagementResourceName,
			cluster.KubeconfigFile,
		)
		if err != nil {
			return fmt.Errorf("could not delete machines crd: %v", err)
		}
	}

	return nil
}

func (p *Provider) handleRufioUnreleasedCRDs(ctx context.Context, cluster *types.Cluster) error {
	// Firstly, retrieve all BaseboardManagement CRs and convert them to Machine CRs.
	bm, err := p.providerKubectlClient.AllBaseboardManagements(
		ctx,
		cluster.KubeconfigFile,
	)
	if err != nil {
		return fmt.Errorf("retrieving baseboardmanagement resources: %v", err)
	}

	if len(bm) > 0 {
		serialized, err := yaml.Serialize(toRufioMachines(bm)...)
		if err != nil {
			return fmt.Errorf("serializing machines: %v", err)
		}

		err = p.providerKubectlClient.ApplyKubeSpecFromBytesWithNamespace(
			ctx,
			cluster,
			yaml.Join(serialized),
			p.stackInstaller.GetNamespace(),
		)
		if err != nil {
			return fmt.Errorf("applying machines: %v", err)
		}
	}

	// Secondly, iterate over all Hardwarfe CRs and update the BMCRef to point to the new Machine
	// CR.
	hardware, err := p.providerKubectlClient.AllTinkerbellHardware(ctx, cluster.KubeconfigFile)
	if err != nil {
		return fmt.Errorf("retrieving hardware resources: %v", err)
	}

	var updatedHardware []tinkv1alpha1.Hardware
	for _, h := range hardware {
		if h.Spec.BMCRef != nil {
			h.Spec.BMCRef.Kind = "Machine"
			updatedHardware = append(updatedHardware, h)
		}
	}

	if len(updatedHardware) > 0 {
		serialized, err := yaml.Serialize(updatedHardware...)
		if err != nil {
			return fmt.Errorf("serializing hardware: %v", err)
		}

		err = p.providerKubectlClient.ApplyKubeSpecFromBytesForce(ctx, cluster, yaml.Join(serialized))
		if err != nil {
			return fmt.Errorf("applying hardware: %v", err)
		}
	}

	return nil
}

func toRufioMachines(items []rufiounreleased.BaseboardManagement) []rufiov1.Machine {
	var machines []rufiov1.Machine
	for _, item := range items {
		machines = append(machines, rufiov1.Machine{
			// We need to populate type meta because we apply with kubectl (leakage).
			TypeMeta: metav1.TypeMeta{
				Kind:       "Machine",
				APIVersion: rufiov1.GroupVersion.String(),
			},
			ObjectMeta: item.ObjectMeta,
			Spec: rufiov1.MachineSpec{
				Connection: rufiov1.Connection{
					AuthSecretRef: item.Spec.Connection.AuthSecretRef,
					Host:          item.Spec.Connection.Host,
					Port:          item.Spec.Connection.Port,
					InsecureTLS:   item.Spec.Connection.InsecureTLS,
				},
			},
		})
	}
	return machines
}
