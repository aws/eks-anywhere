package framework

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const ownerAnnotation = "cluster.x-k8s.io/owner-name"

func (e *ClusterE2ETest) VaidateWorkerNodeTaints() {
	wn := e.ClusterConfig.Spec.WorkerNodeGroupConfigurations
	ctx := context.Background()
	nodes, err := e.KubectlClient.GetNodes(ctx, e.cluster().KubeconfigFile)
	if err != nil {
		e.T.Fatal(err)
	}
	for _, w := range wn {
		md, err := e.KubectlClient.GetMachineDeployment(ctx, w.Name)
		if err != nil {
			e.T.Fatal(fmt.Errorf("failed to get machine deployment for worker node %s when validating taints: %v", w.Name, err))
		}
		templateName := md.Spec.Template.Spec.Bootstrap.ConfigRef.Name
		template, err := e.KubectlClient.GetKubeadmConfigTemplate(ctx, templateName)
		if err != nil {
			e.T.Fatal(fmt.Errorf("failed to get kubeadmconfigtemplate for worker node %s when validating taints: %v", w.Name, err))
		}
		nodeTaints := template.Spec.Template.Spec.JoinConfiguration.NodeRegistration.Taints
		if !v1alpha1.TaintsSliceEqual(w.Taints, nodeTaints) {
			e.T.Fatal(fmt.Errorf("failed to validate taints for node %v: taints in kubeadmconfigtemplate are not equal to taints in worker node group configuration spec. Expected %v, got %v", w.Name, w.Taints, nodeTaints))
		}

		ms, err := e.KubectlClient.GetMachineSets(ctx, md.Name, e.cluster())
		if err != nil {
			e.T.Fatal(fmt.Errorf("failed to get machine sets when validating taints: %v", err))
		}

		if len(ms) != 1 {
			e.T.Fatal(fmt.Errorf("invalid number of machine sets associated with worker node %v", w.Name))
		}

		for _, node := range nodes {
			ownerName, ok := node.Annotations[ownerAnnotation]
			if ok {
				if ownerName == ms[0].Name {
					if !v1alpha1.TaintsSliceEqual(node.Spec.Taints, w.Taints) {
						e.T.Fatal(fmt.Errorf("taints on node %v and corresponding worker node group configuration %v do not match", node.Name, w.Name))
					}
				}
			}
		}
	}
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
