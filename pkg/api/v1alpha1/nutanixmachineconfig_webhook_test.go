package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func nutanixMachineConfig() *v1alpha1.NutanixMachineConfig {
	return &v1alpha1.NutanixMachineConfig{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-nmc",
		},
		Spec: v1alpha1.NutanixMachineConfigSpec{
			OSFamily:       v1alpha1.Ubuntu,
			VCPUsPerSocket: 2,
			VCPUSockets:    4,
			MemorySize:     resource.MustParse("8Gi"),
			Image: v1alpha1.NutanixResourceIdentifier{
				Type: v1alpha1.NutanixIdentifierName,
				Name: ptr.String("ubuntu-image"),
			},
			Cluster: v1alpha1.NutanixResourceIdentifier{
				Type: v1alpha1.NutanixIdentifierName,
				Name: ptr.String("cluster-1"),
			},
			Subnet: v1alpha1.NutanixResourceIdentifier{
				Type: v1alpha1.NutanixIdentifierName,
				Name: ptr.String("subnet-1"),
			},
			SystemDiskSize: resource.MustParse("100Gi"),
			Users: []v1alpha1.UserConfiguration{
				{
					Name: "test-user",
				},
			},
		},
	}
}

func TestNutanixMachineConfigWebhooksValidateCreateReconcilePaused(t *testing.T) {
	g := NewWithT(t)
	conf := nutanixMachineConfig()
	conf.Annotations = map[string]string{
		"anywhere.eks.amazonaws.com/paused": "true",
	}
	g.Expect(conf.ValidateCreate()).To(Succeed())
}

func TestValidateCreate_Valid(t *testing.T) {
	g := NewWithT(t)
	config := nutanixMachineConfig()
	g.Expect(config.ValidateCreate()).To(Succeed())
}

func TestValidateCreate_Invalid(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name string
		fn   func(*v1alpha1.NutanixMachineConfig)
	}{
		{
			name: "invalid name",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Name = ""
			},
		},
		{
			name: "invalid os family",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.OSFamily = "invalid"
			},
		},
		{
			name: "invalid vcpus per socket",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.VCPUsPerSocket = 0
			},
		},
		{
			name: "invalid vcpus sockets",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.VCPUSockets = 0
			},
		},
		{
			name: "invalid memory size",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.MemorySize = resource.MustParse("0Gi")
			},
		},
		{
			name: "invalid image type",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.Image.Type = "invalid"
			},
		},
		{
			name: "invalid image name",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.Image.Name = nil
			},
		},
		{
			name: "invalid cluster type",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.Cluster.Type = "invalid"
			},
		},
		{
			name: "invalid cluster name",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.Cluster.Name = nil
			},
		},
		{
			name: "invalid subnet type",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.Subnet.Type = "invalid"
			},
		},
		{
			name: "invalid subnet name",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.Subnet.Name = nil
			},
		},
		{
			name: "invalid system disk size",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.SystemDiskSize = resource.MustParse("0Gi")
			},
		},
		{
			name: "no user",
			fn: func(config *v1alpha1.NutanixMachineConfig) {
				config.Spec.Users = []v1alpha1.UserConfiguration{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := nutanixMachineConfig()
			tt.fn(config)
			err := config.ValidateCreate()
			g.Expect(err).To(HaveOccurred(), "expected error for %s", tt.name)
		})
	}
}

func TestNutanixMachineConfigWebhooksValidateUpdateReconcilePaused(t *testing.T) {
	g := NewWithT(t)
	oldConfig := nutanixMachineConfig()
	newConfig := nutanixMachineConfig()
	oldConfig.Annotations = map[string]string{
		"anywhere.eks.amazonaws.com/paused": "true",
	}
	g.Expect(newConfig.ValidateUpdate(oldConfig)).To(Succeed())
}

func TestValidateUpdate_Valid(t *testing.T) {
	g := NewWithT(t)
	oldConfig := nutanixMachineConfig()
	newConfig := nutanixMachineConfig()
	newConfig.Spec.VCPUSockets = 8
	g.Expect(newConfig.ValidateUpdate(oldConfig)).To(Succeed())
}

func TestValidateUpdate_Invalid(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name string
		fn   func(new *v1alpha1.NutanixMachineConfig)
	}{
		{
			name: "different os family",
			fn: func(new *v1alpha1.NutanixMachineConfig) {
				new.Spec.OSFamily = v1alpha1.Bottlerocket
			},
		},
		{
			name: "different cluster",
			fn: func(new *v1alpha1.NutanixMachineConfig) {
				new.Spec.Cluster = v1alpha1.NutanixResourceIdentifier{
					Type: v1alpha1.NutanixIdentifierName,
					Name: ptr.String("cluster-2"),
				}
			},
		},
		{
			name: "different subnet",
			fn: func(new *v1alpha1.NutanixMachineConfig) {
				new.Spec.Subnet = v1alpha1.NutanixResourceIdentifier{
					Type: v1alpha1.NutanixIdentifierName,
					Name: ptr.String("subnet-2"),
				}
			},
		},
		{
			name: "different image",
			fn: func(new *v1alpha1.NutanixMachineConfig) {
				new.Spec.Image = v1alpha1.NutanixResourceIdentifier{
					Type: v1alpha1.NutanixIdentifierName,
					Name: ptr.String("image-2"),
				}
			},
		},
		{
			name: "invalid vcpus per socket",
			fn: func(new *v1alpha1.NutanixMachineConfig) {
				new.Spec.VCPUsPerSocket = 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldConfig := nutanixMachineConfig()
			newConfig := nutanixMachineConfig()
			tt.fn(newConfig)
			err := newConfig.ValidateUpdate(oldConfig)
			g.Expect(err).To(HaveOccurred(), "expected error for %s", tt.name)
		})
	}
}

func TestValidateUpdate_OldObjectNotMachineConfig(t *testing.T) {
	g := NewWithT(t)
	oldConfig := nutanixDatacenterConfig()
	newConfig := nutanixMachineConfig()
	err := newConfig.ValidateUpdate(oldConfig)
	g.Expect(err).To(HaveOccurred())
}

func TestNutanixMachineConfigSetupWebhookWithManager(t *testing.T) {
	t.Setenv(features.FullLifecycleAPIEnvVar, "true")
	g := NewWithT(t)
	conf := nutanixMachineConfig()
	g.Expect(conf.SetupWebhookWithManager(env.Manager())).To(Succeed())
}

func TestNutanixMachineConfigWebhooksValidateDelete(t *testing.T) {
	g := NewWithT(t)
	config := nutanixMachineConfig()
	g.Expect(config.ValidateDelete()).To(Succeed())
}
