package framework

import (
	"os"
	"strings"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	httpProxyVar  = "T_HTTP_PROXY"
	httpsProxyVar = "T_HTTPS_PROXY"
	noProxyVar    = "T_NO_PROXY"
)

var proxyRequiredEnvVars = []string{
	httpProxyVar,
	httpsProxyVar,
	noProxyVar,
}

func WithProxy() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, proxyRequiredEnvVars)
		httpProxy := os.Getenv(httpProxyVar)
		httpsProxy := os.Getenv(httpsProxyVar)
		noProxies := os.Getenv(noProxyVar)
		var noProxy []string
		for _, data := range strings.Split(noProxies, ",") {
			noProxy = append(noProxy, strings.TrimSpace(data))
		}

		e.clusterFillers = append(e.clusterFillers,
			api.WithProxyConfig(httpProxy, httpsProxy, noProxy),
		)
	}
}

func RequiredProxyEnvVars() []string {
	return proxyRequiredEnvVars
}
