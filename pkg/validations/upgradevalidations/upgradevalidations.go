package upgradevalidations

import (
	"github.com/aws/eks-anywhere/pkg/validations"
)

// string values of supported validation names that can be skipped.
const (
	PDB = "pod-disruption"
)

// SkippableValidations represents all the validations we offer for users to skip.
var SkippableValidations = []string{
	PDB,
}

func New(opts *validations.Opts) *UpgradeValidations {
	opts.SetDefaults()
	return &UpgradeValidations{Opts: opts}
}

type UpgradeValidations struct {
	Opts *validations.Opts
}
