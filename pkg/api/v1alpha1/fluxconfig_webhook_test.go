package v1alpha1_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestClusterValidateUpdateFluxRepoImmutable(t *testing.T) {
	ctx := context.Background()
	fOld := fluxConfig()
	fOld.Spec.Github = &v1alpha1.GithubProviderConfig{
		Repository: "oldRepo",
	}
	c := fOld.DeepCopy()

	c.Spec.Github.Repository = "fancyNewRepo"
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(ctx, &fOld, c)).Error().To(MatchError(ContainSubstring("Forbidden: config is immutable")))
}

func TestClusterValidateUpdateFluxRepoUrlImmutable(t *testing.T) {
	ctx := context.Background()
	fOld := fluxConfig()
	fOld.Spec.Git = &v1alpha1.GitProviderConfig{
		RepositoryUrl: "https://test.git/test",
	}
	c := fOld.DeepCopy()

	c.Spec.Git.RepositoryUrl = "https://test.git/test2"
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(ctx, &fOld, c)).Error().To(MatchError(ContainSubstring("Forbidden: config is immutable")))
}

func TestClusterValidateUpdateFluxSshKeyAlgoImmutable(t *testing.T) {
	ctx := context.Background()
	fOld := fluxConfig()
	fOld.Spec.Git = &v1alpha1.GitProviderConfig{
		RepositoryUrl:   "https://test.git/test",
		SshKeyAlgorithm: "rsa",
	}
	c := fOld.DeepCopy()

	c.Spec.Git.SshKeyAlgorithm = "rsa2"
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(ctx, &fOld, c)).Error().To(MatchError(ContainSubstring("Forbidden: config is immutable")))
}

func TestClusterValidateUpdateFluxBranchImmutable(t *testing.T) {
	ctx := context.Background()
	fOld := fluxConfig()
	fOld.Spec.Branch = "oldMain"
	c := fOld.DeepCopy()

	c.Spec.Branch = "newMain"
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(ctx, &fOld, c)).Error().To(MatchError(ContainSubstring("Forbidden: config is immutable")))
}

func TestClusterValidateUpdateFluxSubtractionImmutable(t *testing.T) {
	ctx := context.Background()
	fOld := fluxConfig()
	fOld.Spec.Github = &v1alpha1.GithubProviderConfig{
		Repository: "oldRepo",
	}
	c := fOld.DeepCopy()

	c.Spec = v1alpha1.FluxConfigSpec{}
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(ctx, &fOld, c)).Error().To(MatchError(ContainSubstring("Forbidden: config is immutable")))
}

func TestValidateCreateHasValidatedSpec(t *testing.T) {
	ctx := context.Background()
	fNew := fluxConfig()
	fNew.Spec.Git = &v1alpha1.GitProviderConfig{}
	fNew.Spec.Github = &v1alpha1.GithubProviderConfig{}

	f := NewWithT(t)
	warnings, err := fNew.ValidateCreate(ctx, &fNew)
	f.Expect(warnings).To(BeEmpty())

	f.Expect(apierrors.IsInvalid(err)).Error().To(BeTrue())
	f.Expect(err).To(MatchError(ContainSubstring("must specify only one provider")))
}

func TestValidateUpdateHasValidatedSpec(t *testing.T) {
	ctx := context.Background()
	fOld := fluxConfig()
	fOld.Spec.Github = &v1alpha1.GithubProviderConfig{
		Repository: "oldRepo",
	}
	c := fOld.DeepCopy()
	c.Spec.Git = &v1alpha1.GitProviderConfig{}

	f := NewWithT(t)
	warnings, err := c.ValidateUpdate(ctx, &fOld, c)
	f.Expect(warnings).To(BeEmpty())
	f.Expect(apierrors.IsInvalid(err)).Error().To(BeTrue())
	f.Expect(err).To(MatchError(ContainSubstring("must specify only one provider")))
}

func fluxConfig() v1alpha1.FluxConfig {
	return v1alpha1.FluxConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.FluxConfigSpec{},
		Status:     v1alpha1.FluxConfigStatus{},
	}
}

func TestFluxConfigValidateCreateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.FluxConfig{}

	// Call ValidateCreate with the wrong type
	warnings, err := config.ValidateCreate(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a FluxConfig"))
}

func TestFluxConfigValidateUpdateCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.FluxConfig{}

	// Call ValidateUpdate with the wrong type
	warnings, err := config.ValidateUpdate(context.TODO(), wrongType, &v1alpha1.FluxConfig{})

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a FluxConfig"))
}

func TestFluxConfigValidateDeleteCastFail(t *testing.T) {
	g := NewWithT(t)

	// Create a different type that will cause the cast to fail
	wrongType := &v1alpha1.Cluster{}

	// Create the config object that implements CustomValidator
	config := &v1alpha1.FluxConfig{}

	// Call ValidateDelete with the wrong type
	warnings, err := config.ValidateDelete(context.TODO(), wrongType)

	// Verify that an error is returned
	g.Expect(warnings).To(BeNil())
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("expected a FluxConfig"))
}
