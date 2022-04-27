package curatedpackages_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestBundleSourceSet(t *testing.T) {
	type acceptCase struct {
		// name of the test
		name string
		// input from the user
		input string
		// expected result
		expected curatedpackages.BundleSource
	}

	accepts := []acceptCase{
		{"golden path", "cluster", curatedpackages.Cluster},
		{"case insensitivity", "cLuSTEr", curatedpackages.Cluster},
		{"golden path", "registry", curatedpackages.Registry},
		{"case insensitivity #2", "Registry", curatedpackages.Registry},
		{"whitespace before", " registry", curatedpackages.Registry},
		{"whitespace after", "registry ", curatedpackages.Registry},
	}
	for _, testcase := range accepts {
		t.Run(testcase.name, func(t *testing.T) {
			g := NewWithT(t)
			bs := curatedpackages.BundleSource("")
			err := bs.Set(testcase.input)
			g.Expect(err).To(BeNil())
			g.Expect(bs).To(Equal(testcase.expected))
		})
	}
}

func TestBundleSourceRejectCase(t *testing.T) {
	type rejectCase struct {
		// name of the test
		name string
		// input from the user
		input string
	}

	rejects := []rejectCase{
		{"empty", ""},
		{"double dash", "--"},
		{"exclamation point", "!"},
		{"something", "something"},
		{"junk", "junk"},
		{"kubeVersion", "1.21"},
		{"random space in the middle", "reg istry"},
	}
	for _, testcase := range rejects {
		t.Run(testcase.name, func(t *testing.T) {
			g := NewWithT(t)
			bs := curatedpackages.BundleSource("")
			err := bs.Set(testcase.input)
			g.Expect(err).NotTo(BeNil())
			g.Expect(err).To(MatchError(ContainSubstring("unknown source:")))
		})
	}
}
