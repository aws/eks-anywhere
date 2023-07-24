package clusters_test

import (
	"context"
	"testing"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test/envtest"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestReconcileControlPlaneStackedEtcd(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	cp := controlPlaneStackedEtcd(ns)

	g.Expect(clusters.ReconcileControlPlane(ctx, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
}

func TestReconcileControlPlaneExternalEtcdNewCluster(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	cp := controlPlaneExternalEtcd(ns)

	g.Expect(clusters.ReconcileControlPlane(ctx, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)
}

func TestReconcileControlPlaneExternalEtcdUpgradeWithDiff(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	cp := controlPlaneExternalEtcd(ns)
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	cp.EtcdCluster.Spec.Replicas = ptr.Int32(5)

	g.Expect(clusters.ReconcileControlPlane(ctx, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)

	etcdadmCluster := &etcdv1.EtcdadmCluster{ObjectMeta: cp.EtcdCluster.ObjectMeta}
	api.ShouldEventuallyMatch(
		ctx,
		etcdadmCluster,
		func(g Gomega) {
			g.Expect(etcdadmCluster.Spec.Replicas).To(HaveValue(BeEquivalentTo(5)), "etcdadm replicas should have been updated")
			g.Expect(etcdadmCluster.Annotations).To(
				HaveKeyWithValue(etcdv1.UpgradeInProgressAnnotation, "true"),
				"etcdadm upgrading annotation should have been added",
			)
		},
	)
	kcp := &etcdv1.EtcdadmCluster{ObjectMeta: cp.KubeadmControlPlane.ObjectMeta}
	api.ShouldEventuallyMatch(
		ctx,
		etcdadmCluster,
		func(g Gomega) {
			g.Expect(kcp.Annotations).To(
				HaveKeyWithValue(clusterv1.PausedAnnotation, "true"),
				"kcp paused annotation should have been added",
			)
		},
	)
}

func TestReconcileControlPlaneExternalEtcdUpgradeWithNoDiff(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	cp := controlPlaneExternalEtcd(ns)
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	g.Expect(clusters.ReconcileControlPlane(ctx, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)
}

func TestReconcileControlPlaneExternalEtcdUpgradeWithNoNamespace(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	cp := controlPlaneExternalEtcd(ns)
	cp.Cluster.Spec.ManagedExternalEtcdRef.Namespace = ""
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	g.Expect(clusters.ReconcileControlPlane(ctx, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)
}

func TestReconcileControlPlaneExternalEtcdWithExistingEndpoints(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	cp := controlPlaneExternalEtcd(ns)
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = []string{"https://1.1.1.1:2379"}
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = []string{}
	g.Expect(clusters.ReconcileControlPlane(ctx, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)

	kcp := envtest.CloneNameNamespace(cp.KubeadmControlPlane)
	api.ShouldEventuallyMatch(ctx, kcp, func(g Gomega) {
		endpoints := kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints
		g.Expect(endpoints).To(ContainElement("https://1.1.1.1:2379"))
	})
}

func controlPlaneStackedEtcd(namespace string) *clusters.ControlPlane {
	clusterName := "my-cluster"
	return &clusters.ControlPlane{
		Cluster: &clusterv1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "cluster.x-k8s.io/v1beta1",
				Kind:       "Cluster",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: clusterv1.ClusterSpec{
				ControlPlaneRef: &corev1.ObjectReference{
					Name:      clusterName,
					Namespace: namespace,
				},
			},
		},
		KubeadmControlPlane: &controlplanev1.KubeadmControlPlane{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
				Kind:       "KubeadmControlPlane",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: controlplanev1.KubeadmControlPlaneSpec{
				KubeadmConfigSpec: v1beta1.KubeadmConfigSpec{
					ClusterConfiguration: &v1beta1.ClusterConfiguration{},
				},
			},
		},
		ProviderCluster: &dockerv1.DockerCluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				Kind:       "DockerCluster",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
		},
		ControlPlaneMachineTemplate: &dockerv1.DockerMachineTemplate{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				Kind:       "DockerMachineTemplate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      "my-cluster-cp",
			},
		},
	}
}

func controlPlaneExternalEtcd(namespace string) *clusters.ControlPlane {
	cp := controlPlaneStackedEtcd(namespace)
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External = &v1beta1.ExternalEtcd{
		// Endpoints of etcd members. Required for ExternalEtcd.
		Endpoints: []string{},
		CAFile:    "",
		CertFile:  "",
		KeyFile:   "",
	}

	cp.EtcdCluster = &etcdv1.EtcdadmCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
			Kind:       "EtcdadmCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: namespace,
		},
	}
	cp.Cluster.Spec.ManagedExternalEtcdRef = &corev1.ObjectReference{
		Name:      cp.EtcdCluster.Name,
		Namespace: cp.EtcdCluster.Namespace,
	}

	cp.EtcdMachineTemplate = &dockerv1.DockerMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			Kind:       "DockerMachineTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "my-cluster-etcd",
		},
	}

	return cp
}

func TestControlPlaneAllObjects(t *testing.T) {
	stackedCP := controlPlaneStackedEtcd("my-ns")
	withOtherCP := controlPlaneStackedEtcd("my-ns")
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "s",
			Namespace: "eksa-system",
		},
	}
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm",
			Namespace: "eksa-system",
		},
	}
	withOtherCP.Other = append(withOtherCP.Other, secret, cm)
	externalCP := controlPlaneExternalEtcd("my-ns")
	tests := []struct {
		name string
		cp   *clusters.ControlPlane
		want []client.Object
	}{
		{
			name: "stacked etcd",
			cp:   stackedCP,
			want: []client.Object{
				stackedCP.Cluster,
				stackedCP.KubeadmControlPlane,
				stackedCP.ProviderCluster,
				stackedCP.ControlPlaneMachineTemplate,
			},
		},
		{
			name: "external etcd",
			cp:   externalCP,
			want: []client.Object{
				externalCP.Cluster,
				externalCP.KubeadmControlPlane,
				externalCP.ProviderCluster,
				externalCP.ControlPlaneMachineTemplate,
				externalCP.EtcdCluster,
				externalCP.EtcdMachineTemplate,
			},
		},
		{
			name: "stacked etcd with other",
			cp:   withOtherCP,
			want: []client.Object{
				stackedCP.Cluster,
				stackedCP.KubeadmControlPlane,
				stackedCP.ProviderCluster,
				stackedCP.ControlPlaneMachineTemplate,
				secret,
				cm,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cp.AllObjects()).To(ConsistOf(tt.want))
		})
	}
}
