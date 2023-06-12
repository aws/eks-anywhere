package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestClusterDefault(t *testing.T) {
	cOld := &v1alpha1.Cluster{}
	cOld.SetSelfManaged()
	cOld.Spec.RegistryMirrorConfiguration = &v1alpha1.RegistryMirrorConfiguration{
		Port: "",
	}
	cOld.Default()
	g := NewWithT(t)
	g.Expect(cOld.Spec.ClusterNetwork.CNIConfig).To(Equal(&v1alpha1.CNIConfig{}))
	g.Expect(cOld.Spec.RegistryMirrorConfiguration.Port).To(Equal(constants.DefaultHttpsPort))
}

func TestClusterValidateUpdateManagementValueMutableExperimental(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "true")
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration.Labels = map[string]string{"Key1": "Val1"}
	cOld.Spec.ControlPlaneConfiguration.Taints = []v1.Taint{
		{
			Key:    "Key1",
			Value:  "Val1",
			Effect: "PreferNoSchedule",
		},
	}
	cOld.SetSelfManaged()
	cNew := cOld.DeepCopy()
	cNew.Spec.ControlPlaneConfiguration.Labels = map[string]string{"Key2": "Val2"}
	cNew.Spec.ControlPlaneConfiguration.Taints = []v1.Taint{
		{
			Key:    "Key2",
			Value:  "Val2",
			Effect: "PreferNoSchedule",
		},
	}
	cNew.Spec.ControlPlaneConfiguration.Count = 1
	cNew.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "test"

	g := NewWithT(t)
	g.Expect(cNew.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateManagementValueImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.SetManagedBy("management-cluster")

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("field is immutable")))
}

func TestClusterValidateUpdateManagementOldNilNewTrueSuccess(t *testing.T) {
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.SetSelfManaged()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateManagementOldNilNewFalseImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.SetManagedBy("")
	c := cOld.DeepCopy()
	c.SetManagedBy("management-cluster")

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("field is immutable")))
}

func TestClusterValidateUpdateManagementBothNilImmutable(t *testing.T) {
	cOld := baseCluster()
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestManagementClusterValidateUpdateKubernetesVersionImmutable(t *testing.T) {
	features.ClearCache()
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion:         v1alpha1.Kube119,
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
		},
	}
	cOld.SetSelfManaged()
	c := cOld.DeepCopy()
	c.Spec.KubernetesVersion = v1alpha1.Kube120

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("field is immutable 1.20")))
}

func TestManagementNilClusterValidateUpdateKubernetesVersionImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion:         v1alpha1.Kube122,
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.KubernetesVersion = v1alpha1.Kube120

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("field is immutable 1.20")))
}

func TestWorkloadClusterValidateUpdateKubernetesVersionSuccess(t *testing.T) {
	cOld := baseCluster()
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.KubernetesVersion = v1alpha1.Kube122

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestWorkloadClusterValidateUpdateNoUpdateSuccess(t *testing.T) {
	cOld := baseCluster()
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationEqual(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count:           3,
		Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count:           3,
		Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestWorkloadClusterValidateUpdateControlPlaneConfigurationEqual(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count:           3,
		Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count:           3,
		Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateControlPlaneConfigurationImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count:           3,
		Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count:           10,
		Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/2"},
		MachineGroupRef: &v1alpha1.Ref{Name: "test2", Kind: "SecondMachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration.endpoint: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateControlPlaneConfigurationOldEndpointImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
	}
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/2"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration.endpoint: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateControlPlaneConfigurationOldEndpointNilImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: nil,
	}
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration.endpoint: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateControlPlaneConfigurationNewEndpointNilImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
	}
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: nil,
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration.endpoint: Forbidden: field is immutable")))
}

func TestCloudStackClusterValidateUpdateControlPlaneConfigurationOldDefaultPortNewNoPort(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "1.1.1.1:6443"}
	cOld.Spec.DatacenterRef.Kind = v1alpha1.CloudStackDatacenterKind
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "1.1.1.1"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestCloudStackClusterValidateUpdateControlPlaneConfigurationOldNoPortNewDefaultPort(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "1.1.1.1"}
	cOld.Spec.DatacenterRef.Kind = v1alpha1.CloudStackDatacenterKind
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "1.1.1.1:6443"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestCloudStackClusterValidateUpdateControlPlaneConfigurationOldPortImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "1.1.1.1"}
	cOld.Spec.DatacenterRef.Kind = v1alpha1.CloudStackDatacenterKind
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "1.1.1.2"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration.endpoint: Forbidden: field is immutable")))
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationTaintsImmutable(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "false")
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Taints: []v1.Taint{
			{
				Key:    "Key1",
				Value:  "Val1",
				Effect: "PreferNoSchedule",
			},
		},
	}
	cOld.SetSelfManaged()
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Taints: []v1.Taint{
			{
				Key:    "Key2",
				Value:  "Val2",
				Effect: "PreferNoSchedule",
			},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration: Forbidden: field is immutable")))
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationLabelsImmutable(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "false")
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Labels: map[string]string{
			"Key1": "Val1",
		},
	}
	cOld.SetSelfManaged()
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Labels: map[string]string{
			"Key2": "Val2",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration: Forbidden: field is immutable")))
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationOldMachineGroupRefImmutable(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "false")
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test1", Kind: "MachineConfig"},
	}
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test2", Kind: "MachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration: Forbidden: field is immutable")))
}

