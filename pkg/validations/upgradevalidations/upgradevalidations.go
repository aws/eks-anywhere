package upgradevalidations

import (
	"github.com/aws/eks-anywhere/pkg/validations"
)

// SkippableValidations represents all the validations we offer for users to skip.
var SkippableValidations = []string{
	validations.PDB,
	validations.VSphereUserPriv,
	validations.EksaVersionSkew,
}

func New(opts *validations.Opts) *UpgradeValidations {
	opts.SetDefaults()
	return &UpgradeValidations{Opts: opts}
}

type UpgradeValidations struct {
	Opts *validations.Opts
}
