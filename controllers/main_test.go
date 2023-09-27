package controllers_test

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
)

var env *envtest.Environment

func TestMain(m *testing.M) {
	if os.Getenv(test.IntegrationTestEnvVar) == "true" {
		os.Exit(envtest.RunWithEnvironment(m, envtest.WithAssignment(&env)))
	} else {
		os.Exit(m.Run())
	}
}
