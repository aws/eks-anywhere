package config

import (
	"os"
)

const (
	HttpsProxyKey = "HTTPS_PROXY"
	HttpProxyKey  = "HTTP_PROXY"
	NoProxyKey    = "NO_PROXY"
)

func GetProxyConfigFromEnv() map[string]string {
	return map[string]string{
		HttpsProxyKey: os.Getenv(HttpsProxyKey),
		HttpProxyKey:  os.Getenv(HttpProxyKey),
		NoProxyKey:    os.Getenv(NoProxyKey),
	}
}
