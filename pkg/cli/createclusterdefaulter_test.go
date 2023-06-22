package cli_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/cli"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/defaulting"
)

func TestNewCreateClusterDefaulter(t *testing.T) {
	g := NewWithT(t)

	skipIPCheck := cluster.NewControlPlaneIPCheckAnnotationDefaulter(false)

	r := defaulting.NewRunner[*cluster.Spec]()
	r.Register(
		skipIPCheck.ControlPlaneIPCheckDefault,
	)

	got := cli.NewCreateClusterDefaulter(skipIPCheck)

	g.Expect(got).NotTo(BeNil())
}
