package controllers_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
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
	client := env.Client()

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
