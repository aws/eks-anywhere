package clusterapi

import (
	"strconv"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// Autoscaler annotation constants.
const (
	NodeGroupMinSizeAnnotation = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size"
	NodeGroupMaxSizeAnnotation = "cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size"
)

func ConfigureAutoscalingInMachineDeployment(md *clusterv1.MachineDeployment, autoscalingConfig *anywherev1.AutoScalingConfiguration) {
	if autoscalingConfig == nil {
		return
	}

	if md.ObjectMeta.Annotations == nil {
		md.ObjectMeta.Annotations = map[string]string{}
	}

	md.ObjectMeta.Annotations[NodeGroupMinSizeAnnotation] = strconv.Itoa(autoscalingConfig.MinCount)
	md.ObjectMeta.Annotations[NodeGroupMaxSizeAnnotation] = strconv.Itoa(autoscalingConfig.MaxCount)
}
