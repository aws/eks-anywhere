package framework

import (
	"context"
	"fmt"
	"strings"
	"testing"

	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	// VSphereMultiTemplateUbuntu127 is to test multiple vSphere templates.
	VSphereMultiTemplateUbuntu127 = "T_VSPHERE_TEMPLATE_UBUNTU_2204_1_27"
	// ControlPlaneMachineLabel is to get control plane vSphere machines from machine list.
	ControlPlaneMachineLabel = "cluster.x-k8s.io/control-plane"
	// EtcdMachineLabel is to get etcd vSphere machines from machine list.
	EtcdMachineLabel = "cluster.x-k8s.io/etcd-cluster"
	// workerMachineLabel is to get worker vSphere machines from machine list.
	workerMachineLabel = "cluster.x-k8s.io/deployment-name"
)

// VSphereMultiTemplateUbuntu127RequiredEnvVars is required for TestVSphereMultipleTemplatesUbuntu127.
var VSphereMultiTemplateUbuntu127RequiredEnvVars = []string{
	VSphereMultiTemplateUbuntu127,
}

// RequiredVsphereMultiTemplateUbuntu127EnvVars return required env vars for TestVSphereMultipleTemplatesUbuntu127.
func RequiredVsphereMultiTemplateUbuntu127EnvVars() []string {
	return VSphereMultiTemplateUbuntu127RequiredEnvVars
}

// CheckVsphereMultiTemplateUbuntu127EnvVars checks is required env vars are present.
func CheckVsphereMultiTemplateUbuntu127EnvVars(t *testing.T) {
	checkRequiredEnvVars(t, VSphereMultiTemplateUbuntu127RequiredEnvVars)
}

// VsphereMachineValidation should return an error if either an error is encountered during execution or the validation logically fails.
// This validation function will be executed by ValidateVsphereMachine and ValidateWorkerNodeVsphereMachine
// with a vSphere machine config and a corresponding vSphere machine.
type VsphereMachineValidation func(machineConfig *v1alpha1.VSphereMachineConfig, machine vspherev1.VSphereMachine) (err error)

// ValidateVsphereMachine deduces the control plane or etcd configuration to machine mapping
// and for each configuration/machine pair executes the provided validation functions.
func (e *ClusterE2ETest) ValidateVsphereMachine(selector string, machineConfig *v1alpha1.VSphereMachineConfig, validations ...VsphereMachineValidation) {
	e.T.Log("Validating VSphere machine template")
	ctx := context.Background()
	machines, err := e.KubectlClient.GetVsphereMachine(ctx, e.Cluster().KubeconfigFile, selector)
	if err != nil {
		e.T.Fatal(err)
	}

	for _, machine := range machines {
		for _, validation := range validations {
			err = validation(machineConfig, machine)
			if err != nil {
				e.T.Errorf("VSphere machine %v is not valid: %v", machine.Name, err)
			}
		}
	}
	e.StopIfFailed()
}

// ValidateWorkerNodeVsphereMachine deduces the worker node group configuration to machine mapping
// and for each configuration/machine pair executes the provided validation functions.
func (e *ClusterE2ETest) ValidateWorkerNodeVsphereMachine(validations ...VsphereMachineValidation) {
	e.T.Log("Validating VSphere worker machine template")
	machineConfigToMachines := e.getMachineConfigToMachine()

	for machineConfig, machines := range machineConfigToMachines {
		for _, validation := range validations {
			for _, machine := range machines {
				err := validation(machineConfig, machine)
				if err != nil {
					e.T.Errorf("VSphere machine %v is not valid: %v", machine.Name, err)
				}
			}
		}
	}
	e.StopIfFailed()
}

func (e *ClusterE2ETest) getMachineConfigToMachine() map[*v1alpha1.VSphereMachineConfig][]vspherev1.VSphereMachine {
	ctx := context.Background()
	machines, err := e.KubectlClient.GetVsphereMachine(ctx, e.Cluster().KubeconfigFile, workerMachineLabel)
	if err != nil {
		e.T.Fatal(err)
	}
	wngNameToWng := getWngNameToWng(e.ClusterConfig.Cluster.Spec.WorkerNodeGroupConfigurations)

	machineConfigToMachine := make(map[*v1alpha1.VSphereMachineConfig][]vspherev1.VSphereMachine)
	for _, machine := range machines {
		if strings.Contains(machine.GetName(), "control-plane") || strings.Contains(machine.Name, "etcd") {
			continue
		}
		if !strings.HasPrefix(machine.GetName(), e.ClusterName) {
			continue
		}
		wngName, err := getWngNameFromMachine(machine.GetName(), e.ClusterName)
		if err != nil {
			e.T.Fatal(err)
		}
		machineConfigName := wngNameToWng[wngName].MachineGroupRef.Name
		machineConfig := e.ClusterConfig.VSphereMachineConfigs[machineConfigName]
		if _, exists := machineConfigToMachine[machineConfig]; !exists {
			machineConfigToMachine[machineConfig] = []vspherev1.VSphereMachine{machine}
		} else {
			machineConfigToMachine[machineConfig] = append(machineConfigToMachine[machineConfig], machine)
		}
	}
	return machineConfigToMachine
}

func getWngNameToWng(wngConfigs []v1alpha1.WorkerNodeGroupConfiguration) map[string]v1alpha1.WorkerNodeGroupConfiguration {
	wngNameToWng := make(map[string]v1alpha1.WorkerNodeGroupConfiguration)
	for _, wng := range wngConfigs {
		wngNameToWng[wng.Name] = wng
	}
	return wngNameToWng
}

// getWngNameFromMachine gets worker node group name from machine name by trimming cluster name prefix and two unix nano time suffix.
func getWngNameFromMachine(machineName string, clusterName string) (string, error) {
	trimmedMachineName := strings.TrimPrefix(machineName, clusterName+"-")
	wngBaseParts := strings.Split(trimmedMachineName, "-")
	if len(wngBaseParts) < 3 {
		return "", fmt.Errorf("invalid machine name %v", machineName)
	}
	return strings.Join(wngBaseParts[:len(wngBaseParts)-2], "-"), nil
}

// ValidateMachineTemplate validates if template configured in machine config matches the vSphere machine.
func ValidateMachineTemplate(machineConfig *v1alpha1.VSphereMachineConfig, machine vspherev1.VSphereMachine) (err error) {
	if machineConfig.Spec.Template != machine.Spec.Template {
		return fmt.Errorf("template on machine %v does not match; configured template: %v; machine actual template: %v",
			machine.Name, machineConfig.Spec.Template, machine.Spec.Template)
	}
	logger.V(4).Info("expected template is present on the corresponding VSphere machine", "machine", machine.Name, "machine template", machine.Spec.Template, "configuration template", machineConfig.Spec.Template)
	return nil
}
