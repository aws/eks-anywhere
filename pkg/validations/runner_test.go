package validations_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/validations"
)

func TestRunnerRunError(t *testing.T) {
	g := NewWithT(t)
	r := validations.NewRunner()
	r.Register(func() *validations.ValidationResult {
		return &validations.ValidationResult{
			Err: nil,
		}
	})
	r.Register(func() *validations.ValidationResult {
		return &validations.ValidationResult{
			Err: errors.New("failed"),
		}
	})

	err := r.Run()
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(Equal("validations failed"))
}

func TestRunnerRunSuccess(t *testing.T) {
	g := NewWithT(t)
	r := validations.NewRunner()
	r.Register(func() *validations.ValidationResult {
		return &validations.ValidationResult{
			Err: nil,
		}
	})
	r.Register(func() *validations.ValidationResult {
		return &validations.ValidationResult{
			Err: nil,
		}
	})

	g.Expect(r.Run()).To(Succeed())
}
