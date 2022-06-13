package cluster_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
)

func TestDefaultConfigClientBuilderFluxConfig(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			GitOpsRef: &anywherev1.Ref{
				Kind: anywherev1.FluxConfigKind,
				Name: "my-flux",
			},
		},
	}
	fluxConfig := &anywherev1.FluxConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-flux",
			Namespace: "default",
		},
	}

	client.EXPECT().Get(ctx, "my-flux", "default", &anywherev1.FluxConfig{}).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			c := obj.(*anywherev1.FluxConfig)
			c.ObjectMeta = fluxConfig.ObjectMeta
			return nil
		},
	)

	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.Cluster).To(Equal(cluster))
	g.Expect(config.FluxConfig).To(Equal(fluxConfig))
}
