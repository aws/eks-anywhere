package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestNutanixDatacenterConfigWebhooksValidateCreate(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.ValidateCreate(ctx, dcConf)).Error().To(Succeed())
}

func TestNutanixDatacenterConfigWebhooksValidateCreateReconcilePaused(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	dcConf.Annotations = map[string]string{
		"anywhere.eks.amazonaws.com/paused": "true",
	}
	g.Expect(dcConf.ValidateCreate(ctx, dcConf)).Error().To(Succeed())
}

func TestNutanixDatacenterConfigWebhookValidateCreateNoCredentialRef(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	dcConf.Spec.CredentialRef = nil
	_, err := dcConf.ValidateCreate(ctx, dcConf)
	g.Expect(err).To(MatchError(ContainSubstring("credentialRef is required to be set to create a new NutanixDatacenterConfig")))
}

func TestNutanixDatacenterConfigWebhooksValidateCreateValidaitonFailure(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	dcConf.Spec.Endpoint = ""
	g.Expect(dcConf.ValidateCreate(ctx, dcConf)).Error().To(Not(Succeed()))
}

func TestNutanixDatacenterConfigWebhooksValidateUpdate(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.ValidateCreate(ctx, dcConf)).Error().To(Succeed())
	newSpec := nutanixDatacenterConfig()
	newSpec.Spec.CredentialRef.Name = "new-credential"
	g.Expect(dcConf.ValidateUpdate(ctx, dcConf, newSpec)).Error().To(Succeed())
}

func TestNutanixDatacenterConfigWebhooksValidateUpdateReconcilePaused(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.ValidateCreate(ctx, dcConf)).Error().To(Succeed())
	oldSpec := nutanixDatacenterConfig()
	oldSpec.Annotations = map[string]string{
		"anywhere.eks.amazonaws.com/paused": "true",
	}
	oldSpec.Spec.CredentialRef.Name = "new-credential"
	g.Expect(dcConf.ValidateUpdate(ctx, dcConf, oldSpec)).Error().To(Succeed())
}

func TestNutanixDatacenterConfigWebhooksValidateUpdateValidationFailure(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.ValidateCreate(ctx, dcConf)).Error().To(Succeed())
	newSpec := nutanixDatacenterConfig()
	newSpec.Spec.Endpoint = ""
	g.Expect(dcConf.ValidateUpdate(ctx, dcConf, newSpec)).Error().To(Not(Succeed()))
}

func TestNutanixDatacenterConfigWebhooksValidateUpdateInvalidOldObject(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	newConf := nutanixDatacenterConfig()
	newConf.Spec.CredentialRef = nil
	_, err := newConf.ValidateUpdate(ctx, newConf, &v1alpha1.NutanixMachineConfig{})
	g.Expect(err).To(MatchError(ContainSubstring("expected a NutanixDatacenterConfig but got a *v1alpha1.NutanixMachineConfig")))
}

func TestNutanixDatacenterConfigWebhooksValidateUpdateCredentialRefRemoved(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	oldConf := nutanixDatacenterConfig()
	g.Expect(oldConf.ValidateCreate(ctx, oldConf)).Error().To(Succeed())
	newConf := nutanixDatacenterConfig()
	newConf.Spec.CredentialRef = nil
	_, err := newConf.ValidateUpdate(ctx, newConf, oldConf)
	g.Expect(err).To(MatchError(ContainSubstring("credentialRef cannot be removed from an existing NutanixDatacenterConfig")))
}

func TestNutanixDatacenterConfigWebhooksValidateDelete(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.ValidateCreate(ctx, dcConf)).Error().To(Succeed())
	g.Expect(dcConf.ValidateDelete(ctx, dcConf)).Error().To(Succeed())
}

func nutanixDatacenterConfig() *v1alpha1.NutanixDatacenterConfig {
	return &v1alpha1.NutanixDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nutanix-datacenter-config",
		},
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "prism.nutanix.com",
			Port:     constants.DefaultNutanixPrismCentralPort,
			CredentialRef: &v1alpha1.Ref{
				Kind: constants.SecretKind,
				Name: constants.NutanixCredentialsName,
			},
		},
	}
}

func TestNutanixDatacenterConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.NutanixDatacenterConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a NutanixDatacenterConfig"))
}

func TestNutanixDatacenterConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.NutanixDatacenterConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), wrongType, &v1alpha1.NutanixDatacenterConfig{})

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a NutanixDatacenterConfig"))
}

func TestNutanixDatacenterConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.NutanixDatacenterConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a NutanixDatacenterConfig"))
}
