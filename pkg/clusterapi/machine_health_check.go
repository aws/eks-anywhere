package clusterapi

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	machineHealthCheckKind = "MachineHealthCheck"
)

// durationToSeconds converts a *metav1.Duration to *int32 seconds for v1beta2.
func durationToSeconds(d *metav1.Duration) *int32 {
	if d == nil {
		return nil
	}
	s := int32(d.Duration.Seconds())
	return &s
}

func machineHealthCheck(clusterName string, unhealthyTimeout, nodeStartupTimeout *metav1.Duration) *clusterv1beta2.MachineHealthCheck {
	var unhealthyNodeConditions []clusterv1beta2.UnhealthyNodeCondition
	if unhealthyTimeout != nil {
		timeoutSeconds := durationToSeconds(unhealthyTimeout)
		unhealthyNodeConditions = []clusterv1beta2.UnhealthyNodeCondition{
			{
				Type:           "Ready",
				Status:         "Unknown",
				TimeoutSeconds: timeoutSeconds,
			},
			{
				Type:           "Ready",
				Status:         "False",
				TimeoutSeconds: timeoutSeconds,
			},
		}
	}

	return &clusterv1beta2.MachineHealthCheck{
		TypeMeta: metav1.TypeMeta{
			APIVersion: machineHealthCheckAPIVersion,
			Kind:       machineHealthCheckKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: clusterv1beta2.MachineHealthCheckSpec{
			ClusterName: clusterName,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Checks: clusterv1beta2.MachineHealthCheckChecks{
				NodeStartupTimeoutSeconds: durationToSeconds(nodeStartupTimeout),
				UnhealthyNodeConditions:   unhealthyNodeConditions,
			},
		},
	}
}

// MachineHealthCheckForControlPlane creates MachineHealthCheck resources for the control plane.
func MachineHealthCheckForControlPlane(cluster *v1alpha1.Cluster) *clusterv1beta2.MachineHealthCheck {
	unhealthyMachineTimeout := cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout
	if cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck != nil && cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck.UnhealthyMachineTimeout != nil {
		unhealthyMachineTimeout = cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck.UnhealthyMachineTimeout
	}
	nodeStartupTimeout := cluster.Spec.MachineHealthCheck.NodeStartupTimeout
	if cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck != nil && cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck.NodeStartupTimeout != nil {
		nodeStartupTimeout = cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck.NodeStartupTimeout
	}
	mhc := machineHealthCheck(ClusterName(cluster), unhealthyMachineTimeout, nodeStartupTimeout)
	mhc.SetName(ControlPlaneMachineHealthCheckName(cluster))
	mhc.Spec.Selector.MatchLabels[clusterv1beta2.MachineControlPlaneLabel] = ""
	maxUnhealthy := cluster.Spec.MachineHealthCheck.MaxUnhealthy
	if cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck != nil && cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck.MaxUnhealthy != nil {
		maxUnhealthy = cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck.MaxUnhealthy
	}
	if maxUnhealthy != nil {
		mhc.Spec.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo = maxUnhealthy
	}
	return mhc
}

// MachineHealthCheckForWorkers creates MachineHealthCheck resources for the workers.
func MachineHealthCheckForWorkers(cluster *v1alpha1.Cluster) []*clusterv1beta2.MachineHealthCheck {
	m := make([]*clusterv1beta2.MachineHealthCheck, 0, len(cluster.Spec.WorkerNodeGroupConfigurations))
	for _, workerNodeGroupConfig := range cluster.Spec.WorkerNodeGroupConfigurations {
		mhc := machineHealthCheckForWorker(cluster, workerNodeGroupConfig)
		m = append(m, mhc)
	}
	return m
}

func machineHealthCheckForWorker(cluster *v1alpha1.Cluster, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) *clusterv1beta2.MachineHealthCheck {
	unhealthyMachineTimeout := cluster.Spec.MachineHealthCheck.UnhealthyMachineTimeout
	if workerNodeGroupConfig.MachineHealthCheck != nil && workerNodeGroupConfig.MachineHealthCheck.UnhealthyMachineTimeout != nil {
		unhealthyMachineTimeout = workerNodeGroupConfig.MachineHealthCheck.UnhealthyMachineTimeout
	}
	nodeStartupTimeout := cluster.Spec.MachineHealthCheck.NodeStartupTimeout
	if workerNodeGroupConfig.MachineHealthCheck != nil && workerNodeGroupConfig.MachineHealthCheck.NodeStartupTimeout != nil {
		nodeStartupTimeout = workerNodeGroupConfig.MachineHealthCheck.NodeStartupTimeout
	}
	mhc := machineHealthCheck(ClusterName(cluster), unhealthyMachineTimeout, nodeStartupTimeout)
	mhc.SetName(WorkerMachineHealthCheckName(cluster, workerNodeGroupConfig))
	mhc.Spec.Selector.MatchLabels[clusterv1beta2.MachineDeploymentNameLabel] = MachineDeploymentName(cluster, workerNodeGroupConfig)
	maxUnhealthy := cluster.Spec.MachineHealthCheck.MaxUnhealthy
	if workerNodeGroupConfig.MachineHealthCheck != nil && workerNodeGroupConfig.MachineHealthCheck.MaxUnhealthy != nil {
		maxUnhealthy = workerNodeGroupConfig.MachineHealthCheck.MaxUnhealthy
	}
	if maxUnhealthy != nil {
		mhc.Spec.Remediation.TriggerIf.UnhealthyLessThanOrEqualTo = maxUnhealthy
	}
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
