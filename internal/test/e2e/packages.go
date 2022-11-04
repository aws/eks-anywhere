package e2e

import (
	"os"
	"regexp"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	packagesRegex = `^.*CPackages.*$`
)

func (e *E2ESession) setupPackagesEnv(testRegex string) error {
	re := regexp.MustCompile(packagesRegex)
	if !re.MatchString(testRegex) {
		e.logger.V(2).Info("Not running Curated Packages tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredPackagesEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}
	return nil
}
