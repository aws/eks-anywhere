package curatedpackages_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestBundleSourceSet(t *testing.T) {
	h := newBSHelper(t)
	h.Parallel()

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
		h.runAcceptCase(testcase.name, testcase.input, testcase.expected)
	}

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
		h.runRejectCase(testcase.name, testcase.input)
	}
}

//
// Helpers
//

type bsHelper struct{ *testing.T }

func newBSHelper(t *testing.T) *bsHelper {
	return &bsHelper{T: t}
}

func (h *bsHelper) runAcceptCase(name, input string, expected curatedpackages.BundleSource) {
	h.Helper()
	h.Run("accepts "+name, func(t *testing.T) {
		h.assertAccepts(assert.New(t), input, expected)
	})
}

func (h *bsHelper) assertAccepts(t *assert.Assertions, input string, expected curatedpackages.BundleSource) {
	bs := curatedpackages.BundleSource("")
	if t.NoError(bs.Set(input)) {
		t.Equal(expected, bs)
	}
}

func (h *bsHelper) runRejectCase(name, input string) {
	h.Helper()
	h.Run("rejects "+name, func(t *testing.T) {
		h.assertRejects(assert.New(t), input)
	})
}

func (h *bsHelper) assertRejects(t *assert.Assertions, input string) {
	bs := curatedpackages.BundleSource("")
	t.Error(bs.Set(input))
}
