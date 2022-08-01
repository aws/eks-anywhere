package diagnostics_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/diagnostics"
)

func TestCloudStackDataCenterConfigCollectors(t *testing.T) {
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.CloudStackDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory()
	collectors := factory.DataCenterConfigCollectors(datacenter)
	assert.Equal(t, len(collectors), 10, "DataCenterConfigCollectors() mismatch between desired collectors and actual")
	assert.Equal(t, constants.CapcSystemNamespace, collectors[0].Logs.Namespace)
	assert.Equal(t, fmt.Sprintf("logs/%s", constants.CapcSystemNamespace), collectors[0].Logs.Name)
	for _, collector := range collectors[1:] {
		assert.Equal(t, []string{"kubectl"}, collector.Run.PodSpec.Containers[0].Command)
		assert.Equal(t, "eksa-diagnostics", collector.Run.Namespace)
	}
}

func TestTinkerbellDataCenterConfigCollectors(t *testing.T) {
	datacenter := eksav1alpha1.Ref{Kind: eksav1alpha1.TinkerbellDatacenterKind}
	factory := diagnostics.NewDefaultCollectorFactory()
	collectors := factory.DataCenterConfigCollectors(datacenter)
	assert.Equal(t, len(collectors), 13, "DataCenterConfigCollectors() mismatch between desired collectors and actual")
	assert.Equal(t, constants.CaptSystemNamespace, collectors[0].Logs.Namespace)
	assert.Equal(t, fmt.Sprintf("logs/%s", constants.CaptSystemNamespace), collectors[0].Logs.Name)
	for _, collector := range collectors[1:] {
		assert.Equal(t, []string{"kubectl"}, collector.Run.PodSpec.Containers[0].Command)
		assert.Equal(t, "eksa-diagnostics", collector.Run.Namespace)
	}
}

func TestAPIServerCollectors(t *testing.T) {
	controlPlaneIP := "1.1.1.1"
	factory := diagnostics.NewDefaultCollectorFactory()
	collectors := factory.APIServerCollectors(controlPlaneIP)
	assert.Equal(t, len(collectors), 2, "APIServerCollector() should return two collectors:hostPortCollector and pingHostCollector ")
	assert.Equal(t, collectors[0].RunPod.Name, "run-ip", "First of APIserverCollectors should be type of runPod named run-ip")
	assert.Equal(t, collectors[1].RunPod.Name, "run-ping", "Second of APIserverCollectors should be type of runPod named run-ping")
}
