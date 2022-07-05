package unstructuredutil_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/pkg/unstructuredutil"
)

func TestYamlToClientObjects(t *testing.T) {
	tests := []struct {
		name string
		yaml []byte
		want map[string]unstructured.Unstructured
	}{
		{
			name: "two objects",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: ns-1
spec:
  paused: true
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-2
  namespace: ns-1
spec:
  controlPlaneEndpoint:
    host: 1.1.1.1
    port: 8080`),
			want: map[string]unstructured.Unstructured{
				"cluster-1": {
					Object: map[string]interface{}{
						"apiVersion": "cluster.x-k8s.io/v1beta1",
						"kind":       "Cluster",
						"metadata": map[string]interface{}{
							"name":      "cluster-1",
							"namespace": "ns-1",
						},
						"spec": map[string]interface{}{
							"paused": true,
						},
					},
				},
				"cluster-2": {
					Object: map[string]interface{}{
						"apiVersion": "cluster.x-k8s.io/v1beta1",
						"kind":       "Cluster",
						"metadata": map[string]interface{}{
							"name":      "cluster-2",
							"namespace": "ns-1",
						},
						"spec": map[string]interface{}{
							"controlPlaneEndpoint": map[string]interface{}{
								"host": "1.1.1.1",
								"port": float64(8080),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := unstructuredutil.YamlToUnstructured(tt.yaml)
			g.Expect(err).To(BeNil(), "YamlToClientObjects() returned an error")
			g.Expect(len(got)).To(Equal(len(tt.want)), "Should have got %d objects", len(tt.want))
			for _, obj := range got {
				g.Expect(obj).To(Equal(tt.want[obj.GetName()]))
			}
		})
	}
}
