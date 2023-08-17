package reconciler_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywhereCluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi/machinehealthcheck/reconciler"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	tt := newReconciler(t)
	tt.withFakeClient()
	newReconciler := reconciler.New(tt.client, tt.mhcDefaulter)
	tt.cluster.Spec.MachineHealthCheck = nil

	err := newReconciler.Reconcile(ctx, nullLog(), tt.cluster)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestReconcilerReconcileSuccessNotNil(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	tt := newReconciler(t)
	tt.withFakeClient()
	newReconciler := reconciler.New(tt.client, tt.mhcDefaulter)

	err := newReconciler.Reconcile(ctx, nullLog(), tt.cluster)

	g.Expect(err).ToNot(HaveOccurred())
}

func (tt *reconcilerTest) withFakeClient() {
	tt.client = fake.NewClientBuilder().WithObjects(clientutil.ObjectsToClientObjects(tt.eksaSupportObjs)...).Build()
}

func newReconciler(t testing.TB) *reconcilerTest {
	mhcDefaulter := anywhereCluster.NewMachineHealthCheckDefaulter(constants.DefaultNodeStartupTimeout, constants.DefaultUnhealthyMachineTimeout)
	bundle := test.Bundle()
	version := test.DevEksaVersion()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ClusterSpec{
			MachineHealthCheck: &anywherev1.MachineHealthCheck{
				NodeStartupTimeout: &metav1.Duration{
					Duration: 20 * time.Minute,
				},
				UnhealthyMachineTimeout: &metav1.Duration{
					Duration: 20 * time.Minute,
				},
			},
			BundlesRef: &anywherev1.BundlesRef{
				Name:      bundle.Name,
				Namespace: bundle.Namespace,
			},
			KubernetesVersion: anywherev1.Kube119,
		},
	}

	mhc := &clusterv1.MachineHealthCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kcp-unhealthy",
			Namespace: "eksa-system",
		},
		Spec: clusterv1.MachineHealthCheckSpec{
			NodeStartupTimeout: &metav1.Duration{
				Duration: 20 * time.Minute,
			},
			UnhealthyConditions: []clusterv1.UnhealthyCondition{
				{
					Timeout: metav1.Duration{
						Duration: 15 * time.Minute,
					},
				},
			},
		},
	}

	cluster.Spec.BundlesRef = &anywherev1.BundlesRef{
		Name:       bundle.Name,
		Namespace:  bundle.Namespace,
		APIVersion: bundle.APIVersion,
	}
	cluster.Spec.EksaVersion = &version

	tt := &reconcilerTest{
		WithT:        NewWithT(t),
		t:            t,
		mhcDefaulter: mhcDefaulter,
		cluster:      cluster,
		ctx:          context.Background(),
		mhc:          mhc,
		eksaSupportObjs: []client.Object{
			mhc,
			bundle,
			test.EksdRelease("1-19"),
			test.EKSARelease(),
		},
	}

	return tt
}

func nullLog() logr.Logger {
	return logr.New(logf.NullLogSink{})
}

type reconcilerTest struct {
	t testing.TB
	*WithT
	ctx             context.Context
	client          client.Client
	mhcDefaulter    anywhereCluster.MachineHealthCheckDefaulter
	mhc             *clusterv1.MachineHealthCheck
	eksaSupportObjs []client.Object
	cluster         *anywherev1.Cluster
}
