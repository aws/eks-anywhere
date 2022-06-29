package e2e

import (
	"fmt"
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupProxyEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*Proxy.*$`)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running Proxy tests, skipping Env variable setup")
		return nil
	}
	var requiredEnvVars e2etests.ProxyRequiredEnvVars
	if isTestProvider(testRegex, "VSphere") {
		requiredEnvVars = e2etests.RequiredVSphereProxyEnvVars()
	} else if isTestProvider(testRegex, "CloudStack") {
		requiredEnvVars = e2etests.RequiredCloudStackProxyEnvVars()
	} else if isTestProvider(testRegex, "Snow") {
		// TODO: provide separate Proxy Env Vars for Snow provider. Leaving VSphere for backwards compatibility
		requiredEnvVars = e2etests.RequiredVSphereProxyEnvVars()
	} else {
		return fmt.Errorf("proxy config for provider test %s was not found", testRegex)
	}
	for _, eVar := range []string{requiredEnvVars.HttpProxy, requiredEnvVars.HttpsProxy, requiredEnvVars.NoProxy} {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	return nil
}

func isTestProvider(testRegex string, providerName string) bool {
	reProvider := regexp.MustCompile(fmt.Sprintf(`^.*%s.*$`, providerName))
	return reProvider.MatchString(testRegex)
}
