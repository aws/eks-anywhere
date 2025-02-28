package e2e

import (
	"os"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupCommonEnv() error {
	requiredEnvVars := e2etests.RequiredCommonEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	// overwrite license token env variables for staging
	for _, eVar := range requiredEnvVars {
		if e.stage == "staging" {
			if val, ok := os.LookupEnv("STAGING_" + eVar); ok {
				e.testEnvVars[eVar] = val
			}
		}
	}
	return nil
}
