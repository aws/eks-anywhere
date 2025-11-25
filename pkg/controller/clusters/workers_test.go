package clusters_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestReconcileWorkersSuccess(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	w := workers(ns)
	cluster := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: ns,
		},
	}

	existingMachineDeployment1 := machineDeployment("my-cluster-md-1", ns)
	existingMachineDeployment2 := machineDeployment("my-cluster-md-2", ns)
	existingMachineDeployment3 := machineDeployment("my-other-cluster-md-1", ns)
	existingMachineDeployment3.Labels[clusterv1.ClusterNameLabel] = "my-other-cluster"
	envtest.CreateObjs(ctx, t, c,
		existingMachineDeployment1,
		existingMachineDeployment2,
		existingMachineDeployment3,
	)

	g.Expect(clusters.ReconcileWorkers(ctx, c, cluster, w)).To(Equal(controller.Result{}))

	api.ShouldEventuallyExist(ctx, w.Groups[0].MachineDeployment)
	api.ShouldEventuallyExist(ctx, w.Groups[0].KubeadmConfigTemplate)
	api.ShouldEventuallyExist(ctx, w.Groups[0].ProviderMachineTemplate)
	api.ShouldEventuallyExist(ctx, w.Groups[1].MachineDeployment)
	api.ShouldEventuallyExist(ctx, w.Groups[1].KubeadmConfigTemplate)
	api.ShouldEventuallyExist(ctx, w.Groups[1].ProviderMachineTemplate)

	api.ShouldEventuallyNotExist(ctx, existingMachineDeployment1)
	api.ShouldEventuallyNotExist(ctx, existingMachineDeployment2)
	api.ShouldEventuallyExist(ctx, existingMachineDeployment3)
}

func TestReconcileWorkersErrorApplyingObjects(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	ctx := context.Background()
	ns := "fake-ns"
	// ns doesn't exist, it will fail
	w := workers(ns)
	cluster := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: ns,
		},
	}

	g.Expect(clusters.ReconcileWorkers(ctx, c, cluster, w)).Error().To(
		MatchError(ContainSubstring("applying worker nodes CAPI objects")),
	)
}

func TestToWorkers(t *testing.T) {
	g := NewWithT(t)
	namespace := constants.EksaSystemNamespace
	w := &clusterapi.Workers[*dockerv1.DockerMachineTemplate]{
		Groups: []clusterapi.WorkerGroup[*dockerv1.DockerMachineTemplate]{
			{
				MachineDeployment:       machineDeployment("my-cluster-md-0", namespace),
				KubeadmConfigTemplate:   kubeadmConfigTemplate("my-cluster-md-0-1", namespace),
				ProviderMachineTemplate: dockerMachineTemplate("my-cluster-md-0-1", namespace),
			},
			{
				MachineDeployment:       machineDeployment("my-cluster-md-3", namespace),
				KubeadmConfigTemplate:   kubeadmConfigTemplate("my-cluster-md-3-1", namespace),
				ProviderMachineTemplate: dockerMachineTemplate("my-cluster-md-3-1", namespace),
			},
		},
	}

	want := &clusters.Workers{
		Groups: []clusters.WorkerGroup{
			{
				MachineDeployment:       w.Groups[0].MachineDeployment,
				KubeadmConfigTemplate:   w.Groups[0].KubeadmConfigTemplate,
				ProviderMachineTemplate: w.Groups[0].ProviderMachineTemplate,
			},
			{
				MachineDeployment:       w.Groups[1].MachineDeployment,
				KubeadmConfigTemplate:   w.Groups[1].KubeadmConfigTemplate,
				ProviderMachineTemplate: w.Groups[1].ProviderMachineTemplate,
			},
		},
	}

	g.Expect(clusters.ToWorkers(w)).To(Equal(want))
}

func workers(namespace string) *clusters.Workers {
	return &clusters.Workers{
		Groups: []clusters.WorkerGroup{
			{
				MachineDeployment:       machineDeployment("my-cluster-md-0", namespace),
				KubeadmConfigTemplate:   kubeadmConfigTemplate("my-cluster-md-0-1", namespace),
				ProviderMachineTemplate: dockerMachineTemplate("my-cluster-md-0-1", namespace),
			},
			{
				MachineDeployment:       machineDeployment("my-cluster-md-3", namespace),
				KubeadmConfigTemplate:   kubeadmConfigTemplate("my-cluster-md-3-1", namespace),
				ProviderMachineTemplate: dockerMachineTemplate("my-cluster-md-3-1", namespace),
			},
		},
	}
}

func machineDeployment(name, namespace string) *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "MachineDeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: "my-cluster",
			},
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: "my-cluster",
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					ClusterName: "my-cluster",
				},
			},
		},
	}
}

func kubeadmConfigTemplate(name, namespace string) *bootstrapv1.KubeadmConfigTemplate {
	return &bootstrapv1.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
			Kind:       "KubeadmConfigTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func dockerMachineTemplate(name, namespace string) *dockerv1.DockerMachineTemplate {
	return &dockerv1.DockerMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			Kind:       "DockerMachineTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func TestReconcileWorkersForEKSAErrorGettingCAPICluster(t *testing.T) {
	g := NewWithT(t)
	c := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()
	ctx := context.Background()
	ns := "ns"
	w := workers(ns)
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: ns,
		},
	}

	g.Expect(
		clusters.ReconcileWorkersForEKSA(ctx, env.Manager().GetLogger(), c, cluster, w),
	).Error().To(MatchError(ContainSubstring("reconciling workers for EKS-A cluster")))
}

func TestReconcileWorkersForEKSANoCAPICluster(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	ctx := context.Background()
	ns := "ns"
	w := workers(ns)
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: ns,
		},
	}

	g.Expect(
		clusters.ReconcileWorkersForEKSA(ctx, env.Manager().GetLogger(), c, cluster, w),
	).To(Equal(controller.Result{Result: &reconcile.Result{RequeueAfter: 5 * time.Second}}))
}

func TestReconcileWorkersForEKSASuccess(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	w := workers(ns)
	capiCluster := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: constants.EksaSystemNamespace,
		},
	}
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: ns,
		},
	}

	envtest.CreateObjs(ctx, t, c,
		test.Namespace(constants.EksaSystemNamespace),
		capiCluster,
	)

	g.Expect(
		clusters.ReconcileWorkersForEKSA(ctx, env.Manager().GetLogger(), c, cluster, w),
	).To(Equal(controller.Result{}))

	api.ShouldEventuallyExist(ctx, w.Groups[0].MachineDeployment)
	api.ShouldEventuallyExist(ctx, w.Groups[0].KubeadmConfigTemplate)
	api.ShouldEventuallyExist(ctx, w.Groups[0].ProviderMachineTemplate)
	api.ShouldEventuallyExist(ctx, w.Groups[1].MachineDeployment)
	api.ShouldEventuallyExist(ctx, w.Groups[1].KubeadmConfigTemplate)
	api.ShouldEventuallyExist(ctx, w.Groups[1].ProviderMachineTemplate)

	api.DeleteAndWait(ctx, capiCluster)
}
