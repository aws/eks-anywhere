package clusters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	dockerv1beta2 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	"github.com/aws/eks-anywhere/pkg/constants"
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

func TestReconcileControlPlaneUpdateAfterClusterCreation(t *testing.T) {
	g := NewWithT(t)
	c := env.Client()
	api := envtest.NewAPIExpecter(t, c)
	ctx := context.Background()
	ns := env.CreateNamespaceForTest(ctx, t)
	log := test.NewNullLogger()
	cp := controlPlaneStackedEtcd(ns)

	originalCPEndpoint := clusterv1beta2.APIEndpoint{
		Host: "my-server.example.com",
		Port: 6443,
	}
	cp.Cluster.Spec.ControlPlaneEndpoint = originalCPEndpoint

	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	// We never set the endpoint ourselves, so we mimic that here
	// we want to test the original one set by capi is preserved
	cp.Cluster.Spec.ControlPlaneEndpoint = clusterv1beta2.APIEndpoint{}

	g.Expect(clusters.ReconcileControlPlane(ctx, log, c, cp)).To(Equal(controller.Result{}))
	api.ShouldEventuallyExist(ctx, cp.Cluster)
	api.ShouldEventuallyMatch(ctx, cp.Cluster, func(g Gomega) {
		g.Expect(cp.Cluster.Spec.ControlPlaneEndpoint).To(Equal(originalCPEndpoint))
	})
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

func TestReconcileControlPlanePartialCreate(t *testing.T) {
	tests := []struct {
		name     string
		external bool
		prefix   int
	}{
		{name: "stacked before provider", prefix: 1},
		{name: "external before KCP", external: true, prefix: 2},
		{name: "external before etcd", external: true, prefix: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			c := env.Client()
			ns := env.CreateNamespaceForTest(ctx, t)
			cp := controlPlaneStackedEtcd(ns)
			if tt.external {
				cp = controlPlaneExternalEtcd(ns)
			}
			endpoint := clusterv1beta2.APIEndpoint{Host: "control.example.test", Port: 6443}
			cp.Cluster.Spec.ControlPlaneEndpoint = endpoint
			objects := cp.AllObjects()
			envtest.CreateObjs(ctx, t, c, objects[:tt.prefix]...)
			g.Eventually(func() error {
				return c.Get(ctx, client.ObjectKeyFromObject(cp.Cluster), &clusterv1beta2.Cluster{})
			}).Should(Succeed())
			cp.Cluster.Spec.ControlPlaneEndpoint = clusterv1beta2.APIEndpoint{}

			result, err := clusters.ReconcileControlPlane(ctx, test.NewNullLogger(), c, cp)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(result).To(Equal(controller.Result{}))
			api := envtest.NewAPIExpecter(t, c)
			for _, object := range cp.AllObjects() {
				api.ShouldEventuallyExist(ctx, object)
			}
			g.Expect(cp.Cluster.Spec.ControlPlaneEndpoint).To(Equal(endpoint))
		})
	}
}

func TestControlPlaneInitializedChildMissing(t *testing.T) {
	for _, missing := range []string{"KubeadmControlPlane", "EtcdadmCluster"} {
		t.Run(missing, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			c := env.Client()
			cp := controlPlaneExternalEtcd(env.CreateNamespaceForTest(ctx, t))
			envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)
			current := &clusterv1beta2.Cluster{}
			g.Eventually(func() error {
				return c.Get(ctx, client.ObjectKeyFromObject(cp.Cluster), current)
			}).Should(Succeed())
			current.Status.Initialization.ControlPlaneInitialized = ptr.Bool(true)
			g.Expect(c.Status().Update(ctx, current)).To(Succeed())
			g.Eventually(func() bool {
				got := &clusterv1beta2.Cluster{}
				if err := c.Get(ctx, client.ObjectKeyFromObject(cp.Cluster), got); err != nil {
					return false
				}
				return got.Status.Initialization.ControlPlaneInitialized != nil && *got.Status.Initialization.ControlPlaneInitialized
			}).Should(BeTrue())
			reader := &controlPlaneReadErrorClient{
				Client: c,
				kind:   missing,
				err:    apierrors.NewNotFound(schema.GroupResource{Resource: missing}, cp.Cluster.Name),
			}
			_, err := clusters.ReconcileControlPlane(ctx, test.NewNullLogger(), reader, cp)
			g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
			g.Expect(reader.patches).To(BeZero())
		})
	}
}

