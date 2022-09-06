package cluster_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestParseConfigMissingVSphereDatacenter(t *testing.T) {
	g := NewWithT(t)
	got, err := cluster.ParseConfigFromFile("testdata/cluster_vsphere_missing_datacenter.yaml")

	g.Expect(err).To(Not(HaveOccurred()))

	g.Expect(got.VSphereDatacenter).To(BeNil())
}
