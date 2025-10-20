package serverside_test

import (
	"context"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
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
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
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
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
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
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#
spec:
  paused: true
---
apiVersion: cluster.x-k8s.io/v1beta1
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

			g.Expect(serverside.ReconcileYaml(ctx, c, tt.yaml)).To(Succeed(), "Failed to reconcile with ReconcileYaml()")

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

func TestReconcileUpdateObject(t *testing.T) {
	cluster1 := newCluster("cluster-1")

	yaml := []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#
spec:
  controlPlaneEndpoint:
    host: 1.1.1.1
    port: 8081`)

	initial := []*clusterapiv1.Cluster{
		updatedCluster(cluster1, func(c capiCluster) {
			c.Spec.ControlPlaneEndpoint.Port = 8080
			c.Spec.ControlPlaneEndpoint.Host = "1.1.1.1"
		}),
	}

	expected := []*clusterapiv1.Cluster{
		updatedCluster(cluster1, func(c capiCluster) {
			c.Spec.ControlPlaneEndpoint.Port = 8081
			c.Spec.ControlPlaneEndpoint.Host = "1.1.1.1"
		}),
	}

	c := env.Client()
	reader := env.APIReader()
	ctx := context.Background()

	g := NewWithT(t)
	ns := env.CreateNamespaceForTest(ctx, t)

	for _, o := range initial {
		o.SetNamespace(ns)

		if err := c.Create(ctx, o); err != nil {
			t.Fatal(err)
		}
	}

	yaml = []byte(strings.ReplaceAll(string(yaml), "#namespace#", ns))

	objs, err := clientutil.YamlToClientObjects(yaml)
	if err != nil {
		t.Fatal(err)
	}

	objs[0].SetResourceVersion(initial[0].GetResourceVersion())
	for _, o := range objs {
		g.Expect(serverside.UpdateObject(ctx, c, o)).To(Succeed(), "Failed to reconcile with UpdateObject()")
	}

	for _, o := range expected {
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
}

func TestReconcileUpdateObjectError(t *testing.T) {
	cluster1 := newCluster("cluster-1")

	yaml := []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#
spec:
  controlPlaneEndpoint:
    host: 1.1.1.1
    port: 8081`)

	initial := []*clusterapiv1.Cluster{
		updatedCluster(cluster1, func(c capiCluster) {
			c.Spec.ControlPlaneEndpoint.Port = 8080
			c.Spec.ControlPlaneEndpoint.Host = "1.1.1.1"
		}),
	}

	c := env.Client()
	ctx := context.Background()

	g := NewWithT(t)
	ns := env.CreateNamespaceForTest(ctx, t)

	for _, o := range initial {
		o.SetNamespace(ns)

		if err := c.Create(ctx, o); err != nil {
			t.Fatal(err)
		}
	}

	yaml = []byte(strings.ReplaceAll(string(yaml), "#namespace#", ns))

	objs, err := clientutil.YamlToClientObjects(yaml)
	if err != nil {
		t.Fatal(err)
	}

	for _, o := range objs {
		err := serverside.UpdateObject(ctx, c, o)
		g.Expect(err).To(MatchError(ContainSubstring("must be specified for an update")))
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

func TestDeleteYaml(t *testing.T) {
	cluster1 := newCluster("cluster-1")
	cluster2 := newCluster("cluster-2")
	tests := []struct {
		name          string
		initialObjs   []*clusterapiv1.Cluster
		yaml          []byte
		expectError   bool
		errorContains string
	}{
		{
			name: "delete existing object",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#`),
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
			},
			expectError: false,
		},
		{
			name: "delete non-existent object idempotent",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-nonexistent
  namespace: #namespace#`),
			expectError: false,
		},
		{
			name: "delete multiple objects",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-2
  namespace: #namespace#`),
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
				cluster2.DeepCopy(),
			},
			expectError: false,
		},
		{
			name: "delete mixed existing and non-existing",
			yaml: []byte(`apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-1
  namespace: #namespace#
---
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: cluster-nonexistent
  namespace: #namespace#`),
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
			},
			expectError: false,
		},
		{
			name:          "invalid yaml",
			yaml:          []byte(`invalid: yaml: content: [[[`),
			expectError:   true,
			errorContains: "failed to unmarshal",
		},
	}

	c := env.Client()
	reader := env.APIReader()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ns := env.CreateNamespaceForTest(ctx, t)

			// Create initial objects
			for _, o := range tt.initialObjs {
				o.SetNamespace(ns)
				if err := c.Create(ctx, o); err != nil {
					t.Fatal(err)
				}
			}

			// Replace namespace placeholder in YAML
			tt.yaml = []byte(strings.ReplaceAll(string(tt.yaml), "#namespace#", ns))

			// Execute delete
			err := serverside.DeleteYaml(ctx, c, tt.yaml)

			if tt.expectError {
				g.Expect(err).To(HaveOccurred())
				if tt.errorContains != "" {
					g.Expect(err.Error()).To(ContainSubstring(tt.errorContains))
				}
				return
			}

			g.Expect(err).ToNot(HaveOccurred())

			// Verify objects are deleted
			for _, o := range tt.initialObjs {
				key := client.ObjectKey{
					Namespace: ns,
					Name:      o.GetName(),
				}
				cluster := &clusterapiv1.Cluster{}
				err := reader.Get(ctx, key, cluster)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("not found"))
			}
		})
	}
}

