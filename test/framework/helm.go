package framework

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/helm"
)

type HelmInstallConfig struct {
	chartName    string
	chartURI     string
	chartVersion string
	chartValues  []string
	HelmClient   helm.Client
}

func WithHelmInstallConfig(t *testing.T, chartName, chartURI, chartVersion string, chartValues []string) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.HelmInstallConfig = &HelmInstallConfig{
			chartName:    chartName,
			chartURI:     chartURI,
			chartVersion: chartVersion,
			chartValues:  chartValues,
			HelmClient:   buildHelm(t),
		}
	}
}
