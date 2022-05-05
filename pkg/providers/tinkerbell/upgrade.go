package tinkerbell

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
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

func (p *tinkerbellProvider) SetupAndValidateUpgradeCluster(ctx context.Context, cluster *types.Cluster, clusterSpec *cluster.Spec) error {
	logger.Info("Warning: The tinkerbell infrastructure provider is still in development and should not be used in production")

	hardware, err := p.providerTinkClient.GetHardware(ctx)
	if err != nil {
		return fmt.Errorf("retrieving tinkerbell hardware: %v", err)
	}
	logger.MarkPass("Connected to tinkerbell stack")

	if err := setupEnvVars(p.datacenterConfig); err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}

	tinkerbellClusterSpec := newSpec(clusterSpec, p.machineConfigs, p.datacenterConfig)

	if err := p.configureSshKeys(); err != nil {
		return err
	}

	// ValidateHardwareCatalogue performs a lazy load of hardware configuration. Given subsequent steps need the hardware
	// read into memory it needs to be done first. It also needs connection to
	// Tinkerbell steps to verify hardware availability on the stack
	if err := p.validator.ValidateHardwareCatalogue(ctx, p.catalogue, hardware, p.skipPowerActions, p.force); err != nil {
		return err
	}

	if err := p.validator.ValidateTinkerbellConfig(ctx, tinkerbellClusterSpec.datacenterConfig); err != nil {
		return err
	}

	if err := p.validator.ValidateClusterMachineConfigs(ctx, tinkerbellClusterSpec); err != nil {
		return err
	}

	if err := p.validator.ValidateAndPopulateTemplateForUpgrade(ctx, tinkerbellClusterSpec.datacenterConfig, tinkerbellClusterSpec.Spec.TinkerbellTemplateConfigs[tinkerbellClusterSpec.controlPlaneMachineConfig().Spec.TemplateRef.Name], clusterSpec.VersionsBundle.EksD.Raw.Ubuntu.URI, string(clusterSpec.Cluster.Spec.KubernetesVersion)); err != nil {
		return fmt.Errorf("failed validating control plane template config: %v", err)
	}

	if err := p.validator.ValidateMinHardwareAvailableForUpgrade(tinkerbellClusterSpec.Spec.Cluster.Spec, 1); err != nil {
		return fmt.Errorf("minimum hardware not available: %v", err)
	}

	if err := p.validator.ValidateAndPopulateTemplateForUpgrade(ctx, tinkerbellClusterSpec.datacenterConfig, tinkerbellClusterSpec.Spec.TinkerbellTemplateConfigs[tinkerbellClusterSpec.firstWorkerMachineConfig().Spec.TemplateRef.Name], clusterSpec.VersionsBundle.EksD.Raw.Ubuntu.URI, string(clusterSpec.Cluster.Spec.KubernetesVersion)); err != nil {
		return fmt.Errorf("failed validating worker node template config: %v", err)
	}

	// TODO: Add validations when this is supported
	return nil
}

func (p *tinkerbellProvider) RunPostControlPlaneUpgrade(ctx context.Context, oldClusterSpec *cluster.Spec, clusterSpec *cluster.Spec, workloadCluster *types.Cluster, managementCluster *types.Cluster) error {
	// @TODO: do we need this for bare metal upgrade?

	// Use retrier so that cluster upgrade does not fail due to any intermittent failure while connecting to kube-api server

	// This is unfortunate, but ClusterResourceSet's don't support any type of reapply of the resources they manage
	// Even if we create a new ClusterResourceSet, if such resources already exist in the cluster, they won't be reapplied
	// The long term solution is to add this capability to the cluster-api controller,
	// with a new mode like "ReApplyOnChanges" or "ReApplyOnCreate" vs the current "ReApplyOnce"
	/* err := p.Retrier.Retry(
		func() error {
			return p.resourceSetManager.ForceUpdate(ctx, resourceSetName(clusterSpec), constants.EksaSystemNamespace, managementCluster, workloadCluster)
		},
	)
	if err != nil {
		return fmt.Errorf("failed updating the tinkerbell provider resource set post upgrade: %v", err)
	} */
	return nil
}

func (p *tinkerbellProvider) UpgradeNeeded(_ context.Context, _, _ *cluster.Spec, _ *types.Cluster) (bool, error) {
	// TODO: Figure out if something is needed here
	return false, nil
}
