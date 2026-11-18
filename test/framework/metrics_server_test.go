package framework

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNodesWithUsage(t *testing.T) {
	tests := map[string]struct {
		data    string
		nodes   []string
		wantErr string
	}{
		"all nodes have usage": {
			data: `{"kind":"NodeMetricsList","items":[
				{"metadata":{"name":"cp"},"usage":{"cpu":"123456789n","memory":"1048576Ki"}},
				{"metadata":{"name":"md-0"},"usage":{"cpu":"2m","memory":"2Mi"}}]}`,
			nodes: []string{"cp", "md-0"},
		},
		"no nodes": {
			data:    `{"kind":"NodeMetricsList","items":[]}`,
			wantErr: "returned no nodes",
		},
		"node without memory": {
			data:    `{"items":[{"metadata":{"name":"cp"},"usage":{"cpu":"1m"}}]}`,
			wantErr: "node cp has no CPU or memory usage",
		},
		"invalid json": {
			data:    `not json`,
			wantErr: "parsing node metrics",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			g := NewWithT(t)
			nodes, err := nodesWithUsage([]byte(tc.data))
			if tc.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tc.wantErr)))
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(nodes).To(Equal(tc.nodes))
		})
	}
}
