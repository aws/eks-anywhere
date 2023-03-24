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

const (
	eksaPackagesRegion    = "EKSA_AWS_SECRET_ACCESS_KEY"
	eksaPackagesAccessKey = "EKSA_AWS_ACCESS_KEY_ID"
	eksaPackagesSecretKey = "EKSA_AWS_REGION"
	route53AccessKey      = "ROUTE53_ACCESS_KEY_ID"
	route53SecretKey      = "ROUTE53_SECRET_ACCESS_KEY"
	route53Region         = "ROUTE53_REGION"
	route53ZoneID         = "ROUTE53_ZONEID"
)

var requiredPackagesEnvVars = []string{
	eksaPackagesRegion,
	eksaPackagesAccessKey,
	eksaPackagesSecretKey,
}

var requiredCertManagerEnvVars = []string{
	route53Region,
	route53AccessKey,
	route53SecretKey,
	route53ZoneID,
}

// RequiredPackagesEnvVars returns the list of packages env vars.
func RequiredPackagesEnvVars() []string {
	return requiredPackagesEnvVars
}

// RequiredCertManagerEnvVars returns the list of cert manager env vars.
func RequiredCertManagerEnvVars() []string {
	return requiredCertManagerEnvVars
}

// CheckCuratedPackagesCredentials will exit out if the Curated Packages environment variables are not set.
func CheckCuratedPackagesCredentials(t *testing.T) {
	for _, env := range requiredPackagesEnvVars {
		_, ok := os.LookupEnv(env)
		if !ok {
			t.Fatalf("Error Unset Packages environment variable: %v is required", env)
		}
	}
}

// CheckCertManagerCredentials will exit if route53 credentials are not set.
func CheckCertManagerCredentials(t *testing.T) {
	for _, env := range requiredCertManagerEnvVars {
		_, ok := os.LookupEnv(env)
		if !ok {
			t.Fatalf("Error Unset Cert Manager environment variable: %v is required", env)
		}
	}
}

// GetRoute53Configs returns route53 configurations for cert-manager.
func GetRoute53Configs() (string, string, string, string) {
	return os.Getenv(route53AccessKey), os.Getenv(route53SecretKey),
		os.Getenv(route53Region), os.Getenv(route53ZoneID)
}