func TestWorkloadClusterValidateUpdateControlPlaneConfigurationMachineGroupRef(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "test1", Kind: "MachineConfig"}
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "test2", Kind: "MachineConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationOldMachineGroupRefNilImmutable(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "false")
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: nil,
	}
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration: Forbidden: field is immutable")))
}

func TestWorkloadClusterValidateUpdateControlPlaneConfigurationOldMachineGroupRefNilSuccess(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration.MachineGroupRef = nil
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationNewMachineGroupRefNilImmutable(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "false")
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: nil,
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration: Forbidden: field is immutable")))
}

func TestWorkloadClusterValidateUpdateControlPlaneConfigurationNewMachineGroupRefChangedSuccess(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"}
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.MachineGroupRef = &v1alpha1.Ref{Name: "test-2", Kind: "MachineConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestWorkloadClusterValidateUpdateControlPlaneConfigurationNewMachineGroupRefNilError(t *testing.T) {
	cOld := baseCluster()
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.MachineGroupRef = nil

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("must specify machineGroupRef control plane machines")))
}

func TestWorkloadClusterValidateUpdateWorkerNodeConfigurationNewMachineGroupRefNilError(t *testing.T) {
	cOld := baseCluster()
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef = nil

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("must specify machineGroupRef for worker nodes")))
}

func TestWorkloadClusterValidateUpdateExternalEtcdConfigurationNewMachineGroupRefNilError(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
		Count:           3,
	}
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{
		MachineGroupRef: nil,
		Count:           3,
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("must specify machineGroupRef for etcd machines")))
}

