package framework

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/internal/pkg/conformance"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/version"
)

const kubeConformanceImage = "k8s.gcr.io/conformance"

func (e *ClusterE2ETest) RunConformanceTests() {
	ctx := context.Background()
	cluster := e.cluster()
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
	e.T.Log("Downloading Sonobuoy binary for testing")
	err = conformance.Download()
	if err != nil {
		e.T.Errorf("Error downloading Sonobuoy binary: %v", err)
		return
	}
	kubeConformanceImageTagged := fmt.Sprintf("%s:%s", kubeConformanceImage, kubeVersion)
	args := []string{"--kube-conformance-image", kubeConformanceImageTagged}
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
	eksdRelease, _, err := cluster.GetEksdRelease(version.Get(), c)
	if err != nil {
		return "", fmt.Errorf("getting EKS-D release spec from bundle: %v", err)
	}
	if kubeVersion := eksdRelease.KubeVersion; kubeVersion != "" {
		return kubeVersion, nil
	}
	return "", fmt.Errorf("getting KubeVersion from EKS-D release spec: value empty")
}

// Function to parse the conformace test results and look for any failed tests.
// By default we run 2 plugins so we check for failed tests in twice.
func hasFailed(results string) bool {
	failedLog := "Failed: 0"
	count := strings.Count(results, failedLog)
	return count != 2
}
