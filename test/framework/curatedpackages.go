package framework

import (
	"os"
	"testing"
)

type PackageConfig struct {
	*HelmInstallConfig
	bundleURI string
}

func WithPackageConfig(t *testing.T, bundleURI, chartName, chartURI,
	chartVersion string, chartValues []string,
) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.PackageConfig = &PackageConfig{
			HelmInstallConfig: &HelmInstallConfig{
				chartName:    chartName,
				chartURI:     chartURI,
				chartVersion: chartVersion,
				chartValues:  chartValues,
				HelmClient:   buildHelm(t),
			},
			bundleURI: bundleURI,
		}
	}
}

// CheckCuratedPackagesCredentials will exit out if the Curated Packages environment variables are not set.
func CheckCuratedPackagesCredentials(t *testing.T) {
	requiredEnvVars := []string{
		"EKSA_AWS_SECRET_ACCESS_KEY",
		"EKSA_AWS_ACCESS_KEY_ID",
		"EKSA_AWS_REGION",
	}
	for _, env := range requiredEnvVars {
		_, ok := os.LookupEnv(env)
		if !ok {
			t.Fatalf("Error Unset Packages environment variable: %v is required", env)
		}
	}
}
