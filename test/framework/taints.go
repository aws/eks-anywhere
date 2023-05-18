package framework

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const ownerAnnotation = "cluster.x-k8s.io/owner-name"

// ValidateControlPlaneTaints will validate that a controlPlane node has the expected taints.
func ValidateControlPlaneTaints(controlPlane v1alpha1.ControlPlaneConfiguration, node corev1.Node) (err error) {
	if err := api.ValidateControlPlaneTaints(controlPlane, node); err != nil {
		return err
	}
	logger.V(4).Info("Expected taints from cluster spec control plane configuration are present on corresponding node", "node", node.Name, "node taints", node.Spec.Taints, "control plane configuration taints", controlPlane.Taints)
	return nil
}

// ValidateControlPlaneNoTaints will validate that a controlPlane has no taints, for example in the case of a single node cluster.
func ValidateControlPlaneNoTaints(controlPlane v1alpha1.ControlPlaneConfiguration, node corev1.Node) (err error) {
	if err := api.ValidateControlPlaneNoTaints(controlPlane, node); err != nil {
		return err
	}
	logger.V(4).Info("expected no taints on cluster spec control plane configuration and on corresponding node", "node", node.Name, "node taints", node.Spec.Taints, "control plane configuration taints", controlPlane.Taints)
	return nil
}

// ValidateWorkerNodeTaints will validate that a worker node has the expected taints in the worker node group configuration.
func ValidateWorkerNodeTaints(w v1alpha1.WorkerNodeGroupConfiguration, node corev1.Node) (err error) {
	if err := api.ValidateWorkerNodeTaints(w, node); err != nil {
		return err
	}
	logger.V(4).Info("expected taints from cluster spec are present on corresponding node", "worker node group", w.Name, "worker node group taints", w.Taints, "node", node.Name, "node taints", node.Spec.Taints)
	return nil
}

func NoExecuteTaint() corev1.Taint {
	return corev1.Taint{
		Key:    "key1",
		Value:  "value1",
		Effect: corev1.TaintEffectNoExecute,
	}
}

func NoScheduleTaint() corev1.Taint {
	return corev1.Taint{
		Key:    "key1",
		Value:  "value1",
		Effect: corev1.TaintEffectNoSchedule,
	}
}

func PreferNoScheduleTaint() corev1.Taint {
	return corev1.Taint{
		Key:    "key1",
		Value:  "value1",
		Effect: corev1.TaintEffectPreferNoSchedule,
	}
}

func NoScheduleWorkerNodeGroup(name string, count int) *WorkerNodeGroup {
	return WithWorkerNodeGroup(name, api.WithCount(count), api.WithTaint(NoScheduleTaint()))
}

func PreferNoScheduleWorkerNodeGroup(name string, count int) *WorkerNodeGroup {
	return WithWorkerNodeGroup(name, api.WithCount(count), api.WithTaint(PreferNoScheduleTaint()))
}
