package e2e

import (
	"os"
	"regexp"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	cloudstackRegex   = `^.*CloudStack.*$`
	cloudstackCidrVar = "T_CLOUDSTACK_CIDR"
)

func (e *E2ESession) setupCloudStackEnv(testRegex string) error {
	re := regexp.MustCompile(cloudstackRegex)
	if !re.MatchString(testRegex) {
		return nil
	}

	requiredEnvVars := e2etests.RequiredCloudstackEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}
	return nil
}
