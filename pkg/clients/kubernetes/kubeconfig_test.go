package kubernetes_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes/mocks"
)

func TestKubeconfigClientGet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"

	name := "eksa cluster"
	namespace := "eksa-system"
	obj := &anywherev1.Cluster{}
	wantResourceType := "Cluster.v1alpha1.anywhere.eks.amazonaws.com"

	kubectl.EXPECT().Get(
		ctx, wantResourceType, kubeconfig, obj,
		&kubernetes.KubectlGetOptions{Name: name, Namespace: namespace},
	)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())
	kc, err := c.BuildClientFromKubeconfig(kubeconfig)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(kc.Get(ctx, name, namespace, obj)).To(Succeed())
}

func TestKubeconfigClientList(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	list := &corev1.NodeList{}

	kubectl.EXPECT().Get(ctx, "Node", kubeconfig, list)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())
	kc, err := c.BuildClientFromKubeconfig(kubeconfig)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(kc.List(ctx, list)).To(Succeed())
}

func TestKubeconfigClientCreate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	obj := &corev1.Pod{}

	kubectl.EXPECT().Create(ctx, kubeconfig, obj)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())
	kc, err := c.BuildClientFromKubeconfig(kubeconfig)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(kc.Create(ctx, obj)).To(Succeed())
}

func TestKubeconfigClientUpdate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	obj := &corev1.Pod{}

	kubectl.EXPECT().Replace(ctx, kubeconfig, obj)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())
	kc, err := c.BuildClientFromKubeconfig(kubeconfig)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(kc.Update(ctx, obj)).To(Succeed())
}

func TestKubeconfigClientApplyServerSide(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	fieldManager := "my-manager"
	kubeconfig := "k.kubeconfig"
	obj := &corev1.Pod{}

	opts := kubernetes.KubectlApplyOptions{
		ServerSide:   true,
		FieldManager: fieldManager,
	}
	kubectl.EXPECT().Apply(ctx, kubeconfig, obj, opts)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())
	kc, err := c.BuildClientFromKubeconfig(kubeconfig)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(kc.ApplyServerSide(ctx, fieldManager, obj)).To(Succeed())
}

func TestKubeconfigClientDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	obj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "my-ns",
		},
	}

	opts := &kubernetes.KubectlDeleteOptions{
		Name:      "my-pod",
		Namespace: "my-ns",
	}
	kubectl.EXPECT().Delete(ctx, "Pod", kubeconfig, opts)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())
	kc, err := c.BuildClientFromKubeconfig(kubeconfig)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(kc.Delete(ctx, obj)).To(Succeed())
}

func TestKubeconfigClientDeleteAllOf(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	kubectl := mocks.NewMockKubectl(ctrl)
	kubeconfig := "k.kubeconfig"
	obj := &corev1.Pod{}

	kubectlOpts := &kubernetes.KubectlDeleteOptions{
		Namespace: "my-ns",
		HasLabels: map[string]string{
			"k": "v",
		},
	}
	kubectl.EXPECT().Delete(ctx, "Pod", kubeconfig, kubectlOpts)

	c := kubernetes.NewUnAuthClient(kubectl)
	g.Expect(c.Init()).To(Succeed())
	kc, err := c.BuildClientFromKubeconfig(kubeconfig)
	g.Expect(err).NotTo(HaveOccurred())

	deleteOpts := &kubernetes.DeleteAllOfOptions{
		Namespace: "my-ns",
		HasLabels: map[string]string{
			"k": "v",
		},
	}
	g.Expect(kc.DeleteAllOf(ctx, obj, deleteOpts)).To(Succeed())
}
