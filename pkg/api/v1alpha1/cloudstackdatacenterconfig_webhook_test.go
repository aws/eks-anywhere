package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCloudStackDatacenterDatacenterConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)

	originalDatacenter := cloudstackDatacenterConfig()
	originalDatacenter.Spec.AvailabilityZones = nil
	originalDatacenter.Spec.Domain = "domain"
	originalDatacenter.Spec.Account = "admin"
	originalDatacenter.Spec.ManagementApiEndpoint = "https://127.0.0.1:8080/client/api"
	originalDatacenter.Spec.Zones = []v1alpha1.CloudStackZone{
		{Name: "test_zone", Network: v1alpha1.CloudStackResourceIdentifier{Name: "test_zone"}},
	}

	expectedDatacenter := originalDatacenter.DeepCopy()
	expectedDatacenter.Spec.AvailabilityZones = []v1alpha1.CloudStackAvailabilityZone{
		{
			Name:                  "default-az-0",
			CredentialsRef:        "global",
			Zone:                  originalDatacenter.Spec.Zones[0],
			Domain:                originalDatacenter.Spec.Domain,
			Account:               originalDatacenter.Spec.Account,
			ManagementApiEndpoint: originalDatacenter.Spec.ManagementApiEndpoint,
		},
	}
	expectedDatacenter.Spec.Zones = nil
	expectedDatacenter.Spec.Domain = ""
	expectedDatacenter.Spec.Account = ""
	expectedDatacenter.Spec.ManagementApiEndpoint = ""

	originalDatacenter.Default()

	g.Expect(originalDatacenter).To(Equal(*expectedDatacenter))
}

func TestCloudStackDatacenterValidateUpdateDomainImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].Domain = "oldCruftyDomain"
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].Domain = "shinyNewDomain"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateV1beta1ToV1beta2Upgrade(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].Name = "default-az-0"
	vNew := vOld.DeepCopy()

	vNew.Spec.AvailabilityZones[0].Name = "12345678-abcd-4abc-abcd-abcd12345678"
	g := NewWithT(t)
	g.Expect(vNew.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCloudStackDatacenterValidateUpdateV1beta1ToV1beta2UpgradeAddAzInvalid(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].Name = "default-az-0"
	vNew := vOld.DeepCopy()

	vNew.Spec.AvailabilityZones[0].Name = "12345678-abcd-4abc-abcd-abcd12345678"
	vNew.Spec.AvailabilityZones = append(vNew.Spec.AvailabilityZones, vNew.Spec.AvailabilityZones[0])
	g := NewWithT(t)
	g.Expect(vNew.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateRenameAzInvalid(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].Name = "default-az-0"
	vNew := vOld.DeepCopy()

	vNew.Spec.AvailabilityZones[0].Name = "shinyNewAzName"
	g := NewWithT(t)
	g.Expect(vNew.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateManagementApiEndpointImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].ManagementApiEndpoint = "oldCruftyManagementApiEndpoint"
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].ManagementApiEndpoint = "shinyNewManagementApiEndpoint"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateZonesImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].Zone.Name = "shinyNewZone"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateAccountImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].Account = "shinyNewAccount"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateNetworkImmutable(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].Zone.Network.Name = "GuestNet2"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).NotTo(Succeed())
}

func TestCloudStackDatacenterValidateUpdateWithPausedAnnotation(t *testing.T) {
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.Zones = []v1alpha1.CloudStackZone{
		{
			Name: "oldCruftyZone",
			Network: v1alpha1.CloudStackResourceIdentifier{
				Name: "GuestNet1",
			},
		},
	}
	c := vOld.DeepCopy()

	c.Spec.Zones = []v1alpha1.CloudStackZone{
		{
			Name: "oldCruftyZone",
			Network: v1alpha1.CloudStackResourceIdentifier{
				Name: "GuestNet2",
			},
		},
	}

	vOld.PauseReconcile()

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&vOld)).To(Succeed())
}

func TestCloudStackDatacenterValidateUpdateInvalidType(t *testing.T) {
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.CloudStackDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(vOld)).NotTo(Succeed())
}

func cloudstackDatacenterConfig() v1alpha1.CloudStackDatacenterConfig {
	return v1alpha1.CloudStackDatacenterConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec: v1alpha1.CloudStackDatacenterConfigSpec{
			AvailabilityZones: []v1alpha1.CloudStackAvailabilityZone{
				{
					Name: "default-az-0",
					Zone: v1alpha1.CloudStackZone{
						Name: "oldCruftyZone",
						Network: v1alpha1.CloudStackResourceIdentifier{
							Name: "GuestNet1",
						},
					},
				},
			},
		},
		Status: v1alpha1.CloudStackDatacenterConfigStatus{},
	}
}
