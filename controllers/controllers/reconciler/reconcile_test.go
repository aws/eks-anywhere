package reconciler_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/controllers/controllers/reconciler"
)

func TestReconcileYaml(t *testing.T) {
	cluster1 := newCluster("cluster-1")
	cluster2 := newCluster("cluster-2")
	tests := []struct {
		name         string
		initialObjs  []*clusterapiv1.Cluster
		yaml         []byte
		expectedObjs []*clusterapiv1.Cluster
	}{
		{
			name: "new object",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#
spec:
  controlPlaneEndpoint:
    host: 1.1.1.1
    port: 8080`),
			expectedObjs: []*clusterapiv1.Cluster{
				updatedCluster(cluster1, func(c capiCluster) {
					c.Spec.ControlPlaneEndpoint.Port = 8080
					c.Spec.ControlPlaneEndpoint.Host = "1.1.1.1"
				}),
			},
		},
		{
			name: "existing object",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#
spec:
  paused: true`),
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
			},
			expectedObjs: []*clusterapiv1.Cluster{
				updatedCluster(cluster1, func(c capiCluster) { c.Spec.Paused = true }),
			},
		},
		{
			name: "new and existing object",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#
spec:
  paused: true
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: cluster-2
  namespace: #namespace#
spec:
  paused: true`),
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
			},
			expectedObjs: []*clusterapiv1.Cluster{
				updatedCluster(cluster1, func(c capiCluster) { c.Spec.Paused = true }),
				updatedCluster(cluster2, func(c capiCluster) { c.Spec.Paused = true }),
			},
		},
	}

	c := env.Client()
	reader := env.APIReader()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ns := env.CreateNamespaceForTest(ctx, t)

			for _, o := range tt.initialObjs {
				o.SetNamespace(ns)

				if err := c.Create(ctx, o); err != nil {
					t.Fatal(err)
				}
			}

			tt.yaml = []byte(strings.ReplaceAll(string(tt.yaml), "#namespace#", ns))

			g.Expect(reconciler.ReconcileYaml(ctx, c, tt.yaml)).To(Succeed(), "Failed to reconcile with ReconcileYaml()")

			for _, o := range tt.expectedObjs {
				key := client.ObjectKey{
					Namespace: ns,
					Name:      o.GetName(),
				}

				cluster := &clusterapiv1.Cluster{}

				g.Expect(reader.Get(ctx, key, cluster)).To(Succeed(), "Failed getting obj from cluster")
				g.Expect(
					equality.Semantic.DeepDerivative(o.Spec, cluster.Spec),
				).To(BeTrue(), "Object spec in cluster is not equal to expected object spec:\n Actual:\n%#v\n Expected:\n%#v", cluster.Spec, o.Spec)
			}
		})
	}
}

type capiCluster = *clusterapiv1.Cluster

func newCluster(name string, changes ...func(capiCluster)) *clusterapiv1.Cluster {
	c := &clusterapiv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterapiv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	for _, change := range changes {
		change(c)
	}

	return c
}

func updatedCluster(cluster *clusterapiv1.Cluster, f func(*clusterapiv1.Cluster)) *clusterapiv1.Cluster {
	copy := cluster.DeepCopy()
	f(copy)
	return copy
}
