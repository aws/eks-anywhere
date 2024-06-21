package framework

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	maxPod50 = 50
	maxPod60 = 60
)

// WithKubeletConfig returns the default kubelet config set for e2e test.
func WithKubeletConfig() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.addClusterConfigFillers(WithKubeletClusterConfig())
	}
}

// WithKubeletClusterConfig returns a ClusterConfigFiller that adds the default
// KubeletConfig for E2E tests to the cluster Config.
func WithKubeletClusterConfig() api.ClusterConfigFiller {
	cpKubeletConfiguration := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"maxPods": maxPod50,
			"kind":    "KubeletConfiguration",
		},
	}

	wnKubeletConfiguration := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"maxPods": maxPod60,
			"kind":    "KubeletConfiguration",
		},
	}

	return api.JoinClusterConfigFillers(
		api.ClusterToConfigFiller(api.WithControlPlaneKubeletConfig(cpKubeletConfiguration)),
		api.ClusterToConfigFiller(api.WithWorkerNodeKubeletConfig(wnKubeletConfiguration)),
	)
}

// ValidateKubeletConfig validates the kubelet config specified in the cluster spec has been applied to the nodes.
func (e *ClusterE2ETest) ValidateKubeletConfig() {
	ctx := context.Background()
	kubectlClient := buildLocalKubectl()

	e.T.Log("Getting control plane nodes for kubelet max pod verification")
	nodes, err := kubectlClient.GetControlPlaneNodes(ctx,
		e.KubeconfigFilePath(),
	)
	if err != nil {
		e.T.Fatalf("Error getting nodes: %v", err)
	}
	if len(nodes) == 0 {
		e.T.Fatalf("no control plane nodes found")
	}

	got, _ := nodes[0].Status.Capacity.Pods().AsInt64()
	if got != int64(maxPod50) {
		e.T.Fatalf("Node capacity for control plane pods not equal to %v", maxPod50)
	}

	e.T.Log("Successfully verified Kubelet Configuration for control plane nodes")

	e.T.Log("Getting control plane nodes for kubelet max pod verification")
	allNodes, err := kubectlClient.GetNodes(ctx,
		e.KubeconfigFilePath(),
	)
	if err != nil {
		e.T.Fatalf("Error getting nodes: %v", err)
	}
	if len(allNodes) == 0 {
		e.T.Fatalf("no nodes found")
	}

	e.T.Log("Getting worker nodes for kubelet max pod verification")
	var workerNode corev1.Node
	for i := range allNodes {
		if strings.Contains(allNodes[i].Name, "-md-") {
			workerNode = allNodes[i]
		}
	}
	if err != nil {
		e.T.Fatalf("Error getting nodes: %v", err)
	}

	got, _ = workerNode.Status.Capacity.Pods().AsInt64()
	if got != int64(maxPod60) {
		e.T.Fatalf("Node capacity for worker node pods not equal to %v", maxPod60)
	}

	e.T.Log("Successfully verified Kubelet Configuration for worker nodes")
}
