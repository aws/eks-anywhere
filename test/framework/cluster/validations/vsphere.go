package validations

import (
	"context"
	"fmt"

	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

// ValidateCSI checks that the CSI is enabled/disabled based on the settings in the cluster.Spec by
// by checking if CSI related objects components exist in the cluster or not.
func ValidateCSI(ctx context.Context, vc clusterf.StateValidationConfig) error {
	clusterClient := vc.ClusterClient

	disableCSI := vc.ClusterSpec.Config.VSphereDatacenter.Spec.DisableCSI

	deployment := &v1.Deployment{}
	deployKey := types.NamespacedName{Namespace: "kube-system", Name: "vsphere-csi-controller"}
	deployErr := clusterClient.Get(ctx, deployKey, deployment)
	deployRes := handleCSIError(deployErr, disableCSI)
	if deployRes != nil {
		return deployRes
	}

	ds := &v1.DaemonSet{}
	dsKey := types.NamespacedName{Namespace: "kube-system", Name: "vsphere-csi-node"}
	dsErr := clusterClient.Get(ctx, dsKey, ds)
	dsRes := handleCSIError(dsErr, disableCSI)
	if dsRes != nil {
		return dsRes
	}

	return nil
}

func handleCSIError(err error, disabled bool) error {
	if (disabled && err == nil) || (!disabled && err != nil) {
		return fmt.Errorf("CSI state does not match disableCSI %t, %v", disabled, err)
	}

	return nil
}
