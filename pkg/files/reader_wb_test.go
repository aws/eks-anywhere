package files

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestWithEKSAUserAgent(t *testing.T) {
	g := NewWithT(t)
	r := NewReader(WithEKSAUserAgent("cli", "v0.10.0"))
	g.Expect(r.userAgent).To(Equal("eks-a-cli/v0.10.0"))
}
