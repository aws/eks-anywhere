package cli_test

import (
	"testing"

	"github.com/go-logr/logr/funcr"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/cli"
)

func TestValidationPassed(t *testing.T) {
	g := NewWithT(t)

	logged := ""
	log := funcr.New(func(prefix, args string) {
		logged += prefix + args
	}, funcr.Options{})

	cli.ValidationPassed(log, "My message")
	g.Expect(logged).To(Equal("\"level\"=0 \"msg\"=\"✅ My message\""))
}

func TestValidationFailed(t *testing.T) {
	g := NewWithT(t)

	logged := ""
	log := funcr.New(func(prefix, args string) {
		logged += prefix + args
	}, funcr.Options{})

	cli.ValidationFailed(log, "My message")
	g.Expect(logged).To(Equal("\"level\"=0 \"msg\"=\"❌ My message\""))
}
