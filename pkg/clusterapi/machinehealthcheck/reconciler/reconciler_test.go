package reconciler_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywhereCluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi/machinehealthcheck/reconciler"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	envClient := env.Client()
	tt := newReconciler(t, "-success")
	envtest.CreateObjs(ctx, t, envClient,
		tt.mhc,
	)
	envtest.CreateObjs(ctx, t, envClient, test.EksdRelease("1-19"), test.EKSARelease(), tt.bundle)
	newReconciler := reconciler.New(envClient, tt.mhcDefaulter)
	tt.cluster.Spec.MachineHealthCheck = nil

	err := newReconciler.Reconcile(ctx, nullLog(), tt.cluster)

	g.Expect(err).ToNot(HaveOccurred())
}

func TestReconcilerReconcileSuccessNotNil(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	envClient := env.Client()
	tt := newReconciler(t, "-notnil")
	envtest.CreateObjs(ctx, t, envClient,
		tt.eksaSupportObjs...,
	)
	newReconciler := reconciler.New(envClient, tt.mhcDefaulter)

	err := newReconciler.Reconcile(ctx, nullLog(), tt.cluster)

	g.Expect(err).ToNot(HaveOccurred())
}

func (tt *reconcilerTest) withFakeClient() {
	tt.client = fake.NewClientBuilder().WithObjects(clientutil.ObjectsToClientObjects(tt.eksaSupportObjs)...).Build()
}

func newReconciler(t testing.TB, suffix string) *reconcilerTest {
	mhcDefaulter := anywhereCluster.NewMachineHealthCheckDefaulter(constants.DefaultNodeStartupTimeout, constants.DefaultUnhealthyMachineTimeout, intstr.Parse(constants.DefaultMaxUnhealthy), intstr.Parse(constants.DefaultWorkerMaxUnhealthy))
	bundle := test.Bundle()
	version := test.DevEksaVersion()
	clusterName := fmt.Sprintf("my-cluster-%s", suffix)
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: "eksa-system",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
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
			Name:      fmt.Sprintf("y-cluster-kcp-unhealthy-%s", suffix),
			Namespace: "eksa-system",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "MachineHealthCheck",
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		Spec: clusterv1.MachineHealthCheckSpec{
			ClusterName: clusterName,
			NodeStartupTimeout: &metav1.Duration{
				Duration: 20 * time.Minute,
			},
			UnhealthyConditions: []clusterv1.UnhealthyCondition{
				{
					Timeout: metav1.Duration{
						Duration: 15 * time.Minute,
					},
					Status: v1.ConditionUnknown,
					Type:   v1.NodeReady,
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
		},
		bundle: bundle,
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
	bundle          *releasev1.Bundles
}
