package framework

import (
	"log"

	"github.com/aws/eks-anywhere/pkg/semver"
)

const (
	v060 = "0.6.0"
	v050 = "0.5.0"
)

func Eksa060() *semver.Version {
	return newVersion(v060)
}

func Eksa050() *semver.Version {
	return newVersion(v050)
}

func newVersion(version string) *semver.Version {
	v, err := semver.New(version)
	if err != nil {
		log.Fatalf("error creating semver for EKS-A version %s: %v", version, err)
	}
	return v
}
