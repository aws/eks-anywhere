package config_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/config"
)

func TestGetProxyConfigFromEnv(t *testing.T) {
	wantHttpsProxy := "FOO"
	wantHttpProxy := "BAR"
	wantNoProxy := "localhost,anotherhost"
	wantEnv := map[string]string{
		config.HttpsProxyKey: wantHttpsProxy,
		config.HttpProxyKey:  wantHttpProxy,
		config.NoProxyKey:    wantNoProxy,
	}
	for k, v := range wantEnv {
		t.Setenv(k, v)
	}
	env := config.GetProxyConfigFromEnv()

	for k, target := range wantEnv {
		if val := env[k]; val != target {
			t.Fatalf("config.GetProxyConfigFromEnv %s = %s, want %s", k, val, target)
		}
	}
}
