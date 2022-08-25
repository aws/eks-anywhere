package tinkerbell

import (
	"context"
	"fmt"

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

	fieldsChanged := AnyImmutableFieldChanged(oldVdc, newVdc, oldTmc, newTmc)
	return fieldsChanged
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
	// @TODO: Add immutable fields check here
	return false
}

func (p *Provider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec, currentClusterSpec *cluster.Spec) error {
	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		return ErrExternalEtcdUnsupported
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

	return p.validateAvailableHardwareForUpgrade(ctx, currentClusterSpec, clusterSpec)
}

func (p *Provider) validateAvailableHardwareForUpgrade(ctx context.Context, currentSpec, newClusterSpec *cluster.Spec) (err error) {
	clusterSpecValidator := NewClusterSpecValidator(
		HardwareSatisfiesOnlyOneSelectorAssertion(p.catalogue),
	)

	rollingUpgrade := false
	if currentSpec.Cluster.Spec.KubernetesVersion != newClusterSpec.Cluster.Spec.KubernetesVersion {
		clusterSpecValidator.Register(ExtraHardwareAvailableAssertionForRollingUpgrade(p.catalogue, maxSurgeForRollingUpgrade))
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
	err = p.providerKubectlClient.WaitForBaseboardManagements(ctx, cluster, "5m", "Contactable", constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("waiting for baseboard management to be contactable: %v", err)
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

func (p *Provider) UpgradeNeeded(_ context.Context, _, _ *cluster.Spec, _ *types.Cluster) (bool, error) {
	// TODO: Figure out if something is needed here
	return false, nil
}

func (p *Provider) hardwareCSVIsProvided() bool {
	return p.hardwareCSVFile != ""
}

func (p *Provider) isScaleUpDown(currentSpec *cluster.Spec, newSpec *cluster.Spec) bool {
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
}
