package framework

import (
	"context"
	"fmt"
	"strings"

	versionutil "k8s.io/apimachinery/pkg/util/version"

	"github.com/aws/eks-anywhere/internal/pkg/conformance"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/version"
)

const (
	kubeConformanceImage                  = "registry.k8s.io/conformance"
	minKubernetesVersionRequiringTestSkip = "v1.29.0"
	skippedTestName                       = "Services should serve endpoints on same port and different protocols"
)

func (e *ClusterE2ETest) RunConformanceTests() {
	ctx := context.Background()
	cluster := e.Cluster()
	setKubeconfigEnvVar(e.T, e.ClusterName)
	contextName, err := e.KubectlClient.GetCurrentClusterContext(ctx, cluster)
	if err != nil {
		e.T.Errorf("Error getting context name: %v", err)
		return
	}
	kubeVersion, err := e.getEksdReleaseKubeVersion()
	if err != nil {
		e.T.Errorf("Error getting EKS-D release KubeVersion from bundle: %v", err)
		return
	}
	kubeVersionSemver, err := versionutil.ParseSemantic(kubeVersion)
	if err != nil {
		e.T.Errorf("Error getting semver for Kubernetes version %s: %v", kubeVersion, err)
		return
	}
	k8s129Compare, err := kubeVersionSemver.Compare(minKubernetesVersionRequiringTestSkip)
	if err != nil {
		e.T.Errorf("Error comparing cluster Kubernetes version with %s", minKubernetesVersionRequiringTestSkip)
		return
	}

	e.T.Log("Downloading Sonobuoy binary for testing")
	err = conformance.Download()
	if err != nil {
		e.T.Errorf("Error downloading Sonobuoy binary: %v", err)
		return
	}
	kubeConformanceImageTagged := fmt.Sprintf("%s:%s", kubeConformanceImage, kubeVersion)
	args := []string{"--kube-conformance-image", kubeConformanceImageTagged}
	// If running conformance tests for Kubernetes 1.29 or above, skip this particular test
	// because it will not pass with our deployment of Cilium.
	// References:
	// 1. https://github.com/kubernetes/kubernetes/pull/120069
	// 2. https://github.com/cilium/cilium/issues/29913
	// 3. https://github.com/cncf/k8s-conformance/pull/3049
	if k8s129Compare != -1 {
		args = append(args, fmt.Sprintf("--e2e-skip='%s'", skippedTestName))
	} else {
		// Only mode or --e2e-skip can be used at a time. Because we are using --e2e-skip
		// for k8s 1.29 and higher, we need to skip --mode=certified-conformance for those versions.
		// Once we stop skipping e2e, we can add the mode back to k8s 1.29 and higher.
		args = append(args, "--mode=certified-conformance")
	}
	e.T.Logf("Running k8s conformance tests with Image: %s", kubeConformanceImageTagged)
	output, err := conformance.RunTests(ctx, contextName, args...)
	if err != nil {
		e.T.Errorf("Error running k8s conformance tests: %v", err)
		return
	}
	e.T.Logf("Conformance Test run:\n %v", output)

	results, err := conformance.GetResults(ctx, contextName, args...)
	if err != nil {
		e.T.Errorf("Error running k8s conformance tests: %v", err)
		return
	}
	e.T.Logf("Conformance Test results:\n %v", results)
	if hasFailed(results) {
		e.T.Errorf("Conformance run has failed tests")
		return
	}
}

func (e *ClusterE2ETest) getEksdReleaseKubeVersion() (string, error) {
	c, err := v1alpha1.GetClusterConfig(e.ClusterConfigLocation)
	if err != nil {
		return "", fmt.Errorf("fetching cluster config from file: %v", err)
	}
	r := manifests.NewReader(newFileReader())
	b, err := r.ReadBundlesForVersion(version.Get().GitVersion)
	if err != nil {
		return "", fmt.Errorf("getting EKS-D release spec from bundle: %v", err)
	}
	versionsBundle := bundles.VersionsBundleForKubernetesVersion(b, string(c.Spec.KubernetesVersion))
	versionsBundleKubeVersion := versionsBundle.EksD.KubeVersion
	if versionsBundleKubeVersion == "" {
		return "", fmt.Errorf("getting KubeVersion from EKS-D release spec: value empty")
	}

	return versionsBundleKubeVersion, nil
}

// Function to parse the conformace test results and look for any failed tests.
// By default we run 2 plugins so we check for failed tests in twice.
func hasFailed(results string) bool {
	failedLog := "Failed: 0"
	count := strings.Count(results, failedLog)
	return count != 2
}
