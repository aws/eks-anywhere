package cluster_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestGetBundlesForCluster(t *testing.T) {
	testCases := []struct {
		testName                string
		cluster                 *v1alpha1.Cluster
		wantName, wantNamespace string
	}{
		{
			testName: "no bundles ref",
			cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eksa-cluster",
					Namespace: "eksa",
				},
				Spec: v1alpha1.ClusterSpec{},
			},
			wantName:      "eksa-cluster",
			wantNamespace: "eksa",
		},
		{
			testName: "bundles ref",
			cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eksa-cluster",
					Namespace: "eksa",
				},
				Spec: v1alpha1.ClusterSpec{
					BundlesRef: &v1alpha1.BundlesRef{
						Name:       "bundles-1",
						Namespace:  "eksa-system",
						APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
					},
				},
			},
			wantName:      "bundles-1",
			wantNamespace: "eksa-system",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			wantBundles := &v1alpha1release.Bundles{}
			mockFetch := func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error) {
				g.Expect(name).To(Equal(tt.wantName))
				g.Expect(namespace).To(Equal(tt.wantNamespace))

				return wantBundles, nil
			}

			gotBundles, err := cluster.GetBundlesForCluster(context.Background(), tt.cluster, mockFetch)
			g.Expect(err).To(BeNil())
			g.Expect(gotBundles).To(Equal(wantBundles))
		})
	}
}

func TestGetFluxConfigForClusterIsNil(t *testing.T) {
	g := NewWithT(t)
	c := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-cluster",
			Namespace: "eksa",
		},
		Spec: v1alpha1.ClusterSpec{
			GitOpsRef: nil,
		},
	}
	var wantFlux *v1alpha1.FluxConfig
	mockFetch := func(ctx context.Context, name, namespace string) (*v1alpha1.FluxConfig, error) {
		g.Expect(name).To(Equal(c.Name))
		g.Expect(namespace).To(Equal(c.Namespace))

		return wantFlux, nil
	}

	gotFlux, err := cluster.GetFluxConfigForCluster(context.Background(), c, mockFetch)
	g.Expect(err).To(BeNil())
	g.Expect(gotFlux).To(Equal(wantFlux))
}

func TestGetFluxConfigForCluster(t *testing.T) {
	g := NewWithT(t)
	c := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-cluster",
			Namespace: "eksa",
		},
		Spec: v1alpha1.ClusterSpec{
			GitOpsRef: &v1alpha1.Ref{
				Kind: v1alpha1.FluxConfigKind,
				Name: "eksa-cluster",
			},
		},
	}
	wantFlux := &v1alpha1.FluxConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.FluxConfigSpec{},
		Status:     v1alpha1.FluxConfigStatus{},
	}
	mockFetch := func(ctx context.Context, name, namespace string) (*v1alpha1.FluxConfig, error) {
		g.Expect(name).To(Equal(c.Name))
		g.Expect(namespace).To(Equal(c.Namespace))

		return wantFlux, nil
	}

	gotFlux, err := cluster.GetFluxConfigForCluster(context.Background(), c, mockFetch)
	g.Expect(err).To(BeNil())
	g.Expect(gotFlux).To(Equal(wantFlux))
}
