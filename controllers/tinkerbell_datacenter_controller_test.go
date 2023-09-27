package controllers_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/internal/test"
)

func TestTinkerbellDatacenterReconcilerSetupWithManager(t *testing.T) {
	test.MarkIntegration(t)
	client := env.Client()
	r := controllers.NewTinkerbellDatacenterReconciler(client)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}
