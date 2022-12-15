package framework

import (
	"log"

	"github.com/aws/eks-anywhere/pkg/semver"
)

func newVersion(version string) *semver.Version {
	v, err := semver.New(version)
	if err != nil {
		log.Fatalf("error creating semver for EKS-A version %s: %v", version, err)
	}
	return v
}
