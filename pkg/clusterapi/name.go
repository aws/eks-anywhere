package clusterapi

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

var nameRegex = regexp.MustCompile(`(.*?)(-)(\d+)$`)

func IncrementName(name string) (string, error) {
	match := nameRegex.FindStringSubmatch(name)
	if match == nil {
		return "", fmt.Errorf(`invalid format of name [name=%s]. Name has to follow regex pattern "(-)(\d+)$", e.g. machinetemplate-cp-1`, name)
	}

	n, err := strconv.Atoi(match[3])
	if err != nil {
		return "", fmt.Errorf("converting object suffix to int: %v", err)
	}

	return ObjectName(match[1], n+1), nil
}

func ObjectName(baseName string, version int) string {
	return fmt.Sprintf("%s-%d", baseName, version)
}

func DefaultObjectName(baseName string) string {
	return ObjectName(baseName, 1)
}

func KubeadmControlPlaneName(clusterSpec *cluster.Spec) string {
	return clusterSpec.Cluster.GetName()
}

func MachineDeploymentName(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) string {
	// Adding cluster name prefix guarantees the machine deployment name uniqueness
	// among clusters under the same management cluster setting.
	return clusterWorkerNodeGroupName(clusterSpec, workerNodeGroupConfig)
}

func DefaultKubeadmConfigTemplateName(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) string {
	return DefaultObjectName(clusterWorkerNodeGroupName(clusterSpec, workerNodeGroupConfig))
}

func clusterWorkerNodeGroupName(clusterSpec *cluster.Spec, workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration) string {
	return fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, workerNodeGroupConfig.Name)
}
