package createvalidations

import (
	"github.com/aws/eks-anywhere/pkg/validations"
)

// string values of supported validation names that can be skipped.
const (
	VSphereUserPriv = "vsphere-user-privilege"
)

// SkippableValidations represents all the validations we offer for users to skip.
var SkippableValidations = []string{
	VSphereUserPriv,
}

func New(opts *validations.Opts) *CreateValidations {
	opts.SetDefaults()
	return &CreateValidations{Opts: opts}
}

type CreateValidations struct {
	Opts *validations.Opts
}