func TestDeleteObjects(t *testing.T) {
	cluster1 := newCluster("cluster-1")
	cluster2 := newCluster("cluster-2")
	tests := []struct {
		name        string
		initialObjs []*clusterapiv1.Cluster
		deleteObjs  []*clusterapiv1.Cluster
		expectError bool
	}{
		{
			name: "delete existing objects",
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
				cluster2.DeepCopy(),
			},
			deleteObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
				cluster2.DeepCopy(),
			},
			expectError: false,
		},
		{
			name: "delete non-existent objects idempotent",
			deleteObjs: []*clusterapiv1.Cluster{
				newCluster("nonexistent-1"),
				newCluster("nonexistent-2"),
			},
			expectError: false,
		},
		{
			name: "delete mixed existing and non-existing",
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
			},
			deleteObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
				newCluster("nonexistent"),
			},
			expectError: false,
		},
		{
			name:        "delete empty object slice",
			deleteObjs:  []*clusterapiv1.Cluster{},
			expectError: false,
		},
		{
			name: "delete single object",
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
			},
			deleteObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
			},
			expectError: false,
		},
	}

	c := env.Client()
	reader := env.APIReader()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ns := env.CreateNamespaceForTest(ctx, t)

			// Create initial objects
			for _, o := range tt.initialObjs {
				o.SetNamespace(ns)
				if err := c.Create(ctx, o); err != nil {
					t.Fatal(err)
				}
			}

			// Set namespace for delete objects
			var deleteObjs []client.Object
			for _, o := range tt.deleteObjs {
				o.SetNamespace(ns)
				deleteObjs = append(deleteObjs, o)
			}

			// Execute delete
			err := serverside.DeleteObjects(ctx, c, deleteObjs)

			if tt.expectError {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).ToNot(HaveOccurred())

			// Verify initially existing objects are deleted
			for _, o := range tt.initialObjs {
				key := client.ObjectKey{
					Namespace: ns,
					Name:      o.GetName(),
				}
				cluster := &clusterapiv1.Cluster{}
				err := reader.Get(ctx, key, cluster)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("not found"))
			}
		})
	}
}

func TestDeleteObject(t *testing.T) {
	cluster1 := newCluster("cluster-1")
	tests := []struct {
		name          string
		initialObjs   []*clusterapiv1.Cluster
		deleteObj     *clusterapiv1.Cluster
		expectError   bool
		errorContains string
	}{
		{
			name: "delete existing object",
			initialObjs: []*clusterapiv1.Cluster{
				cluster1.DeepCopy(),
			},
			deleteObj:   cluster1.DeepCopy(),
			expectError: false,
		},
		{
			name:        "delete non-existent object idempotent",
			deleteObj:   newCluster("nonexistent"),
			expectError: false,
		},
	}

	c := env.Client()
	reader := env.APIReader()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ns := env.CreateNamespaceForTest(ctx, t)

			// Create initial objects
			for _, o := range tt.initialObjs {
				o.SetNamespace(ns)
				if err := c.Create(ctx, o); err != nil {
					t.Fatal(err)
				}
			}

			// Set namespace for delete object
			tt.deleteObj.SetNamespace(ns)

			// Execute delete
			err := serverside.DeleteObject(ctx, c, tt.deleteObj)

			if tt.expectError {
				g.Expect(err).To(HaveOccurred())
				if tt.errorContains != "" {
					g.Expect(err.Error()).To(ContainSubstring(tt.errorContains))
				}
				return
			}

			g.Expect(err).ToNot(HaveOccurred())

			// If there were initial objects, verify they are deleted
			if len(tt.initialObjs) > 0 {
				key := client.ObjectKey{
					Namespace: ns,
					Name:      tt.deleteObj.GetName(),
				}
				cluster := &clusterapiv1.Cluster{}
				err := reader.Get(ctx, key, cluster)
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("not found"))
			}
		})
	}
}
