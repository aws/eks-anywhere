package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCloudStackDatacenterDatacenterConfigSetDefaults(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

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

	// Using the new CustomDefaulter interface
	err := originalDatacenter.Default(ctx, &originalDatacenter)
	g.Expect(err).To(BeNil())

	g.Expect(originalDatacenter).To(Equal(*expectedDatacenter))
}

func TestCloudStackDatacenterValidateUpdateDomainImmutable(t *testing.T) {
	ctx := context.Background()
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].Domain = "oldCruftyDomain"
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].Domain = "shinyNewDomain"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, &vOld)).Error().To(HaveOccurred())
}

func TestCloudStackDatacenterValidateUpdateV1beta1ToV1beta2Upgrade(t *testing.T) {
	ctx := context.Background()
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].Name = "default-az-0"
	vNew := vOld.DeepCopy()

	vNew.Spec.AvailabilityZones[0].Name = "12345678-abcd-4abc-abcd-abcd12345678"
	g := NewWithT(t)
	g.Expect(vNew.ValidateUpdate(ctx, vNew, &vOld)).Error().To(Succeed())
}

func TestCloudStackDatacenterValidateUpdateV1beta1ToV1beta2UpgradeAddAzInvalid(t *testing.T) {
	ctx := context.Background()
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].Name = "default-az-0"
	vNew := vOld.DeepCopy()

	vNew.Spec.AvailabilityZones[0].Name = "12345678-abcd-4abc-abcd-abcd12345678"
	vNew.Spec.AvailabilityZones = append(vNew.Spec.AvailabilityZones, vNew.Spec.AvailabilityZones[0])
	g := NewWithT(t)
	g.Expect(vNew.ValidateUpdate(ctx, vNew, &vOld)).Error().To(HaveOccurred())
}

func TestCloudStackDatacenterValidateUpdateRenameAzInvalid(t *testing.T) {
	ctx := context.Background()
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].Name = "default-az-0"
	vNew := vOld.DeepCopy()

	vNew.Spec.AvailabilityZones[0].Name = "shinyNewAzName"
	g := NewWithT(t)
	g.Expect(vNew.ValidateUpdate(ctx, vNew, &vOld)).Error().To(HaveOccurred())
}

func TestCloudStackDatacenterValidateUpdateManagementApiEndpointImmutable(t *testing.T) {
	ctx := context.Background()
	vOld := cloudstackDatacenterConfig()
	vOld.Spec.AvailabilityZones[0].ManagementApiEndpoint = "oldCruftyManagementApiEndpoint"
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].ManagementApiEndpoint = "shinyNewManagementApiEndpoint"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, &vOld)).Error().To(HaveOccurred())
}

func TestCloudStackDatacenterValidateUpdateZonesImmutable(t *testing.T) {
	ctx := context.Background()
	vOld := cloudstackDatacenterConfig()
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].Zone.Name = "shinyNewZone"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, &vOld)).Error().To(HaveOccurred())
}

func TestCloudStackDatacenterValidateUpdateAccountImmutable(t *testing.T) {
	ctx := context.Background()
	vOld := cloudstackDatacenterConfig()
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].Account = "shinyNewAccount"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, &vOld)).Error().To(HaveOccurred())
}

func TestCloudStackDatacenterValidateUpdateNetworkImmutable(t *testing.T) {
	ctx := context.Background()
	vOld := cloudstackDatacenterConfig()
	c := vOld.DeepCopy()

	c.Spec.AvailabilityZones[0].Zone.Network.Name = "GuestNet2"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, &vOld)).Error().To(HaveOccurred())
}

func TestCloudStackDatacenterValidateUpdateWithPausedAnnotation(t *testing.T) {
	ctx := context.Background()
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
	g.Expect(c.ValidateUpdate(ctx, c, &vOld)).Error().To(Succeed())
}

func TestCloudStackDatacenterValidateUpdateInvalidType(t *testing.T) {
	ctx := context.Background()
	vOld := &v1alpha1.Cluster{}
	c := &v1alpha1.CloudStackDatacenterConfig{}

	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, vOld)).Error().To(HaveOccurred())
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

func TestCloudStackDatacenterConfigDefaultCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomDefaulter
	config := &v1alpha1.CloudStackDatacenterConfig{}

	// Call Default with the wrong type
	err := config.Default(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a CloudStackDatacenterConfig"))
}

func TestCloudStackDatacenterConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.CloudStackDatacenterConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a CloudStackDatacenterConfig"))
}

func TestCloudStackDatacenterConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.CloudStackDatacenterConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), wrongType, &v1alpha1.CloudStackDatacenterConfig{})

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a CloudStackDatacenterConfig"))
}

func TestCloudStackDatacenterConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.CloudStackDatacenterConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a CloudStackDatacenterConfig"))
}
