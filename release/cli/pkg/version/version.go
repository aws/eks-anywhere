// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	"github.com/aws/eks-anywhere/release/cli/pkg/git"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

const FakeComponentChecksum = "abcdef1"

type ProjectVersioner interface {
	patchVersion() (string, error)
}

func BuildComponentVersion(versioner ProjectVersioner, componentCheckSum string) (string, error) {
	patchVersion, err := versioner.patchVersion()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s+%s", patchVersion, componentCheckSum), nil
}

type Versioner struct {
	repoSource    string
	pathToProject string
}

func NewVersioner(pathToProject string) *Versioner {
	return &Versioner{pathToProject: pathToProject}
}

func (v *Versioner) patchVersion() (string, error) {
	projectSource := filepath.Join(v.repoSource, v.pathToProject)
	out, err := git.DescribeTag(projectSource)
	if err != nil {
		return "", errors.Wrapf(err, "failed executing git describe to get version in [%s]", projectSource)
	}

	gitVersion := strings.Split(out, "-")
	gitTag := gitVersion[0]

	return gitTag, nil
}

type VersionerWithGITTAG struct {
	Versioner
	folderWithGITTAG  string
	sourcedFromBranch string
	releaseConfig     *releasetypes.ReleaseConfig
}

func NewVersionerWithGITTAG(repoSource, pathToProject, sourcedFromBranch string, releaseConfig *releasetypes.ReleaseConfig) *VersionerWithGITTAG {
	return &VersionerWithGITTAG{
		folderWithGITTAG:  pathToProject,
		Versioner:         Versioner{repoSource: repoSource, pathToProject: pathToProject},
		sourcedFromBranch: sourcedFromBranch,
		releaseConfig:     releaseConfig,
	}
}

func NewMultiProjectVersionerWithGITTAG(repoSource, pathToRootFolder, pathToMainProject, sourcedFromBranch string, releaseConfig *releasetypes.ReleaseConfig) *VersionerWithGITTAG {
	return &VersionerWithGITTAG{
		folderWithGITTAG:  pathToMainProject,
		Versioner:         Versioner{repoSource: repoSource, pathToProject: pathToRootFolder},
		sourcedFromBranch: sourcedFromBranch,
		releaseConfig:     releaseConfig,
	}
}

func (v *VersionerWithGITTAG) patchVersion() (string, error) {
	return filereader.ReadGitTag(v.folderWithGITTAG, v.releaseConfig.BuildRepoSource, v.sourcedFromBranch)
}

type cliVersioner struct {
	Versioner
	cliVersion string
}

func NewCliVersioner(cliVersion, pathToProject string) *cliVersioner {
	return &cliVersioner{
		cliVersion: cliVersion,
		Versioner:  Versioner{pathToProject: pathToProject},
	}
}

func (v *cliVersioner) patchVersion() (string, error) {
	return v.cliVersion, nil
}

func GenerateComponentHash(hashes []string, dryRun bool) string {
	b := make([][]byte, len(hashes))
	for i, str := range hashes {
		b[i] = []byte(str)
	}
	joinByteArrays := bytes.Join(b, []byte(""))
	hash := sha256.Sum256(joinByteArrays)
	hashStr := hex.EncodeToString(hash[:])[:7]

	return hashStr
}

func GenerateManifestHash(r *releasetypes.ReleaseConfig, manifestArtifact *releasetypes.ManifestArtifact) (string, error) {
	if r.DryRun {
		return FakeComponentChecksum, nil
	}

	manifestContents, err := os.ReadFile(filepath.Join(manifestArtifact.ArtifactPath, manifestArtifact.ReleaseName))
	if err != nil {
		return "", errors.Wrapf(err, "failed reading manifest contents from [%s]", manifestArtifact.ArtifactPath)
	}
	hash := sha256.Sum256(manifestContents)
	hashStr := hex.EncodeToString(hash[:])

	return hashStr, nil
}