func TestClusterValidateUpdateDatacenterRefImmutableEqual(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.DatacenterRef = v1alpha1.Ref{
		Name: "test", Kind: "DatacenterConfig",
	}
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateDatacenterRefImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.DatacenterRef = v1alpha1.Ref{
		Name: "test", Kind: "DatacenterConfig",
	}
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef = v1alpha1.Ref{Name: "test2", Kind: "SecondDatacenterConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.datacenterRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateDatacenterRefImmutableName(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.DatacenterRef = v1alpha1.Ref{
		Name: "test", Kind: "DatacenterConfig",
	}
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef = v1alpha1.Ref{Name: "test2", Kind: "DatacenterConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.datacenterRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateDatacenterRefNilImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.DatacenterRef = v1alpha1.Ref{
		Name: "test", Kind: "DatacenterConfig",
	}
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef = v1alpha1.Ref{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.datacenterRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateExternalEtcdReplicasImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
	cOld.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count:    3,
		Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
	}
	c := cOld.DeepCopy()
	c.Spec.ExternalEtcdConfiguration.Count = 5

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.externalEtcdConfiguration.count: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateDataCenterRefNameImmutable(t *testing.T) {
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef.Name = "FancyNewDataCenter"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.datacenterRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateDataCenterRefKindImmutable(t *testing.T) {
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef.Name = v1alpha1.DockerDatacenterKind

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.datacenterRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateClusterNetworkPodsImmutable(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"1.2.3.4/5"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.clusterNetwork.pods: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateClusterNetworkServicesImmutable(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork.Services.CidrBlocks = []string{"1.2.3.4/9", "1.2.3.4/10"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.clusterNetwork.services: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateClusterNetworkDNSImmutable(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
		DNS: v1alpha1.DNS{
			ResolvConf: &v1alpha1.ResolvConf{
				Path: "my-path",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork.DNS.ResolvConf.Path = "other-path"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.clusterNetwork.dns: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateClusterNetworkNodesImmutable(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork.Nodes = &v1alpha1.Nodes{
		CIDRMaskSize: ptr.Int(10),
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.clusterNetwork.nodes: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateProxyConfigurationEqualOrder(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "http://test.com:1",
		HttpsProxy: "https://test.com:1",
		NoProxy: []string{
			"noproxy1",
			"noproxy2",
		},
	}

	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "http://test.com:1",
		HttpsProxy: "https://test.com:1",
		NoProxy: []string{
			"noproxy2",
			"noproxy1",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateProxyConfigurationImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "http://test.com",
		HttpsProxy: "https://test.com",
		NoProxy:    []string{"noproxy1"},
	}
	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "http://test.com",
		HttpsProxy: "https://test.com",
		NoProxy:    []string{"noproxy1", "noproxy2"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ProxyConfiguration: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateProxyConfigurationNoProxyImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "httpproxy",
		HttpsProxy: "httpsproxy",
	}
	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "httpproxy",
		HttpsProxy: "httpsproxy",
		NoProxy:    []string{"noproxy"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ProxyConfiguration: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateProxyConfigurationOldNilImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ProxyConfiguration = nil

	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "httpproxy",
		HttpsProxy: "httpsproxy",
		NoProxy:    []string{"noproxy"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ProxyConfiguration: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateProxyConfigurationNewNilImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "httpproxy",
		HttpsProxy: "httpsproxy",
		NoProxy:    []string{"noproxy"},
	}
	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = nil
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.ProxyConfiguration: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateGitOpsRefImmutableNilEqual(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.GitOpsRef = nil

	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateGitOpsRefImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.GitOpsRef = &v1alpha1.Ref{}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = &v1alpha1.Ref{Name: "test2", Kind: "GitOpsConfig2"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.GitOpsRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateGitOpsRefImmutableName(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.GitOpsRef = &v1alpha1.Ref{
		Name: "test1", Kind: "GitOpsConfig",
	}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = &v1alpha1.Ref{Name: "test2", Kind: "GitOpsConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.GitOpsRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateGitOpsRefImmutableKind(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.GitOpsRef = &v1alpha1.Ref{
		Name: "test", Kind: "GitOpsConfig1",
	}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = &v1alpha1.Ref{Name: "test", Kind: "GitOpsConfig2"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.GitOpsRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateGitOpsRefOldNilImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.GitOpsRef = nil

	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = &v1alpha1.Ref{Name: "test", Kind: "GitOpsConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.GitOpsRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateGitOpsRefNewNilImmutable(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.GitOpsRef = &v1alpha1.Ref{
		Name: "test", Kind: "GitOpsConfig",
	}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = nil

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.GitOpsRef: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateAWSIamNameImmutableUpdateSameName(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}
	c := baseCluster()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateAWSIamNameImmutableUpdateSameNameWorkloadCluster(t *testing.T) {
	cOld := baseCluster()
	cOld.SetManagedBy("mgmt2")
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}
	c := baseCluster()
	c.SetManagedBy("mgmt2")
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateAWSIamNameImmutableUpdateName(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}
	c := baseCluster()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name2",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.IdentityProviderRefs: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateAWSIamNameImmutableEmpty(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}
	c := baseCluster()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.IdentityProviderRefs: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateAWSIamNameImmutableAddConfig(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{}
	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.IdentityProviderRefs: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateUnsetBundlesRefImmutable(t *testing.T) {
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.Spec.BundlesRef = nil

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.BundlesRef: Invalid value: \"null\": field cannot be removed after setting")))
}

func TestClusterValidateUpdateOIDCNameMutableUpdateNameWorkloadCluster(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}
	cOld.SetManagedBy("management-cluster")
	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs[0].Name = "name2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableUpdateNameMgmtCluster(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}
	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs[0].Name = "name2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.IdentityProviderRefs: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateOIDCNameMutableUpdateNameUnchanged(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableWorkloadCluster(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}
	cOld.SetManagedBy("mgmt2")

	c := baseCluster()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{}

	c.SetManagedBy("mgmt2")
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableMgmtCluster(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}
	c := baseCluster()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.IdentityProviderRefs: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateOIDCNameMutableAddConfigWorkloadCluster(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{}

	cOld.SetManagedBy("mgmt2")

	c := baseCluster()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}
	c.SetManagedBy("mgmt2")

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableAddConfigMgmtCluster(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{}

	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.IdentityProviderRefs: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateSwapIdentityProviders(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}
	c := baseCluster()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateSwapIdentityProvidersWorkloadCluster(t *testing.T) {
	cOld := baseCluster()
	cOld.SetManagedBy("mgmt2")
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}
	c := baseCluster()
	c.SetManagedBy("mgmt2")
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
		{
			Kind: v1alpha1.OIDCConfigKind,
			Name: "name1",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateEmptyIdentityProviders(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.IdentityProviderRefs = []v1alpha1.Ref{}
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateGitOpsRefOldEmptyImmutable(t *testing.T) {
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: "identity",
			Name: "name",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("spec.IdentityProviderRefs: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateWithPausedAnnotation(t *testing.T) {
	cOld := baseCluster()
	cOld.ObjectMeta.Annotations = make(map[string]string, 1)
	cOld.PauseReconcile()
	c := cOld.DeepCopy()
	c.Spec.KubernetesVersion = v1alpha1.Kube122

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateInvalidType(t *testing.T) {
	cOld := &v1alpha1.VSphereDatacenterConfig{}
	c := &v1alpha1.Cluster{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("expected a Cluster but got a *v1alpha1.VSphereDatacenterConfig")))
}

func TestClusterValidateUpdateSuccess(t *testing.T) {
	features.ClearCache()
	workerConfiguration := append([]v1alpha1.WorkerNodeGroupConfiguration{}, v1alpha1.WorkerNodeGroupConfiguration{Count: ptr.Int(5), Name: "test", MachineGroupRef: &v1alpha1.Ref{Name: "ref-name"}})
	cOld := baseCluster()
	cOld.Spec.WorkerNodeGroupConfigurations = workerConfiguration
	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(10)

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterCreateManagementCluster(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	workerConfiguration := append([]v1alpha1.WorkerNodeGroupConfiguration{}, v1alpha1.WorkerNodeGroupConfiguration{Count: ptr.Int(5)})
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: workerConfiguration,
			KubernetesVersion:             v1alpha1.Kube119,
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
		},
	}

	g := NewWithT(t)
	g.Expect(cluster.ValidateCreate()).To(MatchError(ContainSubstring("creating new cluster on existing cluster is not supported for self managed clusters")))
}

func TestClusterCreateCloudStackMultipleWorkerNodeGroupsValidation(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cluster := baseCluster()
	cluster.Spec.WorkerNodeGroupConfigurations = append([]v1alpha1.WorkerNodeGroupConfiguration{},
		v1alpha1.WorkerNodeGroupConfiguration{Count: ptr.Int(5), Name: "test", MachineGroupRef: &v1alpha1.Ref{Name: "ref-name"}},
		v1alpha1.WorkerNodeGroupConfiguration{Count: ptr.Int(5), Name: "test2", MachineGroupRef: &v1alpha1.Ref{Name: "ref-name"}})

	cluster.Spec.ManagementCluster.Name = "management-cluster"

	g := NewWithT(t)
	g.Expect(cluster.ValidateCreate()).To(Succeed())
}

func TestClusterCreateWorkloadCluster(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cluster := baseCluster()
	cluster.Spec.WorkerNodeGroupConfigurations = append([]v1alpha1.WorkerNodeGroupConfiguration{},
		v1alpha1.WorkerNodeGroupConfiguration{
			Count: ptr.Int(5),
			Name:  "md-0",
			MachineGroupRef: &v1alpha1.Ref{
				Name: "test",
			},
		})
	cluster.Spec.KubernetesVersion = v1alpha1.Kube119
	cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"}, MachineGroupRef: &v1alpha1.Ref{Name: "test"},
	}

	cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test"}}
	cluster.Spec.ManagementCluster.Name = "management-cluster"

	g := NewWithT(t)
	g.Expect(cluster.ValidateCreate()).To(Succeed())
}

func TestClusterUpdateWorkerNodeGroupTaintsAndLabelsSuccess(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{
		Name: "test",
		Taints: []v1.Taint{{
			Key:    "test",
			Value:  "test",
			Effect: "PreferNoSchedule",
		}},
		Labels: map[string]string{
			"test": "val1",
		},
		Count: ptr.Int(1),
		MachineGroupRef: &v1alpha1.Ref{
			Name: "test",
		},
	}}

	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Taints[0].Value = "test2"
	c.Spec.WorkerNodeGroupConfigurations[0].Labels["test"] = "val2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterUpdateWorkerNodeGroupTaintsInvalid(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{
		Count: ptr.Int(1),
		Name:  "test",
		Taints: []v1.Taint{{
			Key:    "test",
			Value:  "test",
			Effect: "PreferNoSchedule",
		}},
		MachineGroupRef: &v1alpha1.Ref{},
	}}

	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Taints[0].Effect = "NoSchedule"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("at least one WorkerNodeGroupConfiguration must not have NoExecute and/or NoSchedule taints")))
}

func TestClusterUpdateWorkerNodeGroupNameInvalid(t *testing.T) {
	cOld := baseCluster()
	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Name = ""

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("must specify name for worker nodes")))
}

func TestClusterUpdateWorkerNodeGroupLabelsInvalid(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{
		Count: ptr.Int(1),
		Name:  "test",
		Labels: map[string]string{
			"test": "val1",
		},
		MachineGroupRef: &v1alpha1.Ref{},
	}}

	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Labels["test"] = "val1/val2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(MatchError(ContainSubstring("labels for worker node group test not valid: found following errors with labels: spec.workerNodeGroupConfigurations[0].labels: Invalid value:")))
}

func TestClusterUpdateControlPlaneTaintsAndLabelsSuccess(t *testing.T) {
	cOld := baseCluster()
	cOld.Spec.ControlPlaneConfiguration.Taints = []v1.Taint{{
		Key:    "test",
		Value:  "test",
		Effect: "PreferNoSchedule",
	}}
	cOld.Spec.ControlPlaneConfiguration.Labels = map[string]string{
		"test": "val1",
	}

	cOld.SetManagedBy("management-cluster")
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.Taints[0].Value = "test2"
	c.Spec.ControlPlaneConfiguration.Labels["test"] = "val2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterUpdateControlPlaneLabelsInvalid(t *testing.T) {
	cluster := baseCluster()
	cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Labels: map[string]string{
			"test": "val1",
		},
		Endpoint:        &v1alpha1.Endpoint{"1.1.1.1"},
		MachineGroupRef: &v1alpha1.Ref{},
		Count:           1,
	}
	cluster.SetManagedBy("management-cluster")
	c := cluster.DeepCopy()
	c.Spec.ControlPlaneConfiguration.Labels["test"] = "val1/val2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cluster)).To(MatchError(ContainSubstring("spec.controlPlaneConfiguration.labels: Invalid value")))
}

func TestClusterValidateCreateSelfManagedUnpaused(t *testing.T) {
	features.ClearCache()
	cluster := baseCluster()
	g := NewWithT(t)
	cluster.SetSelfManaged()
	err := cluster.ValidateCreate()
	g.Expect(err).To(MatchError(ContainSubstring("creating new cluster on existing cluster is not supported for self managed clusters")))
}

func TestClusterValidateCreateManagedUnpaused(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.FullLifecycleAPIEnvVar, "")
	cluster := baseCluster()
	g := NewWithT(t)
	cluster.SetManagedBy("mgmt2")
	err := cluster.ValidateCreate()
	g.Expect(err.Error()).To(ContainSubstring("creating new managed cluster on existing cluster is not supported"))
}

func TestClusterValidateCreateSelfManagedNotPaused(t *testing.T) {
	features.ClearCache()
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cluster := baseCluster()
	cluster.SetSelfManaged()

	g := NewWithT(t)
	err := cluster.ValidateCreate()
	g.Expect(err).To(MatchError(ContainSubstring("creating new cluster on existing cluster is not supported for self managed clusters")))
}

func TestClusterValidateCreateInvalidCluster(t *testing.T) {
	tests := []struct {
		name               string
		featureGateEnabled bool
		cluster            *v1alpha1.Cluster
	}{
		{
			name: "Paused self-managed cluster, feature gate off",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
				c.PauseReconcile()
			}),
			featureGateEnabled: false,
		},
		{
			name: "Paused workload cluster, feature gate off",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
				c.PauseReconcile()
			}),
			featureGateEnabled: false,
		},
		{
			name: "Paused self-managed cluster, feature gate on",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
				c.PauseReconcile()
			}),
			featureGateEnabled: true,
		},
		{
			name: "Paused workload cluster, feature gate on",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
				c.PauseReconcile()
			}),
			featureGateEnabled: true,
		},
		{
			name: "Unpaused workload cluster, feature gate on",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
			}),
			featureGateEnabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features.ClearCache()
			if tt.featureGateEnabled {
				t.Setenv(features.FullLifecycleAPIEnvVar, "true")
			}

			// Invalid control plane configuration
			tt.cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{Endpoint: &v1alpha1.Endpoint{Host: "test-ip"}, MachineGroupRef: &v1alpha1.Ref{Name: "test"}}

			g := NewWithT(t)
			err := tt.cluster.ValidateCreate()
			g.Expect(err).To(MatchError(ContainSubstring("control plane node count must be positive")))
		})
	}
}

