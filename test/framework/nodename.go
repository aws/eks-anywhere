package framework

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
)

// ValidateControlPlaneNodeNameMatchCAPIMachineName validate if node name is same as CAPI machine name.
func ValidateControlPlaneNodeNameMatchCAPIMachineName(controlPlane v1alpha1.ControlPlaneConfiguration, node corev1.Node) error {
	if controlPlane.MachineGroupRef.Kind == "CloudStackMachineConfig" {
		logger.V(4).Info("Validating control plane node matches CAPI machine name")
		return validateNodeNameMatchCAPIMachineName(node)
	}
	return nil
}

// ValidateWorkerNodeNameMatchCAPIMachineName validate if node name is same as CAPI machine name.
func ValidateWorkerNodeNameMatchCAPIMachineName(w v1alpha1.WorkerNodeGroupConfiguration, node corev1.Node) error {
	if w.MachineGroupRef.Kind == "CloudStackMachineConfig" {
		logger.V(4).Info("Validating worker node matches CAPI machine name")
		return validateNodeNameMatchCAPIMachineName(node)
	}
	return nil
}

func validateNodeNameMatchCAPIMachineName(node corev1.Node) error {
	capiMachineName, ok := node.Annotations["cluster.x-k8s.io/machine"]
	if !ok {
		return fmt.Errorf("CAPI machine name not found for node %s", node.Name)
	}

	if node.Name != capiMachineName {
		return fmt.Errorf("node name %s not match CAPI machine name %s", node.Name, capiMachineName)
	}
	return nil
}
