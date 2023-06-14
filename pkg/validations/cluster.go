package validations

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/types"
)

// ValidateOSForRegistryMirror checks if the OS is valid for the provided registry mirror configuration.
func ValidateOSForRegistryMirror(clusterSpec *cluster.Spec, provider providers.Provider) error {
	cluster := clusterSpec.Cluster
	if cluster.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}

	machineConfigs := provider.MachineConfigs(clusterSpec)
	if !cluster.Spec.RegistryMirrorConfiguration.InsecureSkipVerify || machineConfigs == nil {
		return nil
	}

	for _, mc := range machineConfigs {
		if mc.OSFamily() == v1alpha1.Bottlerocket {
			return errors.New("InsecureSkipVerify is not supported for bottlerocket")
		}
	}
	return nil
}

func ValidateCertForRegistryMirror(clusterSpec *cluster.Spec, tlsValidator TlsValidator) error {
	cluster := clusterSpec.Cluster
	if cluster.Spec.RegistryMirrorConfiguration == nil {
		return nil
	}

	if cluster.Spec.RegistryMirrorConfiguration.InsecureSkipVerify {
		logger.V(1).Info("Warning: skip registry certificate verification is enabled", "registryMirrorConfiguration.insecureSkipVerify", true)
		return nil
	}

	host, port := cluster.Spec.RegistryMirrorConfiguration.Endpoint, cluster.Spec.RegistryMirrorConfiguration.Port
	authorityUnknown, err := tlsValidator.IsSignedByUnknownAuthority(host, port)
	if err != nil {
		return fmt.Errorf("validating registry mirror endpoint: %v", err)
	}
	if authorityUnknown {
		logger.V(1).Info(fmt.Sprintf("Warning: registry mirror endpoint %s is using self-signed certs", cluster.Spec.RegistryMirrorConfiguration.Endpoint))
	}

	certContent := cluster.Spec.RegistryMirrorConfiguration.CACertContent
	if certContent == "" && authorityUnknown {
		return fmt.Errorf("registry %s is using self-signed certs, please provide the certificate using caCertContent field. Or use insecureSkipVerify field to skip registry certificate verification", cluster.Spec.RegistryMirrorConfiguration.Endpoint)
	}

	if certContent != "" {
		if err = tlsValidator.ValidateCert(host, port, certContent); err != nil {
			return fmt.Errorf("invalid registry certificate: %v", err)
		}
	}

	return nil
}

