package e2e

import (
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

const (
	cidrVar               = "T_VSPHERE_CIDR"
	privateNetworkCidrVar = "T_VSPHERE_PRIVATE_NETWORK_CIDR"
	vsphereRegex          = `^.*VSphere.*$`
)

func (e *E2ESession) setupVSphereEnv(testRegex string) error {
	re := regexp.MustCompile(vsphereRegex)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running VSphere tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredVsphereEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	ipPool := e.ipPool.ToString()
	if ipPool != "" {
		e.testEnvVars[e2etests.VsphereClusterIPPoolEnvVar] = ipPool
	}

	return nil
}
