package framework

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/internal/pkg/conformance"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/version"
)

const kubeConformanceImage = "k8s.gcr.io/conformance"

func (e *E2ETest) RunConformanceTests() {
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
	results, err := conformance.RunTests(ctx, contextName, args...)
	if err != nil {
		e.T.Errorf("Error running k8s conformance tests: %v", err)
		return
	}
	e.T.Logf("Conformance Test results:\n %v", results)
}

func (e *E2ETest) getEksdReleaseKubeVersion() (string, error) {
	c, err := v1alpha1.GetClusterConfig(e.ClusterConfigLocation)
	if err != nil {
		return "", fmt.Errorf("error fetching cluster config from file: %v", err)
	}
	eksdRelease, err := cluster.GetEksdRelease(version.Get(), c)
	if err != nil {
		return "", fmt.Errorf("error getting EKS-D release spec from bundle: %v", err)
	}
	if kubeVersion := eksdRelease.KubeVersion; kubeVersion != "" {
		return kubeVersion, nil
	}
	return "", fmt.Errorf("error getting KubeVersion from EKS-D release spec: value empty")
}
