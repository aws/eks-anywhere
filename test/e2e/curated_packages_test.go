//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	eksAnywherePackagesHelmChartName = "eks-anywhere-packages"
	eksAnywherePackagesHelmUri       = "oci://public.ecr.aws/l0g8r8j6/eks-anywhere-packages"
	eksAnywherePackagesHelmVersion   = "0.1.6-eks-a-v0.0.0-dev-build.2404"
)

var eksAnywherePackagesHelmValues = []string{"sourceRegistry=public.ecr.aws/l0g8r8j6"}

func TestKubernetes122PackagesInstallSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithHelmInstallConfig(t, eksAnywherePackagesHelmChartName, eksAnywherePackagesHelmUri, eksAnywherePackagesHelmVersion, eksAnywherePackagesHelmValues),
	)
	runHelmInstallSimpleFlow(test)
}
