package bundles

import releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"

func Charts(bundles *releasev1.Bundles) []releasev1.Image {
	var charts []releasev1.Image
	for _, v := range bundles.Spec.VersionsBundles {
		versionsBundleCharts := v.Charts()
		for _, c := range versionsBundleCharts {
			charts = append(charts, *c)
		}
	}

	return charts
}
