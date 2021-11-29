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
	pathToProject string
}

func newVersioner(pathToProject string) *versioner {
	return &versioner{pathToProject: pathToProject}
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