func TestControlPlaneChildReadError(t *testing.T) {
	for _, kind := range []string{"KubeadmControlPlane", "EtcdadmCluster"} {
		for _, failure := range []struct {
			name string
			err  error
		}{
			{name: "forbidden", err: apierrors.NewForbidden(schema.GroupResource{Resource: kind}, "test", errors.New("denied"))},
			{name: "transport", err: errors.New("transport unavailable")},
		} {
			t.Run(kind+" "+failure.name, func(t *testing.T) {
				g := NewWithT(t)
				ctx := context.Background()
				c := env.Client()
				cp := controlPlaneExternalEtcd(env.CreateNamespaceForTest(ctx, t))
				envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)
				g.Eventually(func() error {
					return c.Get(ctx, client.ObjectKeyFromObject(cp.Cluster), &clusterv1beta2.Cluster{})
				}).Should(Succeed())
				reader := &controlPlaneReadErrorClient{
					Client: c,
					kind:   kind,
					err:    failure.err,
				}
				_, err := clusters.ReconcileControlPlane(ctx, test.NewNullLogger(), reader, cp)
				g.Expect(errors.Is(err, failure.err)).To(BeTrue())
				g.Expect(reader.patches).To(BeZero())
			})
		}
	}
}

type controlPlaneReadErrorClient struct {
	client.Client
	kind    string
	err     error
	patches int
}

func (c *controlPlaneReadErrorClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if c.kind == "KubeadmControlPlane" {
		if _, ok := obj.(*controlplanev1beta2.KubeadmControlPlane); ok {
			return c.err
		}
	}
	if c.kind == "EtcdadmCluster" {
		if _, ok := obj.(*etcdv1.EtcdadmCluster); ok {
			return c.err
		}
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c *controlPlaneReadErrorClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	c.patches++
	return c.Client.Patch(ctx, obj, patch, opts...)
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
				HaveKeyWithValue(clusterv1beta2.PausedAnnotation, "true"),
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
	clientutil.AddAnnotation(cp.KubeadmControlPlane, clusterv1beta2.PausedAnnotation, "true")
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

func TestReconcileControlPlaneExternalEtcdWithPlaceholderEndpoints(t *testing.T) {
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
	clientutil.AddAnnotation(cp.KubeadmControlPlane, clusterv1beta2.PausedAnnotation, "true")
	// Set placeholder endpoint in existing kcp
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = []string{constants.PlaceholderExternalEtcdEndpoint}
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(newCPReplicas)
	// Desired kcp has placeholder endpoints (matching actual provider behavior)
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = []string{constants.PlaceholderExternalEtcdEndpoint}

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
			// When current has placeholder and desired has placeholder, placeholder should stay
			// (no real endpoints to preserve, placeholder is not overwritten)
			g.Expect(kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints).To(
				HaveExactElements(constants.PlaceholderExternalEtcdEndpoint),
				"Placeholder etcd endpoints should be preserved when no real endpoints exist",
			)
		},
	)
}

func TestReconcileControlPlaneExternalEtcdWithMultipleEndpoints(t *testing.T) {
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
	clientutil.AddAnnotation(cp.KubeadmControlPlane, clusterv1beta2.PausedAnnotation, "true")
	// Set multiple real etcd endpoints in existing kcp
	multipleEndpoints := []string{"https://1.1.1.1:2379", "https://2.2.2.2:2379", "https://3.3.3.3:2379"}
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = multipleEndpoints
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(newCPReplicas)
	// Desired kcp has empty endpoints
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
			// Multiple real endpoints should be preserved
			g.Expect(kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints).To(
				HaveExactElements(multipleEndpoints),
				"Multiple real etcd endpoints should be preserved",
			)
		},
	)
}