func TestClusterValidateUpdateInvalidManagementCluster(t *testing.T) {
	features.ClearCache()
	tests := []struct {
		name               string
		featureGateEnabled bool
		clusterNew         *v1alpha1.Cluster
	}{
		{
			name: "Paused self-managed cluster, feature gate off",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
				c.PauseReconcile()
			}),
			featureGateEnabled: false,
		},
		{
			name: "Unpaused self-managed cluster, feature gate off",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
			}),
			featureGateEnabled: false,
		},
		{
			name: "Paused self-managed cluster, feature gate on",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
				c.PauseReconcile()
			}),
			featureGateEnabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features.ClearCache()
			if tt.featureGateEnabled {
				t.Setenv(features.FullLifecycleAPIEnvVar, "true")
			}
			clusterOld := baseCluster()
			clusterOld.SetSelfManaged()
			tt.clusterNew.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "md-0",
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			}}

			g := NewWithT(t)
			err := tt.clusterNew.ValidateUpdate(clusterOld)
			g.Expect(err).To(MatchError(ContainSubstring("worker node count must be >= 0")))
		})
	}
}

func TestClusterValidateUpdateInvalidWorkloadCluster(t *testing.T) {
	features.ClearCache()
	tests := []struct {
		name               string
		featureGateEnabled bool
		clusterNew         *v1alpha1.Cluster
	}{
		{
			name: "Paused workload cluster, feature gate off",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
				c.PauseReconcile()
			}),
			featureGateEnabled: false,
		},
		{
			name: "Unpaused workload cluster, feature gate off",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
			}),
			featureGateEnabled: false,
		},
		{
			name: "Paused workload cluster, feature gate on",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
				c.PauseReconcile()
			}),
			featureGateEnabled: true,
		},
		{
			name: "Unpaused workload cluster, feature gate on",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
			}),
			featureGateEnabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features.ClearCache()
			if tt.featureGateEnabled {
				t.Setenv(features.FullLifecycleAPIEnvVar, "true")
			}
			clusterOld := baseCluster()
			clusterOld.SetManagedBy("my-management-cluster")

			// Invalid control plane configuration
			tt.clusterNew.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			}

			g := NewWithT(t)
			err := tt.clusterNew.ValidateUpdate(clusterOld)
			g.Expect(err).To(MatchError(ContainSubstring("control plane node count must be positive")))
		})
	}
}

