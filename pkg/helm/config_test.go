package helm_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

func TestWithRegistryMirror(t *testing.T) {
	g := NewWithT(t)
	cluster := test.Cluster(func(c *v1alpha1.Cluster) {
		c.Name = "test-cluster"
		c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
			Authenticate: true,
			Endpoint:     "1.2.3.4",
			Port:         "5000",
		}
	})

	r := registrymirror.FromCluster(cluster)
	config := helm.NewConfig(helm.WithRegistryMirror(r))
	g.Expect(config.RegistryMirror).To(Equal(r))
}

func TestWithInsecure(t *testing.T) {
	g := NewWithT(t)
	tests := map[string]struct {
		insecure bool
	}{
		"WithoutInsecure": {
			insecure: false,
		},
		"WithInsecure": {
			insecure: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			opts := []helm.Opt{}
			if tt.insecure {
				opts = append(opts, helm.WithInsecure())
			}
			config := helm.NewConfig(opts...)
			g.Expect(config.Insecure).To(Equal(tt.insecure))
		})
	}
}

func TestWithProxyConfig(t *testing.T) {
	g := NewWithT(t)
	proxyConfigMap := map[string]string{
		"HTTP_PROXY":  "http://1.2.3.4:5050",
		"HTTPS_PROXY": "https://1.2.3.4:5050",
		"NO_PROXY":    "test hostname",
	}
	config := helm.NewConfig(helm.WithProxyConfig(proxyConfigMap))
	g.Expect(config.ProxyConfig).To(Equal(proxyConfigMap))
}
