package clientutil_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

func TestDeleteYamlSuccess(t *testing.T) {
	tests := []struct {
		name        string
		initialObjs []client.Object
		yaml        []byte
	}{
		{
			name: "delete single object",
			initialObjs: []client.Object{
				cluster("cluster-1"),
			},
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: default
spec:
  controlPlaneEndpoint:
    host: 1.1.1.1
    port: 8080`),
		},
		{
			name: "delete multiple objects",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: default
spec:
  paused: true
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-2
  namespace: default
spec:
  paused: true`),
			initialObjs: []client.Object{
				cluster("cluster-1"),
				cluster("cluster-2"),
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			c := fake.NewClientBuilder().WithObjects(tt.initialObjs...).Build()

			g.Expect(clientutil.DeleteYaml(ctx, c, tt.yaml)).To(Succeed(), "Failed to delete with DeleteYaml()")

			for _, o := range tt.initialObjs {
				key := client.ObjectKey{
					Namespace: "default",
					Name:      o.GetName(),
				}

				cluster := &clusterapiv1.Cluster{}
				err := c.Get(ctx, key, cluster)
				g.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "Object should have been deleted")
			}
		})
	}
}

func cluster(name string) *clusterapiv1.Cluster {
	c := &clusterapiv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterapiv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}

	return c
}

func TestDeleteYamlError(t *testing.T) {
	tests := []struct {
		name    string
		yaml    []byte
		wantErr string
	}{
		{
			name:    "invalid yaml",
			yaml:    []byte(`x`),
			wantErr: "error unmarshaling JSON",
		},
		{
			name: "error deleting",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: default
spec:
  paused: true`),
			wantErr: "deleting object cluster.x-k8s.io/v1beta1, Kind=Cluster, default/cluster-1",
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			c := fake.NewClientBuilder().Build()

			g.Expect(clientutil.DeleteYaml(ctx, c, tt.yaml)).To(MatchError(ContainSubstring(tt.wantErr)))
		})
	}
}