func TestClusterValidateCreateValidCluster(t *testing.T) {
	tests := []struct {
		name               string
		featureGateEnabled bool
		cluster            *v1alpha1.Cluster
	}{
		{
			name: "Paused self-managed cluster, feature gate off",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
				c.PauseReconcile()
			}),
			featureGateEnabled: false,
		},
		{
			name: "Paused workload cluster, feature gate off",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
				c.PauseReconcile()
			}),
			featureGateEnabled: false,
		},
		{
			name: "Paused self-managed cluster, feature gate on",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
				c.PauseReconcile()
			}),
			featureGateEnabled: true,
		},
		{
			name: "Paused workload cluster, feature gate on",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
				c.PauseReconcile()
			}),
			featureGateEnabled: true,
		},
		{
			name: "Unpaused workload cluster, feature gate on",
			cluster: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
			}),
			featureGateEnabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features.ClearCache()
			if tt.featureGateEnabled {
				t.Setenv(features.FullLifecycleAPIEnvVar, "true")
			}
			g := NewWithT(t)
			g.Expect(tt.cluster.ValidateCreate()).To(Succeed())
		})
	}
}

func TestClusterValidateUpdateValidManagementCluster(t *testing.T) {
	features.ClearCache()
	tests := []struct {
		name               string
		featureGateEnabled bool
		oldCluster         *v1alpha1.Cluster
		updateCluster      clusterOpt
	}{
		{
			name:       "Paused self-managed cluster, feature gate off",
			oldCluster: baseCluster(),
			updateCluster: func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
				c.PauseReconcile()
			},
			featureGateEnabled: false,
		},
		{
			name:       "Unpaused self-managed cluster, feature gate off",
			oldCluster: baseCluster(),
			updateCluster: func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
			},
			featureGateEnabled: false,
		},
		{
			name:       "Paused self-managed cluster, feature gate on",
			oldCluster: baseCluster(),
			updateCluster: func(c *v1alpha1.Cluster) {
				c.SetSelfManaged()
				c.PauseReconcile()
			},
			featureGateEnabled: true,
		},
		{
			name: "Unpaused self-managed cluster, feature gate on, no changes",
			oldCluster: baseCluster(
				func(c *v1alpha1.Cluster) {
					c.SetSelfManaged()
				},
			),
			updateCluster:      func(c *v1alpha1.Cluster) {},
			featureGateEnabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features.ClearCache()
			if tt.featureGateEnabled {
				t.Setenv(features.FullLifecycleAPIEnvVar, "true")
			}

			tt.oldCluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{
				Name:  "md-0",
				Count: ptr.Int(4),
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			}}

			newCluster := tt.oldCluster.DeepCopy()
			tt.updateCluster(newCluster)

			g := NewWithT(t)
			err := newCluster.ValidateUpdate(tt.oldCluster)
			g.Expect(err).To(Succeed())
		})
	}
}

