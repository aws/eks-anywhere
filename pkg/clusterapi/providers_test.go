package clusterapi_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestGetProvidersEmpty(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	client := fake.NewClientBuilder().
		WithRuntimeObjects().
		Build()

	providers, err := clusterapi.GetProviders(ctx, client)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(providers).To(BeEmpty())
}

func TestGetProvidersMultipleProviders(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	providersWant := []clusterctlv1.Provider{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "kubeadm-controlplane",
				ResourceVersion: "1",
			},
			Type:         string(clusterctlv1.ControlPlaneProviderType),
			ProviderName: "kubeadm",
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "vsphere",
				ResourceVersion: "1",
			},
			Type:         string(clusterctlv1.InfrastructureProviderType),
			ProviderName: "vsphere",
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "unknown",
				ResourceVersion: "1",
			},
			Type:         string(clusterctlv1.InfrastructureProviderType),
			ProviderName: "unknown-provider",
		},
	}

	providerObjs := make([]runtime.Object, 0, len(providersWant))
	for _, p := range providersWant {
		provider := p
		providerObjs = append(providerObjs, &provider)
	}

	client := fake.NewClientBuilder().
		WithRuntimeObjects(providerObjs...).
		Build()

	providers, err := clusterapi.GetProviders(ctx, client)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(providers).To(ConsistOf(providersWant))
}

func TestGetProvidersError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	client := fake.NewClientBuilder().
		WithRuntimeObjects().
		// using an empty scheme will fail since it doesn't have the clusterctlv1 api
		WithScheme(runtime.NewScheme()).
		Build()

	_, err := clusterapi.GetProviders(ctx, client)
	g.Expect(err).To(HaveOccurred())
}
