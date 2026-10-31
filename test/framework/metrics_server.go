package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/retrier"
)

// VerifyMetricsAPIServing checks that the resource metrics API, which the metrics-server package serves,
// reports CPU and memory usage for the cluster's nodes.
func (e *ClusterE2ETest) VerifyMetricsAPIServing() {
	ctx := context.Background()
	var nodes []string
	err := retrier.New(5*time.Minute, retrier.WithRetryPolicy(retrier.BackOffPolicy(10*time.Second))).Retry(func() error {
		out, err := e.KubectlClient.ExecuteCommand(ctx, "get", "--raw", "/apis/metrics.k8s.io/v1beta1/nodes",
			"--kubeconfig", e.KubeconfigFilePath())
		if err != nil {
			return err
		}
		nodes, err = nodesWithUsage(out.Bytes())
		return err
	})
	if err != nil {
		e.T.Fatalf("Waiting for the metrics API to report node usage: %v", err)
	}
	e.T.Logf("The metrics API reports CPU and memory usage for nodes %v", nodes)
}

type nodeMetricsList struct {
	Items []struct {
		Metadata metav1.ObjectMeta `json:"metadata"`
		Usage    map[string]string `json:"usage"`
	} `json:"items"`
}

// nodesWithUsage parses a metrics.k8s.io NodeMetricsList and returns its node names. It fails if the list is
// empty or a node has no CPU or memory usage.
func nodesWithUsage(data []byte) ([]string, error) {
	list := &nodeMetricsList{}
	if err := json.Unmarshal(data, list); err != nil {
		return nil, fmt.Errorf("parsing node metrics: %v", err)
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("the metrics API returned no nodes")
	}
	nodes := make([]string, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Usage["cpu"] == "" || item.Usage["memory"] == "" {
			return nil, fmt.Errorf("node %s has no CPU or memory usage: %v", item.Metadata.Name, item.Usage)
		}
		nodes = append(nodes, item.Metadata.Name)
	}
	return nodes, nil
}
