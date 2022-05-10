package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestClusterValidateUpdateFluxRepoImmutable(t *testing.T) {
	fOld := fluxConfig()
	fOld.Spec.Github = &v1alpha1.GithubProviderConfig{
		Repository: "oldRepo",
	}
	c := fOld.DeepCopy()

	c.Spec.Github.Repository = "fancyNewRepo"
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(&fOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateFluxRepoUrlImmutable(t *testing.T) {
	fOld := fluxConfig()
	fOld.Spec.Git = &v1alpha1.GitProviderConfig{
		RepositoryUrl: "https://test.git/test",
	}
	c := fOld.DeepCopy()

	c.Spec.Git.RepositoryUrl = "https://test.git/test2"
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(&fOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateFluxSshKeyAlgoImmutable(t *testing.T) {
	fOld := fluxConfig()
	fOld.Spec.Git = &v1alpha1.GitProviderConfig{
		RepositoryUrl:   "https://test.git/test",
		SshKeyAlgorithm: "rsa",
	}
	c := fOld.DeepCopy()

	c.Spec.Git.SshKeyAlgorithm = "rsa2"
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(&fOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateFluxBranchImmutable(t *testing.T) {
	fOld := fluxConfig()
	fOld.Spec.Branch = "oldMain"
	c := fOld.DeepCopy()

	c.Spec.Branch = "newMain"
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(&fOld)).NotTo(Succeed())
}

func TestClusterValidateUpdateFluxSubtractionImmutable(t *testing.T) {
	fOld := fluxConfig()
	fOld.Spec.Github = &v1alpha1.GithubProviderConfig{
		Repository: "oldRepo",
	}
	c := fOld.DeepCopy()

	c.Spec = v1alpha1.FluxConfigSpec{}
	f := NewWithT(t)
	f.Expect(c.ValidateUpdate(&fOld)).NotTo(Succeed())
}

func fluxConfig() v1alpha1.FluxConfig {
	return v1alpha1.FluxConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.FluxConfigSpec{},
		Status:     v1alpha1.FluxConfigStatus{},
	}
}
