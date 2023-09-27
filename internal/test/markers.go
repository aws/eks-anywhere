package test

import (
	"os"
	"testing"
)

const IntegrationTestEnvVar = "INTEGRATION_TESTS_ENABLED"

func MarkIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv(IntegrationTestEnvVar) != "true" {
		t.Skipf("set env var '%v=true' to run this test", IntegrationTestEnvVar)
	}
}
