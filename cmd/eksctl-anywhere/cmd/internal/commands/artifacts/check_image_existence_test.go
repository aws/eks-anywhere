package artifacts_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/internal/commands/artifacts"
)

type checkImageExistenceTest struct {
	*WithT
	ctx     context.Context
	command *artifacts.CheckImageExistence
}

func newCheckImageExistenceTest(t *testing.T) *checkImageExistenceTest {
	return &checkImageExistenceTest{
		WithT:   NewWithT(t),
		ctx:     context.Background(),
		command: &artifacts.CheckImageExistence{},
	}
}

func TestCheckImageExistenceRun(t *testing.T) {
	tt := newCheckImageExistenceTest(t)

	// Default image URI
	tt.command.ImageUri = "public.ecr.aws/bottlerocket/bottlerocket-admin:v0.8.0"
	tt.Expect(tt.command.Run(tt.ctx)).To(Succeed())

	// Mirrored image URI
	tt.command.ImageUri = "public.ecr.aws:443/bottlerocket/bottlerocket-admin:v0.8.0"
	tt.Expect(tt.command.Run(tt.ctx)).To(Succeed())

	// Nonexisting mirrored image URI
	tt.command.ImageUri = "public.ecr.aws:443/xxx"
	tt.Expect(tt.command.Run(tt.ctx)).NotTo(Succeed())

	// Invalid URI #1
	tt.command.ImageUri = ""
	tt.Expect(tt.command.Run(tt.ctx)).NotTo(Succeed())

	// Invalid URI #2
	tt.command.ImageUri = "public.ecr.aws/"
	tt.Expect(tt.command.Run(tt.ctx)).NotTo(Succeed())

	// Invalid URI #3
	tt.command.ImageUri = "public.ecr.aws:443/"
	tt.Expect(tt.command.Run(tt.ctx)).NotTo(Succeed())
}
