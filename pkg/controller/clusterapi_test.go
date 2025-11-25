package controller_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

func TestGetCAPIClusterSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	capiCluster := capiCluster()
	client := fake.NewClientBuilder().WithObjects(eksaCluster, capiCluster).Build()

	g.Expect(controller.GetCAPICluster(ctx, client, eksaCluster)).To(Equal(capiCluster))
}

func TestGetCAPIClusterNoCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	client := fake.NewClientBuilder().WithObjects(eksaCluster).Build()

	g.Expect(controller.GetCAPICluster(ctx, client, eksaCluster)).To(BeNil())
}

func TestGetCAPIClusterError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	// This should make the client fail because CRDs are not registered
	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()

	_, err := controller.GetCAPICluster(ctx, client, eksaCluster)
	g.Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
}

func TestGetKubeadmControlPlaneSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	kubeadmControlPlane := kubeadmControlPlane()
	client := fake.NewClientBuilder().WithObjects(eksaCluster, kubeadmControlPlane).Build()

	g.Expect(controller.GetKubeadmControlPlane(ctx, client, eksaCluster)).To(Equal(kubeadmControlPlane))
}

func TestGetKubeadmControlPlaneMissingKubeadmControlPlane(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	client := fake.NewClientBuilder().WithObjects(eksaCluster).Build()

	g.Expect(controller.GetKubeadmControlPlane(ctx, client, eksaCluster)).To(BeNil())
}

func TestGetKubeadmControlPlaneError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	// This should make the client fail because CRDs are not registered
	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()

	_, err := controller.GetKubeadmControlPlane(ctx, client, eksaCluster)
	g.Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
}

func TestGetMachineDeploymentSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	machineDeployment := machineDeployment()
	client := fake.NewClientBuilder().WithObjects(eksaCluster, machineDeployment).Build()

	g.Expect(controller.GetMachineDeployment(ctx, client, "my-cluster")).To(Equal(machineDeployment))
}

func TestGetMachineDeploymentMissingMD(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	client := fake.NewClientBuilder().WithObjects(eksaCluster).Build()

	g.Expect(controller.GetMachineDeployment(ctx, client, "test")).To(BeNil())
}

func TestGetMachineDeploymentsSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	md1 := machineDeployment()
	md1.Name = "md-1"

	md1.Labels = map[string]string{
		clusterv1.ClusterNameLabel: eksaCluster.Name,
	}

	md2 := md1.DeepCopy()
	md2.Name = "md-2"

	client := fake.NewClientBuilder().WithObjects(eksaCluster, md1, md2).Build()

	g.Expect(controller.GetMachineDeployments(ctx, client, eksaCluster)).To(Equal([]clusterv1.MachineDeployment{*md1, *md2}))
}

func TestGetMachineDeploymentsMachineDeploymentsInDifferentClusters(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	machineDeployment1 := machineDeployment()
	machineDeployment1.Name = "md-1"
	machineDeployment1.Labels = map[string]string{
		clusterv1.ClusterNameLabel: eksaCluster.Name,
	}

	machineDeployment2 := machineDeployment()
	machineDeployment2.Name = "md-2"
	machineDeployment2.Labels = map[string]string{
		clusterv1.ClusterNameLabel: "other-cluster",
	}

	client := fake.NewClientBuilder().WithObjects(eksaCluster, machineDeployment1, machineDeployment2).Build()

	g.Expect(controller.GetMachineDeployments(ctx, client, eksaCluster)).To(Equal([]clusterv1.MachineDeployment{*machineDeployment1}))
}

func TestGetMachineDeploymentsError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	// This should make the client fail because CRDs are not registered
	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()

	_, err := controller.GetMachineDeployments(ctx, client, eksaCluster)
	g.Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
}

func TestGetMachineDeploymentError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	// This should make the client fail because CRDs are not registered
	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()

	_, err := controller.GetMachineDeployment(ctx, client, "test")
	g.Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
}

func TestGetCapiClusterObjectKey(t *testing.T) {
	g := NewWithT(t)
	eksaCluster := eksaCluster()

	expected := types.NamespacedName{
		Name:      "my-cluster",
		Namespace: "eksa-system",
	}

	key := controller.CapiClusterObjectKey(eksaCluster)
	g.Expect(key).To(Equal(expected))
}

func eksaCluster() *anywherev1.Cluster {
	return &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
	}
}

func capiCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}
}

func kubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	return &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: controlplanev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}
}

func machineDeployment() *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineDeployment",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}
}
