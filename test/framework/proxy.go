package framework

import (
	"os"
	"strings"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	httpProxyVar  = "T_HTTP_PROXY"
	httspProxyVar = "T_HTTPS_PROXY"
	noProxyVar    = "T_NO_PROXY"
)

var proxyRequiredEnvVars = []string{
	httpProxyVar,
	httspProxyVar,
	noProxyVar,
}

func WithProxy() E2ETestOpt {
	return func(e *E2ETest) {
		checkRequiredEnvVars(e.T, proxyRequiredEnvVars)
		httpProxy := os.Getenv(httpProxyVar)
		httpsProxy := os.Getenv(httpProxyVar)
		noProxies := os.Getenv(noProxyVar)
		noProxy := []string{}
		for _, data := range strings.Split(noProxies, ",") {
			noProxy = append(noProxy, strings.TrimSpace(data))
		}

		e.clusterFillers = append(e.clusterFillers,
			api.WithProxyConfig(httpProxy, httpsProxy, noProxy),
		)
	}
}
