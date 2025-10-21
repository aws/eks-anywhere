package tinkerbell

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestTinkerbellTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionEnabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_stacked_etcd.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(true)

	cpMachineCfg, err := getControlPlaneMachineSpec(spec)
	g.Expect(err).ToNot(HaveOccurred())

	wngMachineCfgs, err := getWorkerNodeGroupMachineSpec(spec)
	g.Expect(err).ToNot(HaveOccurred())

	builder := NewTemplateBuilder(&spec.TinkerbellDatacenter.Spec, cpMachineCfg, nil, wngMachineCfgs, "0.0.0.0", test.FakeNow)
	data, err := builder.GenerateCAPISpecControlPlane(spec)
	g.Expect(err).ToNot(HaveOccurred())

	test.AssertYAMLSubset(t, data, "testdata/admission_exclusion_enabled.yaml")
}

func TestTinkerbellTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionDisabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_stacked_etcd.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(false)

	cpMachineCfg, err := getControlPlaneMachineSpec(spec)
	g.Expect(err).ToNot(HaveOccurred())

	wngMachineCfgs, err := getWorkerNodeGroupMachineSpec(spec)
	g.Expect(err).ToNot(HaveOccurred())

	builder := NewTemplateBuilder(&spec.TinkerbellDatacenter.Spec, cpMachineCfg, nil, wngMachineCfgs, "0.0.0.0", test.FakeNow)
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

func TestTinkerbellTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionNil(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_tinkerbell_stacked_etcd.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = nil

	cpMachineCfg, err := getControlPlaneMachineSpec(spec)
	g.Expect(err).ToNot(HaveOccurred())

	wngMachineCfgs, err := getWorkerNodeGroupMachineSpec(spec)
	g.Expect(err).ToNot(HaveOccurred())

	builder := NewTemplateBuilder(&spec.TinkerbellDatacenter.Spec, cpMachineCfg, nil, wngMachineCfgs, "0.0.0.0", test.FakeNow)
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
