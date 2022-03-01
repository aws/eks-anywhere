package cluster_test

import (
	"context"
	"testing"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestGetBundlesForCluster(t *testing.T) {
	g := NewWithT(t)
	c := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-cluster",
			Namespace: "eksa",
		},
	}
	wantBundles := &v1alpha1release.Bundles{}
	mockFetch := func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error) {
		g.Expect(name).To(Equal(c.Name))
		g.Expect(namespace).To(Equal(c.Namespace))

		return wantBundles, nil
	}

	gotBundles, err := cluster.GetBundlesForCluster(context.Background(), c, mockFetch)
	g.Expect(err).To(BeNil())
	g.Expect(gotBundles).To(Equal(wantBundles))
}

func TestGetEksdReleaseForCluster(t *testing.T) {
	g := NewWithT(t)
	c := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-cluster",
			Namespace: "eksa",
		},
		Status: v1alpha1.ClusterStatus{
			EksdReleaseRef: &v1alpha1.EksdReleaseRef{
				Name:      "eks-d",
				Namespace: "eksa-system",
			},
		},
	}
	wantRelease := &eksdv1alpha1.Release{}
	mockFetch := func(ctx context.Context, name, namespace string) (*eksdv1alpha1.Release, error) {
		g.Expect(name).To(Equal(c.Status.EksdReleaseRef.Name))
		g.Expect(namespace).To(Equal(c.Status.EksdReleaseRef.Namespace))

		return wantRelease, nil
	}

	gotRelease, err := cluster.GetEksdReleaseForCluster(context.Background(), c, mockFetch)
	g.Expect(err).To(BeNil())
	g.Expect(gotRelease).To(Equal(wantRelease))
}
