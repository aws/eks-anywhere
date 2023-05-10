package e2e

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

var proxyVarsByProvider = map[string]e2etests.ProxyRequiredEnvVars{
	"CloudStack": e2etests.CloudstackProxyRequiredEnvVars,
	"VSphere":    e2etests.VsphereProxyRequiredEnvVars,
	"Tinkerbell": e2etests.TinkerbellProxyRequiredEnvVars,
}

func (e *E2ESession) setupProxyEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*Proxy.*$`)
	if !re.MatchString(testRegex) {
		return nil
	}
	var requiredEnvVars e2etests.ProxyRequiredEnvVars
	for key, vars := range proxyVarsByProvider {
		if strings.Contains(testRegex, key) {
			requiredEnvVars = vars
			break
		}
	}
	if reflect.ValueOf(requiredEnvVars).IsZero() {
		return fmt.Errorf("proxy config for provider test %s was not found", testRegex)
	}

	for _, eVar := range []string{requiredEnvVars.HttpProxy, requiredEnvVars.HttpsProxy, requiredEnvVars.NoProxy} {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	return nil
}
