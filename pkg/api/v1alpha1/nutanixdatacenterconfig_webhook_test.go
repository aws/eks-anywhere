package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestNutanixDatacenterConfigWebhooks(t *testing.T) {
	g := NewWithT(t)
	dcConf := nutanixDatacenterConfig()
	dcConf.Default()
	g.Expect(dcConf.Spec.CredentialRef.Kind).To(Equal(constants.SecretKind))
	g.Expect(dcConf.Spec.CredentialRef.Name).To(Equal(constants.NutanixCredentialsName))
	g.Expect(dcConf.ValidateCreate()).To(Succeed())
	newConf := dcConf.DeepCopy()
	newConf.Spec.CredentialRef.Name = "new-creds"
	g.Expect(dcConf.ValidateUpdate(dcConf)).To(Succeed())
	g.Expect(newConf.ValidateDelete()).To(Succeed())
}

func nutanixDatacenterConfig() *NutanixDatacenterConfig {
	return &NutanixDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nutanix-datacenter-config",
		},
		Spec: NutanixDatacenterConfigSpec{
			Endpoint: "prism.nutanix.com",
			Port:     9440,
		},
	}
}
