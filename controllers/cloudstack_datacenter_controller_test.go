package controllers_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/controllers"
)

func TestCloudStackDatacenterReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewCloudStackDatacenterReconciler(client)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}
