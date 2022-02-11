package tinkerbell

import (
	"context"
	"errors"
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type Validator struct{}

func (v *Validator) ValidateTinkerbellConfig(ctx context.Context, datacenterConfig *anywherev1.TinkerbellDatacenterConfig) error {
	// TODO: add validations for tinkerbellAccess
	if err := v.validateTinkerbellIP(ctx, datacenterConfig.Spec.TinkerbellIP); err != nil {
		return err
	}

	if err := v.validateTinkerbellCertURL(ctx, datacenterConfig.Spec.TinkerbellCertURL); err != nil {
		return err
	}

	if err := v.validatetinkerbellGRPCAuth(ctx, datacenterConfig.Spec.TinkerbellGRPCAuth); err != nil {
		return err
	}

	if err := v.validatetinkerbellPBnJGRPCAuth(ctx, datacenterConfig.Spec.TinkerbellPBnJGRPCAuth); err != nil {
		return err
	}

	return nil
}

// TODO: dry out machine configs validations
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, tinkerbellClusterSpec *spec) error {
	// TODO: move this to api Cluster validations
	if len(tinkerbellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}

	if tinkerbellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for control plane")
	}

	controlPlaneMachineConfig := tinkerbellClusterSpec.controlPlaneMachineConfig()
	if controlPlaneMachineConfig == nil {
		return fmt.Errorf("cannot find tinkerbellClusterSpec %v for control plane", tinkerbellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}
	if len(controlPlaneMachineConfig.Spec.OSFamily) <= 0 {
		return errors.New("tinkerbellMachineConfig osFamily for control plane is not set or is empty")
	}

	if tinkerbellClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for worker nodes")
	}

	workerNodeGroupMachineConfig := tinkerbellClusterSpec.firstWorkerMachineConfig()
	if workerNodeGroupMachineConfig == nil {
		return fmt.Errorf("cannot find tinkerbellMachineConfig %v for worker nodes", tinkerbellClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name)
	}
	if len(workerNodeGroupMachineConfig.Spec.OSFamily) <= 0 {
		return errors.New("tinkerbellMachineConfig osFamily for worker nodes is not set or is empty")
	}

	// TODO: move this to api Cluster validations
	controlPlaneEndpointIp := tinkerbellClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	if err := networkutils.ValidateIP(controlPlaneEndpointIp); err != nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host %s", err)
	}

	if controlPlaneMachineConfig.Spec.OSFamily != anywherev1.Ubuntu {
		return fmt.Errorf("control plane osFamily: %s is not supported, please use %s", controlPlaneMachineConfig.Spec.OSFamily, anywherev1.Ubuntu)
	}

	if workerNodeGroupMachineConfig.Spec.OSFamily != anywherev1.Ubuntu {
		return fmt.Errorf("worker node osFamily: %s is not supported, please use %s", workerNodeGroupMachineConfig.Spec.OSFamily, anywherev1.Ubuntu)
	}

	if controlPlaneMachineConfig.Spec.OSFamily != workerNodeGroupMachineConfig.Spec.OSFamily {
		return errors.New("control plane and worker nodes must have the same osFamily specified")
	}

	for _, machineConfig := range tinkerbellClusterSpec.machineConfigsLookup {
		if machineConfig.Namespace != tinkerbellClusterSpec.Cluster.Namespace {
			return errors.New("TinkerbellMachineConfig and Cluster objects must have the same namespace specified")
		}
	}

	if tinkerbellClusterSpec.datacenterConfig.Namespace != tinkerbellClusterSpec.Cluster.Namespace {
		return errors.New("TinkerbellDatacenterConfig and Cluster objects must have the same namespace specified")
	}

	return nil
}

func (v *Validator) validateTinkerbellIP(ctx context.Context, ip string) error {
	// check if tinkerbellIP is valid
	if err := networkutils.ValidateIP(ip); err != nil {
		return fmt.Errorf("cluster tinkerbellDatacenterConfig.tinkerbellIP %s", err)
	}
	return nil
}

func (v *Validator) validateTinkerbellCertURL(ctx context.Context, tinkerbellCertURL string) error {
	if tinkerbellCertURL == "" {
		return fmt.Errorf("tinkerbellCertURL is required")
	}

	return nil
}

func (v *Validator) validatetinkerbellGRPCAuth(ctx context.Context, tinkerbellGRPCAuth string) error {
	if tinkerbellGRPCAuth == "" {
		return fmt.Errorf("tinkerbellGRPCAuth is required")
	}

	return nil
}

func (v *Validator) validatetinkerbellPBnJGRPCAuth(ctx context.Context, tinkerbellPBnJGRPCAuth string) error {
	if tinkerbellPBnJGRPCAuth == "" {
		return fmt.Errorf("tinkerbellPBnJGRPCAuth is required")
	}

	return nil
}
