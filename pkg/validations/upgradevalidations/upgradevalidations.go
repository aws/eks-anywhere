package upgradevalidations

import (
	"github.com/aws/eks-anywhere/pkg/validations"
)

func New(opts *validations.Opts) *UpgradeValidations {
	return &UpgradeValidations{Opts: opts}
}

type UpgradeValidations struct {
	Opts *validations.Opts
}
