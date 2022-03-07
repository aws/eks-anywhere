package envtest

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestModuleWithCRDRegexes(t *testing.T) {
	requireDirective := `	sigs.k8s.io/cluster-api v1.0.2`
	replaceDirective := `	sigs.k8s.io/cluster-api => github.com/mrajashree/cluster-api v1.0.3-0.20220301005127-382d70d4a76f`
	g := NewWithT(t)
	m, err := buildModuleWithCRD("sigs.k8s.io/cluster-api")
	g.Expect(err).To(BeNil())

	matchesRequire := m.requireRegex.FindStringSubmatch(requireDirective)
	g.Expect(len(matchesRequire)).To(Equal(3))
	g.Expect(matchesRequire[2]).To(Equal("1.0.2"))

	matchesRequire = m.requireRegex.FindStringSubmatch(replaceDirective)
	g.Expect(matchesRequire).To(BeNil())

	matchesReplace := m.replaceRegex.FindStringSubmatch(replaceDirective)
	g.Expect(len(matchesReplace)).To(Equal(4))
	g.Expect(matchesReplace[2]).To(Equal("github.com/mrajashree/cluster-api"))
	g.Expect(matchesReplace[3]).To(Equal("1.0.3-0.20220301005127-382d70d4a76f"))

	matchesReplace = m.replaceRegex.FindStringSubmatch(requireDirective)
	g.Expect(matchesReplace).To(BeNil())
}
