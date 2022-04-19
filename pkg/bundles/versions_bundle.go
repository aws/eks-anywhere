package bundles

import (
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func VersionsBundleForKubernetesVersion(bundles *releasev1.Bundles, kubeVersion string) *releasev1.VersionsBundle {
	for _, versionsBundle := range bundles.Spec.VersionsBundles {
		if versionsBundle.KubeVersion == kubeVersion {
			return &versionsBundle
		}
	}
	return nil
}