func TestClusterValidateUpdateValidWorkloadCluster(t *testing.T) {
	features.ClearCache()
	tests := []struct {
		name               string
		featureGateEnabled bool
		clusterNew         *v1alpha1.Cluster
	}{
		{
			name: "Paused workload cluster, feature gate off",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
				c.PauseReconcile()
			}),
			featureGateEnabled: false,
		},
		{
			name: "Unpaused workload cluster, feature gate off",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
			}),
			featureGateEnabled: false,
		},
		{
			name: "Paused workload cluster, feature gate on",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
				c.PauseReconcile()
			}),
			featureGateEnabled: true,
		},
		{
			name: "Unpaused workload cluster, feature gate on",
			clusterNew: newCluster(func(c *v1alpha1.Cluster) {
				c.SetManagedBy("my-management-cluster")
			}),
			featureGateEnabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features.ClearCache()
			if tt.featureGateEnabled {
				t.Setenv(features.FullLifecycleAPIEnvVar, "true")
			}
			clusterOld := baseCluster()
			clusterOld.SetManagedBy("my-management-cluster")
			tt.clusterNew.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{
				Name:  "md-0",
				Count: ptr.Int(4),
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			}}

			g := NewWithT(t)
			err := tt.clusterNew.ValidateUpdate(clusterOld)
			g.Expect(err).To(Succeed())
		})
	}
}

