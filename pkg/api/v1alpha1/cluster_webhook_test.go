package v1alpha1_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
)

func TestClusterValidateUpdateManagementValueImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{}
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.SetManagedBy("management-cluster")

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateManagementOldNilNewTrueSuccess(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()
	c.SetSelfManaged()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateManagementOldNilNewFalseImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{}
	c := cOld.DeepCopy()
	c.SetManagedBy("management-cluster")

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateManagementBothNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestManagementClusterValidateUpdateKubernetesVersionImmutable(t *testing.T) {
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
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestManagementNilClusterValidateUpdateKubernetesVersionImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion:         v1alpha1.Kube119,
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.KubernetesVersion = v1alpha1.Kube120

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestWorkloadClusterValidateUpdateKubernetesVersionSuccess(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion:         v1alpha1.Kube119,
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.KubernetesVersion = v1alpha1.Kube120

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationEqual(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count:           3,
				Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
				MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
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
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count:           3,
				Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
				MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
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
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count:           3,
				Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
				MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Count:           10,
		Endpoint:        &v1alpha1.Endpoint{Host: "1.1.1.1/2"},
		MachineGroupRef: &v1alpha1.Ref{Name: "test2", Kind: "SecondMachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateControlPlaneConfigurationOldEndpointImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/2"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateControlPlaneConfigurationOldEndpointNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Endpoint: nil,
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateControlPlaneConfigurationNewEndpointNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		Endpoint: nil,
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationTaintsImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Taints: []v1.Taint{
					{
						Key:    "Key1",
						Value:  "Val1",
						Effect: "PreferNoSchedule",
					},
				},
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
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationLabelsImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Labels: map[string]string{
					"Key1": "Val1",
				},
			},
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
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationOldMachineGroupRefImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				MachineGroupRef: &v1alpha1.Ref{Name: "test1", Kind: "MachineConfig"},
			},
		},
	}
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test2", Kind: "MachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestWorkloadClusterValidateUpdateControlPlaneConfigurationMachineGroupRef(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				MachineGroupRef: &v1alpha1.Ref{Name: "test1", Kind: "MachineConfig"},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test2", Kind: "MachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationOldMachineGroupRefNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				MachineGroupRef: nil,
			},
		},
	}
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestWorkloadClusterValidateUpdateControlPlaneConfigurationOldMachineGroupRefNilSuccess(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				MachineGroupRef: nil,
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestManagementClusterValidateUpdateControlPlaneConfigurationNewMachineGroupRefNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
			},
		},
	}
	cOld.SetSelfManaged()

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: nil,
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestWorkloadClusterValidateUpdateControlPlaneConfigurationNewMachineGroupRefNilSuccess(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				MachineGroupRef: &v1alpha1.Ref{Name: "test", Kind: "MachineConfig"},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("management-cluster")

	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
		MachineGroupRef: nil,
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateDatacenterRefImmutableEqual(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Name: "test", Kind: "DatacenterConfig",
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateDatacenterRefImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Name: "test", Kind: "DatacenterConfig",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef = v1alpha1.Ref{Name: "test2", Kind: "SecondDatacenterConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateDatacenterRefImmutableName(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Name: "test", Kind: "DatacenterConfig",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef = v1alpha1.Ref{Name: "test2", Kind: "DatacenterConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateDatacenterRefNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Name: "test", Kind: "DatacenterConfig",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef = v1alpha1.Ref{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateExternalEtcdReplicasImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ExternalEtcdConfiguration.Count = 5

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateDataCenterRefNameImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Name: "oldBadDatacetner",
				Kind: v1alpha1.VSphereDatacenterKind,
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef.Name = "FancyNewDataCenter"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateDataCenterRefKindImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			DatacenterRef: v1alpha1.Ref{
				Name: "oldBadDatacetner",
				Kind: v1alpha1.VSphereDatacenterKind,
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.DatacenterRef.Name = v1alpha1.DockerDatacenterKind

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateClusterNetworkEqualOrder(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{"1.2.3.4/5", "1.2.3.4/6"},
				},
				Services: v1alpha1.Services{
					CidrBlocks: []string{"1.2.3.4/7", "1.2.3.4/8"},
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
		Pods: v1alpha1.Pods{
			CidrBlocks: []string{"1.2.3.4/6", "1.2.3.4/5"},
		},
		Services: v1alpha1.Services{
			CidrBlocks: []string{"1.2.3.4/8", "1.2.3.4/7"},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateClusterNetworkImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{"1.2.3.4/5", "1.2.3.4/6"},
				},
				Services: v1alpha1.Services{
					CidrBlocks: []string{"1.2.3.4/7", "1.2.3.4/8"},
				},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
		Pods: v1alpha1.Pods{
			CidrBlocks: []string{"1.2.3.4/5"},
		},
		Services: v1alpha1.Services{
			CidrBlocks: []string{"1.2.3.4/9", "1.2.3.4/10"},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateClusterNetworkOldEmptyImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ClusterNetwork: v1alpha1.ClusterNetwork{},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
		Pods: v1alpha1.Pods{
			CidrBlocks: []string{"1.2.3.4/5"},
		},
		Services: v1alpha1.Services{
			CidrBlocks: []string{"1.2.3.4/9", "1.2.3.4/10"},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateClusterNetworkNewEmptyImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{"1.2.3.4/5", "1.2.3.4/6"},
				},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
		Pods: v1alpha1.Pods{},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateClusterNetworkCiliumConfigImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ClusterNetwork: v1alpha1.ClusterNetwork{
				CNIConfig: &v1alpha1.CNIConfig{
					Cilium: &v1alpha1.CiliumConfig{
						PolicyEnforcementMode: "default",
					},
				},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
		CNIConfig: &v1alpha1.CNIConfig{
			Cilium: &v1alpha1.CiliumConfig{
				PolicyEnforcementMode: "always",
			},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateProxyConfigurationEqualOrder(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ProxyConfiguration: &v1alpha1.ProxyConfiguration{
				HttpsProxy: "httpsproxy",
				NoProxy: []string{
					"noproxy1",
					"noproxy2",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	c := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ProxyConfiguration: &v1alpha1.ProxyConfiguration{
				HttpProxy:  "",
				HttpsProxy: "httpsproxy",
				NoProxy: []string{
					"noproxy2",
					"noproxy1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateProxyConfigurationImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ProxyConfiguration: &v1alpha1.ProxyConfiguration{
				HttpProxy:  "httpproxy1",
				HttpsProxy: "httpsproxy1",
				NoProxy:    []string{"noproxy1"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "httpproxy2",
		HttpsProxy: "httpsproxy2",
		NoProxy:    []string{"noproxy1", "noproxy2"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateProxyConfigurationNoProxyImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ProxyConfiguration: &v1alpha1.ProxyConfiguration{
				HttpProxy:  "httpproxy",
				HttpsProxy: "httpsproxy",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "httpproxy",
		HttpsProxy: "httpsproxy",
		NoProxy:    []string{"noproxy"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateProxyConfigurationOldNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{ProxyConfiguration: nil},
	}
	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{
		HttpProxy:  "httpproxy",
		HttpsProxy: "httpsproxy",
		NoProxy:    []string{"noproxy"},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateProxyConfigurationNewNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ProxyConfiguration: &v1alpha1.ProxyConfiguration{
				HttpProxy:  "httpproxy",
				HttpsProxy: "httpsproxy",
				NoProxy:    []string{"noproxy"},
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.ProxyConfiguration = nil
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateGitOpsRefImmutableNilEqual(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			GitOpsRef: nil,
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateGitOpsRefImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			GitOpsRef: &v1alpha1.Ref{
				Name: "test1", Kind: "GitOpsConfig1",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = &v1alpha1.Ref{Name: "test2", Kind: "GitOpsConfig2"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateGitOpsRefImmutableName(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			GitOpsRef: &v1alpha1.Ref{
				Name: "test1", Kind: "GitOpsConfig",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = &v1alpha1.Ref{Name: "test2", Kind: "GitOpsConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateGitOpsRefImmutableKind(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			GitOpsRef: &v1alpha1.Ref{
				Name: "test", Kind: "GitOpsConfig1",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = &v1alpha1.Ref{Name: "test", Kind: "GitOpsConfig2"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateGitOpsRefOldNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{GitOpsRef: nil},
	}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = &v1alpha1.Ref{Name: "test", Kind: "GitOpsConfig"}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateGitOpsRefNewNilImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			GitOpsRef: &v1alpha1.Ref{
				Name: "test", Kind: "GitOpsConfig",
			},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.GitOpsRef = nil

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateAWSIamNameImmutableUpdateSameName(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.AWSIamConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.AWSIamConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateAWSIamNameImmutableUpdateName(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.AWSIamConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.AWSIamConfigKind,
					Name: "name2",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateAWSIamNameImmutableEmpty(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.AWSIamConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateAWSIamNameImmutableAddConfig(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: v1alpha1.AWSIamConfigKind,
			Name: "name1",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableUpdateNameWorkloadCluster(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("management-cluster")
	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs[0].Name = "name2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableUpdateNameMgmtCluster(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs[0].Name = "name2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableUpdateNameUnchanged(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableWorkloadCluster(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("mgmt")

	c := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	c.SetManagedBy("mgmt")
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableMgmtCluster(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	c := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableAddConfigWorkloadCluster(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("mgmt")

	c := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c.SetManagedBy("mgmt")

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateOIDCNameMutableAddConfigMgmtCluster(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	c := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{
				{
					Kind: v1alpha1.OIDCConfigKind,
					Name: "name1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateEmptyIdentityProviders(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateGitOpsRefOldEmptyImmutable(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			IdentityProviderRefs: []v1alpha1.Ref{},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.IdentityProviderRefs = []v1alpha1.Ref{
		{
			Kind: "identity",
			Name: "name",
		},
	}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateWithPausedAnnotation(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: make(map[string]string, 1),
		},
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.Kube119,
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.PauseReconcile()
	c := cOld.DeepCopy()
	c.Spec.KubernetesVersion = v1alpha1.Kube120

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterValidateUpdateInvalidType(t *testing.T) {
	cOld := &v1alpha1.VSphereDatacenterConfig{}
	c := &v1alpha1.Cluster{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateSuccess(t *testing.T) {
	workerConfiguration := append([]v1alpha1.WorkerNodeGroupConfiguration{}, v1alpha1.WorkerNodeGroupConfiguration{Count: 5, Name: "test"})
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: workerConfiguration,
			KubernetesVersion:             v1alpha1.Kube119,
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Count = 10

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterCreateManagementCluster(t *testing.T) {
	os.Setenv("FULL_LIFECYCLE_API", "true")
	workerConfiguration := append([]v1alpha1.WorkerNodeGroupConfiguration{}, v1alpha1.WorkerNodeGroupConfiguration{Count: 5})
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
	g.Expect(cluster.ValidateCreate()).NotTo(Succeed())
	os.Unsetenv("FULL_LIFECYCLE_API")
}

func TestClusterCreateCloudStackMultipleWorkerNodeGroupsValidation(t *testing.T) {
	os.Setenv(features.CloudStackProviderEnvVar, "true")
	workerConfiguration := append([]v1alpha1.WorkerNodeGroupConfiguration{}, v1alpha1.WorkerNodeGroupConfiguration{Count: 5, Name: "test"},
		v1alpha1.WorkerNodeGroupConfiguration{Count: 5, Name: "test2"})
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: workerConfiguration,
			KubernetesVersion:             v1alpha1.Kube119,
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.CloudStackDatacenterKind,
			},
		},
	}

	g := NewWithT(t)
	g.Expect(cluster.ValidateCreate()).NotTo(Succeed())
	os.Unsetenv(features.CloudStackProviderEnvVar)
}

func TestClusterCreateWorkloadCluster(t *testing.T) {
	os.Setenv("FULL_LIFECYCLE_API", "true")
	workerConfiguration := append([]v1alpha1.WorkerNodeGroupConfiguration{}, v1alpha1.WorkerNodeGroupConfiguration{Count: 5})
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: workerConfiguration,
			KubernetesVersion:             v1alpha1.Kube119,
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3, Endpoint: &v1alpha1.Endpoint{Host: "1.1.1.1/1"},
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{Count: 3},
			ClusterNetwork:            v1alpha1.ClusterNetwork{CNIConfig: &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{}}},
		},
	}
	cluster.Spec.ManagementCluster.Name = "management-cluster"

	g := NewWithT(t)
	g.Expect(cluster.ValidateCreate()).To(Succeed())
	os.Unsetenv("FULL_LIFECYCLE_API")
}

func TestClusterUpdateWorkerNodeGroupTaintsAndLabelsSuccess(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
				Taints: []v1.Taint{{
					Key:    "test",
					Value:  "test",
					Effect: "PreferNoSchedule",
				}},
				Labels: map[string]string{
					"test": "val1",
				},
			}},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Taints[0].Value = "test2"
	c.Spec.WorkerNodeGroupConfigurations[0].Labels["test"] = "val2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterUpdateWorkerNodeGroupTaintsInvalid(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
				Taints: []v1.Taint{{
					Key:    "test",
					Value:  "test",
					Effect: "PreferNoSchedule",
				}},
			}},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Taints[0].Effect = "NoSchedule"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterUpdateWorkerNodeGroupNameInvalid(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Name = ""

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterUpdateWorkerNodeGroupLabelsInvalid(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
				Labels: map[string]string{
					"test": "val1",
				},
			}},
		},
	}
	c := cOld.DeepCopy()
	c.Spec.WorkerNodeGroupConfigurations[0].Labels["test"] = "val1/val2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}

func TestClusterUpdateControlPlaneTaintsAndLabelsSuccess(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Taints: []v1.Taint{{
					Key:    "test",
					Value:  "test",
					Effect: "PreferNoSchedule",
				}},
				Labels: map[string]string{
					"test": "val1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("management-cluster")
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.Taints[0].Value = "test2"
	c.Spec.ControlPlaneConfiguration.Labels["test"] = "val2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).To(Succeed())
}

func TestClusterUpdateControlPlaneLabelsInvalid(t *testing.T) {
	cOld := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Labels: map[string]string{
					"test": "val1",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{{
				Name: "test",
			}},
		},
	}
	cOld.SetManagedBy("management-cluster")
	c := cOld.DeepCopy()
	c.Spec.ControlPlaneConfiguration.Labels["test"] = "val1/val2"

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(cOld)).NotTo(Succeed())
}
