package cloudstack

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestGenerateTemplateBuilder(t *testing.T) {
	g := NewWithT(t)
	clusterSpec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	spec := v1alpha1.ClusterSpec{
		ControlPlaneConfiguration:     clusterSpec.Cluster.Spec.ControlPlaneConfiguration,
		WorkerNodeGroupConfigurations: clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations,
		ExternalEtcdConfiguration:     clusterSpec.Cluster.Spec.ExternalEtcdConfiguration,
	}

	expectedControlPlaneMachineSpec := &v1alpha1.CloudStackMachineConfigSpec{
		Template: v1alpha1.CloudStackResourceIdentifier{
			Id:   "",
			Name: "centos7-k8s-118",
		},
		ComputeOffering: v1alpha1.CloudStackResourceIdentifier{Id: "", Name: "m4-large"},
		DiskOffering: &v1alpha1.CloudStackResourceDiskOffering{
			CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{Id: "", Name: "Small"},
			CustomSize:                   0,
			MountPath:                    "/data-small",
			Device:                       "/dev/vdb",
			Filesystem:                   "ext4",
			Label:                        "data_disk",
		},
		Users: []v1alpha1.UserConfiguration{
			{
				Name: "mySshUsername",
				SshAuthorizedKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com",
				},
			},
		},
		Affinity: "",
		AffinityGroupIds: []string{
			"control-plane-anti-affinity",
		},
		UserCustomDetails: nil,
		Symlinks: v1alpha1.SymlinkMaps{
			"/var/log/kubernetes": "/data-small/var/log/kubernetes",
		},
	}

	gotExpectedControlPlaneMachineSpec, err := getControlPlaneMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotExpectedControlPlaneMachineSpec).To(Equal(expectedControlPlaneMachineSpec))

	expectedWorkerNodeGroupMachineSpec := map[string]v1alpha1.CloudStackMachineConfigSpec{
		"test": {
			Template: v1alpha1.CloudStackResourceIdentifier{
				Id:   "",
				Name: "centos7-k8s-118",
			},
			ComputeOffering: v1alpha1.CloudStackResourceIdentifier{Id: "", Name: "m4-large"},
			DiskOffering: &v1alpha1.CloudStackResourceDiskOffering{
				CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{Id: "", Name: "Small"},
				CustomSize:                   0,
				MountPath:                    "/data-small",
				Device:                       "/dev/vdb",
				Filesystem:                   "ext4",
				Label:                        "data_disk",
			},
			Users: []v1alpha1.UserConfiguration{
				{
					Name: "mySshUsername",
					SshAuthorizedKeys: []string{
						"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com",
					},
				},
			},
			Affinity:          "",
			AffinityGroupIds:  []string{"worker-affinity"},
			UserCustomDetails: map[string]string{"foo": "bar"},
			Symlinks: v1alpha1.SymlinkMaps{
				"/var/log/containers": "/data-small/var/log/containers",
				"/var/log/pods":       "/data-small/var/log/pods",
			},
		},
	}
	gotWorkerNodeGroupMachineSpec, err := getWorkerNodeGroupMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotWorkerNodeGroupMachineSpec).To(Equal(expectedWorkerNodeGroupMachineSpec))

	gotEtcdMachineSpec, err := getEtcdMachineSpec(spec, clusterSpec.CloudStackMachineConfigs)
	expectedEtcdMachineSpec := &v1alpha1.CloudStackMachineConfigSpec{
		Template: v1alpha1.CloudStackResourceIdentifier{
			Id:   "",
			Name: "centos7-k8s-118",
		},
		ComputeOffering: v1alpha1.CloudStackResourceIdentifier{Id: "", Name: "m4-large"},
		DiskOffering: &v1alpha1.CloudStackResourceDiskOffering{
			CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{Id: "", Name: "Small"},
			CustomSize:                   0,
			MountPath:                    "/data-small",
			Device:                       "/dev/vdb",
			Filesystem:                   "ext4",
			Label:                        "data_disk",
		},
		Users: []v1alpha1.UserConfiguration{
			{
				Name: "mySshUsername",
				SshAuthorizedKeys: []string{
					"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ== testemail@test.com",
				},
			},
		},
		Affinity: "",
		AffinityGroupIds: []string{
			"etcd-affinity",
		},
		UserCustomDetails: nil,
		Symlinks: v1alpha1.SymlinkMaps{
			"/var/lib/": "/data-small/var/lib",
		},
	}
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(gotEtcdMachineSpec).To(Equal(expectedEtcdMachineSpec))
}
