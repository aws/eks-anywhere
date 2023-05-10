package upgradevalidations

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
	cliversion "github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// ValidateEKSAVersionSkew ensures that the eksa version is incremented by one minor version exactly (e.g. 0.14 -> 0.15).
func ValidateEKSAVersionSkew(ctx context.Context, upgradeVersion string, k validations.KubectlClient, cluster *types.Cluster) error {
	c, err := k.GetEksaCluster(ctx, cluster, cluster.Name)
	if err != nil {
		return err
	}
	if c.Spec.BundlesRef == nil {
		return fmt.Errorf("cluster bundlesRef cannot be nil")
	}
	bundles, err := k.GetBundles(ctx, cluster.KubeconfigFile, c.Spec.BundlesRef.Name, c.Spec.BundlesRef.Namespace)
	if err != nil {
		return err
	}

	clusterVersion, err := parseTags(bundles)
	if err != nil {
		return err
	}

	parsedClusterVersion, err := cliversion.New(clusterVersion)
	if err != nil {
		return fmt.Errorf("parsing cluster cli version: %v", err)
	}

	parsedUpgradeVersion, err := cliversion.New(upgradeVersion)
	if err != nil {
		return fmt.Errorf("parsing upgrade cli version: %v", err)
	}

	logger.V(3).Info("calculating version differences", "upgradeVersion", parsedUpgradeVersion, "clusterVersion", parsedClusterVersion)
	majorVersionDifference := math.Abs(float64(parsedUpgradeVersion.Major) - float64(parsedClusterVersion.Major))
	minorVersionDifference := float64(parsedUpgradeVersion.Minor) - float64(parsedClusterVersion.Minor)
	logger.V(3).Info("calculated version differences", "majorVersionDifference", majorVersionDifference, "minorVersionDifference", minorVersionDifference)

	if majorVersionDifference > 0 || !(minorVersionDifference <= supportedMinorVersionIncrement && minorVersionDifference >= 0) {
		msg := fmt.Sprintf("WARNING: version difference between upgrade version (%d.%d) and cluster version (%d.%d) do not meet the supported version increment of +%d",
			parsedUpgradeVersion.Major, parsedUpgradeVersion.Minor, parsedClusterVersion.Major, parsedClusterVersion.Minor, supportedMinorVersionIncrement)
		return fmt.Errorf(msg)
	}
	return nil
}

func parseTags(bundles *releasev1alpha1.Bundles) (string, error) {
	image := bundles.Spec.VersionsBundles[0].Eksa.ClusterController.URI
	if !strings.Contains(image, ":") {
		return "", fmt.Errorf("could not find tag in Eksa Cluster Controller Image")
	}
	tag := strings.Split(image, ":")[1]
	version := strings.Split(tag, "-")[0]
	return version, nil
}
