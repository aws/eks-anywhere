package tinkerbell

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestGenerateTemplateBuilder(t *testing.T) {
	g := NewWithT(t)
	clusterSpec := test.NewFullClusterSpec(t, testClusterConfigFilename)

	expectedControlPlaneMachineSpec := &v1alpha1.TinkerbellMachineConfigSpec{
		HardwareSelector: map[string]string{"type": "cp"},
		TemplateRef: v1alpha1.Ref{
			Kind: "TinkerbellTemplateConfig",
			Name: "tink-test",
		},
		OSFamily: "ubuntu",
		Users: []v1alpha1.UserConfiguration{
			{
				Name:              "tink-user",
				SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
			},
		},
	}
	gotExpectedControlPlaneMachineSpec, err := getControlPlaneMachineSpec(clusterSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotExpectedControlPlaneMachineSpec).To(Equal(expectedControlPlaneMachineSpec))

	expectedWorkerNodeGroupMachineSpec := map[string]v1alpha1.TinkerbellMachineConfigSpec{
		"test-md": {
			HardwareSelector: map[string]string{"type": "worker"},
			TemplateRef: v1alpha1.Ref{
				Kind: "TinkerbellTemplateConfig",
				Name: "tink-test",
			},
			OSFamily: "ubuntu",
			Users: []v1alpha1.UserConfiguration{
				{
					Name:              "tink-user",
					SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"},
				},
			},
		},
	}
	gotWorkerNodeGroupMachineSpec, err := getWorkerNodeGroupMachineSpec(clusterSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotWorkerNodeGroupMachineSpec).To(Equal(expectedWorkerNodeGroupMachineSpec))

	gotEtcdMachineSpec, err := getEtcdMachineSpec(clusterSpec)
	var expectedEtcdMachineSpec *v1alpha1.TinkerbellMachineConfigSpec
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotEtcdMachineSpec).To(Equal(expectedEtcdMachineSpec))
}

func TestBuildTemplateMapCPFailAuditPolicy(t *testing.T) {
	g := NewWithT(t)
	clusterSpec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	controlPlaneMachineSpec, err := getControlPlaneMachineSpec(clusterSpec)
	g.Expect(err).ToNot(HaveOccurred())

	cpTemplateOverride := "test"
	etcdTemplateOverride := "test"

	etcdMachineSpec := &v1alpha1.TinkerbellMachineConfigSpec{
		HardwareSelector: map[string]string{"type": "etcd"},
		TemplateRef: v1alpha1.Ref{
			Kind: "TinkerbellTemplateConfig",
			Name: "tink-test",
		},
		OSFamily: "ubuntu",
		Users: []v1alpha1.UserConfiguration{
			{
				Name:              "tink-user",
				SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
			},
		},
	}

	clusterSpec.Cluster.Spec.KubernetesVersion = "invalid"
	_, err = buildTemplateMapCP(clusterSpec, *controlPlaneMachineSpec, *etcdMachineSpec, cpTemplateOverride, etcdTemplateOverride, *clusterSpec.TinkerbellDatacenter.Spec.DeepCopy())
	g.Expect(err).To(HaveOccurred())
}

func TestGenerateTemplateBuilderForMachineConfigOsImageURL(t *testing.T) {
	g := NewWithT(t)
	testFile := "testdata/cluster_osimage_machine_config.yaml"
	clusterSpec := test.NewFullClusterSpec(t, testFile)

	expectedControlPlaneMachineSpec := &v1alpha1.TinkerbellMachineConfigSpec{
		HardwareSelector: map[string]string{"type": "cp"},
		OSFamily:         "ubuntu",
		OSImageURL:       "https://ubuntu-1-21.gz",
		Users: []v1alpha1.UserConfiguration{
			{
				Name:              "tink-user",
				SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
			},
		},
		HostOSConfiguration: nil,
	}
	gotExpectedControlPlaneMachineSpec, err := getControlPlaneMachineSpec(clusterSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotExpectedControlPlaneMachineSpec).To(Equal(expectedControlPlaneMachineSpec))

	expectedWorkerNodeGroupMachineSpec := map[string]v1alpha1.TinkerbellMachineConfigSpec{
		"test-md": {
			HardwareSelector: map[string]string{"type": "worker"},
			OSFamily:         "ubuntu",
			OSImageURL:       "https://ubuntu-1-21.gz",
			Users: []v1alpha1.UserConfiguration{
				{
					Name:              "tink-user",
					SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com"},
				},
			},
			HostOSConfiguration: nil,
		},
	}
	gotWorkerNodeGroupMachineSpec, err := getWorkerNodeGroupMachineSpec(clusterSpec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotWorkerNodeGroupMachineSpec).To(Equal(expectedWorkerNodeGroupMachineSpec))

	gotEtcdMachineSpec, err := getEtcdMachineSpec(clusterSpec)
	var expectedEtcdMachineSpec *v1alpha1.TinkerbellMachineConfigSpec
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotEtcdMachineSpec).To(Equal(expectedEtcdMachineSpec))
}

