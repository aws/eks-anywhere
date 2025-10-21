package cloudstack_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestCloudStackTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionEnabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(true)

	builder := cloudstack.NewTemplateBuilder(test.FakeNow)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = "test-cp-template"
		values["etcdTemplateName"] = "test-etcd-template"
	})
	g.Expect(err).ToNot(HaveOccurred())

	test.AssertYAMLSubset(t, data, "testdata/admission_exclusion_enabled.yaml")
}

func TestCloudStackTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionDisabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(false)

	builder := cloudstack.NewTemplateBuilder(test.FakeNow)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = "test-cp-template"
		values["etcdTemplateName"] = "test-etcd-template"
	})
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

func TestCloudStackTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionNil(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_multiple_worker_node_groups.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = nil

	builder := cloudstack.NewTemplateBuilder(test.FakeNow)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = "test-cp-template"
		values["etcdTemplateName"] = "test-etcd-template"
	})
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