// ValidateAuthenticationForRegistryMirror checks if REGISTRY_USERNAME and REGISTRY_PASSWORD is set if authenticated registry mirrors are used.
func ValidateAuthenticationForRegistryMirror(clusterSpec *cluster.Spec) error {
	cluster := clusterSpec.Cluster
	if cluster.Spec.RegistryMirrorConfiguration != nil && cluster.Spec.RegistryMirrorConfiguration.Authenticate {
		_, _, err := config.ReadCredentials()
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateManagementClusterName checks if the management cluster specified in the workload cluster spec is valid.
func ValidateManagementClusterName(ctx context.Context, k KubectlClient, mgmtCluster *types.Cluster, mgmtClusterName string) error {
	cluster, err := k.GetEksaCluster(ctx, mgmtCluster, mgmtClusterName)
	if err != nil {
		return err
	}
	if cluster.IsManaged() {
		return fmt.Errorf("%s is not a valid management cluster", mgmtClusterName)
	}
	return nil
}

// ValidateManagementClusterBundlesVersion checks if management cluster's bundle version
// is greater than or equal to the bundle version used to upgrade a workload cluster.
func ValidateManagementClusterBundlesVersion(ctx context.Context, k KubectlClient, mgmtCluster *types.Cluster, workload *cluster.Spec) error {
	cluster, err := k.GetEksaCluster(ctx, mgmtCluster, mgmtCluster.Name)
	if err != nil {
		return err
	}

	if cluster.Spec.BundlesRef == nil {
		return fmt.Errorf("management cluster bundlesRef cannot be nil")
	}

	mgmtBundles, err := k.GetBundles(ctx, mgmtCluster.KubeconfigFile, cluster.Spec.BundlesRef.Name, cluster.Spec.BundlesRef.Namespace)
	if err != nil {
		return err
	}

	if mgmtBundles.Spec.Number < workload.Bundles.Spec.Number {
		return fmt.Errorf("cannot upgrade workload cluster with bundle spec.number %d while management cluster %s is on older bundle spec.number %d", workload.Bundles.Spec.Number, mgmtCluster.Name, mgmtBundles.Spec.Number)
	}

	return nil
}

func ValidateEKSAVersionSkew(ctx context.Context, k KubectlClient, cluster *types.Cluster, workload *cluster.Spec) error {
	c, err := k.GetEksaCluster(ctx, cluster, cluster.Name)
	if err != nil {
		return err
	}

	parsedClusterVersion, err := semver.New(string(*c.Spec.EksaVersion))
	if err != nil {
		return fmt.Errorf("parsing cluster cli version: %v", err)
	}

	parsedUpgradeVersion, err := semver.New(string(*workload.Cluster.Spec.EksaVersion))
	if err != nil {
		return fmt.Errorf("parsing upgrade cli version: %v", err)
	}

	majorVersionDifference := math.Abs(float64(parsedUpgradeVersion.Major) - float64(parsedClusterVersion.Major))
	minorVersionDifference := float64(parsedUpgradeVersion.Minor) - float64(parsedClusterVersion.Minor)
	supportedMinorVersionIncrement := float64(1)

	if majorVersionDifference > 0 || !(minorVersionDifference <= supportedMinorVersionIncrement && minorVersionDifference >= 0) {
		msg := fmt.Sprintf("WARNING: version difference between upgrade version (%d.%d) and cluster version (%d.%d) do not meet the supported version increment of +%f",
			parsedUpgradeVersion.Major, parsedUpgradeVersion.Minor, parsedClusterVersion.Major, parsedClusterVersion.Minor, supportedMinorVersionIncrement)
		return fmt.Errorf(msg)
	}
	return nil
}

func ValidateManagementClusterEksaVersion(ctx context.Context, k KubectlClient, mgmtCluster *types.Cluster, workload *cluster.Spec) error {
	cluster, err := k.GetEksaCluster(ctx, mgmtCluster, mgmtCluster.Name)
	if err != nil {
		return err
	}

	if cluster.Spec.EksaVersion == nil {
		return fmt.Errorf("management cluster eksaVersion cannot be nil")
	}

	mVersion, err := semver.New(string(*cluster.Spec.EksaVersion))
	if err != nil {
		return fmt.Errorf("management cluster eksaVersion is invalid: %w", err)
	}

	wVersion, err := semver.New(string(*workload.Cluster.Spec.EksaVersion))
	if err != nil {
		return fmt.Errorf("workload cluster eksaVersion is invalid: %w", err)
	}

	if wVersion.GreaterThan(mVersion) {
		return fmt.Errorf("Cannot upgrade workload cluster with version %v while management cluster is an older version %v", wVersion, mVersion)
	}

	return nil
}

func ValidateManagementWorkloadSkew(ctx context.Context, k KubectlClient, mgmtCluster *types.Cluster, target *cluster.Spec) error {
	cluster, err := k.GetEksaCluster(ctx, mgmtCluster, mgmtCluster.Name)
	if err != nil {
		return err
	}

	workloads, err := k.GetEksaClusters(ctx, mgmtCluster)

	if err != nil {
		return err
	}

	mVersion, err := semver.New(string(*cluster.Spec.EksaVersion))
	if err != nil {
		return fmt.Errorf("management cluster eksaVersion is invalid: %w", err)
	}

	newVersion, err := semver.New(string(*target.Cluster.Spec.EksaVersion))
	if err != nil {
		return fmt.Errorf("upgrade eksaVersion is invalid: %w", err)
	}

	for _, w := range workloads {
		if w.Spec.ManagementCluster.Name != mgmtCluster.Name || w.Name == cluster.Name {
			continue
		}

		if w.Spec.EksaVersion == nil {
			return fmt.Errorf("workload cluster eksaVersion cannot be nil: %v", w)
		}

		wVersion, err := semver.New(string(*w.Spec.EksaVersion))
		if err != nil {
			return fmt.Errorf("workload cluster eksaVersion is invalid: %v", w)
		}

		if !newVersion.Equal(mVersion) && mVersion.GreaterThan(wVersion) {
			return fmt.Errorf("Cannot upgrade management cluster to %v. There can only be a skew of one eksa minor version against workload cluster %s: %v", newVersion, w.Name, wVersion)
		}
	}

	return nil
}
