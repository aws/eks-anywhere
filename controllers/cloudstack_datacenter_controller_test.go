package controllers_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCloudStackDatacenterReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewCloudStackDatacenterReconciler(client)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestCloudStackDatacenterReconcilerSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	objs := []runtime.Object{dcConfig}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client)

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestCloudStackDatacenterReconcilerSetDefaultSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	dcConfig.Spec.AvailabilityZones = nil
	dcConfig.Spec.Zones = []anywherev1.CloudStackZone{
		{
			Id:      "",
			Name:    "",
			Network: anywherev1.CloudStackResourceIdentifier{},
		},
	}
	objs := []runtime.Object{dcConfig}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	getDcConfig := &anywherev1.CloudStackDatacenterConfig{}
	err = client.Get(ctx, req.NamespacedName, getDcConfig)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(getDcConfig.Spec.AvailabilityZones)).ToNot(Equal(0))
	g.Expect(getDcConfig.Spec.AvailabilityZones[0].Name).To(Equal(anywherev1.DefaultCloudStackAZPrefix + "-0"))
}

func TestCloudstackDatacenterConfigReconcilerDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	dcConfig := createCloudstackDatacenterConfig()
	dcConfig.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	objs := []runtime.Object{dcConfig}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	r := controllers.NewCloudStackDatacenterReconciler(client)

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
}

func createCloudstackDatacenterConfig() *anywherev1.CloudStackDatacenterConfig {
	return &anywherev1.CloudStackDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.CloudStackDatacenterKind,
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: anywherev1.CloudStackDatacenterConfigSpec{
			AvailabilityZones: []anywherev1.CloudStackAvailabilityZone{
				{
					Name:           "testAz",
					CredentialsRef: "testCred",
					Zone: anywherev1.CloudStackZone{
						Name: "zone1",
						Network: anywherev1.CloudStackResourceIdentifier{
							Name: "SharedNet1",
						},
					},
					Domain:                "testDomain",
					Account:               "testAccount",
					ManagementApiEndpoint: "testApiEndpoint",
				},
			},
		},
	}
}
