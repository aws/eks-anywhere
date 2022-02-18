package framework

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
)

const ownerAnnotation = "cluster.x-k8s.io/owner-name"

func (e *ClusterE2ETest) VaidateWorkerNodeTaints() {
	ctx := context.Background()
	nodes, err := e.KubectlClient.GetNodes(ctx, e.cluster().KubeconfigFile)
	if err != nil {
		e.T.Fatal(err)
	}

	c, err := v1alpha1.GetClusterConfigFromContent(e.ClusterConfigB)
	if err != nil {
		e.T.Fatal(err)
	}
	wn := c.Spec.WorkerNodeGroupConfigurations
	// deduce the worker node group configuration to node mapping via the machine deployment and machine set
	for _, w := range wn {
		mdName := fmt.Sprintf("%v-%v", e.ClusterName, w.Name)
		md, err := e.KubectlClient.GetMachineDeployment(ctx, mdName, executables.WithKubeconfig(e.cluster().KubeconfigFile), executables.WithNamespace(constants.EksaSystemNamespace))
		if err != nil {
			e.T.Fatal(fmt.Errorf("failed to get machine deployment for worker node %s when validating taints: %v", w.Name, err))
		}
		ms, err := e.KubectlClient.GetMachineSets(ctx, md.Name, e.cluster())
		if err != nil {
			e.T.Fatal(fmt.Errorf("failed to get machine sets when validating taints: %v", err))
		}
		if len(ms) == 0 {
			e.T.Fatal(fmt.Errorf("invalid number of machine sets associated with worker node configuration %v", w.Name))
		}

		for _, node := range nodes {
			ownerName, ok := node.Annotations[ownerAnnotation]
			if ok {
				// there will be multiple machineSets present on a cluster following an upgrade.
				// find the one that is associated with this worker node, and compare the taints.
				for _, machineSet := range ms {
					if ownerName == machineSet.Name {
						if !v1alpha1.TaintsSliceEqual(node.Spec.Taints, w.Taints) {
							e.T.Fatal(fmt.Errorf("taints on node %v and corresponding worker node group configuration %v do not match", node.Name, w.Name))
						}
					}
				}
			}
		}
	}
	e.T.Log("validated that expected taints are present on the workload cluster nodes")
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

func NoScheduleWorkerNodeGroup(name string, count int) *WorkerNodeGroup {
	return WithWorkerNodeGroup(name, api.WithCount(count), api.WithTaint(NoScheduleTaint()))
}

func PreferNoScheduleWorkerNodeGroup(name string, count int) *WorkerNodeGroup {
	return WithWorkerNodeGroup(name, api.WithCount(count), api.WithTaint(PreferNoScheduleTaint()))
}
