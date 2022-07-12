package clusterapi

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	unhealthyConditionTimeout = 5 * time.Minute
	machineHealthCheckKind    = "MachineHealthCheck"
	maxUnhealthyControlPlane  = "100%"
	maxUnhealthyWorker        = "40%"
)

func machineHealthCheck(clusterName string) *clusterv1.MachineHealthCheck {
	return &clusterv1.MachineHealthCheck{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clusterAPIVersion,
			Kind:       machineHealthCheckKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: clusterv1.MachineHealthCheckSpec{
			ClusterName: clusterName,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			UnhealthyConditions: []clusterv1.UnhealthyCondition{
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionUnknown,
					Timeout: metav1.Duration{Duration: unhealthyConditionTimeout},
				},
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionFalse,
					Timeout: metav1.Duration{Duration: unhealthyConditionTimeout},
				},
			},
		},
	}
}

func MachineHealthCheckForControlPlane(clusterSpec *cluster.Spec) *clusterv1.MachineHealthCheck {
	mhc := machineHealthCheck(ClusterName(clusterSpec.Cluster))
	mhc.SetName(ControlPlaneMachineHealthCheckName(clusterSpec))
	mhc.Spec.Selector.MatchLabels[clusterv1.MachineControlPlaneLabelName] = ""
	maxUnhealthy := intstr.Parse(maxUnhealthyControlPlane)
	mhc.Spec.MaxUnhealthy = &maxUnhealthy
	return mhc
}

func MachineHealthCheckForWorkers(clusterSpec *cluster.Spec) map[string]*clusterv1.MachineHealthCheck {
	m := make(map[string]*clusterv1.MachineHealthCheck, len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		mhc := machineHealthCheckForWorker(clusterSpec, workerNodeGroupConfig)
		m[workerNodeGroupConfig.Name] = mhc
	}
	return m
}

func machineHealthCheckForWorker(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) *clusterv1.MachineHealthCheck {
	mhc := machineHealthCheck(ClusterName(clusterSpec.Cluster))
	mhc.SetName(WorkerMachineHealthCheckName(clusterSpec, workerNodeGroupConfig))
	mhc.Spec.Selector.MatchLabels[clusterv1.MachineDeploymentLabelName] = MachineDeploymentName(clusterSpec, workerNodeGroupConfig)
	maxUnhealthy := intstr.Parse(maxUnhealthyWorker)
	mhc.Spec.MaxUnhealthy = &maxUnhealthy
	return mhc
}
