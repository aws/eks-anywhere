package tinkerbell_test

import (
	"errors"
	"testing"

	"github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
)

func TestClusterSpecValidator_AssertionsWithoutError(t *testing.T) {
	g := gomega.NewWithT(t)
	assertions := []*MockAssertion{{}, {}, {}}

	validator := tinkerbell.NewClusterSpecValidator()
	for _, assertion := range assertions {
		validator.Register(assertion.ClusterSpecAsseritonFunc())
	}

	g.Expect(validator.Validate(NewDefaultValidClusterSpecBuilder().Build())).To(gomega.Succeed())
	for _, assertion := range assertions {
		g.Expect(assertion.Called).To(gomega.BeTrue())
	}
}

func TestClusterSpecValidator_AssertionsWithError(t *testing.T) {
	g := gomega.NewWithT(t)
	assertions := []*MockAssertion{{}, {}, {}}
	assertions[0].Return = errors.New("assertion error")

	validator := tinkerbell.NewClusterSpecValidator()
	for _, assertion := range assertions {
		validator.Register(assertion.ClusterSpecAsseritonFunc())
	}

	g.Expect(validator.Validate(NewDefaultValidClusterSpecBuilder().Build())).ToNot(gomega.Succeed())
	g.Expect(assertions[0].Called).To(gomega.BeTrue())
	g.Expect(assertions[1].Called).To(gomega.BeFalse())
	g.Expect(assertions[1].Called).To(gomega.BeFalse())
}

type MockAssertion struct {
	Called bool
	Return error
}

func (a *MockAssertion) ClusterSpecAsseritonFunc() tinkerbell.ClusterSpecAssertion {
	return func(*tinkerbell.ClusterSpec) error {
		a.Called = true
		return a.Return
	}
}
