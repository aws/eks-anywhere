package framework

import (
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const LabelPrefix = "eksa.e2e"

func ValidateControlPlaneLabels(controlPlane v1alpha1.ControlPlaneConfiguration, node corev1.Node) error {
	logger.V(4).Info("Validating control plane labels")
	return validateLabels(controlPlane.Labels, node)
}

// ValidateControlPlaneFailureDomainLabels validate if Cloudstack provider replaces ds.meta_data.failuredomain with proper failuredomain name
// in control plane node label 'cluster.x-k8s.io/failure-domain'
func ValidateControlPlaneFailureDomainLabels(controlPlane v1alpha1.ControlPlaneConfiguration, node corev1.Node) error {
	if controlPlane.MachineGroupRef.Kind == "CloudStackMachineConfig" {
		logger.V(4).Info("Validating control plane node failuredomain label")
		return validateFailuredomainLabel(controlPlane.Labels, node)
	}
	return nil
}

func ValidateWorkerNodeLabels(w v1alpha1.WorkerNodeGroupConfiguration, node corev1.Node) error {
	logger.V(4).Info("Validating worker node labels", "worker node group", w.Name)
	return validateLabels(w.Labels, node)
}

// ValidateWorkerNodeFailureDomainLabels validate if Cloudstack provider replaces ds.meta_data.failuredomain with proper failuredomain name
// in worker group node label 'cluster.x-k8s.io/failure-domain'
func ValidateWorkerNodeFailureDomainLabels(w v1alpha1.WorkerNodeGroupConfiguration, node corev1.Node) error {
	if w.MachineGroupRef.Kind == "CloudStackMachineConfig" {
		logger.V(4).Info("Validating worker node failuredomain label", "worker node group", w.Name)
		return validateFailuredomainLabel(w.Labels, node)
	}
	return nil
}

func validateLabels(expectedLabels map[string]string, node corev1.Node) error {
	actualLabels := retrieveTestNodeLabels(node.Labels)
	expectedBytes, _ := json.Marshal(expectedLabels)
	actualBytes, _ := json.Marshal(actualLabels)
	if !v1alpha1.LabelsMapEqual(expectedLabels, actualLabels) {
		return fmt.Errorf("labels on node %v and corresponding configuration do not match; configured labels: %v; node labels: %v",
			node.Name, string(expectedBytes), string(actualBytes))
	}
	logger.V(4).Info("expected labels from cluster spec configuration are present on the corresponding node", "node", node.Name, "node labels", string(actualBytes), "configuration labels", string(expectedBytes))
	return nil
}

func retrieveTestNodeLabels(nodeLabels map[string]string) map[string]string {
	labels := map[string]string{}
	for key, val := range nodeLabels {
		if strings.HasPrefix(key, LabelPrefix) {
			labels[key] = val
		}
	}
	return labels
}

func validateFailureDomainLabel(expectedLabels map[string]string, node corev1.Node) error {
	if failuredomainSpecified, ok := expectedLabels[constants.FailuredomainLabelName]; ok {
		if failuredomain, exist := node.Labels[constants.FailuredomainLabelName]; exist {
			logger.V(4).Info("node label: ", constants.FailuredomainLabelName, failuredomain)
			if failuredomainSpecified == constants.CloudstackFailuredomainPlaceholder && failuredomain == failuredomainSpecified {
				return fmt.Errorf("value %s of label %s on node %s is not replaced with a failurdomain name by CloudStack provider",
					constants.CloudstackFailuredomainPlaceholder,
					constants.FailuredomainLabelName,
					node.Name)
			}
		} else {
			return fmt.Errorf("expected labels %s not found on node %s", constants.FailuredomainLabelName, node.Name)
		}
	}
	return nil
}
