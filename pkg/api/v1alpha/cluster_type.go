package v1alpha

import (
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/semver"
)

// KubeVersionToValidSemver converts kube version to semver for comparisons.
func KubeVersionToValidSemver(kubeVersion v1alpha1.KubernetesVersion) (*semver.Version, error) {
	// appending the ".0" as the patch version to have a valid semver string and use those semvers for comparison
	return semver.New(string(kubeVersion) + ".0")
}
