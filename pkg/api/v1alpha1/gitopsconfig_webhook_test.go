package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestClusterValidateUpdateGitOpsRepoImmutable(t *testing.T) {
	ctx := context.Background()
	gOld := gitOpsConfig()
	gOld.Spec.Flux.Github.Repository = "oldRepo"
	c := gOld.DeepCopy()

	c.Spec.Flux.Github.Repository = "fancyNewRepo"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, &gOld)).Error().To(MatchError(ContainSubstring("GitOpsConfig: Forbidden: config is immutable")))
}

func TestClusterValidateUpdateGitOpsBranchImmutable(t *testing.T) {
	ctx := context.Background()
	gOld := gitOpsConfig()
	gOld.Spec.Flux.Github.Branch = "oldMain"
	c := gOld.DeepCopy()

	c.Spec.Flux.Github.Repository = "newMain"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, &gOld)).Error().To(MatchError(ContainSubstring("GitOpsConfig: Forbidden: config is immutable")))
}

func TestClusterValidateUpdateGitOpsSubtractionImmutable(t *testing.T) {
	ctx := context.Background()
	gOld := gitOpsConfig()
	gOld.Spec.Flux.Github.Repository = "oldRepo"
	c := gOld.DeepCopy()

	c.Spec = v1alpha1.GitOpsConfigSpec{}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(ctx, c, &gOld)).Error().To(MatchError(ContainSubstring("GitOpsConfig: Forbidden: config is immutable")))
}

func gitOpsConfig() v1alpha1.GitOpsConfig {
	return v1alpha1.GitOpsConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.GitOpsConfigSpec{},
		Status:     v1alpha1.GitOpsConfigStatus{},
	}
}

func TestGitOpsConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.GitOpsConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a GitOpsConfig"))
}

func TestGitOpsConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.GitOpsConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), wrongType, &v1alpha1.GitOpsConfig{})

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a GitOpsConfig"))
}

func TestGitOpsConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.GitOpsConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a GitOpsConfig"))
}
