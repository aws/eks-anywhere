package upgradevalidations

import (
	"context"
	"fmt"
	"math"

	"k8s.io/apimachinery/pkg/util/version"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

const supportedMinorVersionIncrement = 1

func ValidateServerVersionSkew(ctx context.Context, compareVersion v1alpha1.KubernetesVersion, cluster *types.Cluster, kubectl validations.KubectlClient) error {
	versions, err := kubectl.Version(ctx, cluster)
	if err != nil {
		return fmt.Errorf("fetching cluster version: %v", err)
	}

	parsedInputVersion, err := version.ParseGeneric(string(compareVersion))
	if err != nil {
		return fmt.Errorf("parsing comparison version: %v", err)
	}

	parsedServerVersion, err := version.ParseSemantic(versions.ServerVersion.GitVersion)
	if err != nil {
		return fmt.Errorf("parsing cluster version: %v", err)
	}

	logger.V(3).Info("calculating version differences", "inputVersion", parsedInputVersion, "clusterVersion", parsedServerVersion)
	majorVersionDifference := math.Abs(float64(parsedInputVersion.Major()) - float64(parsedServerVersion.Major()))
	minorVersionDifference := float64(parsedInputVersion.Minor()) - float64(parsedServerVersion.Minor())
	logger.V(3).Info("calculated version differences", "majorVersionDifference", majorVersionDifference, "minorVersionDifference", minorVersionDifference)

	if majorVersionDifference > 0 || !(minorVersionDifference <= supportedMinorVersionIncrement && minorVersionDifference >= 0) {
		msg := fmt.Sprintf("WARNING: version difference between upgrade version (%d.%d) and server version (%d.%d) do not meet the supported version increment of +%d",
			parsedInputVersion.Major(), parsedInputVersion.Minor(), parsedServerVersion.Major(), parsedServerVersion.Minor(), supportedMinorVersionIncrement)
		return fmt.Errorf(msg)
	}
	return nil
}
