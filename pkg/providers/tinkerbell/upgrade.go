package tinkerbell

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/types"
)

func NeedsNewControlPlaneTemplate(oldSpec, newSpec *cluster.Spec, oldVdc, newVdc *v1alpha1.TinkerbellDatacenterConfig, oldTmc, newTmc *v1alpha1.TinkerbellMachineConfig) bool {
	// Another option is to generate MachineTemplates based on the old and new eksa spec,
	// remove the name field and compare them with DeepEqual
	// We plan to approach this way since it's more flexible to add/remove fields and test out for validation
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host != newSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}

	return AnyImmutableFieldChanged(oldVdc, newVdc, oldTmc, newTmc)
}

func NeedsNewWorkloadTemplate(oldSpec, newSpec *cluster.Spec, oldVdc, newVdc *v1alpha1.TinkerbellDatacenterConfig, oldTmc, newTmc *v1alpha1.TinkerbellMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	if !v1alpha1.WorkerNodeGroupConfigurationSliceTaintsEqual(oldSpec.Cluster.Spec.WorkerNodeGroupConfigurations, newSpec.Cluster.Spec.WorkerNodeGroupConfigurations) ||
		!v1alpha1.WorkerNodeGroupConfigurationsLabelsMapEqual(oldSpec.Cluster.Spec.WorkerNodeGroupConfigurations, newSpec.Cluster.Spec.WorkerNodeGroupConfigurations) {
		return true
	}
	return AnyImmutableFieldChanged(oldVdc, newVdc, oldTmc, newTmc)
}

func NeedsNewKubeadmConfigTemplate(newWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration, oldWorkerNodeGroup *v1alpha1.WorkerNodeGroupConfiguration) bool {
	return !v1alpha1.TaintsSliceEqual(newWorkerNodeGroup.Taints, oldWorkerNodeGroup.Taints) || !v1alpha1.LabelsMapEqual(newWorkerNodeGroup.Labels, oldWorkerNodeGroup.Labels)
}

func NeedsNewEtcdTemplate(oldSpec, newSpec *cluster.Spec, oldVdc, newVdc *v1alpha1.TinkerbellDatacenterConfig, oldTmc, newTmc *v1alpha1.TinkerbellMachineConfig) bool {
	if oldSpec.Cluster.Spec.KubernetesVersion != newSpec.Cluster.Spec.KubernetesVersion {
		return true
	}
	if oldSpec.Bundles.Spec.Number != newSpec.Bundles.Spec.Number {
		return true
	}
	return AnyImmutableFieldChanged(oldVdc, newVdc, oldTmc, newTmc)
}

