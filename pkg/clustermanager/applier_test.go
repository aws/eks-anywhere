package clustermanager_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

type applierTest struct {
	Gomega
	tb            testing.TB
	clientFactory *mocks.MockClientFactory
	ctx           context.Context
	spec          *cluster.Spec
	client        kubernetes.Client
	log           logr.Logger
	mgmtCluster   types.Cluster
}

func newApplierTest(tb testing.TB) *applierTest {
	ctrl := gomock.NewController(tb)
	return &applierTest{
		tb:            tb,
		Gomega:        NewWithT(tb),
		clientFactory: mocks.NewMockClientFactory(ctrl),
		ctx:           context.Background(),
		spec:          test.VSphereClusterSpec(tb, tb.Name()),
		log:           test.NewNullLogger(),
		mgmtCluster: types.Cluster{
			KubeconfigFile: "my-config",
		},
	}
}

func (a *applierTest) buildClient(objs ...kubernetes.Object) {
	a.client = test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(objs)...)
	a.clientFactory.EXPECT().BuildClientFromKubeconfig(a.mgmtCluster.KubeconfigFile).Return(a.client, nil)
}

func (a *applierTest) updateFailureMessage(c *anywherev1.Cluster, err string) {
	c.Status.FailureMessage = ptr.String(err)
	a.Expect(a.client.Update(a.ctx, c)).To(Succeed())
}

func (a *applierTest) markCPReady(c *anywherev1.Cluster) {
	conditions.MarkTrue(c, anywherev1.ControlPlaneReadyCondition)
	a.Expect(a.client.Update(a.ctx, c)).To(Succeed())
}

func (a *applierTest) markCNIConfigured(c *anywherev1.Cluster) {
	conditions.MarkTrue(c, anywherev1.DefaultCNIConfiguredCondition)
	a.Expect(a.client.Update(a.ctx, c)).To(Succeed())
}

func (a *applierTest) markWorkersReady(c *anywherev1.Cluster) {
	conditions.MarkTrue(c, anywherev1.WorkersReadyCondition)
	a.Expect(a.client.Update(a.ctx, c)).To(Succeed())
}

func (a *applierTest) markClusterReady(c *anywherev1.Cluster) {
	conditions.MarkTrue(c, anywherev1.ReadyCondition)
	a.Expect(a.client.Update(a.ctx, c)).To(Succeed())
}

const updateMarkerAnnotation = "anywhere.aws.amazon.com/cluster-updated"

func (a *applierTest) startFakeController() {
	a.tb.Helper()
	if a.client == nil {
		a.tb.Fatal("Client needs to be initialized before starting the fake controller")
	}
	// Before starting the controller, we add an annotation to the cluster so we can
	// check from the controller when it gets updated
	clientutil.AddAnnotation(a.spec.Cluster, updateMarkerAnnotation, "true")

	go func() {
		c := &anywherev1.Cluster{}
		for {
			err := a.client.Get(a.ctx, a.spec.Cluster.Name, a.spec.Cluster.Namespace, c)
			if apierrors.IsNotFound(err) {
				continue
			}
			a.Expect(err).NotTo(HaveOccurred())

			// If the annotation is not present, keep trying
			if _, ok := c.Annotations[updateMarkerAnnotation]; !ok {
				continue
			}

			a.markCPReady(c)
			// We wait after each condition update to "simulate" multiple reconcile loops
			time.Sleep(5 * time.Millisecond)
			a.markCNIConfigured(c)
			time.Sleep(5 * time.Millisecond)
			a.markWorkersReady(c)
			time.Sleep(5 * time.Millisecond)
			a.markClusterReady(c)

			break
		}
	}()
}

func TestApplierRunClusterCreateSuccess(t *testing.T) {
	tt := newApplierTest(t)
	tt.buildClient()
	tt.startFakeController()
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierRetryBackOff(time.Millisecond),
		clustermanager.WithApplierNoTimeouts(),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(Succeed())
}

func TestApplierRunClusterUpdateSuccess(t *testing.T) {
	tt := newApplierTest(t)
	tt.buildClient(tt.spec.ClusterAndChildren()...)
	tt.startFakeController()
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierRetryBackOff(time.Millisecond),
		clustermanager.WithApplierNoTimeouts(),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(Succeed())
}

func TestApplierRunCClusterUpdatedWithCNINotManaged(t *testing.T) {
	tt := newApplierTest(t)
	tt.spec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(true)
	tt.buildClient(tt.spec.ClusterAndChildren()...)
	tt.markCPReady(tt.spec.Cluster)
	tt.markWorkersReady(tt.spec.Cluster)
	tt.markClusterReady(tt.spec.Cluster)
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierWaitForClusterReconcile(0),
		clustermanager.WithApplierWaitForFailureMessage(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(Succeed())
}

func TestApplierRunErrorBuildingClient(t *testing.T) {
	tt := newApplierTest(t)
	tt.client = test.NewFakeKubeClientAlwaysError()
	tt.clientFactory.EXPECT().BuildClientFromKubeconfig(tt.mgmtCluster.KubeconfigFile).Return(nil, errors.New("bad client"))
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierApplyClusterTimeout(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("building client to apply cluster spec changes")))
}

func TestApplierRunErrorApplying(t *testing.T) {
	tt := newApplierTest(t)
	tt.client = test.NewFakeKubeClientAlwaysError()
	tt.clientFactory.EXPECT().BuildClientFromKubeconfig(tt.mgmtCluster.KubeconfigFile).Return(tt.client, nil)
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierApplyClusterTimeout(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("applying cluster spec")))
}

func TestApplierRunFailureMessage(t *testing.T) {
	tt := newApplierTest(t)
	tt.buildClient(tt.spec.ClusterAndChildren()...)
	tt.updateFailureMessage(tt.spec.Cluster, "error")
	tt.startFakeController()
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierRetryBackOff(0),
		clustermanager.WithApplierWaitForFailureMessage(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("cluster has a validation error that doesn't seem transient")))
}

func TestApplierRunControlPlaneNotReady(t *testing.T) {
	tt := newApplierTest(t)
	tt.buildClient()
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierWaitForClusterReconcile(0),
		clustermanager.WithApplierWaitForFailureMessage(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("waiting for cluster's control plane to be ready")))
}

func TestApplierRunCNINotConfigured(t *testing.T) {
	tt := newApplierTest(t)
	tt.buildClient(tt.spec.ClusterAndChildren()...)
	tt.markCPReady(tt.spec.Cluster)
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierWaitForClusterReconcile(0),
		clustermanager.WithApplierWaitForFailureMessage(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("waiting for cluster's CNI to be configured")))
}

func TestApplierRunWorkersNotReady(t *testing.T) {
	tt := newApplierTest(t)
	tt.buildClient(tt.spec.ClusterAndChildren()...)
	tt.markCPReady(tt.spec.Cluster)
	tt.markCNIConfigured(tt.spec.Cluster)
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierWaitForClusterReconcile(0),
		clustermanager.WithApplierWaitForFailureMessage(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("waiting for cluster's workers to be ready")))
}

func TestApplierRunClusterNotReady(t *testing.T) {
	tt := newApplierTest(t)
	tt.buildClient(tt.spec.ClusterAndChildren()...)
	tt.markCPReady(tt.spec.Cluster)
	tt.markCNIConfigured(tt.spec.Cluster)
	tt.markWorkersReady(tt.spec.Cluster)
	a := clustermanager.NewApplier(tt.log, tt.clientFactory,
		clustermanager.WithApplierWaitForClusterReconcile(0),
		clustermanager.WithApplierWaitForFailureMessage(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("waiting for cluster to be ready")))
}
