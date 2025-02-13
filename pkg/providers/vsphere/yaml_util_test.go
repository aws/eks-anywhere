package vsphere

import (
	"testing"

	gomega "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

func TestYamlProcessorSuccessParsing(t *testing.T) {
	yaml := []byte(`apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: VSphereFailureDomain
metadata:
  name: fd-az1
spec:
  region:
    name: datacenter
    type: Datacenter
    tagCategory: k8s-region
  zone:
    name: az1
    type: ComputeCluster
    tagCategory: k8s-zone
  topology:
    datacenter: myDatacenter
    computeCluster: myCluster
    datastore: myDatastore
    networks:
    - myNetwork
---
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: VSphereDeploymentZone
metadata:
  name: dz-az1
  labels:
    region: "network-poc"
spec:
  server: myServer
  failureDomain: fd-az1
  placementConstraint:
    resourcePool: myResourcepool
    folder: myFolder`)
	processor := NewFailureDomainsYamlProcessor(test.NewNullLogger())
	failureDomains, err := processor.ProcessYAML(yaml)
	assert.Nil(t, err)
	assert.Equal(t, "dz-az1", failureDomains.Groups[0].VsphereDeploymentZone.Name)
}

func TestYamlProcessorFailedParsing(t *testing.T) {
	yaml := []byte(`invalid
	invalid
`)
	processor := NewFailureDomainsYamlProcessor(test.NewNullLogger())
	_, err := processor.ProcessYAML(yaml)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parsing vsphere failure domains yaml")
}

func TestRegisterWorkerMappingsError(t *testing.T) {
	g := gomega.NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	g.Expect(
		parser.RegisterMapping("VSphereDeploymentZone", func() yamlutil.APIObject { return nil }),
	).To(gomega.Succeed())
	yamlProcessor := &FailureDomainsYamlProcessor{}
	yamlProcessor.parser = parser
	_, err := yamlProcessor.ProcessYAML([]byte{})
	assert.Contains(t, err.Error(), "registering failure domain mappings")
}
