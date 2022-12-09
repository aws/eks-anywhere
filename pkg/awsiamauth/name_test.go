package awsiamauth_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/awsiamauth"
)

func TestKubeconfigSecretName(t *testing.T) {
	g := NewWithT(t)
	g.Expect(awsiamauth.KubeconfigSecretName("my-cluster")).To(Equal("my-cluster-aws-iam-kubeconfig"))
}
