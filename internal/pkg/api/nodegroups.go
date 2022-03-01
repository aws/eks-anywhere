package api

import (
	corev1 "k8s.io/api/core/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type WorkerNodeGroupFiller func(w *anywherev1.WorkerNodeGroupConfiguration)

func FillWorkerNodeGroup(w *anywherev1.WorkerNodeGroupConfiguration, fillers ...WorkerNodeGroupFiller) {
	for _, f := range fillers {
		f(w)
	}
}

func WithTaint(taint corev1.Taint) WorkerNodeGroupFiller {
	return func(w *anywherev1.WorkerNodeGroupConfiguration) {
		w.Taints = append(w.Taints, taint)
	}
}

func WithNoTaints() WorkerNodeGroupFiller {
	return func(w *anywherev1.WorkerNodeGroupConfiguration) {
		w.Taints = nil
	}
}

func WithLabel(key, value string) WorkerNodeGroupFiller {
	return func(w *anywherev1.WorkerNodeGroupConfiguration) {
		if w.Labels == nil {
			w.Labels = map[string]string{}
		}
		w.Labels[key] = value
	}
}

func WithCount(count int) WorkerNodeGroupFiller {
	return func(w *anywherev1.WorkerNodeGroupConfiguration) {
		w.Count = count
	}
}

func WithMachineGroupRef(name, kind string) WorkerNodeGroupFiller {
	return func(w *anywherev1.WorkerNodeGroupConfiguration) {
		w.MachineGroupRef = &anywherev1.Ref{
			Name: name,
			Kind: kind,
		}
	}
}