func TestClusterValidateUpdateInvalidRequest(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.SetSelfManaged()
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	t.Setenv(features.ExperimentalSelfManagedClusterUpgradeEnvVar, "false")

	cNew := cOld.DeepCopy()
	cNew.Spec.ControlPlaneConfiguration.Count = cNew.Spec.ControlPlaneConfiguration.Count + 1
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateRollingAndScalingTinkerbellRequest(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.IsManaged()
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cOld.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind
	cOld.Spec.KubernetesVersion = "1.22"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "1.23"
	cNew.Spec.ControlPlaneConfiguration.Count = cNew.Spec.ControlPlaneConfiguration.Count + 1
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("cannot perform scale up or down during rolling upgrades. Previous control plane node count")))
}

func TestClusterValidateUpdateAddWNConfig(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.IsManaged()
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cOld.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind
	cOld.Spec.KubernetesVersion = "1.22"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "1.23"
	addWNC := v1alpha1.WorkerNodeGroupConfiguration{
		Name:  "md-1",
		Count: ptr.Int(1),
	}
	cNew.Spec.WorkerNodeGroupConfigurations = append(cNew.Spec.WorkerNodeGroupConfigurations, addWNC)
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("cannot perform scale up or down during rolling upgrades. Please remove the new worker node group")))
}

func TestClusterValidateUpdateAddWNCount(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.IsManaged()
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cOld.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind
	cOld.Spec.KubernetesVersion = "1.22"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "1.23"
	cNew.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(3)
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("cannot perform scale up or down during rolling upgrades. Previous worker node count")))
}

func TestClusterValidateUpdateRollingTinkerbellRequest(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "test"
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cOld.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind
	cOld.Spec.KubernetesVersion = "1.22"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "1.23"
	g := NewWithT(t)
	g.Expect(cNew.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateLabelTaintsCPTinkerbellRequest(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "test"
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cOld.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind

	nodeLabels := map[string]string{"label1": "foo", "label2": "bar"}
	var cpTaints []v1.Taint

	cpTaints = append(cpTaints, v1.Taint{Key: "key1", Value: "val1", Effect: "NoSchedule", TimeAdded: nil})
	cOld.Spec.ControlPlaneConfiguration.Labels = nodeLabels
	cOld.Spec.ControlPlaneConfiguration.Taints = cpTaints

	cNew := cOld.DeepCopy()
	cNew.Spec.ControlPlaneConfiguration.Labels = map[string]string{}
	cNew.Spec.ControlPlaneConfiguration.Taints = []v1.Taint{}
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration.labels: Forbidden: field is immutable")))
	g.Expect(err).To(MatchError(ContainSubstring("spec.ControlPlaneConfiguration.taints: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateLabelTaintsWNTinkerbellRequest(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "test"
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cOld.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind

	nodeLabels := map[string]string{"label1": "foo", "label2": "bar"}

	var wnTaints []v1.Taint

	wnTaints = append(wnTaints, v1.Taint{Key: "key1", Value: "val1", Effect: "NoSchedule", TimeAdded: nil})

	cOld.Spec.WorkerNodeGroupConfigurations[0].Labels = nodeLabels
	cOld.Spec.WorkerNodeGroupConfigurations[0].Taints = wnTaints

	cNew := cOld.DeepCopy()
	cNew.Spec.WorkerNodeGroupConfigurations[0].Labels = map[string]string{}
	cNew.Spec.WorkerNodeGroupConfigurations[0].Taints = []v1.Taint{}

	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("spec.WorkerNodeConfiguration.labels: Forbidden: field is immutable")))
	g.Expect(err).To(MatchError(ContainSubstring("spec.WorkerNodeConfiguration.taints: Forbidden: field is immutable")))
}

func TestClusterValidateUpdateLabelTaintsMultiWNTinkerbellRequest(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "test"
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	cOld.Spec.DatacenterRef.Kind = v1alpha1.TinkerbellDatacenterKind

	nodeLabels := map[string]string{"label1": "foo", "label2": "bar"}
	nodeLabels2 := map[string]string{"label3": "foo", "label4": "bar"}

	cOld.Spec.WorkerNodeGroupConfigurations[0].Labels = nodeLabels
	cOld.Spec.WorkerNodeGroupConfigurations[0].Taints = []v1.Taint{{Key: "key1", Value: "val1", Effect: "NoSchedule"}}
	cOld.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Kind = v1alpha1.TinkerbellMachineConfigKind
	cOld.Spec.WorkerNodeGroupConfigurations = append(cOld.Spec.WorkerNodeGroupConfigurations,
		v1alpha1.WorkerNodeGroupConfiguration{
			Name:  "md-1",
			Count: ptr.Int(1),
			MachineGroupRef: &v1alpha1.Ref{
				Kind: v1alpha1.TinkerbellMachineConfigKind,
				Name: "eksa-unit-test",
			},
		})
	cOld.Spec.WorkerNodeGroupConfigurations[1].Labels = nodeLabels2
	cOld.Spec.WorkerNodeGroupConfigurations[1].Taints = []v1.Taint{}

	cNew := cOld.DeepCopy()
	cNew.Spec.WorkerNodeGroupConfigurations[0].Taints = []v1.Taint{{Key: "key1", Value: "val1", Effect: "NoSchedule", TimeAdded: nil}}

	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(BeNil())
}

func TestClusterValidateUpdateSkipUpgradeImmutability(t *testing.T) {
	tests := []struct {
		Name  string
		Old   *v1alpha1.Cluster
		New   *v1alpha1.Cluster
		Error bool
	}{
		{
			Name: "NilToFalse",
			Old: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = nil
			}),
			New: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(false)
			}),
		},
		{
			Name: "FalseToNil",
			Old: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(false)
			}),
			New: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = nil
			}),
		},
		{
			Name: "NilToTrue",
			Old: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = nil
			}),
			New: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(true)
			}),
		},
		{
			Name: "FalseToTrue",
			Old: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(false)
			}),
			New: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(true)
			}),
		},
		{
			Name: "TrueToNil",
			Old: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(true)
			}),
			New: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = nil
			}),
			Error: true,
		},
		{
			Name: "TrueToFalse",
			Old: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(true)
			}),
			New: baseCluster(func(c *v1alpha1.Cluster) {
				c.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(false)
			}),
			Error: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			g := NewWithT(t)

			err := tc.New.ValidateUpdate(tc.Old)
			if !tc.Error {
				g.Expect(err).To(Succeed())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(
					"spec.clusterNetwork.cniConfig.cilium.skipUpgrade: Forbidden: cannot toggle " +
						"off skipUpgrade once enabled",
				)))
			}
		})
	}
}

