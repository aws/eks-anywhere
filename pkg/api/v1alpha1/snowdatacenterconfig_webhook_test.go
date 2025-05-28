package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestSnowDatacenterConfigValidateCreateValid(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	snowDC := snowDatacenterConfig()
	snowDC.Spec.IdentityRef.Name = "refName"
	snowDC.Spec.IdentityRef.Kind = v1alpha1.SnowIdentityKind

	g.Expect(snowDC.ValidateCreate(ctx, &snowDC)).Error().To(Succeed())
}

func TestSnowDatacenterConfigValidateCreateEmptyIdentityRef(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	snowDC := snowDatacenterConfig()

	g.Expect(snowDC.ValidateCreate(ctx, &snowDC)).Error().To(MatchError(ContainSubstring("IdentityRef name must not be empty")))
}

func TestSnowDatacenterConfigValidateCreateEmptyIdentityKind(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	snowDC := snowDatacenterConfig()
	snowDC.Spec.IdentityRef.Name = "refName"

	g.Expect(snowDC.ValidateCreate(ctx, &snowDC)).Error().To(MatchError(ContainSubstring("IdentityRef kind must not be empty")))
}

func TestSnowDatacenterConfigValidateCreateIdentityKindNotSnow(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	snowDC := snowDatacenterConfig()
	snowDC.Spec.IdentityRef.Name = "refName"
	snowDC.Spec.IdentityRef.Kind = v1alpha1.OIDCConfigKind

	g.Expect(snowDC.ValidateCreate(ctx, &snowDC)).Error().To(MatchError(ContainSubstring("is invalid, the only supported kind is Secret")))
}

func TestSnowDatacenterConfigValidateValidateEmptyIdentityRef(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	snowDCOld := snowDatacenterConfig()
	snowDCNew := snowDCOld.DeepCopy()
	g.Expect(snowDCNew.ValidateUpdate(ctx, &snowDCOld, snowDCNew)).Error().To(MatchError(ContainSubstring("IdentityRef name must not be empty")))
}

func TestSnowDatacenterConfigValidateValidateEmptyIdentityKind(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	snowDCOld := snowDatacenterConfig()
	snowDCNew := snowDCOld.DeepCopy()
	snowDCNew.Spec.IdentityRef.Name = "refName"

	g.Expect(snowDCNew.ValidateUpdate(ctx, &snowDCOld, snowDCNew)).Error().To(MatchError(ContainSubstring("IdentityRef kind must not be empty")))
}

func TestSnowDatacenterConfigValidateValidateIdentityKindNotSnow(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	snowDCOld := snowDatacenterConfig()
	snowDCNew := snowDCOld.DeepCopy()
	snowDCNew.Spec.IdentityRef.Name = "refName"
	snowDCNew.Spec.IdentityRef.Kind = v1alpha1.OIDCConfigKind

	g.Expect(snowDCNew.ValidateUpdate(ctx, &snowDCOld, snowDCNew)).Error().To(MatchError(ContainSubstring("is invalid, the only supported kind is Secret")))
}

func snowDatacenterConfig() v1alpha1.SnowDatacenterConfig {
	return v1alpha1.SnowDatacenterConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 2)},
		Spec:       v1alpha1.SnowDatacenterConfigSpec{},
		Status:     v1alpha1.SnowDatacenterConfigStatus{},
	}
}

func TestSnowDatacenterConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.SnowDatacenterConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a SnowDatacenterConfig"))
}

func TestSnowDatacenterConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.SnowDatacenterConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), &v1alpha1.SnowDatacenterConfig{}, wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a SnowDatacenterConfig"))
}

func TestSnowDatacenterConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.SnowDatacenterConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a SnowDatacenterConfig"))
}
