package framework

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/retrier"
)

func (e *ClusterE2ETest) ValidateNodeTaints(expectedTaints map[corev1.Taint]int) {
	ctx := context.Background()
	e.T.Log("Validating cluster node taints")
	r := retrier.New(time.Minute)
	var nodes []corev1.Node
	var err error
	err = r.Retry(func() error {
		nodes, err = e.KubectlClient.GetNodes(ctx, e.cluster().KubeconfigFile)
		if err != nil {
			return fmt.Errorf("error validating nodes taint presence: %v", err)
		}
		return nil
	})
	if err != nil {
		e.T.Fatal(err)
	}
	for _, node := range nodes {
		for _, taint := range node.Spec.Taints {
			if taint == masterNodeTaint() {
				continue
			}
			_, ok := expectedTaints[taint]
			if !ok {
				e.T.Fatal(fmt.Errorf("node %s has taint %v, but none are expected", node.Name, taint))
			}
			expectedTaints[taint] -= 1
			if expectedTaints[taint] == 0 {
				delete(expectedTaints, taint)
			}
		}
	}
	if len(expectedTaints) > 0 {
		e.T.Fatal(fmt.Errorf("expected taints %v missing from cluster", expectedTaints))
	}
}

func masterNodeTaint() corev1.Taint {
	return corev1.Taint{
		Key:    "node-role.kubernetes.io/master",
		Effect: "NoSchedule",
	}
}
