//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	bottlerocketOSFileName         = "bottlerocket"
	airgapUsername                 = "airgap"
	bundleReleasePathFromArtifacts = "./eks-anywhere-downloads/bundle-release.yaml"
)

// runAirgapConfigFlow run airgap deployment but allow bootstrap cluster to access local peers.
func runAirgapConfigFlow(test *framework.ClusterE2ETest, localCIDRs string) {
	test.GenerateClusterConfig()
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()
	test.DownloadImages()
	test.AirgapDockerContainers(localCIDRs)
	test.CreateAirgappedUser(localCIDRs)
	test.AssertAirgappedNetwork()
	test.ImportImages()
	test.CreateCluster(
		framework.WithSudo(airgapUsername),
		framework.WithBundlesOverride(bundleReleasePathFromArtifacts), // generated by ExtractDownloadArtifacts
	)
	test.DeleteCluster()
}

func runTinkerbellAirgapConfigFlow(test *framework.ClusterE2ETest, localCIDRs, kubeVersion string) {
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()

	brContent, err := os.ReadFile(bundleReleasePathFromArtifacts)
	if err != nil {
		test.T.Fatalf("Cannot read bundleRelease file: %v", err)
	}

	server := downloadAndServeTinkerbellArtifacts(test.T, brContent, kubeVersion)
	defer server.Shutdown(context.Background())

	test.GenerateClusterConfig()
	test.DownloadImages(
		framework.WithBundlesOverride(bundleReleasePathFromArtifacts),
	)
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.AirgapDockerContainers(localCIDRs)
	test.CreateAirgappedUser(localCIDRs)
	test.AssertAirgappedNetwork()
	test.ImportImages()
	test.CreateCluster(
		// airgap user should be airgapped through iptables
		framework.WithSudo(airgapUsername),
		framework.WithBundlesOverride(bundleReleasePathFromArtifacts), // generated by ExtractDownloadArtifacts
		framework.WithControlPlaneWaitTimeout("20m"))
	test.StopIfFailed()
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func runTinkerbellAirgapConfigProxyFlow(test *framework.ClusterE2ETest, localCIDRs, kubeVersion string) {
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()
	// for testing proxy feature in an air gapped env
	// we have to download hook images on the admin host
	brContent, err := os.ReadFile(bundleReleasePathFromArtifacts)
	if err != nil {
		test.T.Fatalf("Cannot read bundleRelease file: %v", err)
	}

	server := downloadAndServeTinkerbellArtifacts(test.T, brContent, kubeVersion)
	defer server.Shutdown(context.Background())

	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOffHardware()
	test.AirgapDockerContainers(localCIDRs)
	test.CreateAirgappedUser(localCIDRs)
	test.AssertAirgappedNetwork()
	test.CreateCluster(
		// airgap user should be airgapped through iptables
		framework.WithSudo(airgapUsername),
		framework.WithBundlesOverride(bundleReleasePathFromArtifacts), // generated by ExtractDownloadArtifacts
		framework.WithControlPlaneWaitTimeout("30m"))
	test.StopIfFailed()
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}

func downloadFile(url string, output string) error {
	out, err := os.Create(output)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func downloadAndServeTinkerbellArtifacts(t framework.T, bundleRelease []byte, kubeVersion string) *http.Server {
	initramfsUrl := regexp.MustCompile(`https://.*/hook/.*initramfs-x86_64`).Find(bundleRelease)
	if initramfsUrl == nil {
		t.Fatalf("Cannot find initramfsUrl from release bundle")
	}

	vmlinuzUrl := regexp.MustCompile(`https://.*/hook/.*vmlinuz-x86_64`).Find(bundleRelease)
	if vmlinuzUrl == nil {
		t.Fatalf("Cannot find vmlinuzUrl from release bundle")
	}

	brOsUrl := regexp.MustCompile(fmt.Sprintf("https://.*/raw/%s/.*bottlerocket.*amd64.img.gz", kubeVersion)).Find(bundleRelease)
	if brOsUrl == nil {
		t.Fatalf("Cannot find bottlerocketOS url from release bundle")
	}

	dir, err := os.MkdirTemp("", "tinkerbell_artifacts_")
	if err != nil {
		t.Fatalf("Cannot create temporary directory to serve Tinkerbell artifacts %v", err)
	}
	t.Logf("Created directory for holding local tinkerbell artifacts: %s", dir)

	t.Logf("Download Initramfs from: %s", initramfsUrl)
	err = downloadFile(string(initramfsUrl), dir+"/initramfs-x86_64")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Download vmlinuz from %s", vmlinuzUrl)
	err = downloadFile(string(vmlinuzUrl), dir+"/vmlinuz-x86_64")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Download bottlerocket OS from %s", brOsUrl)
	// Save image file with kube version in the image name to satisfy condition to have kube version in the template name.
	err = downloadFile(string(brOsUrl), dir+"/"+bottlerocketOSFileName+"-"+kubeVersion+".img.gz")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Downloaded Bottlerocket OS")

	server := &http.Server{Addr: ":8080", Handler: http.FileServer(http.Dir(dir))}
	go func() {
		t.Log("Start local file server at :8080")
		server.ListenAndServe()
		t.Log("Local file server is shutdown")
		err = os.RemoveAll(dir)
		if err != nil {
			t.Logf("Temporary tinkerbell artifacts cannot be cleaned: %v", err)
		} else {
			t.Log("Temporary tinkerbell artifacts have been cleaned")
		}
	}()

	return server
}

func runDockerAirgapConfigFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()
	test.DownloadImages()
	test.ChangeInstanceSecurityGroup(os.Getenv(framework.RegistryMirrorAirgappedSecurityGroup))
	test.ImportImages()
	test.CreateCluster(framework.WithBundlesOverride(bundleReleasePathFromArtifacts))
	test.DeleteCluster(framework.WithBundlesOverride(bundleReleasePathFromArtifacts))
	test.ChangeInstanceSecurityGroup(os.Getenv(framework.RegistryMirrorDefaultSecurityGroup))
}

func runDockerAirgapUpgradeFromReleaseFlow(test *framework.ClusterE2ETest, latestRelease *releasev1.EksARelease, wantVersion anywherev1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))

	// Downloading and importing the artifacts from the previous version
	test.DownloadArtifacts(framework.ExecuteWithEksaRelease(latestRelease))
	test.ExtractDownloadedArtifacts(framework.ExecuteWithEksaRelease(latestRelease))
	test.DownloadImages(framework.ExecuteWithEksaRelease(latestRelease))
	test.ImportImages(framework.ExecuteWithEksaRelease(latestRelease))
	test.CreateCluster(framework.ExecuteWithEksaRelease(latestRelease), framework.WithBundlesOverride(bundleReleasePathFromArtifacts))

	// Adding this manual wait because old versions of the cli don't wait long enough
	// after creation, which makes the upgrade preflight validations fail
	test.WaitForControlPlaneReady()

	// Downloading and importing the artifacts from the current version
	test.DownloadArtifacts()
	test.ExtractDownloadedArtifacts()
	test.DownloadImages()
	test.ChangeInstanceSecurityGroup(os.Getenv(framework.RegistryMirrorAirgappedSecurityGroup))
	test.ImportImages()

	test.UpgradeClusterWithNewConfig(clusterOpts, framework.WithBundlesOverride(bundleReleasePathFromArtifacts))
	test.ValidateCluster(wantVersion)
	test.StopIfFailed()
	test.DeleteCluster(framework.WithBundlesOverride(bundleReleasePathFromArtifacts))
}