func AnyImmutableFieldChanged(oldVdc, newVdc *v1alpha1.TinkerbellDatacenterConfig, oldTmc, newTmc *v1alpha1.TinkerbellMachineConfig) bool {
	return false
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

		writer := hardware.MultiMachineWriter(machineCatalogueWriter, &p.diskExtractor)

		machines, err := hardware.NewNormalizedCSVReaderFromFile(p.hardwareCSVFile)
		if err != nil {
			return err
		}

		// TODO(chrisdoherty4) Build the selectors slice using the selectors in the Tinkerbell
		// Enabled Management Cluster that we're upgrading.
		var selectors []v1alpha1.HardwareSelector

		machineValidator := hardware.NewDefaultMachineValidator()
		machineValidator.Register(hardware.MatchingDisksForSelectors(selectors))

		if err := hardware.TranslateAll(machines, writer, machineValidator); err != nil {
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
		if err := p.diskExtractor.InsertDisks(&hardware[i]); err != nil {
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
	for i := range hardware {
		if err := p.diskExtractor.InsertProvisionedHardwareDisks(&hardware[i]); err != nil {
			return err
		}
	}

	// Remove all the provisioned hardware from the existing cluster if repeated from the hardware csv input.
	if err := p.catalogue.RemoveHardwares(hardware); err != nil {
		return err
	}

	return p.validateAvailableHardwareForUpgrade(ctx, currentClusterSpec, clusterSpec)
}

func (p *Provider) validateAvailableHardwareForUpgrade(ctx context.Context, currentSpec, newClusterSpec *cluster.Spec) (err error) {
	clusterSpecValidator := NewClusterSpecValidator(
		HardwareSatisfiesOnlyOneSelectorAssertion(p.catalogue),
	)

	rollingUpgrade := false
	if currentSpec.Cluster.Spec.KubernetesVersion != newClusterSpec.Cluster.Spec.KubernetesVersion {
		clusterSpecValidator.Register(ExtraHardwareAvailableAssertionForRollingUpgrade(p.catalogue))
		rollingUpgrade = true
	}

	clusterSpecValidator.Register(AssertionsForScaleUpDown(p.catalogue, currentSpec, rollingUpgrade))

	tinkerbellClusterSpec := NewClusterSpec(newClusterSpec, p.machineConfigs, p.datacenterConfig)

	if err := clusterSpecValidator.Validate(tinkerbellClusterSpec); err != nil {
		return err
	}

	return nil
}

func (p *Provider) PostBootstrapDeleteForUpgrade(ctx context.Context) error {
	if err := p.stackInstaller.UninstallLocal(ctx); err != nil {
		return err
	}
	return nil
}

func (p *Provider) PostBootstrapSetupUpgrade(ctx context.Context, clusterConfig *v1alpha1.Cluster, cluster *types.Cluster) error {
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
		err := p.providerKubectlClient.WaitForBaseboardManagements(ctx, bootstrapCluster, "5m", "Contactable", constants.EksaSystemNamespace)
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

func (p *Provider) ValidateNewSpec(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	prevSpec, err := p.providerKubectlClient.GetEksaCluster(ctx, cluster, clusterSpec.Cluster.Name)
	if err != nil {
		return err
	}

	prevDatacenterConfig, err := p.providerKubectlClient.GetEksaTinkerbellDatacenterConfig(ctx, prevSpec.Spec.DatacenterRef.Name, cluster.KubeconfigFile, prevSpec.Namespace)
	if err != nil {
		return err
	}

	oSpec := prevDatacenterConfig.Spec
	nSpec := p.datacenterConfig.Spec

	prevMachineConfigRefs := machineRefSliceToMap(prevSpec.MachineConfigRefs())

	for _, machineConfigRef := range clusterSpec.Cluster.MachineConfigRefs() {
		machineConfig, ok := p.machineConfigs[machineConfigRef.Name]
		if !ok {
			return fmt.Errorf("cannot find machine config %s in tinkerbell provider machine configs", machineConfigRef.Name)
		}

		if _, ok = prevMachineConfigRefs[machineConfig.Name]; !ok {
			return fmt.Errorf("cannot add or remove MachineConfigs as part of upgrade")
		}
		err = p.validateMachineConfigImmutability(ctx, cluster, machineConfig, clusterSpec)
		if err != nil {
			return err
		}
	}

	if nSpec.TinkerbellIP != oSpec.TinkerbellIP {
		return fmt.Errorf("spec.TinkerbellIP is immutable. Previous value %s,   New value %s", oSpec.TinkerbellIP, nSpec.TinkerbellIP)
	}

	// for any operation other than k8s version change, osImageURL and hookImageURL are immutable
	if prevSpec.Spec.KubernetesVersion == clusterSpec.Cluster.Spec.KubernetesVersion {
		if nSpec.OSImageURL != oSpec.OSImageURL {
			return fmt.Errorf("spec.OSImageURL is immutable. Previous value %s,   New value %s", oSpec.OSImageURL, nSpec.OSImageURL)
		}
		if nSpec.HookImagesURLPath != oSpec.HookImagesURLPath {
			return fmt.Errorf("spec.HookImagesURLPath is immutable. Previous value %s,   New value %s", oSpec.HookImagesURLPath, nSpec.HookImagesURLPath)
		}
	}

	return nil
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
	for _, workerNodeGroupConfiguration := range oldCluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMap[workerNodeGroupConfiguration.Name] = &workerNodeGroupConfiguration
	}

	for _, nodeGroupNewSpec := range newCluster.Spec.WorkerNodeGroupConfigurations {
		if workerNodeGrpOldSpec, ok := workerNodeGroupMap[nodeGroupNewSpec.Name]; ok {
			if nodeGroupNewSpec.Count != workerNodeGrpOldSpec.Count {
				return true
			}
		}
	}

	return false
}

/* func (p *Provider) isScaleUpDown(currentSpec *cluster.Spec, newSpec *cluster.Spec) bool {
	if currentSpec.Cluster.Spec.ControlPlaneConfiguration.Count != newSpec.Cluster.Spec.ControlPlaneConfiguration.Count {
		return true
	}

	workerNodeGroupMap := make(map[string]*v1alpha1.WorkerNodeGroupConfiguration)
	for _, workerNodeGroupConfiguration := range currentSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMap[workerNodeGroupConfiguration.Name] = &workerNodeGroupConfiguration
	}

	for _, nodeGroupNewSpec := range newSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if workerNodeGrpOldSpec, ok := workerNodeGroupMap[nodeGroupNewSpec.Name]; ok {
			if nodeGroupNewSpec.Count != workerNodeGrpOldSpec.Count {
				return true
			}
		}
	}

	return false
} */

func (p *Provider) validateMachineConfigImmutability(ctx context.Context, cluster *types.Cluster, newConfig *v1alpha1.TinkerbellMachineConfig, clusterSpec *cluster.Spec) error {
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

func machineRefSliceToMap(machineRefs []v1alpha1.Ref) map[string]v1alpha1.Ref {
	refMap := make(map[string]v1alpha1.Ref, len(machineRefs))
	for _, ref := range machineRefs {
		refMap[ref.Name] = ref
	}
	return refMap
}
