package pkg

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func BuildComponentVersion(versioner projectVersioner, componentCheckSum string) (string, error) {
	patchVersion, err := versioner.patchVersion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s+%s", patchVersion, componentCheckSum), nil
}

type versioner struct {
	repoSource    string
	pathToProject string
}

func newVersioner(pathToProject string) *versioner {
	return &versioner{pathToProject: pathToProject}
}

func (v *versioner) patchVersion() (string, error) {
	projectSource := filepath.Join(v.repoSource, v.pathToProject)
	cmd := exec.Command("git", "-C", projectSource, "describe", "--tag")
	out, err := execCommand(cmd)
	if err != nil {
		return "", errors.Wrapf(err, "failed executing git describe to get version in [%s]", projectSource)
	}

	gitVersion := strings.Split(out, "-")
	gitTag := gitVersion[0]

	return gitTag, nil
}

type versionerWithGITTAG struct {
	versioner
	folderWithGITTAG  string
	sourcedFromBranch string
	releaseConfig     *ReleaseConfig
}

func newVersionerWithGITTAG(repoSource, pathToProject, sourcedFromBranch string, releaseConfig *ReleaseConfig) *versionerWithGITTAG {
	return &versionerWithGITTAG{
		folderWithGITTAG:  pathToProject,
		versioner:         versioner{repoSource: repoSource, pathToProject: pathToProject},
		sourcedFromBranch: sourcedFromBranch,
		releaseConfig:     releaseConfig,
	}
}

func newMultiProjectVersionerWithGITTAG(repoSource, pathToRootFolder, pathToMainProject, sourcedFromBranch string, releaseConfig *ReleaseConfig) *versionerWithGITTAG {
	return &versionerWithGITTAG{
		folderWithGITTAG:  pathToMainProject,
		versioner:         versioner{repoSource: repoSource, pathToProject: pathToRootFolder},
		sourcedFromBranch: sourcedFromBranch,
		releaseConfig:     releaseConfig,
	}
}

func (v *versionerWithGITTAG) patchVersion() (string, error) {
	return v.releaseConfig.readGitTag(v.folderWithGITTAG, v.sourcedFromBranch)
}

type cliVersioner struct {
	versioner
	cliVersion string
}

func newCliVersioner(cliVersion, pathToProject string) *cliVersioner {
	return &cliVersioner{
		cliVersion: cliVersion,
		versioner:  versioner{pathToProject: pathToProject},
	}
}

func (v *cliVersioner) patchVersion() (string, error) {
	return v.cliVersion, nil
}

func GenerateComponentChecksum(hashes []string) string {
	b := make([][]byte, len(hashes))
	if hashes != nil {
		for i, str := range hashes {
			b[i] = []byte(str)
		}
	}
	joinByteArrays := bytes.Join(b, []byte(""))
	sum256 := sha256.Sum256(joinByteArrays)
	sumStr := string(sum256[:])[:4]

	return sumStr
}
