package e2e

import (
	"os"
	"regexp"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupFluxEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*Flux.*$`)
	if !re.MatchString(testRegex) {
		return nil
	}

	requiredEnvVars := e2etests.RequiredFluxGithubEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}
	return nil
}
