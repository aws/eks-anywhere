package kubernetes_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterapiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
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
		{
			name:             "capi kubeadm controlplane",
			namespace:        "eksa-system",
			obj:              &controlplanev1.KubeadmControlPlane{},
			wantResourceType: "KubeadmControlPlane.v1beta1.controlplane.cluster.x-k8s.io",
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

	g.Expect(c.Get(ctx, "name", "namespace", "kubeconfig", &metav1.Status{})).Error()
}

func TestUnAuthClientDeleteSuccess(t *testing.T) {
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

			kubectl.EXPECT().Delete(ctx, tt.wantResourceType, tt.name, tt.namespace, kubeconfig)

			c := kubernetes.NewUnAuthClient(kubectl)
			g.Expect(c.Init()).To(Succeed())

			g.Expect(c.Delete(ctx, tt.name, tt.namespace, kubeconfig, tt.obj)).To(Succeed())
		})
	}
}

func TestUnAuthClientApplySuccess(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		obj       runtime.Object
	}{
		{
			name:      "eksa cluster",
			namespace: "eksa-system",
			obj:       &anywherev1.Cluster{},
		},
		{
			name:      "capi cluster",
			namespace: "eksa-system",
			obj:       &clusterapiv1.Cluster{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			kubectl := mocks.NewMockKubectlGetter(ctrl)
			kubeconfig := "k.kubeconfig"

			kubectl.EXPECT().Apply(ctx, kubeconfig, tt.obj)

			c := kubernetes.NewUnAuthClient(kubectl)
			g.Expect(c.Init()).To(Succeed())

			g.Expect(c.Apply(ctx, kubeconfig, tt.obj)).To(Succeed())
		})
	}
}

func TestUnAuthClientDeleteUnknownObjType(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectlGetter(ctrl)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())

	g.Expect(c.Delete(ctx, "name", "namespace", "kubeconfig", &metav1.Status{})).Error()
}
