package v1alpha1_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test/envtest"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestNutanixDatacenterConfigWebhooksValidateCreate(t *testing.T) {
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.ValidateCreate()).To(Succeed())
}

func TestNutanixDatacenterConfigWebhookValidateCreateNoCredentialRef(t *testing.T) {
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	dcConf.Spec.CredentialRef = nil
	g.Expect(dcConf.ValidateCreate()).To(MatchError("credentialRef is required to be set to create a new NutanixDatacenterConfig"))
}

func TestNutanixDatacenterConfigWebhooksValidateUpdate(t *testing.T) {
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.ValidateCreate()).To(Succeed())
	newSpec := nutanixDatacenterConfig()
	newSpec.Spec.CredentialRef.Name = "new-credential"
	g.Expect(dcConf.ValidateUpdate(newSpec)).To(Succeed())
}

func TestNutanixDatacenterConfigWebhooksValidateUpdateInvalidOldObject(t *testing.T) {
	g := NewWithT(t)
	newConf := nutanixDatacenterConfig()
	newConf.Spec.CredentialRef = nil
	g.Expect(newConf.ValidateUpdate(&v1alpha1.NutanixMachineConfig{})).To(MatchError("old object is not a NutanixDatacenterConfig"))
}

func TestNutanixDatacenterConfigWebhooksValidateUpdateCredentialRefRemoved(t *testing.T) {
	g := NewWithT(t)
	oldConf := nutanixDatacenterConfig()
	g.Expect(oldConf.ValidateCreate()).To(Succeed())
	newConf := nutanixDatacenterConfig()
	newConf.Spec.CredentialRef = nil
	g.Expect(newConf.ValidateUpdate(oldConf)).To(MatchError("credentialRef cannot be removed from an existing NutanixDatacenterConfig"))
}

func TestNutanixDatacenterConfigWebhooksValidateDelete(t *testing.T) {
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.ValidateCreate()).To(Succeed())
	g.Expect(dcConf.ValidateDelete()).To(Succeed())
}

func TestNutanixDatacenterConfigSetupWebhookWithManager(t *testing.T) {
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	g.Expect(dcConf.SetupWebhookWithManager(env.Manager())).To(Succeed())
}

func nutanixDatacenterConfig() *v1alpha1.NutanixDatacenterConfig {
	return &v1alpha1.NutanixDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nutanix-datacenter-config",
		},
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "prism.nutanix.com",
			Port:     9440,
			CredentialRef: &v1alpha1.Ref{
				Kind: constants.SecretKind,
				Name: constants.NutanixCredentialsName,
			},
		},
	}
}

var env *envtest.Environment

func TestMain(m *testing.M) {
	os.Exit(envtest.RunWithEnvironment(m, envtest.WithAssignment(&env)))
}
