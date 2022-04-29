package kubernetes_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clusterapiv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestUnAuthClientGetSuccess(t *testing.T) {
	tests := []struct {
		name             string
		namespace        string
		obj              runtime.Object
		wantResourceType string
	}{
		{
			name:             "eksa cluster",
			namespace:        "eksa-system",
			obj:              &anywherev1.Cluster{},
			wantResourceType: "Cluster.v1alpha1.anywhere.eks.amazonaws.com",
		},
		{
			name:             "capi cluster",
			namespace:        "eksa-system",
			obj:              &clusterapiv1.Cluster{},
			wantResourceType: "Cluster.v1beta1.cluster.x-k8s.io",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			kubectl := mocks.NewMockKubectlGetter(ctrl)
			kubeconfig := "k.kubeconfig"

			kubectl.EXPECT().GetObject(ctx, tt.wantResourceType, tt.name, tt.namespace, kubeconfig, tt.obj)

			c := kubernetes.NewUnAuthClient(kubectl)
			g.Expect(c.Init()).To(Succeed())

			g.Expect(c.Get(ctx, tt.name, tt.namespace, kubeconfig, tt.obj)).To(Succeed())
		})
	}
}

func TestUnAuthClientGetUnknownObjType(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectlGetter(ctrl)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())

	g.Expect(c.Get(ctx, "name", "namespace", "kubeconfig", &releasev1.Release{})).Error()
}
