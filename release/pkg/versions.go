package pkg

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/pkg/git"
)

const fakeComponentChecksum = "abcdef1"

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

func (v *versioner) buildMetadata() (string, error) {
	out, err := git.GetLatestCommitForPath(v.pathToProject, v.pathToProject)
	if err != nil {
		return "", errors.Wrapf(err, "failed executing git log to get build metadata in [%s]", v.pathToProject)
	}

	return out, nil
}

func (v *versioner) patchVersion() (string, error) {
	projectSource := filepath.Join(v.repoSource, v.pathToProject)
	out, err := git.DescribeTag(projectSource)
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

func generateComponentHash(hashes []string) string {
	b := make([][]byte, len(hashes))
	for i, str := range hashes {
		b[i] = []byte(str)
	}
	joinByteArrays := bytes.Join(b, []byte(""))
	hash := sha256.Sum256(joinByteArrays)
	hashStr := hex.EncodeToString(hash[:])[:7]

	return hashStr
}

func (r *ReleaseConfig) GenerateManifestHash(manifestArtifact *ManifestArtifact) (string, error) {
	if r.DryRun {
		return fakeComponentChecksum, nil
	}

	manifestContents, err := ioutil.ReadFile(filepath.Join(manifestArtifact.ArtifactPath, manifestArtifact.ReleaseName))
	if err != nil {
		return "", errors.Wrapf(err, "failed reading manifest contents from [%s]", manifestArtifact.ArtifactPath)
	}
	hash := sha256.Sum256(manifestContents)
	hashStr := hex.EncodeToString(hash[:])

	return hashStr, nil
}
