package docker_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestDockerTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionEnabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_api_server_cert_san_ip.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(true)

	builder := docker.NewDockerTemplateBuilder(test.FakeNow)
	data, err := builder.GenerateCAPISpecControlPlane(spec)
	g.Expect(err).ToNot(HaveOccurred())

	test.AssertYAMLSubset(t, data, "testdata/admission_exclusion_enabled.yaml")
}

func TestDockerTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionDisabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_api_server_cert_san_ip.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(false)

	builder := docker.NewDockerTemplateBuilder(test.FakeNow)
	data, err := builder.GenerateCAPISpecControlPlane(spec)
	g.Expect(err).ToNot(HaveOccurred())

	objects, err := test.ParseMultiDocYAML(data)
	g.Expect(err).ToNot(HaveOccurred())

	kcp, err := test.FindObjectByKind(objects, "KubeadmControlPlane")
	g.Expect(err).ToNot(HaveOccurred())

	test.AssertNotContainsItemAtPath(t, kcp, "spec.kubeadmConfigSpec.clusterConfiguration.apiServer.extraEnvs",
		map[string]interface{}{"name": "EKS_PATCH_EXCLUSION_RULES_FILE"})

	test.AssertNotContainsItemAtPath(t, kcp, "spec.kubeadmConfigSpec.clusterConfiguration.apiServer.extraVolumes",
		map[string]interface{}{"name": "admission-exclusion-rules"})

	test.AssertNotContainsItemAtPath(t, kcp, "spec.kubeadmConfigSpec.files",
		map[string]interface{}{"path": "/etc/kubernetes/admission-plugin-exclusion-rules.json"})
}

func TestDockerTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionNil(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_api_server_cert_san_ip.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = nil

	builder := docker.NewDockerTemplateBuilder(test.FakeNow)
	data, err := builder.GenerateCAPISpecControlPlane(spec)
	g.Expect(err).ToNot(HaveOccurred())

	objects, err := test.ParseMultiDocYAML(data)
	g.Expect(err).ToNot(HaveOccurred())

	kcp, err := test.FindObjectByKind(objects, "KubeadmControlPlane")
	g.Expect(err).ToNot(HaveOccurred())

	test.AssertNotContainsItemAtPath(t, kcp, "spec.kubeadmConfigSpec.clusterConfiguration.apiServer.extraEnvs",
		map[string]interface{}{"name": "EKS_PATCH_EXCLUSION_RULES_FILE"})

	test.AssertNotContainsItemAtPath(t, kcp, "spec.kubeadmConfigSpec.clusterConfiguration.apiServer.extraVolumes",
		map[string]interface{}{"name": "admission-exclusion-rules"})

	test.AssertNotContainsItemAtPath(t, kcp, "spec.kubeadmConfigSpec.files",
		map[string]interface{}{"path": "/etc/kubernetes/admission-plugin-exclusion-rules.json"})
}
