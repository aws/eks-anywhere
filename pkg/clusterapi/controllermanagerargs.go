package clusterapi

import (
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func ControllerManagerArgs(clusterSpec *cluster.Spec) ExtraArgs {
	return SecureTlsCipherSuitesExtraArgs().
		Append(NodeCIDRMaskExtraArgs(&clusterSpec.Cluster.Spec.ClusterNetwork))
}
