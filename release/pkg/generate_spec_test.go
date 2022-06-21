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

package pkg

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/pkg/git"
	"github.com/aws/eks-anywhere/release/pkg/test"
)

const (
	releaseFolder         = "release"
	testdataFolder        = "pkg/test/testdata"
	generatedBundleFolder = "generated-bundles"
)

var releaseConfig = &ReleaseConfig{
	CliRepoSource:            "eks-a-build",
	BuildRepoSource:          "eks-a-cli",
	CliRepoUrl:               "https://github.com/aws/eks-anywhere.git",
	BuildRepoUrl:             "https://github.com/aws/eks-anywhere-build-tooling.git",
	SourceBucket:             "projectbuildpipeline-857-pipelineoutputartifactsb-10ajmk30khe3f",
	ReleaseBucket:            "release-bucket",
	SourceContainerRegistry:  "sourceContainerRegistry",
	ReleaseContainerRegistry: "public.ecr.aws/release-container-registry",
	CDN:                      "https://release-bucket",
	BundleNumber:             1,
	ReleaseNumber:            1,
	ReleaseVersion:           "vDev",
	ReleaseDate:              time.Unix(0, 0),
	DevRelease:               true,
	DryRun:                   true,
}

var update = flag.Bool("update", false, "update the golden files of this test")

func TestGenerateBundleManifest(t *testing.T) {
	testCases := []struct {
		testName            string
		buildRepoBranchName string
		cliRepoBranchName   string
		cliMinVersion       string
		cliMaxVersion       string
	}{
		{
			testName:            "Dev-release from main",
			buildRepoBranchName: "main",
			cliRepoBranchName:   "main",
			cliMinVersion:       "v0.7.2",
			cliMaxVersion:       "v0.7.2",
		},
		{
			testName:            "Dev-release from release-0.9",
			buildRepoBranchName: "release-0.9",
			cliRepoBranchName:   "release-0.9",
			cliMinVersion:       "v0.9.0",
			cliMaxVersion:       "v0.9.0",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				t.Fatalf("Error getting home directory: %v\n", err)
			}

			parentSourceDir := filepath.Join(homeDir, "eks-a-source")
			err = os.RemoveAll(parentSourceDir)
			if err != nil {
				t.Fatalf("Error removing source directory: %v\n", err)
			}

			gitRoot, err := git.GetRepoRoot()
			if err != nil {
				t.Fatalf("Error getting top-level Git directory: %v\n", err)
			}

			generatedBundlePath := filepath.Join(gitRoot, releaseFolder, generatedBundleFolder)
			if err := os.MkdirAll(generatedBundlePath, 0o755); err != nil {
				t.Fatalf("Error creating directory at %s for bundle generation: %v\n", generatedBundleFolder, err)
			}

			releaseConfig.BuildRepoBranchName = tt.buildRepoBranchName
			releaseConfig.CliRepoBranchName = tt.cliRepoBranchName

			releaseVersion, err := releaseConfig.GetCurrentEksADevReleaseVersion(releaseConfig.ReleaseVersion)
			if err != nil {
				t.Fatalf("Error getting previous EKS-A dev release number: %v\n", err)
			}

			releaseConfig.ReleaseVersion = releaseVersion
			releaseConfig.DevReleaseUriVersion = strings.ReplaceAll(releaseVersion, "+", "-")

			err = os.RemoveAll(releaseConfig.ArtifactDir)
			if err != nil {
				t.Fatalf("Error removing local artifacts directory: %v\n", err)
			}

			err = releaseConfig.SetRepoHeads()
			if err != nil {
				t.Fatalf("Error getting heads of code repositories: %v\n", err)
			}

			bundleArtifactsTable, err := releaseConfig.GenerateBundleArtifactsTable()
			if err != nil {
				t.Fatalf("Error getting bundle artifacts data: %v\n", err)
			}
			releaseConfig.BundleArtifactsTable = bundleArtifactsTable

			imageDigests, err := releaseConfig.GenerateImageDigestsTable(bundleArtifactsTable)
			if err != nil {
				t.Fatalf("Error generating image digests table: %+v\n", err)
			}

			bundle := releaseConfig.NewBaseBundles()
			bundle.Spec.CliMinVersion = tt.cliMinVersion
			bundle.Spec.CliMaxVersion = tt.cliMaxVersion

			err = releaseConfig.GenerateBundleSpec(bundle, imageDigests)
			if err != nil {
				t.Fatalf("Error generating bundles manifest: %+v\n", err)
			}

			bundleManifest, err := yaml.Marshal(bundle)
			if err != nil {
				t.Fatalf("Error marshaling bundles manifest: %+v\n", err)
			}

			expectedBundleManifestFile := filepath.Join(gitRoot, releaseFolder, testdataFolder, fmt.Sprintf("%s-bundle-release.yaml", tt.buildRepoBranchName))
			generatedBundleManifestFile := filepath.Join(generatedBundlePath, fmt.Sprintf("%s-dry-run-bundle-release.yaml", tt.buildRepoBranchName))
			err = ioutil.WriteFile(generatedBundleManifestFile, bundleManifest, 0o644)
			if err != nil {
				t.Fatalf("Error writing bundles manifest file to disk: %v\n", err)
			}

			test.CheckFilesEquals(t, generatedBundleManifestFile, expectedBundleManifestFile, *update)
		})
	}
}

func TestReleaseConfigNewBundlesName(t *testing.T) {
	testCases := []struct {
		testName      string
		releaseConfig *ReleaseConfig
		want          string
	}{
		{
			testName: "number 2",
			releaseConfig: &ReleaseConfig{
				BundleNumber: 2,
			},
			want: "bundles-2",
		},
		{
			testName:      "no bundle number",
			releaseConfig: &ReleaseConfig{},
			want:          "bundles-0",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(tt.releaseConfig.NewBundlesName()).To(Equal(tt.want))
		})
	}
}

func TestReleaseConfigNewBaseBundles(t *testing.T) {
	g := NewWithT(t)
	now := time.Now()
	releaseConfig := &ReleaseConfig{
		BundleNumber: 10,
		ReleaseDate:  now,
	}
	wantBundles := &anywherev1alpha1.Bundles{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
			Kind:       "Bundles",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "bundles-10",
			CreationTimestamp: metav1.Time{Time: now},
		},
		Spec: anywherev1alpha1.BundlesSpec{
			Number: 10,
		},
	}

	g.Expect(releaseConfig.NewBaseBundles()).To(Equal(wantBundles))
}
