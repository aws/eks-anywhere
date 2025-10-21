package nutanix

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestNutanixTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionEnabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(true)

	creds := GetCredsFromEnv()
	cpMachineSpec := spec.NutanixMachineConfigs[spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	workerSpecs := make(map[string]anywherev1.NutanixMachineConfigSpec)
	for _, wng := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerSpecs[wng.Name] = spec.NutanixMachineConfigs[wng.MachineGroupRef.Name].Spec
	}
	var etcdMachineSpec *anywherev1.NutanixMachineConfigSpec
	if spec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdSpec := spec.NutanixMachineConfigs[spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		etcdMachineSpec = &etcdSpec
	}
	builder := NewNutanixTemplateBuilder(&spec.NutanixDatacenter.Spec, &cpMachineSpec, etcdMachineSpec, workerSpecs, creds, test.FakeNow)
	data, err := builder.GenerateCAPISpecControlPlane(spec)
	g.Expect(err).ToNot(HaveOccurred())

	test.AssertYAMLSubset(t, data, "testdata/admission_exclusion_enabled.yaml")
}

func TestNutanixTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionDisabled(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = ptr.Bool(false)

	creds := GetCredsFromEnv()
	cpMachineSpec := spec.NutanixMachineConfigs[spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	workerSpecs := make(map[string]anywherev1.NutanixMachineConfigSpec)
	for _, wng := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerSpecs[wng.Name] = spec.NutanixMachineConfigs[wng.MachineGroupRef.Name].Spec
	}
	var etcdMachineSpec *anywherev1.NutanixMachineConfigSpec
	if spec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdSpec := spec.NutanixMachineConfigs[spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		etcdMachineSpec = &etcdSpec
	}
	builder := NewNutanixTemplateBuilder(&spec.NutanixDatacenter.Spec, &cpMachineSpec, etcdMachineSpec, workerSpecs, creds, test.FakeNow)
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

func TestNutanixTemplateBuilderGenerateCAPISpecControlPlaneWithAdmissionExclusionNil(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.SkipAdmissionForSystemResources = nil

	creds := GetCredsFromEnv()
	cpMachineSpec := spec.NutanixMachineConfigs[spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name].Spec
	workerSpecs := make(map[string]anywherev1.NutanixMachineConfigSpec)
	for _, wng := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerSpecs[wng.Name] = spec.NutanixMachineConfigs[wng.MachineGroupRef.Name].Spec
	}
	var etcdMachineSpec *anywherev1.NutanixMachineConfigSpec
	if spec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdSpec := spec.NutanixMachineConfigs[spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name].Spec
		etcdMachineSpec = &etcdSpec
	}
	builder := NewNutanixTemplateBuilder(&spec.NutanixDatacenter.Spec, &cpMachineSpec, etcdMachineSpec, workerSpecs, creds, test.FakeNow)
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
