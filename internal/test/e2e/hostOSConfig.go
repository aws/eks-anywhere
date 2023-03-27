package e2e

import (
	"os"
	"regexp"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupNTPEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*NTP.*$`)
	if !re.MatchString(testRegex) {
		return nil
	}

	for _, eVar := range e2etests.RequiredNTPServersEnvVars() {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}
	return nil
}

func (e *E2ESession) setupBottlerocketKubernetesSettingsEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*BottlerocketKubernetesSettings.*$`)
	if !re.MatchString(testRegex) {
		return nil
	}

	for _, eVar := range e2etests.RequiredBottlerocketKubernetesSettingsEnvVars() {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}
	return nil
}
