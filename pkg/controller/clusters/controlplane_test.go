package clusters_test

import (
	"context"
	"testing"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestReconcileControlPlaneStackedEtcd(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	log := test.NewNullLogger()
	cp := controlPlaneStackedEtcd(ns)

	g.Expect(clusters.ReconcileControlPlane(ctx, log, c, cp)).To(Equal(controller.Result{}))
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
	log := test.NewNullLogger()
	cp := controlPlaneExternalEtcd(ns)

	g.Expect(clusters.ReconcileControlPlane(ctx, log, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)

	kcp := envtest.CloneNameNamespace(cp.KubeadmControlPlane)
	api.ShouldEventuallyMatch(
		ctx,
		kcp,
		func(g Gomega) {
			g.Expect(kcp.Annotations).To(
				HaveKey("cluster.x-k8s.io/skip-pause-cp-managed-etcd"),
				"kcp should have skip pause annotation after being created",
			)
		},
	)
}

func TestReconcileControlPlaneExternalEtcdUpgradeWithDiff(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	log := test.NewNullLogger()
	cp := controlPlaneExternalEtcd(ns)

	var oldCPReplicas int32 = 3
	var newCPReplicas int32 = 4
	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(oldCPReplicas)

	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	cp.EtcdCluster.Spec.Replicas = ptr.Int32(5)
	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(newCPReplicas)

	g.Expect(clusters.ReconcileControlPlane(ctx, log, c, cp)).To(
		Equal(controller.Result{Result: &reconcile.Result{RequeueAfter: 10 * time.Second}}),
	)
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)

	etcdadmCluster := envtest.CloneNameNamespace(cp.EtcdCluster)
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
	kcp := envtest.CloneNameNamespace(cp.KubeadmControlPlane)
	api.ShouldEventuallyMatch(
		ctx,
		kcp,
		func(g Gomega) {
			g.Expect(kcp.Annotations).To(
				HaveKeyWithValue(clusterv1.PausedAnnotation, "true"),
				"kcp paused annotation should have been added",
			)
			g.Expect(kcp.Spec.Replicas).To(
				HaveValue(Equal(oldCPReplicas)),
				"kcp replicas should not have changed",
			)
		},
	)
}

func TestReconcileControlPlaneExternalEtcdNotReady(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	log := test.NewNullLogger()
	cp := controlPlaneExternalEtcd(ns)
	var oldCPReplicas int32 = 3
	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(oldCPReplicas)
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(4)

	g.Expect(clusters.ReconcileControlPlane(ctx, log, c, cp)).To(
		Equal(controller.Result{Result: &reconcile.Result{RequeueAfter: 30 * time.Second}}),
	)
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)

	kcp := envtest.CloneNameNamespace(cp.KubeadmControlPlane)
	api.ShouldEventuallyMatch(
		ctx,
		kcp,
		func(g Gomega) {
			g.Expect(kcp.Spec.Replicas).To(
				HaveValue(Equal(oldCPReplicas)),
				"kcp replicas should not have changed",
			)
		},
	)
}

func TestReconcileControlPlaneExternalEtcdReadyControlPlaneUpgrade(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	log := test.NewNullLogger()
	cp := controlPlaneExternalEtcd(ns)
	cp.EtcdCluster.Status.Ready = true
	cp.EtcdCluster.Status.ObservedGeneration = 1

	var oldCPReplicas int32 = 3
	var newCPReplicas int32 = 4
	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(oldCPReplicas)
	clientutil.AddAnnotation(cp.KubeadmControlPlane, clusterv1.PausedAnnotation, "true")
	// an existing kcp should already have etcd endpoints
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = []string{"https://1.1.1.1:2379"}
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(newCPReplicas)
	// by default providers code will generate kcp with empty endpoints, so we imitate that here
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = []string{}

	g.Expect(clusters.ReconcileControlPlane(ctx, log, c, cp)).To(
		Equal(controller.Result{}),
	)
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)

	kcp := envtest.CloneNameNamespace(cp.KubeadmControlPlane)
	api.ShouldEventuallyMatch(
		ctx,
		kcp,
		func(g Gomega) {
			g.Expect(kcp.Annotations).To(
				HaveKey("cluster.x-k8s.io/skip-pause-cp-managed-etcd"),
				"kcp should have skip pause annotation",
			)
			g.Expect(annotations.HasPaused(kcp)).To(
				BeFalse(), "kcp should not be paused",
			)
			g.Expect(kcp.Spec.Replicas).To(
				HaveValue(Equal(newCPReplicas)),
				"kcp replicas should have been updated",
			)
			g.Expect(kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints).To(
				HaveExactElements("https://1.1.1.1:2379"),
				"Etcd endpoints should remain the same and not be emptied",
			)
		},
	)
}

func TestReconcileControlPlaneExternalEtcdUpgradeWithNoNamespace(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	log := test.NewNullLogger()
	cp := controlPlaneExternalEtcd(ns)
	cp.Cluster.Spec.ManagedExternalEtcdRef.Namespace = ""
	cp.EtcdCluster.Status.Ready = true
	cp.EtcdCluster.Status.ObservedGeneration = 1
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	g.Expect(clusters.ReconcileControlPlane(ctx, log, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyExist(ctx, cp.KubeadmControlPlane)
	api.ShouldEventuallyExist(ctx, cp.ControlPlaneMachineTemplate)
	api.ShouldEventuallyExist(ctx, cp.ProviderCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdCluster)
	api.ShouldEventuallyExist(ctx, cp.EtcdMachineTemplate)
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
