package clusters_test

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/internal/test/envtest"
)

var env *envtest.Environment

func TestMain(m *testing.M) {
	os.Exit(envtest.RunWithEnvironment(m, envtest.WithAssignment(&env)))
}
