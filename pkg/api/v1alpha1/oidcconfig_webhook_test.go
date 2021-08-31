package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestValidateUpdateOIDCClientId(t *testing.T) {
	ocOld := oidcConfig()
	ocOld.Spec.ClientId = "test"
	c := ocOld.DeepCopy()

	c.Spec.ClientId = "test2"
	o := NewWithT(t)
	o.Expect(c.ValidateUpdate(&ocOld)).NotTo(Succeed())
}

func TestValidateUpdateOIDCGroupsClaim(t *testing.T) {
	ocOld := oidcConfig()
	ocOld.Spec.GroupsClaim = "test"
	c := ocOld.DeepCopy()

	c.Spec.GroupsClaim = "test2"
	o := NewWithT(t)
	o.Expect(c.ValidateUpdate(&ocOld)).NotTo(Succeed())
}

func TestValidateUpdateOIDCGroupsPrefix(t *testing.T) {
	ocOld := oidcConfig()
	ocOld.Spec.GroupsPrefix = "test"
	c := ocOld.DeepCopy()

	c.Spec.GroupsPrefix = "test2"
	o := NewWithT(t)
	o.Expect(c.ValidateUpdate(&ocOld)).NotTo(Succeed())
}

func TestValidateUpdateOIDCIssuerUrl(t *testing.T) {
	ocOld := oidcConfig()
	ocOld.Spec.IssuerUrl = "test"
	c := ocOld.DeepCopy()

	c.Spec.IssuerUrl = "test2"
	o := NewWithT(t)
	o.Expect(c.ValidateUpdate(&ocOld)).NotTo(Succeed())
}

func TestValidateUpdateOIDCUsernameClaim(t *testing.T) {
	ocOld := oidcConfig()
	ocOld.Spec.UsernameClaim = "test"
	c := ocOld.DeepCopy()

	c.Spec.UsernameClaim = "test2"
	o := NewWithT(t)
	o.Expect(c.ValidateUpdate(&ocOld)).NotTo(Succeed())
}

func TestValidateUpdateOIDCUsernamePrefix(t *testing.T) {
	ocOld := oidcConfig()
	ocOld.Spec.UsernamePrefix = "test"
	c := ocOld.DeepCopy()

	c.Spec.UsernamePrefix = "test2"
	o := NewWithT(t)
	o.Expect(c.ValidateUpdate(&ocOld)).NotTo(Succeed())
}

func TestValidateUpdateOIDCRequiredClaims(t *testing.T) {
	ocOld := oidcConfig()
	ocOld.Spec.RequiredClaims = []v1alpha1.OIDCConfigRequiredClaim{{Claim: "test", Value: "value"}}
	c := ocOld.DeepCopy()

	c.Spec.RequiredClaims = []v1alpha1.OIDCConfigRequiredClaim{{Claim: "test", Value: "value2"}}
	o := NewWithT(t)
	o.Expect(c.ValidateUpdate(&ocOld)).NotTo(Succeed())
}

func TestValidateUpdateOIDCRequiredClaimsMultiple(t *testing.T) {
	ocOld := oidcConfig()
	ocOld.Spec.RequiredClaims = []v1alpha1.OIDCConfigRequiredClaim{{Claim: "test", Value: "value"}}
	c := ocOld.DeepCopy()

	c.Spec.RequiredClaims = append(c.Spec.RequiredClaims, v1alpha1.OIDCConfigRequiredClaim{
		Claim: "test2",
		Value: "value2",
	})
	o := NewWithT(t)
	o.Expect(c.ValidateUpdate(&ocOld)).NotTo(Succeed())
}

func oidcConfig() v1alpha1.OIDCConfig {
	return v1alpha1.OIDCConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.OIDCConfigSpec{},
		Status:     v1alpha1.OIDCConfigStatus{},
	}
}
