package validations

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/features"
)

func ValidateClusterNameArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", errors.New("please specify a cluster name")
	}
	err := v1alpha1.ValidateClusterName(args[0])
	if err != nil {
		return args[0], err
	}
	err = v1alpha1.ValidateClusterNameLength(args[0])
	if err != nil {
		return args[0], err
	}
	return args[0], nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func KubeConfigExists(dir, clusterName string, kubeConfigFileOverride string, kubeconfigPattern string) bool {
	kubeConfigFile := kubeConfigFileOverride
	if kubeConfigFile == "" {
		kubeConfigFile = filepath.Join(dir, fmt.Sprintf(kubeconfigPattern, clusterName))
	}

	if info, err := os.Stat(kubeConfigFile); err == nil && info.Size() > 0 {
		return true
	}
	return false
}

func ValidateTaintsSupport(ctx context.Context, clusterSpec *cluster.Spec) error {
	if !features.IsActive(features.TaintsSupport()) {
		if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints) > 0 ||
			len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints) > 0 {
			return fmt.Errorf("Taints feature is not enabled. Environment variable TAINTS_SUPPORT needs to be set to true.")
		}
	} else if len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints) > 0 {
		invalidWorkerNodeGroupTaints := false
		for _, slice := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints {
			if slice.Effect == "NoExecute" || slice.Effect == "NoSchedule" {
				invalidWorkerNodeGroupTaints = true
				break
			}
		}

		if invalidWorkerNodeGroupTaints {
			return fmt.Errorf("The first worker node group does not support NoExecute or NoSchedule taints.")
		}
	}
	return nil
}
