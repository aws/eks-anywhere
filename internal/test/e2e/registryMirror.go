package e2e

import (
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupRegistryMirrorEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*RegistryMirror.*$`)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running RegistryMirror tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredRegistryMirrorEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	return nil
}