func TestReconcileControlPlaneExternalEtcdWithEmptyEndpoints(t *testing.T) {
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
	clientutil.AddAnnotation(cp.KubeadmControlPlane, clusterv1beta2.PausedAnnotation, "true")
	// Set placeholder endpoints in existing kcp (v1beta2 requires non-empty endpoints)
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = []string{constants.PlaceholderExternalEtcdEndpoint}
	envtest.CreateObjs(ctx, t, c, cp.AllObjects()...)

	cp.KubeadmControlPlane.Spec.Replicas = ptr.Int32(newCPReplicas)
	// Desired kcp also has placeholder endpoints
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints = []string{constants.PlaceholderExternalEtcdEndpoint}

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
			// Placeholder endpoints should be preserved when desired also has placeholder
			g.Expect(kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External.Endpoints).To(
				HaveExactElements(constants.PlaceholderExternalEtcdEndpoint),
				"Placeholder etcd endpoints should be preserved when desired also has placeholder",
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
		Cluster: &clusterv1beta2.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "cluster.x-k8s.io/v1beta2",
				Kind:       "Cluster",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: clusterv1beta2.ClusterSpec{
				ControlPlaneRef: clusterv1beta2.ContractVersionedObjectReference{
					APIGroup: "controlplane.cluster.x-k8s.io",
					Kind:     "KubeadmControlPlane",
					Name:     clusterName,
				},
				InfrastructureRef: clusterv1beta2.ContractVersionedObjectReference{
					APIGroup: "infrastructure.cluster.x-k8s.io",
					Kind:     "DockerCluster",
					Name:     clusterName,
				},
			},
		},
		KubeadmControlPlane: &controlplanev1beta2.KubeadmControlPlane{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta2",
				Kind:       "KubeadmControlPlane",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: controlplanev1beta2.KubeadmControlPlaneSpec{
				Version: "v1.28.0",
				MachineTemplate: controlplanev1beta2.KubeadmControlPlaneMachineTemplate{
					Spec: controlplanev1beta2.KubeadmControlPlaneMachineTemplateSpec{
						InfrastructureRef: clusterv1beta2.ContractVersionedObjectReference{
							APIGroup: "infrastructure.cluster.x-k8s.io",
							Kind:     "DockerMachineTemplate",
							Name:     clusterName + "-cp",
						},
					},
				},
				KubeadmConfigSpec: bootstrapv1beta2.KubeadmConfigSpec{
					ClusterConfiguration: bootstrapv1beta2.ClusterConfiguration{},
				},
			},
		},
		ProviderCluster: &dockerv1beta2.DockerCluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
				Kind:       "DockerCluster",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: namespace,
			},
			Spec: dockerv1beta2.DockerClusterSpec{
				LoadBalancer: dockerv1beta2.DockerLoadBalancer{
					ImageMeta: dockerv1beta2.ImageMeta{
						ImageRepository: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind",
						ImageTag:        "v0.11.1-eks-a-v0.0.0-dev-build.1464",
					},
				},
			},
			Status: dockerv1beta2.DockerClusterStatus{
				Initialization: dockerv1beta2.DockerClusterInitializationStatus{
					Provisioned: &[]bool{true}[0],
				},
			},
		},
		ControlPlaneMachineTemplate: &dockerv1beta2.DockerMachineTemplate{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
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
	cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.External = bootstrapv1beta2.ExternalEtcd{
		// Endpoints of etcd members. Required for ExternalEtcd.
		Endpoints: []string{constants.PlaceholderExternalEtcdEndpoint},
		CAFile:    "/etc/kubernetes/pki/etcd/ca.crt",
		CertFile:  "/etc/kubernetes/pki/apiserver-etcd-client.crt",
		KeyFile:   "/etc/kubernetes/pki/apiserver-etcd-client.key",
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
	cp.Cluster.Spec.ManagedExternalEtcdRef = &clusterv1beta2.ContractVersionedObjectReference{
		APIGroup: "etcdcluster.cluster.x-k8s.io",
		Kind:     "EtcdadmCluster",
		Name:     cp.EtcdCluster.Name,
	}

	cp.EtcdMachineTemplate = &dockerv1beta2.DockerMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
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
