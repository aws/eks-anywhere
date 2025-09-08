package vsphere_test

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionEnabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(true)

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
	})
	g.Expect(err).ToNot(HaveOccurred())

	test.AssertYAMLSubset(t, data, "testdata/admission_exclusion_enabled.yaml")
}

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionDisabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(false)

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
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

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionNil(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = nil

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
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

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionEnabledBottlerocket(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_br.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(true)

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
	})
	g.Expect(err).ToNot(HaveOccurred())

	test.AssertYAMLSubset(t, data, "testdata/admission_exclusion_enabled_br.yaml")
}
