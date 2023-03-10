package controllers_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/nutanix"
)

func TestNutanixDatacenterConfigReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewNutanixDatacenterReconciler(client, nutanix.NewDefaulter())

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestNutanixDatacenterConfigReconcilerSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	config := nutanixDatacenterConfig()
	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewNutanixDatacenterReconciler(cl, nutanix.NewDefaulter())

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "nutanix-datacenter-config",
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	ndc := &anywherev1.NutanixDatacenterConfig{}
	err = cl.Get(ctx, req.NamespacedName, ndc)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestNutanixDatacenterConfigReconcileDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	config := nutanixDatacenterConfig()
	now := metav1.Now()
	config.DeletionTimestamp = &now
	objs := []runtime.Object{config}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewNutanixDatacenterReconciler(cl, nutanix.NewDefaulter())

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "nutanix-datacenter-config",
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestNutanixDatacenterConfigReconcilerFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects().Build()

	r := controllers.NewNutanixDatacenterReconciler(cl, nutanix.NewDefaulter())
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "nutanix-datacenter-config",
		},
	}

	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
}

func nutanixDatacenterConfig() *anywherev1.NutanixDatacenterConfig {
	return &anywherev1.NutanixDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nutanix-datacenter-config",
		},
		Spec: anywherev1.NutanixDatacenterConfigSpec{
			Endpoint: "prism.nutanix.com",
			Port:     9440,
		},
	}
}
