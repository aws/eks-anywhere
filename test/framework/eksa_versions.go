package framework

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/version"
)

func newVersion(version string) *semver.Version {
	v, err := semver.New(version)
	if err != nil {
		log.Fatalf("error creating semver for EKS-A version %s: %v", version, err)
	}
	return v
}

// localEKSAVersion returns the version of eks-anywhere installed locally.
func localEKSAVersion() (string, error) {
	v, err := localEKSAVersionCommand()
	if err != nil {
		return "", err
	}
	return v.GitVersion, nil
}

// localEKSAVersionCommand returns the output of the eks-anywhere version command.
func localEKSAVersionCommand() (version.Info, error) {
	cmd, err := prepareCommand("eksctl", "anywhere", "version", "--output", "json")
	if err != nil {
		return version.Info{}, err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return version.Info{}, errors.Errorf("failed to run eksctl anywhere version: %v, output: %s", err, out)
	}

	versionOut := &version.Info{}
	err = json.Unmarshal(out, versionOut)
	if err != nil {
		return version.Info{}, err
	}

	return *versionOut, nil
}
