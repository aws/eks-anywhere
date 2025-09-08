package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestControlPlaneConfigurationEqualSkipAdmissionBothNil(t *testing.T) {
	g := NewWithT(t)

	config1 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: nil,
	}
	config2 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: nil,
	}

	g.Expect(config1.Equal(config2)).To(BeTrue())
}

func TestControlPlaneConfigurationEqualSkipAdmissionBothTrue(t *testing.T) {
	g := NewWithT(t)

	config1 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: ptr.Bool(true),
	}
	config2 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: ptr.Bool(true),
	}

	g.Expect(config1.Equal(config2)).To(BeTrue())
}

func TestControlPlaneConfigurationEqualSkipAdmissionBothFalse(t *testing.T) {
	g := NewWithT(t)

	config1 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: ptr.Bool(false),
	}
	config2 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: ptr.Bool(false),
	}

	g.Expect(config1.Equal(config2)).To(BeTrue())
}

func TestControlPlaneConfigurationNotEqualSkipAdmissionOneNilOneTrue(t *testing.T) {
	g := NewWithT(t)

	config1 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: nil,
	}
	config2 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: ptr.Bool(true),
	}

	g.Expect(config1.Equal(config2)).To(BeFalse())
}

func TestControlPlaneConfigurationNotEqualSkipAdmissionOneNilOneFalse(t *testing.T) {
	g := NewWithT(t)

	config1 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: nil,
	}
	config2 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: ptr.Bool(false),
	}

	g.Expect(config1.Equal(config2)).To(BeFalse())
}

func TestControlPlaneConfigurationNotEqualSkipAdmissionDifferentValues(t *testing.T) {
	g := NewWithT(t)

	config1 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: ptr.Bool(true),
	}
	config2 := &ControlPlaneConfiguration{
		Count:                           3,
		SkipAdmissionForSystemResources: ptr.Bool(false),
	}

	g.Expect(config1.Equal(config2)).To(BeFalse())
}

func TestControlPlaneConfigurationEqualSkipAdmissionWithOtherFields(t *testing.T) {
	g := NewWithT(t)

	config1 := &ControlPlaneConfiguration{
		Count: 3,
		MachineGroupRef: &Ref{
			Kind: "VSphereMachineConfig",
			Name: "test-cp",
		},
		SkipAdmissionForSystemResources: ptr.Bool(true),
		Labels: map[string]string{
			"env": "test",
		},
	}
	config2 := &ControlPlaneConfiguration{
		Count: 3,
		MachineGroupRef: &Ref{
			Kind: "VSphereMachineConfig",
			Name: "test-cp",
		},
		SkipAdmissionForSystemResources: ptr.Bool(true),
		Labels: map[string]string{
			"env": "test",
		},
	}

	g.Expect(config1.Equal(config2)).To(BeTrue())
}
