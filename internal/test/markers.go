package test

import (
	"os"
	"testing"
)

// IntegrationTestEnvVar is the name of the environment variable that gates when integration tests are run.
const IntegrationTestEnvVar = "INTEGRATION_TESTS_ENABLED"

// MarkIntegration marks a test as an integration test.
// Integration tests are only run when the environment variable INTEGRATION_TESTS_ENABLED is set to true.
func MarkIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv(IntegrationTestEnvVar) != "true" {
		t.Skipf("set env var '%v=true' to run this test", IntegrationTestEnvVar)
	}
}
