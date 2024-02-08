package framework

import (
	"encoding/json"
	"log"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/pkg/semver"
)

func newVersion(version string) *semver.Version {
	v, err := semver.New(version)
	if err != nil {
		log.Fatalf("error creating semver for EKS-A version %s: %v", version, err)
	}
	return v
}

// versionCommandOutput is the output of the eks-anywhere version command.
type versionCommandOutput struct {
	Version           string `json:"version"`
	BundleManifestURL string `json:"bundleManifestURL"`
}

// localEKSAVersion returns the version of eks-anywhere installed locally.
func localEKSAVersion() (string, error) {
	v, err := localEKSAVersionCommand()
	if err != nil {
		return "", err
	}
	return v.Version, nil
}

// localEKSAVersionCommand returns the output of the eks-anywhere version command.
func localEKSAVersionCommand() (versionCommandOutput, error) {
	cmd, err := prepareCommand("eksctl", "anywhere", "version", "--output", "json")
	if err != nil {
		return versionCommandOutput{}, err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return versionCommandOutput{}, errors.Errorf("failed to run eksctl anywhere version: %v, output: %s", err, out)
	}

	versionOut := &versionCommandOutput{}
	err = json.Unmarshal(out, versionOut)
	if err != nil {
		return versionCommandOutput{}, err
	}

	return *versionOut, nil
}