func TestClusterValidateUpdateVersionSkew(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "mgmt2"
	cOld.Spec.KubernetesVersion = "1.22"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "1.24"
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("only +1 minor version skew is supported")))
}

func TestClusterValidateUpdateVersionSkewDecrement(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "mgmt2"
	cOld.Spec.KubernetesVersion = "1.24"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "1.23"
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("kubernetes version downgrade is not supported")))
}

func TestClusterValidateUpdateVersionInvalidNew(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "mgmt2"
	cOld.Spec.KubernetesVersion = "1.24"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "test"
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("parsing comparison version: could not parse \"test\" as version")))
}

func TestClusterValidateUpdateVersionInvalidOld(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "mgmt2"
	cOld.Spec.KubernetesVersion = "test"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "1.24"
	g := NewWithT(t)
	err := cNew.ValidateUpdate(cOld)
	g.Expect(err).To(MatchError(ContainSubstring("parsing cluster version: could not parse \"test\" as version")))
}

func TestClusterValidateUpdateVersionMinorVersionBump(t *testing.T) {
	features.ClearCache()
	cOld := baseCluster()
	cOld.Spec.ManagementCluster.Name = "mgmt2"
	cOld.Spec.KubernetesVersion = "1.24"

	cNew := cOld.DeepCopy()
	cNew.Spec.KubernetesVersion = "1.25"
	g := NewWithT(t)
	g.Expect(cNew.ValidateUpdate(cOld)).To(Succeed())
}

func newCluster(opts ...func(*v1alpha1.Cluster)) *v1alpha1.Cluster {
	c := baseCluster()
	for _, o := range opts {
		o(c)
	}

	return c
}

type clusterOpt func(c *v1alpha1.Cluster)

func baseCluster(opts ...clusterOpt) *v1alpha1.Cluster {
	c := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.ClusterKind,
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "mgmt",
		},
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.Kube121,
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			},
			BundlesRef: &v1alpha1.BundlesRef{
				Name:       "bundles-1",
				Namespace:  constants.EksaSystemNamespace,
				APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name:  "md-0",
				Count: ptr.Int(1),
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			}},
			ClusterNetwork: v1alpha1.ClusterNetwork{
				CNIConfig: &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{}},
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{"192.168.0.0/16"},
				},
				Services: v1alpha1.Services{
					CidrBlocks: []string{"10.96.0.0/12"},
				},
			},
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.VSphereDatacenterKind,
				Name: "eksa-unit-test",
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
