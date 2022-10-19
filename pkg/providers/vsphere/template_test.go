package vsphere

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/internal/test"
	v1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlane(t *testing.T) {
	type fields struct {
		datacenterSpec          *v1alpha1.VSphereDatacenterConfigSpec
		controlPlaneMachineSpec *v1alpha1.VSphereMachineConfigSpec
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.ControlPlaneConfiguration = v1alpha1.ControlPlaneConfiguration{
			Count: 3,
			Endpoint: &v1alpha1.Endpoint{
				Host: "test-ip",
			},
			MachineGroupRef: &v1alpha1.Ref{
				Kind: v1alpha1.VSphereMachineConfigKind,
				Name: "eksa-unit-test",
			},
		}
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{
			Name:  "md-0",
			Count: ptr.Int(3),
			MachineGroupRef: &v1alpha1.Ref{
				Kind: v1alpha1.VSphereMachineConfigKind,
				Name: "eksa-unit-test",
			},
		}}
		s.Cluster.Spec.ClusterNetwork = v1alpha1.ClusterNetwork{
			CNIConfig: &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{}},
			Pods: v1alpha1.Pods{
				CidrBlocks: []string{"192.168.0.0/16"},
			},
			Services: v1alpha1.Services{
				CidrBlocks: []string{"10.96.0.0/12"},
			},
		}
		s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: "eksa-unit-test",
		}
		s.VSphereDatacenter = &v1alpha1.VSphereDatacenterConfig{
			Spec: v1alpha1.VSphereDatacenterConfigSpec{
				Datacenter: "test",
				Network:    "test",
				Server:     "test",
			},
		}
		s.Cluster.Spec.DatacenterRef = v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: "vsphere test",
		}
	})

	vsphereMachineConfig := &v1alpha1.VSphereMachineConfigSpec{
		Users: []v1alpha1.UserConfiguration{
			{
				Name:              "capv",
				SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC1BK73XhIzjX+meUr7pIYh6RHbvI3tmHeQIXY5lv7aztN1UoX+bhPo3dwo2sfSQn5kuxgQdnxIZ/CTzy0p0GkEYVv3gwspCeurjmu0XmrdmaSGcGxCEWT/65NtvYrQtUE5ELxJ+N/aeZNlK2B7IWANnw/82913asXH4VksV1NYNduP0o1/G4XcwLLSyVFB078q/oEnmvdNIoS61j4/o36HVtENJgYr0idcBvwJdvcGxGnPaqOhx477t+kfJAa5n5dSA5wilIaoXH5i1Tf/HsTCM52L+iNCARvQzJYZhzbWI1MDQwzILtIBEQCJsl2XSqIupleY8CxqQ6jCXt2mhae+wPc3YmbO5rFvr2/EvC57kh3yDs1Nsuj8KOvD78KeeujbR8n8pScm3WDp62HFQ8lEKNdeRNj6kB8WnuaJvPnyZfvzOhwG65/9w13IBl7B1sWxbFnq2rMpm5uHVK7mAmjL0Tt8zoDhcE1YJEnp9xte3/pvmKPkST5Q/9ZtR9P5sI+02jY0fvPkPyC03j2gsPixG7rpOCwpOdbny4dcj0TDeeXJX8er+oVfJuLYz0pNWJcT2raDdFfcqvYA0B0IyNYlj5nWX4RuEcyT3qocLReWPnZojetvAG/H8XwOh7fEVGqHAKOVSnPXCSQJPl6s0H12jPJBDJMTydtYPEszl4/CeQ=="},
			},
		},
	}

	tests := []struct {
		name        string
		fields      fields
		wantContent []byte
		wantErr     error
	}{
		{
			name: "kube version not specified",
			fields: fields{
				datacenterSpec:          &clusterSpec.VSphereDatacenter.Spec,
				controlPlaneMachineSpec: vsphereMachineConfig,
			},
			wantErr: fmt.Errorf("error building template map from CP"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			vs := &VsphereTemplateBuilder{
				datacenterSpec:          tt.fields.datacenterSpec,
				controlPlaneMachineSpec: tt.fields.controlPlaneMachineSpec,
			}
			gotContent, err := vs.GenerateCAPISpecControlPlane(clusterSpec)
			if err != tt.wantErr && !assert.Contains(t, err.Error(), tt.wantErr.Error()) {
				t.Errorf("Got VsphereTemplateBuilder.GenerateCAPISpecControlPlane() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				g.Expect(gotContent).NotTo(BeEmpty())
			}
		})
	}
}
