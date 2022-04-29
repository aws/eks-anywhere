package kubernetes_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
)

func TestKubeconfigClientGet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectlGetter(ctrl)
	kubeconfig := "k.kubeconfig"

	name := "eksa cluster"
	namespace := "eksa-system"
	obj := &anywherev1.Cluster{}
	wantResourceType := "Cluster.v1alpha1.anywhere.eks.amazonaws.com"

	kubectl.EXPECT().GetObject(ctx, wantResourceType, name, namespace, kubeconfig, obj)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())
	kc := c.KubeconfigClient(kubeconfig)

	g.Expect(kc.Get(ctx, name, namespace, obj)).To(Succeed())
}
