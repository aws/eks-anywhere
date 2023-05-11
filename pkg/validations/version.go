package validations

import "github.com/aws/eks-anywhere/pkg/version"

type VersionClient interface {
	Get() version.Info
}
