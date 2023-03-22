package resource

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestClusterReconcilerApplyTemplatesAnnotationsArePreserved(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster-test",
		},
	}
	kcp := &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterapi.KubeadmControlPlaneName(cluster),
			Namespace: constants.EksaSystemNamespace,
			Annotations: map[string]string{
				"my-custom-annotation": "true",
			},
		},
	}
	newKCP := kcp.DeepCopy()
	newKCP.Annotations = map[string]string{
		"eksa-annotation": "false",
	}
	newKCPUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newKCP)
	g.Expect(err).NotTo(HaveOccurred())

	resources := []*unstructured.Unstructured{{Object: newKCPUnstructured}}

	client := fake.NewClientBuilder().WithObjects(cluster, kcp).Build()
	log := logr.Discard()

	r := NewClusterReconciler(
		NewCAPIResourceFetcher(client, log),
		NewCAPIResourceUpdater(client, log),
		time.Now,
		log,
	)

	g.Expect(r.applyTemplates(ctx, cluster, resources, false)).To(Succeed())

	api := envtest.NewAPIExpecter(t, client)
	api.ShouldEventuallyMatch(ctx, kcp, func(g Gomega) {
		g.Expect(kcp.Annotations).To(HaveKeyWithValue("my-custom-annotation", "true"))
		g.Expect(kcp.Annotations).To(HaveKeyWithValue("eksa-annotation", "false"))
	})
}

func TestClusterReconcilerApplyTemplatesNoExistingAnnotations(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster-test",
		},
	}
	kcp := &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterapi.KubeadmControlPlaneName(cluster),
			Namespace: constants.EksaSystemNamespace,
		},
	}
	newKCP := kcp.DeepCopy()
	newKCP.Annotations = map[string]string{
		"eksa-annotation": "false",
	}
	newKCPUnstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newKCP)
	g.Expect(err).NotTo(HaveOccurred())

	resources := []*unstructured.Unstructured{{Object: newKCPUnstructured}}

	client := fake.NewClientBuilder().WithObjects(cluster, kcp).Build()
	log := logr.Discard()

	r := NewClusterReconciler(
		NewCAPIResourceFetcher(client, log),
		NewCAPIResourceUpdater(client, log),
		time.Now,
		log,
	)

	g.Expect(r.applyTemplates(ctx, cluster, resources, false)).To(Succeed())

	api := envtest.NewAPIExpecter(t, client)
	api.ShouldEventuallyMatch(ctx, kcp, func(g Gomega) {
		g.Expect(kcp.Annotations).To(HaveKeyWithValue("eksa-annotation", "false"))
	})
}
