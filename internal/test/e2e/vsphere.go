package e2e

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	vsphereCidrVar               = "T_VSPHERE_CIDR"
	vspherePrivateNetworkCidrVar = "T_VSPHERE_PRIVATE_NETWORK_CIDR"
	vsphereRegex                 = `^.*VSphere.*$`
)

func (e *E2ESession) setupVSphereEnv(testRegex string) error {
	re := regexp.MustCompile(vsphereRegex)
	if !re.MatchString(testRegex) {
		return nil
	}

	requiredEnvVars := e2etests.RequiredVsphereEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	// This algorithm is not very efficient with two nested loops
	// Making the assumption that VSphereExtraEnvVarPrefixes() returns a very small number of prefixes
	// this should be ok and probably not worth the complexity of building a more complex data structure.
	// If in the future we see the need to have a bigger number of prefixes, we will need
	// to change this to avoid the n*m complexity
	envVars := os.Environ()
	for _, envVarPrefix := range e2etests.VSphereExtraEnvVarPrefixes() {
		for _, envVar := range envVars {
			if strings.HasPrefix(envVar, envVarPrefix) {
				split := strings.Split(envVar, "=")
				if len(split) != 2 {
					return fmt.Errorf("invalid vsphere env var format, expected key=value: %s", envVar)
				}
				key := split[0]
				value := split[1]

				e.testEnvVars[key] = value
			}
		}
	}

	return nil
}
