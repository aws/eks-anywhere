package clusters_test

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/test/envtest"
)

const integrationTestEnvVar = "INTEGRATION_TESTS_ENABLED"

var env *envtest.Environment

func TestMain(m *testing.M) {
	if os.Getenv(integrationTestEnvVar) == "true" {
		os.Exit(envtest.RunWithEnvironment(m, envtest.WithAssignment(&env)))
	}
}
