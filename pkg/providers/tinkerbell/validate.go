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

	if config.Spec.OSFamily == "" {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.osFamily: %v", config.Name)
	}

	if config.Spec.OSFamily != v1alpha1.Ubuntu {
		return fmt.Errorf(
			"TinkerbellMachineConfig: unsupported spec.osFamily (%v); '%v' is the only supported family",
			config.Spec.OSFamily,
			v1alpha1.Ubuntu,
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

func validateMinimumExpectedHardware(cluster v1alpha1.ClusterSpec, catalogue *hardware.Catalogue) error {
	requestedNodesCount := cluster.ControlPlaneConfiguration.Count +
		sumWorkerNodeCounts(cluster.WorkerNodeGroupConfigurations)

	// Optional external etcd configuration.
	if cluster.ExternalEtcdConfiguration != nil {
		requestedNodesCount += cluster.ExternalEtcdConfiguration.Count
	}

	if catalogue.TotalHardware() < requestedNodesCount {
		return fmt.Errorf(
			"have %v tinkerbell hardware; cluster spec requires >= %v hardware",
			catalogue.TotalHardware(),
			requestedNodesCount,
		)
	}

	return nil
}

func sumWorkerNodeCounts(nodes []v1alpha1.WorkerNodeGroupConfiguration) int {
	var requestedNodesCount int
	for _, workerSpec := range nodes {
		requestedNodesCount += workerSpec.Count
	}
	return requestedNodesCount
}
