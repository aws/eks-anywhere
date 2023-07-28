package clusterapi

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	machineHealthCheckKind   = "MachineHealthCheck"
	maxUnhealthyControlPlane = "100%"
	maxUnhealthyWorker       = "40%"
)

func machineHealthCheck(clusterName string, unhealthyTimeout, nodeStartupTimeout *metav1.Duration) *clusterv1.MachineHealthCheck {
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
			NodeStartupTimeout: nodeStartupTimeout,
			UnhealthyConditions: []clusterv1.UnhealthyCondition{
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionUnknown,
					Timeout: *unhealthyTimeout,
				},
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionFalse,
					Timeout: *unhealthyTimeout,
				},
			},
		},
	}
}

// MachineHealthCheckForControlPlane creates MachineHealthCheck resources for the control plane.
func MachineHealthCheckForControlPlane(cluster *v1alpha1.Cluster) *clusterv1.MachineHealthCheck {
	mhc := machineHealthCheck(ClusterName(cluster), cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout, cluster.Spec.MachineHealthCheck.NodeStartupTimeout)
	mhc.SetName(ControlPlaneMachineHealthCheckName(cluster))
	mhc.Spec.Selector.MatchLabels[clusterv1.MachineControlPlaneLabel] = ""
	maxUnhealthy := intstr.Parse(maxUnhealthyControlPlane)
	mhc.Spec.MaxUnhealthy = &maxUnhealthy
	return mhc
}

// MachineHealthCheckForWorkers creates MachineHealthCheck resources for the workers.
func MachineHealthCheckForWorkers(cluster *v1alpha1.Cluster) []*clusterv1.MachineHealthCheck {
	m := make([]*clusterv1.MachineHealthCheck, 0, len(cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfig := range cluster.Spec.WorkerNodeGroupConfigurations {
		mhc := machineHealthCheckForWorker(cluster, workerNodeGroupConfig)
		m = append(m, mhc)
	}
	return m
}

func machineHealthCheckForWorker(cluster *v1alpha1.Cluster, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) *clusterv1.MachineHealthCheck {
	mhc := machineHealthCheck(ClusterName(cluster), cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout, cluster.Spec.MachineHealthCheck.NodeStartupTimeout)
	mhc.SetName(WorkerMachineHealthCheckName(cluster, workerNodeGroupConfig))
	mhc.Spec.Selector.MatchLabels[clusterv1.MachineDeploymentNameLabel] = MachineDeploymentName(cluster, workerNodeGroupConfig)
	maxUnhealthy := intstr.Parse(maxUnhealthyWorker)
	mhc.Spec.MaxUnhealthy = &maxUnhealthy
	return mhc
}

// MachineHealthCheckObjects creates MachineHealthCheck resources for control plane and all the worker node groups.
func MachineHealthCheckObjects(cluster *v1alpha1.Cluster) []kubernetes.Object {
	mhcWorkers := MachineHealthCheckForWorkers(cluster)
	o := make([]kubernetes.Object, 0, len(mhcWorkers)+1)
	for _, item := range mhcWorkers {
		o = append(o, item)
	}
	return append(o, MachineHealthCheckForControlPlane(cluster))
}
