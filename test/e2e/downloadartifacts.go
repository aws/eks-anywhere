//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/test/framework"
)

func runDownloadArtifactsFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.DownloadArtifacts()
}
