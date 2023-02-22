//go:build e2e
// +build e2e

package e2e

import (
	"os"

	"github.com/aws/eks-anywhere/test/framework"
)

// runVSphereAirgapConfigFlow run airgap deployment but allow bootstrap cluster to access local peers.
func runVSphereAirgapConfigFlow(test *framework.ClusterE2ETest, localCIDRs string) {
	test.GenerateClusterConfig()
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()
	test.DownloadImages()
	test.ImportImages()
	test.AirgapDockerContainers(localCIDRs)
	test.CreateAirgappedUser(localCIDRs)
	test.AssertAirgappedNetwork()
	test.CreateCluster(
		framework.WithSudo("airgap"),
		framework.WithBundlesOverride("./eks-anywhere-downloads/bundle-release.yaml"), // generated by ExtractDownloadArtifacts
	)
	test.DeleteCluster()
}

func runDockerAirgapConfigFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()
	test.DownloadImages()
	test.ChangeInstanceSecurityGroup(os.Getenv(framework.RegistryMirrorAirgappedSecurityGroup))
	test.ImportImages()
	test.CreateCluster(framework.WithBundlesOverride("./eks-anywhere-downloads/bundle-release.yaml"))
	test.DeleteCluster(framework.WithBundlesOverride("./eks-anywhere-downloads/bundle-release.yaml"))
	test.ChangeInstanceSecurityGroup(os.Getenv(framework.RegistryMirrorDefaultSecurityGroup))
}
