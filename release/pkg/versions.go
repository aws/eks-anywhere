package pkg

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func GetBuildComponentVersionFunc(isDevRelease bool) generateComponentBundleVersion {
	if isDevRelease {
		return buildComponentVersionForDev
	} else {
		return buildComponentVersionForProd
	}
}

func buildComponentVersionForDev(versioner projectVersioner) (string, error) {
	patchVersion, err := versioner.patchVersion()
	if err != nil {
		return "", err
	}

	metadata, err := versioner.buildMetadata()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s+%s", patchVersion, metadata), nil
}

func buildComponentVersionForProd(versioner projectVersioner) (string, error) {
	patchVersion, err := versioner.patchVersion()
	if err != nil {
		return "", err
	}

	return patchVersion, nil
}

type versioner struct {
	pathToProject string
}

func newVersioner(pathToProject string) *versioner {
	return &versioner{pathToProject: pathToProject}
}

func (v *versioner) buildMetadata() (string, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%h", "-n1", v.pathToProject)
	out, err := execCommand(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "failed executing git rev-parse to get build metadata in [%s]", v.pathToProject)
	}

	return out, nil
}

func (v *versioner) patchVersion() (string, error) {
	cmd := exec.Command("git", "-C", v.pathToProject, "describe", "--tag")
	out, err := execCommand(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "failed executing git describe to get version in [%s]", v.pathToProject)
	}

	gitVersion := strings.Split(out, "-")
	gitTag := gitVersion[0]

	return gitTag, nil
}

type versionerWithGITTAG struct {
	versioner
	folderWithGITTAG string
}

func newVersionerWithGITTAG(pathToProject string) *versionerWithGITTAG {
	return &versionerWithGITTAG{
		folderWithGITTAG: pathToProject,
		versioner:        versioner{pathToProject: pathToProject},
	}
}

func newMultiProjectVersionerWithGITTAG(pathToRootFolder, pathToMainProject string) *versionerWithGITTAG {
	return &versionerWithGITTAG{
		folderWithGITTAG: pathToMainProject,
		versioner:        versioner{pathToProject: pathToRootFolder},
	}
}

func (v *versionerWithGITTAG) patchVersion() (string, error) {
	tagFile := filepath.Join(v.folderWithGITTAG, "GIT_TAG")
	gitTag, err := readFile(tagFile)
	if err != nil {
		return "", errors.Wrapf(err, "failed reading GIT_TAG file for [%s]", v.pathToProject)
	}

	return gitTag, nil
}
