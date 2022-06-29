package framework

import (
	"os"
	"strings"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

const (
	vsphereHttpProxyVar     = "T_HTTP_PROXY"
	vsphereHttpsProxyVar    = "T_HTTPS_PROXY"
	vsphereNoProxyVar       = "T_NO_PROXY"
	cloudstackHttpProxyVar  = "T_HTTP_PROXY_CLOUDSTACK"
	cloudstackHttpsProxyVar = "T_HTTPS_PROXY_CLOUDSTACK"
	cloudstackNoProxyVar    = "T_NO_PROXY_CLOUDSTACK"
)

var vsphereProxyRequiredEnvVars = ProxyRequiredEnvVars{
	HttpProxy:  vsphereHttpProxyVar,
	HttpsProxy: vsphereHttpsProxyVar,
	NoProxy:    vsphereNoProxyVar,
}

var cloudstackProxyRequiredEnvVars = ProxyRequiredEnvVars{
	HttpProxy:  cloudstackHttpProxyVar,
	HttpsProxy: cloudstackHttpsProxyVar,
	NoProxy:    cloudstackNoProxyVar,
}

type ProxyRequiredEnvVars struct {
	HttpProxy  string
	HttpsProxy string
	NoProxy    string
}

func WithProxy(requiredEnvVars ProxyRequiredEnvVars) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, []string{requiredEnvVars.HttpProxy, requiredEnvVars.HttpsProxy, requiredEnvVars.NoProxy})
		HttpProxy := os.Getenv(requiredEnvVars.HttpProxy)
		HttpsProxy := os.Getenv(requiredEnvVars.HttpsProxy)
		noProxies := os.Getenv(requiredEnvVars.NoProxy)
		var noProxy []string
		for _, data := range strings.Split(noProxies, ",") {
			noProxy = append(noProxy, strings.TrimSpace(data))
		}

		e.clusterFillers = append(e.clusterFillers,
			api.WithProxyConfig(HttpProxy, HttpsProxy, noProxy),
		)
	}
}

func RequiredVSphereProxyEnvVars() ProxyRequiredEnvVars {
	return vsphereProxyRequiredEnvVars
}

func RequiredCloudStackProxyEnvVars() ProxyRequiredEnvVars {
	return cloudstackProxyRequiredEnvVars
}
