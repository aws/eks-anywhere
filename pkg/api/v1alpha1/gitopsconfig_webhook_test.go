package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestClusterValidateUpdateGitOpsRepoImmutable(t *testing.T) {
	gOld := gitOpsConfig()
	gOld.Spec.Flux.Github.Repository = "oldRepo"
	c := gOld.DeepCopy()

	c.Spec.Flux.Github.Repository = "fancyNewRepo"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&gOld)).To(MatchError(ContainSubstring("GitOpsConfig: Forbidden: config is immutable")))
}

func TestClusterValidateUpdateGitOpsBranchImmutable(t *testing.T) {
	gOld := gitOpsConfig()
	gOld.Spec.Flux.Github.Branch = "oldMain"
	c := gOld.DeepCopy()

	c.Spec.Flux.Github.Repository = "newMain"
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&gOld)).To(MatchError(ContainSubstring("GitOpsConfig: Forbidden: config is immutable")))
}

func TestClusterValidateUpdateGitOpsSubtractionImmutable(t *testing.T) {
	gOld := gitOpsConfig()
	gOld.Spec.Flux.Github.Repository = "oldRepo"
	c := gOld.DeepCopy()

	c.Spec = v1alpha1.GitOpsConfigSpec{}
	g := NewWithT(t)
	g.Expect(c.ValidateUpdate(&gOld)).To(MatchError(ContainSubstring("GitOpsConfig: Forbidden: config is immutable")))
}

func gitOpsConfig() v1alpha1.GitOpsConfig {
	return v1alpha1.GitOpsConfig{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string, 1)},
		Spec:       v1alpha1.GitOpsConfigSpec{},
		Status:     v1alpha1.GitOpsConfigStatus{},
	}
}
