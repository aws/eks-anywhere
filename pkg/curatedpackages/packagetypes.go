package curatedpackages

import (
	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

// DisplayablePackage wraps Package to omit undesired members (like Status).
//
// This is necessary in part because of https://github.com/golang/go/issues/11939
// but also because we just don't want to generate a Status section when we're
// emitting templates for a user to modify.
type DisplayablePackage struct {
	*api.Package
	Status *interface{} `json:"status,omitempty"`
}

func NewDisplayablePackage(p *api.Package) *DisplayablePackage {
	return &DisplayablePackage{Package: p}
}
