package nodeupgrader_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/nodeupgrader"
)

const (
	nodeName          = "my-node"
	upgraderImage     = "public.ecr.aws/eks-anywhere/node-upgrader:latest"
	kubernetesVersion = "v1.28.3-eks-1-28-9"
	etcdVersion       = "v3.5.9-eks-1-28-9"
)

func TestUpgradeFirstControlPlanePod(t *testing.T) {
	g := NewWithT(t)
	pod := nodeupgrader.UpgradeFirstControlPlanePod(nodeName, upgraderImage, kubernetesVersion, etcdVersion)
	g.Expect(pod).ToNot(BeNil())

	data, err := yaml.Marshal(pod)
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(data), "testdata/expected_first_control_plane_upgrader_pod.yaml")
}

func TestUpgradeSecondaryControlPlanePod(t *testing.T) {
	g := NewWithT(t)
	pod := nodeupgrader.UpgradeSecondaryControlPlanePod(nodeName, upgraderImage)
	g.Expect(pod).ToNot(BeNil())

	data, err := yaml.Marshal(pod)
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(data), "testdata/expected_rest_control_plane_upgrader_pod.yaml")
}

func TestUpgradeWorkerPod(t *testing.T) {
	g := NewWithT(t)
	pod := nodeupgrader.UpgradeWorkerPod(nodeName, upgraderImage)
	g.Expect(pod).ToNot(BeNil())

	data, err := yaml.Marshal(pod)
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(data), "testdata/expected_worker_upgrader_pod.yaml")
}
