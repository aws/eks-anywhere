package framework

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
)

const ownerAnnotation = "cluster.x-k8s.io/owner-name"

func ValidateControlPlaneTaints(controlPlane v1alpha1.ControlPlaneConfiguration, node corev1.Node) (err error) {
	cpTaints := controlPlane.Taints

	// if no taints are specified, kubeadm defaults it to a well-known control plane taint.
	// so, we make sure to check for that well-known taint if no taints are provided in the spec.
	var taintsValid bool
	if cpTaints == nil {
		taintsValid = validateDefaultControlPlaneTaints(node)
	} else {
		taintsValid = v1alpha1.TaintsSliceEqual(cpTaints, node.Spec.Taints)
	}

	if !taintsValid {
		return fmt.Errorf("taints on control plane node %v and corresponding control plane configuration do not match; configured taints: %v; node taints: %v",
			node.Name, cpTaints, node.Spec.Taints)
	}
	logger.V(4).Info("Expected taints from cluster spec control plane configuration are present on corresponding node", "node", node.Name, "node taints", node.Spec.Taints, "control plane configuration taints", cpTaints)
	return nil
}

// ValidateControlPlaneNoTaints will validate that a controlPlane has no taints, for example in the case of a single node cluster.
func ValidateControlPlaneNoTaints(controlPlane v1alpha1.ControlPlaneConfiguration, node corev1.Node) (err error) {
	valid := len(controlPlane.Taints) == 0 && len(node.Spec.Taints) == 0
	if !valid {
		return fmt.Errorf("taints on control plane node %v or corresponding control plane configuration found; configured taints: %v; node taints: %v",
			node.Name, controlPlane.Taints, node.Spec.Taints)
	}
	logger.V(4).Info("expected no taints on cluster spec control plane configuration and on corresponding node", "node", node.Name, "node taints", node.Spec.Taints, "control plane configuration taints", controlPlane.Taints)
	return nil
}

func ValidateWorkerNodeTaints(w v1alpha1.WorkerNodeGroupConfiguration, node corev1.Node) (err error) {
	valid := v1alpha1.TaintsSliceEqual(node.Spec.Taints, w.Taints)
	if !valid {
		return fmt.Errorf("taints on node %v and corresponding worker node group configuration %v do not match", node.Name, w.Name)
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

// MasterTaint will be deprecated from kubernetes version 1.25 onwards.
func MasterTaint() corev1.Taint {
	return corev1.Taint{
		Key:    "node-role.kubernetes.io/master",
		Effect: corev1.TaintEffectNoSchedule,
	}
}

// ControlPlaneTaint has been added from 1.24 onwards.
func ControlPlaneTaint() corev1.Taint {
	return corev1.Taint{
		Key:    "node-role.kubernetes.io/control-plane",
		Effect: corev1.TaintEffectNoSchedule,
	}
}

func NoScheduleWorkerNodeGroup(name string, count int) *WorkerNodeGroup {
	return WithWorkerNodeGroup(name, api.WithCount(count), api.WithTaint(NoScheduleTaint()))
}

func PreferNoScheduleWorkerNodeGroup(name string, count int) *WorkerNodeGroup {
	return WithWorkerNodeGroup(name, api.WithCount(count), api.WithTaint(PreferNoScheduleTaint()))
}

func validateDefaultControlPlaneTaints(node corev1.Node) bool {
	// Due to the transition from "master" to "control-plane", CP nodes can have one or both
	// of these taints, depending on the k8s version. So checking that the node has at least one
	// of them.

	masterTaint := MasterTaint()
	cpTaint := ControlPlaneTaint()

	for _, v := range node.Spec.Taints {
		if taintEqual(v, masterTaint) || taintEqual(v, cpTaint) {
			return true
		}
	}

	return false
}

func taintEqual(a, b corev1.Taint) bool {
	return a.Key == b.Key && a.Effect == b.Effect && a.Value == b.Value
}
