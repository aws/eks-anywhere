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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/pkg"
)

var (
	bundleReleaseManifestFile = "/bundle-release.yaml"
	eksAReleaseManifestFile   = "/eks-a-release.yaml"
)

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Cut an eks-anywhere release",
	PreRun: func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			fmt.Printf("Error initializing flags: %v\n", err)
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// TODO validation on these flags
		releaseVersion := viper.GetString("release-version")
		bundleNumber := viper.GetInt("bundle-number")
		cliMinVersion := viper.GetString("min-version")
		cliMaxVersion := viper.GetString("max-version")
		releaseNumber := viper.GetInt("release-number")
		cliRepoDir := viper.GetString("cli-repo-source")
		buildRepoDir := viper.GetString("build-repo-source")
		cliRepoUrl := viper.GetString("cli-repo-url")
		buildRepoUrl := viper.GetString("build-repo-url")
		buildRepoBranchName := viper.GetString("build-repo-branch-name")
		cliRepoBranchName := viper.GetString("cli-repo-branch-name")
		artifactDir := viper.GetString("artifact-dir")
		sourceBucket := viper.GetString("source-bucket")
		releaseBucket := viper.GetString("release-bucket")
		sourceContainerRegistry := viper.GetString("source-container-registry")
		releaseContainerRegistry := viper.GetString("release-container-registry")
		cdn := viper.GetString("cdn")
		devRelease := viper.GetBool("dev-release")
		dryRun := viper.GetBool("dry-run")
		releaseTime := time.Now().UTC()

		var bundleRelease bool
		var releaseEnvironment string

		if !devRelease {
			bundleRelease = viper.GetBool("bundle-release")
			releaseEnvironment = viper.GetString("release-environment")
		}

		if bundleRelease {
			releaseVersion = cliMaxVersion
		}

		releaseConfig := &pkg.ReleaseConfig{
			CliRepoSource:            cliRepoDir,
			BuildRepoSource:          buildRepoDir,
			CliRepoUrl:               cliRepoUrl,
			BuildRepoUrl:             buildRepoUrl,
			BuildRepoBranchName:      buildRepoBranchName,
			CliRepoBranchName:        cliRepoBranchName,
			ArtifactDir:              artifactDir,
			SourceBucket:             sourceBucket,
			ReleaseBucket:            releaseBucket,
			SourceContainerRegistry:  sourceContainerRegistry,
			ReleaseContainerRegistry: releaseContainerRegistry,
			CDN:                      cdn,
			BundleNumber:             bundleNumber,
			ReleaseNumber:            releaseNumber,
			ReleaseVersion:           releaseVersion,
			ReleaseDate:              releaseTime,
			DevRelease:               devRelease,
			DryRun:                   dryRun,
			ReleaseEnvironment:       releaseEnvironment,
		}

		err := releaseConfig.SetRepoHeads()
		if err != nil {
			fmt.Printf("Error getting heads of code repositories: %v\n", err)
			os.Exit(1)
		}

		var sourceClients *pkg.SourceClients
		var releaseClients *pkg.ReleaseClients
		if devRelease {
			sourceClients, releaseClients, err = releaseConfig.CreateDevReleaseClients()
			if err != nil {
				fmt.Printf("Error creating clients: %v\n", err)
				os.Exit(1)
			}
		}
		if releaseEnvironment == "development" {
			sourceClients, releaseClients, err = releaseConfig.CreateStagingReleaseClients()
			if err != nil {
				fmt.Printf("Error creating clients: %v\n", err)
				os.Exit(1)
			}
		}
		if releaseEnvironment == "production" {
			sourceClients, releaseClients, err = releaseConfig.CreateProdReleaseClients()
			if err != nil {
				fmt.Printf("Error creating clients: %v\n", err)
				os.Exit(1)
			}
		}

		releaseConfig.SourceClients = sourceClients
		releaseConfig.ReleaseClients = releaseClients

		if devRelease {
			releaseVersion, err = releaseConfig.GetCurrentEksADevReleaseVersion(releaseVersion)
			if err != nil {
				fmt.Printf("Error getting previous EKS-A dev release number: %v\n", err)
				os.Exit(1)
			}
			releaseConfig.ReleaseVersion = releaseVersion
		}
		releaseConfig.DevReleaseUriVersion = strings.ReplaceAll(releaseVersion, "+", "-")

		if devRelease || bundleRelease {
			bundle := &anywherev1alpha1.Bundles{
				Spec: anywherev1alpha1.BundlesSpec{
					Number:        bundleNumber,
					CliMinVersion: cliMinVersion,
					CliMaxVersion: cliMaxVersion,
				},
			}
			bundle.APIVersion = "anywhere.eks.amazonaws.com/v1alpha1"
			bundle.Kind = anywherev1alpha1.BundlesKind
			bundle.CreationTimestamp = v1.Time{Time: releaseTime}

			fmt.Println("Getting bundle artifacts data")
			artifactsTable, err := releaseConfig.GetBundleArtifactsData()
			if err != nil {
				fmt.Printf("Error getting bundle artifacts data: %v\n", err)
				os.Exit(1)
			}

			// Download ECR images + S3 artifacts and rename them to the
			// proper release URIs for manifest generation.
			err = releaseConfig.PrepareBundleRelease(artifactsTable)
			if err != nil {
				fmt.Printf("Error preparing bundle release: %v\n", err)
				os.Exit(1)
			}

			err = releaseConfig.UploadArtifacts(artifactsTable)
			if err != nil {
				fmt.Printf("Error uploading bundle release artifacts: %v\n", err)
				os.Exit(1)
			}

			imageDigests, err := releaseConfig.UpdateImageDigests(artifactsTable)
			if err != nil {
				fmt.Printf("Error updating image digests in bundles manifest: %+v\n", err)
				os.Exit(1)
			}

			err = releaseConfig.GenerateBundleSpec(bundle, imageDigests)
			if err != nil {
				fmt.Printf("Error generating bundles manifest: %+v\n", err)
				os.Exit(1)
			}

			bundleManifest, err := yaml.Marshal(bundle)
			if err != nil {
				fmt.Printf("Error marshaling bundles manifest: %+v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Generated bundles manifest:\n\n%s\n", string(bundleManifest))

			if !dryRun {
				err = ioutil.WriteFile(bundleReleaseManifestFile, bundleManifest, 0o755)
				if err != nil {
					fmt.Printf("Error writing bundles manifest file to disk: %v\n", err)
					os.Exit(1)
				}

				bundleReleaseManifestKey := releaseConfig.GetManifestFilepaths(anywherev1alpha1.BundlesKind)
				err = releaseConfig.UploadFileToS3(bundleReleaseManifestFile, aws.String(bundleReleaseManifestKey))
				if err != nil {
					fmt.Printf("Error uploading bundle manifest to release bucket: %+v", err)
					os.Exit(1)
				}
				fmt.Println("Bundle release successful")
			}

		}

		if devRelease || !bundleRelease {
			eksAReleaseManifestKey := releaseConfig.GetManifestFilepaths(anywherev1alpha1.ReleaseKind)
			release, err := releaseConfig.GetPreviousReleaseIfExists()
			if err != nil {
				fmt.Printf("Error getting previous EKS-A releases: %v\n", err)
				os.Exit(1)
			}

			release.Name = "eks-anywhere"
			release.APIVersion = "anywhere.eks.amazonaws.com/v1alpha1"
			release.Kind = anywherev1alpha1.ReleaseKind
			release.CreationTimestamp = v1.Time{Time: releaseTime}
			release.Spec.LatestVersion = releaseVersion

			artifactsTable, err := releaseConfig.GetEksAArtifactsData()
			if err != nil {
				fmt.Printf("Error getting EKS-A artifacts data: %v\n", err)
				os.Exit(1)
			}

			err = releaseConfig.PrepareEksARelease()
			if err != nil {
				fmt.Printf("Error preparing EKS-A release: %v\n", err)
				os.Exit(1)
			}

			err = releaseConfig.UploadArtifacts(artifactsTable)
			if err != nil {
				fmt.Printf("Error uploading EKS-A release artifacts: %v\n", err)
				os.Exit(1)
			}

			currentEksARelease, err := releaseConfig.GetEksARelease()
			if err != nil {
				fmt.Printf("Error getting EKS-A release: %v\n", err)
				os.Exit(1)
			}

			currentEksAReleaseYaml, err := yaml.Marshal(currentEksARelease)
			if err != nil {
				fmt.Printf("Error marshaling EKS-A releases manifest: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Generated current EKS-A release:\n\n%s\n", string(currentEksAReleaseYaml))
			if dryRun {
				fmt.Println("Dry-run completed")
				os.Exit(0)
			}

			previousReleases := pkg.EksAReleases(release.Spec.Releases)
			release.Spec.Releases = previousReleases.AppendOrUpdateRelease(currentEksARelease)

			releaseManifest, err := yaml.Marshal(release)
			if err != nil {
				fmt.Printf("Error marshaling EKS-A releases manifest: %v\n", err)
				os.Exit(1)
			}

			// Push the manifest file and other artifacts to release locations
			err = ioutil.WriteFile(eksAReleaseManifestFile, releaseManifest, 0o755)
			if err != nil {
				fmt.Printf("Error writing EKS-A release manifest file to disk: %v\n", err)
				os.Exit(1)
			}

			err = releaseConfig.UploadFileToS3(eksAReleaseManifestFile, aws.String(eksAReleaseManifestKey))
			if err != nil {
				fmt.Printf("Error uploading EKS-A release manifest to release bucket: %v", err)
				os.Exit(1)
			}

			if devRelease {
				err = releaseConfig.PutEksAReleaseVersion(releaseVersion)
				if err != nil {
					fmt.Printf("Error uploading latest EKS-A release version to S3: %v\n", err)
					os.Exit(1)
				}
			}
			fmt.Println("EKS-A release successful")
		}
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	releaseCmd.Flags().String("release-version", "vDev", "The version of eks-a")
	releaseCmd.Flags().Int("bundle-number", 1, "The bundle version number")
	releaseCmd.Flags().String("min-version", "v0.0.0", "The minimum version of eks-a supported by dependency bundles")
	releaseCmd.Flags().String("max-version", "v0.0.0", "The maximum version of eks-a supported by dependency bundles")
	releaseCmd.Flags().Int("release-number", 1, "The release-number to create")
	releaseCmd.Flags().String("cli-repo-url", "", "URL to clone the eks-anywhere repo")
	releaseCmd.Flags().String("build-repo-url", "", "URL to clone the eks-anywhere-build-tooling repo")
	releaseCmd.Flags().String("cli-repo-source", "", "The eks-anywhere-cli source")
	releaseCmd.Flags().String("build-repo-source", "", "The eks-anywhere-build-tooling source")
	releaseCmd.Flags().String("build-repo-branch-name", "main", "The branch name to build bundles from")
	releaseCmd.Flags().String("cli-repo-branch-name", "main", "The branch name to build EKS-A CLI from")
	releaseCmd.Flags().String("artifact-dir", "downloaded-artifacts", "The base directory for artifacts")
	releaseCmd.Flags().String("cdn", "https://anywhere.eks.amazonaws.com", "The URL base for artifacts")
	releaseCmd.Flags().String("source-bucket", "eks-a-source-bucket", "The bucket name where the built/staging artifacts are located to download")
	releaseCmd.Flags().String("release-bucket", "eks-a-release-bucket", "The bucket name where released artifacts live")
	releaseCmd.Flags().String("source-container-registry", "", "The container registry to pull images from for a dev release")
	releaseCmd.Flags().String("release-container-registry", "", "The container registry that images wll be pushed to")
	releaseCmd.Flags().Bool("dev-release", true, "Flag to indicate its a dev release")
	releaseCmd.Flags().Bool("bundle-release", true, "Flag to indicate a bundle release")
	releaseCmd.Flags().String("release-environment", "", "Release environment")
	releaseCmd.Flags().Bool("dry-run", false, "Flag to indicate if the release is a dry run")
}
