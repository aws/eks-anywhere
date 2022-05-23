package tinkerbell

import (
	"context"
	"fmt"
	"strings"
	"time"

	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

func (p *Provider) SetupAndValidateDeleteCluster(ctx context.Context, cluster *types.Cluster) error {
	hardwares, err := p.providerKubectlClient.GetHardwareWithLabel(ctx, tinkerbellOwnerNameLabel, cluster.KubeconfigFile, constants.EksaSystemNamespace)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	filteredHws, err := filterHardwareForCluster(hardwares, cluster.Name)
	if err != nil {
		return fmt.Errorf("failed setup and validations: %v", err)
	}
	p.hardwares = filteredHws

	return nil
}

// filterHardwareForCluster filters hardware with ownerName label that contains cluster name.
func filterHardwareForCluster(hardwares []tinkv1alpha1.Hardware, clusterName string) ([]tinkv1alpha1.Hardware, error) {
	var filteredHardwareList []tinkv1alpha1.Hardware
	for _, hw := range hardwares {
		if strings.Contains(hw.Labels[tinkerbellOwnerNameLabel], clusterName) {
			filteredHardwareList = append(filteredHardwareList, hw)
		}
	}
	// Ensure that there are one or more hardware CRDs presnt in the hardware list for a cluster.
	if len(filteredHardwareList) == 0 {
		return nil, fmt.Errorf("no hardware found for cluster %s", clusterName)
	}
	return filteredHardwareList, nil
}

func (p *Provider) DeleteResources(ctx context.Context, clusterSpec *cluster.Spec) error {
	for _, mc := range p.machineConfigs {
		if err := p.providerKubectlClient.DeleteEksaMachineConfig(ctx, eksaTinkerbellDatacenterResourceType, mc.Name, clusterSpec.ManagementCluster.KubeconfigFile, mc.Namespace); err != nil {
			return err
		}
	}
	return p.providerKubectlClient.DeleteEksaDatacenterConfig(ctx, eksaTinkerbellMachineResourceType, p.datacenterConfig.Name, clusterSpec.ManagementCluster.KubeconfigFile, p.datacenterConfig.Namespace)
}

func (p *Provider) PostClusterDeleteValidate(ctx context.Context, managementCluster *types.Cluster) error {
	// We want to validate cluster nodes are powered off.
	// We wait on BMC status.powerState to check for power off.
	bmcRefs := make([]string, 0, len(p.hardwares))
	for _, hw := range p.hardwares {
		bmcRefs = append(bmcRefs, hw.Spec.BmcRef)
	}

	// TODO (pokearu): The retry logic can be substituted by changing GetBmcsPowerState to use kubectl wait --for
	// In the current version of kubectl in EKSA --for does not support jsonpath.
	err := retrier.Retry(10, 10*time.Second, func() error {
		powerStates, err := p.providerKubectlClient.GetBmcsPowerState(ctx, bmcRefs, managementCluster.KubeconfigFile, constants.EksaSystemNamespace)
		if err != nil {
			return fmt.Errorf("retrieving bmc power state: %w", err)
		}

		for _, state := range powerStates {
			if !strings.Contains(state, bmcStatePowerActionHardoff) {
				return fmt.Errorf("bmc current power state '%s'; expected power state '%s'", state, bmcStatePowerActionHardoff)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
