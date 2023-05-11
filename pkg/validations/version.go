package validations

import "github.com/aws/eks-anywhere/pkg/version"

// VersionClient is used to mock version package.
type VersionClient interface {
	Get() version.Info
}
