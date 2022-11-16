package e2e

import (
	"os"
	"regexp"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	nutanixFeatureGateEnvVar = "NUTANIX_PROVIDER"
	nutanixCidrVar           = "T_NUTANIX_CONTROL_PLANE_CIDR"
	nutanixRegex             = `^.*Nutanix.*$`
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

	// Since Nutanix Provider is feature gated, manually enable the feature gate for all Nutanix tests.
	e.testEnvVars[nutanixFeatureGateEnvVar] = "true"
	return nil
}
