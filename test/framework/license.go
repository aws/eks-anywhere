package framework

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// WithCurrentCLILicense switches a release-upgrade fixture to the CLI under test's signing key.
func WithCurrentCLILicense() ClusterE2ETestOpt {
	return WithClusterUpgrade(func(c *v1alpha1.Cluster) {
		c.Spec.LicenseToken = licenseTokenForCluster(c, false)
	})
}

// WithReleasedCLILicense selects a production-signed token before invoking a released CLI.
func WithReleasedCLILicense() ClusterE2ETestOpt {
	return WithClusterUpgrade(func(c *v1alpha1.Cluster) {
		c.Spec.LicenseToken = licenseTokenForCluster(c, true)
	})
}

func licenseTokenForCluster(c *v1alpha1.Cluster, released bool) string {
	if released {
		if c.IsManaged() {
			return GetStagingLicenseToken2()
		}
		return GetStagingLicenseToken()
	}
	if c.IsManaged() {
		return GetLicenseToken2()
	}
	return GetLicenseToken()
}
