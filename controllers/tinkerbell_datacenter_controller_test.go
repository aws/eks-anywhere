package controllers_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/controllers"
)

func TestTinkerbellDatacenterReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewTinkerbellDatacenterReconciler(client)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}