func TestTemplateBuilder_CertSANs(t *testing.T) {
	for _, tc := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "testdata/cluster_tinkerbell_api_server_cert_san_domain_name.yaml",
			Output: "testdata/expected_cluster_tinkerbell_api_server_cert_san_domain_name.yaml",
		},
		{
			Input:  "testdata/cluster_tinkerbell_api_server_cert_san_ip.yaml",
			Output: "testdata/expected_cluster_tinkerbell_api_server_cert_san_ip.yaml",
		},
	} {
		g := NewWithT(t)
		clusterSpec := test.NewFullClusterSpec(t, tc.Input)

		cpMachineCfg, err := getControlPlaneMachineSpec(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())

		wngMachineCfgs, err := getWorkerNodeGroupMachineSpec(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())

		bldr := NewTemplateBuilder(&clusterSpec.TinkerbellDatacenter.Spec, cpMachineCfg, nil, wngMachineCfgs, "0.0.0.0", time.Now)

		data, err := bldr.GenerateCAPISpecControlPlane(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())
		test.AssertContentToFile(t, string(data), tc.Output)

	}
}

func TestTemplateBuilder(t *testing.T) {
	for _, tc := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "testdata/cluster_tinkerbell_api_server_cert_san_ip.yaml",
			Output: "testdata/expected_cluster_tinkerbell_api_server_cert_san_ip.yaml",
		},
	} {
		g := NewWithT(t)
		clusterSpec := test.NewFullClusterSpec(t, tc.Input)
		clusterSpec.Cluster.AddTinkerbellIPAnnotation("1.1.1.1")

		cpMachineCfg, err := getControlPlaneMachineSpec(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())

		wngMachineCfgs, err := getWorkerNodeGroupMachineSpec(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())

		tinkIPBefore := "0.0.0.0"
		bldr := NewTemplateBuilder(&clusterSpec.TinkerbellDatacenter.Spec, cpMachineCfg, nil, wngMachineCfgs, tinkIPBefore, time.Now)

		data, err := bldr.GenerateCAPISpecControlPlane(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())

		test.AssertContentToFile(t, string(data), tc.Output)

	}
}

func TestTemplateBuilderCPKubeletConfig(t *testing.T) {
	for _, tc := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "testdata/cluster_tinkerbell_api_server_cert_san_ip.yaml",
			Output: "testdata/expected_kcp.yaml",
		},
	} {
		g := NewWithT(t)
		clusterSpec := test.NewFullClusterSpec(t, tc.Input)
		cpMachineCfg, _ := getControlPlaneMachineSpec(clusterSpec)
		wngMachineCfgs, _ := getWorkerNodeGroupMachineSpec(clusterSpec)
		tinkIPBefore := "0.0.0.0"
		bldr := NewTemplateBuilder(&clusterSpec.TinkerbellDatacenter.Spec, cpMachineCfg, nil, wngMachineCfgs, tinkIPBefore, time.Now)

		clusterSpec.Cluster.Spec.ControlPlaneConfiguration.KubeletConfiguration = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"maxPods":    20,
				"apiVersion": "kubelet.config.k8s.io/v1beta1",
				"kind":       "KubeletConfiguration",
			},
		}

		clusterSpec.Cluster.Spec.ClusterNetwork.DNS = v1alpha1.DNS{
			ResolvConf: &v1alpha1.ResolvConf{
				Path: "test-path",
			},
		}

		data, err := bldr.GenerateCAPISpecControlPlane(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(data).To(ContainSubstring("maxPods"))
		test.AssertContentToFile(t, string(data), tc.Output)
	}
}

func TestTemplateBuilderWNKubeletConfig(t *testing.T) {
	for _, tc := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "testdata/cluster_tinkerbell_api_server_cert_san_ip.yaml",
			Output: "testdata/expected_kct.yaml",
		},
	} {
		g := NewWithT(t)
		clusterSpec := test.NewFullClusterSpec(t, tc.Input)
		clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{
			{
				Name:  "test",
				Count: ptr.Int(1),
				KubeletConfiguration: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"maxPods":    20,
						"apiVersion": "kubelet.config.k8s.io/v1beta1",
						"kind":       "KubeletConfiguration",
					},
				},
				MachineGroupRef: &v1alpha1.Ref{
					Name: "wn-ref",
					Kind: v1alpha1.TinkerbellMachineConfigKind,
				},
			},
		}
		clusterSpec.TinkerbellMachineConfigs = map[string]*v1alpha1.TinkerbellMachineConfig{
			"wn-ref": {
				Spec: v1alpha1.TinkerbellMachineConfigSpec{
					Users: []v1alpha1.UserConfiguration{
						{
							SshAuthorizedKeys: []string{"ssh abcdef..."},
							Name:              "user",
						},
					},
				},
			},
		}

		clusterSpec.Cluster.Spec.ClusterNetwork.DNS = v1alpha1.DNS{
			ResolvConf: &v1alpha1.ResolvConf{
				Path: "test-path",
			},
		}

		cpMachineCfg, _ := getControlPlaneMachineSpec(clusterSpec)
		wngMachineCfgs, _ := getWorkerNodeGroupMachineSpec(clusterSpec)
		tinkIPBefore := "0.0.0.0"
		bldr := NewTemplateBuilder(&clusterSpec.TinkerbellDatacenter.Spec, cpMachineCfg, nil, wngMachineCfgs, tinkIPBefore, time.Now)
		workerTemplateNames, kubeadmTemplateNames := clusterapi.InitialTemplateNamesForWorkers(clusterSpec)
		data, err := bldr.GenerateCAPISpecWorkers(clusterSpec, workerTemplateNames, kubeadmTemplateNames)
		g.Expect(err).ToNot(HaveOccurred())
		test.AssertContentToFile(t, string(data), tc.Output)
	}
}
