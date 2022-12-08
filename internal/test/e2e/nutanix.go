package e2e

import (
	"os"
	"regexp"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	nutanixCidrVar = "T_NUTANIX_CONTROL_PLANE_CIDR"
	nutanixRegex   = `^.*Nutanix.*$`
)

func (e *E2ESession) setupNutanixEnv(testRegex string) error {
	re := regexp.MustCompile(nutanixRegex)
	if !re.MatchString(testRegex) {
		return nil
	}

	requiredEnvVars := e2etests.RequiredNutanixEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	return nil
}
