package e2e

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

var proxyVarsByProvider = map[string]e2etests.ProxyRequiredEnvVars{
	"CloudStack": e2etests.CloudstackProxyRequiredEnvVars,
	"VSphere":    e2etests.VsphereProxyRequiredEnvVars,
}

func (e *E2ESession) setupProxyEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*Proxy.*$`)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running Proxy tests, skipping Env variable setup")
		return nil
	}
	var requiredEnvVars e2etests.ProxyRequiredEnvVars
	for key, vars := range proxyVarsByProvider{
		if strings.Contains(testRegex, key) {
			requiredEnvVars = vars
			break
		}
	}
	if reflect.ValueOf(requiredEnvVars).IsZero() {
		return fmt.Errorf("proxy config for provider test %s was not found", testRegex)
	}
	return nil
}