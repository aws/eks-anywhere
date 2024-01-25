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
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/cli/pkg/bundles"
	"github.com/aws/eks-anywhere/release/cli/pkg/clients"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"
	"github.com/aws/eks-anywhere/release/cli/pkg/operations"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	artifactutils "github.com/aws/eks-anywhere/release/cli/pkg/util/artifacts"
	releaseutils "github.com/aws/eks-anywhere/release/cli/pkg/util/release"
)

var (
	bundleReleaseManifestFile = "/bundle-release.yaml"
	eksAReleaseManifestFile   = "/eks-a-release.yaml"
)

// releaseCmd represents the release command.
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
		weekly := viper.GetBool("weekly")
		releaseTime := time.Now().UTC()
		releaseDate := releaseTime.Format(constants.YYYYMMDD)
		awsSignerProfileArn := viper.GetString("aws-signer-profile-arn")

		var bundleRelease bool
		var releaseEnvironment string

		if !devRelease {
			bundleRelease = viper.GetBool("bundle-release")
			releaseEnvironment = viper.GetString("release-environment")
		}

		if bundleRelease {
			releaseVersion = cliMaxVersion
		}

		releaseConfig := &releasetypes.ReleaseConfig{
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
			ReleaseDate:              releaseDate,
			ReleaseTime:              releaseTime,
			DevRelease:               devRelease,
			DryRun:                   dryRun,
			Weekly:                   weekly,
			ReleaseEnvironment:       releaseEnvironment,
			AwsSignerProfileArn:      awsSignerProfileArn,
		}

		err := operations.SetRepoHeads(releaseConfig)
		if err != nil {
			fmt.Printf("Error getting heads of code repositories: %v\n", err)
			os.Exit(1)
		}

		var sourceClients *clients.SourceClients
		var releaseClients *clients.ReleaseClients
		if devRelease {
			sourceClients, releaseClients, err = clients.CreateDevReleaseClients(dryRun)
			if err != nil {
				fmt.Printf("Error creating clients: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s Successfully created dev release clients\n", constants.SuccessIcon)
		}
		if releaseEnvironment == "development" {
			sourceClients, releaseClients, err = clients.CreateStagingReleaseClients()
			if err != nil {
				fmt.Printf("Error creating clients: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s Successfully created staging release clients\n", constants.SuccessIcon)
		}
		if releaseEnvironment == "production" {
			sourceClients, releaseClients, err = clients.CreateProdReleaseClients()
			if err != nil {
				fmt.Printf("Error creating clients: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s Successfully created dev release clients\n", constants.SuccessIcon)
		}

		releaseConfig.SourceClients = sourceClients
		releaseConfig.ReleaseClients = releaseClients

		if devRelease {
			buildNumber, err := filereader.GetNextEksADevBuildNumber(releaseVersion, releaseConfig)
			if err != nil {
				fmt.Printf("Error getting previous EKS-A dev release number: %v\n", err)
				os.Exit(1)
			}
			releaseVersion, err = filereader.GetCurrentEksADevReleaseVersion(releaseVersion, releaseConfig, buildNumber)
			if err != nil {
				fmt.Printf("Error getting previous EKS-A dev release number: %v\n", err)
				os.Exit(1)
			}
			releaseConfig.BundleNumber = buildNumber
			releaseConfig.ReleaseVersion = releaseVersion
		}
		releaseConfig.DevReleaseUriVersion = strings.ReplaceAll(releaseVersion, "+", "-")

		if devRelease || bundleRelease {
			bundle := bundles.NewBaseBundles(releaseConfig)
			bundle.Spec.CliMinVersion = cliMinVersion
			bundle.Spec.CliMaxVersion = cliMaxVersion

			bundleArtifactsTable, err := operations.GenerateBundleArtifactsTable(releaseConfig)
			if err != nil {
				fmt.Printf("Error getting bundle artifacts data: %v\n", err)
				os.Exit(1)
			}
			releaseConfig.BundleArtifactsTable = bundleArtifactsTable

			// Download ECR images + S3 artifacts and rename them to the
			// proper release URIs + Upload them to release destinations
			err = operations.BundleArtifactsRelease(releaseConfig)
			if err != nil {
				fmt.Printf("Error releasing bundle artifacts: %v\n", err)
				os.Exit(1)
			}

			imageDigests, err := operations.GenerateImageDigestsTable(context.Background(), releaseConfig)
			if err != nil {
				fmt.Printf("Error generating image digests table: %+v\n", err)
				os.Exit(1)
			}

			if bundleRelease && releaseEnvironment == "development" {
				err = operations.SignImagesNotation(releaseConfig, imageDigests)
				if err != nil {
					fmt.Printf("Error signing container images using notation CLI and AWS Signer: %v\n", err)
					os.Exit(1)
				}
			}

			if bundleRelease && releaseEnvironment == "production" {
				err = operations.CopyImageSignatureUsingOras(releaseConfig, imageDigests)
				if err != nil {
					fmt.Printf("Error copying image signature: %v\n", err)
					os.Exit(1)
				}
			}

			err = operations.GenerateBundleSpec(releaseConfig, bundle, imageDigests)
			if err != nil {
				fmt.Printf("Error generating bundles manifest: %+v\n", err)
				os.Exit(1)
			}

			bundleManifest, err := yaml.Marshal(bundle)
			if err != nil {
				fmt.Printf("Error marshaling bundles manifest: %+v\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n%s\n", string(bundleManifest))

			if !dryRun {
				err = os.WriteFile(bundleReleaseManifestFile, bundleManifest, 0o644)
				if err != nil {
					fmt.Printf("Error writing bundles manifest file to disk: %v\n", err)
					os.Exit(1)
				}

				bundleReleaseManifestKey := artifactutils.GetManifestFilepaths(releaseConfig.DevRelease, releaseConfig.Weekly, releaseConfig.BundleNumber, constants.BundlesKind, releaseConfig.BuildRepoBranchName, releaseConfig.ReleaseDate)
				err = s3.UploadFile(bundleReleaseManifestFile, aws.String(releaseConfig.ReleaseBucket), aws.String(bundleReleaseManifestKey), releaseConfig.ReleaseClients.S3.Uploader)
				if err != nil {
					fmt.Printf("Error uploading bundle manifest to release bucket: %+v", err)
					os.Exit(1)
				}
				fmt.Printf("%s Successfully completed bundle release\n", constants.SuccessIcon)
			}

		}

		if devRelease || !bundleRelease {
			release, err := releaseutils.GetPreviousReleaseIfExists(releaseConfig)
			if err != nil {
				fmt.Printf("Error getting previous EKS-A releases: %v\n", err)
				os.Exit(1)
			}

			release.Name = "eks-anywhere"
			release.APIVersion = "anywhere.eks.amazonaws.com/v1alpha1"
			release.Kind = constants.ReleaseKind
			release.CreationTimestamp = v1.Time{Time: releaseTime}

			if devRelease {
				release.Spec.LatestVersion = releaseVersion
			} else {
				previousLatestVersion := release.Spec.LatestVersion
				previousLatestVersionSemver, err := semver.New(previousLatestVersion)
				if err != nil {
					fmt.Printf("Error getting semver for previous latest release version %s: %v\n", previousLatestVersion, err)
					os.Exit(1)
				}

				releaseVersionSemver, err := semver.New(releaseVersion)
				if err != nil {
					fmt.Printf("Error getting semver for current release version %s: %v\n", releaseVersion, err)
					os.Exit(1)
				}

				if releaseVersionSemver.GreaterThan(previousLatestVersionSemver) {
					release.Spec.LatestVersion = releaseVersion
				}
			}

			eksAArtifactsTable, err := operations.GenerateEksAArtifactsTable(releaseConfig)
			if err != nil {
				fmt.Printf("Error getting EKS-A artifacts data: %v\n", err)
				os.Exit(1)
			}
			releaseConfig.EksAArtifactsTable = eksAArtifactsTable

			err = operations.EksAArtifactsRelease(releaseConfig)
			if err != nil {
				fmt.Printf("Error releasing EKS-A CLI artifacts: %v\n", err)
				os.Exit(1)
			}

			currentEksARelease, err := bundles.GetEksARelease(releaseConfig)
			if err != nil {
				fmt.Printf("Error getting EKS-A release: %v\n", err)
				os.Exit(1)
			}

			currentEksAReleaseYaml, err := yaml.Marshal(currentEksARelease)
			if err != nil {
				fmt.Printf("Error marshaling EKS-A releases manifest: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n%s\n", string(currentEksAReleaseYaml))

			if dryRun {
				fmt.Printf("%s Successfully completed dry-run of release process\n", constants.SuccessIcon)
				os.Exit(0)
			}

			previousReleases := releaseutils.EksAReleases(release.Spec.Releases)
			release.Spec.Releases = previousReleases.AppendOrUpdateRelease(currentEksARelease)

			releaseManifest, err := yaml.Marshal(release)
			if err != nil {
				fmt.Printf("Error marshaling EKS-A releases manifest: %v\n", err)
				os.Exit(1)
			}

			// Push the manifest file and other artifacts to release locations
			err = os.WriteFile(eksAReleaseManifestFile, releaseManifest, 0o644)
			if err != nil {
				fmt.Printf("Error writing EKS-A release manifest file to disk: %v\n", err)
				os.Exit(1)
			}

			eksAReleaseManifestKey := artifactutils.GetManifestFilepaths(releaseConfig.DevRelease, releaseConfig.Weekly, releaseConfig.BundleNumber, constants.ReleaseKind, releaseConfig.BuildRepoBranchName, releaseConfig.ReleaseDate)
			err = s3.UploadFile(eksAReleaseManifestFile, aws.String(releaseConfig.ReleaseBucket), aws.String(eksAReleaseManifestKey), releaseConfig.ReleaseClients.S3.Uploader)
			if err != nil {
				fmt.Printf("Error uploading EKS-A release manifest to release bucket: %v", err)
				os.Exit(1)
			}

			if !weekly {
				err = filereader.PutEksAReleaseVersion(releaseVersion, releaseConfig)
				if err != nil {
					fmt.Printf("Error uploading latest EKS-A release version to S3: %v\n", err)
					os.Exit(1)
				}
			}
			fmt.Printf("%s Successfully completed EKS-A release\n", constants.SuccessIcon)
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
	releaseCmd.Flags().Bool("dev-release", true, "Flag to indicate a dev release")
	releaseCmd.Flags().Bool("bundle-release", true, "Flag to indicate a bundle release")
	releaseCmd.Flags().String("release-environment", "", "Release environment")
	releaseCmd.Flags().Bool("dry-run", false, "Flag to indicate if the release is a dry run")
	releaseCmd.Flags().Bool("weekly", false, "Flag to indicate a weekly bundle release")
	releaseCmd.Flags().String("aws-signer-profile-arn", "", "Arn of AWS Signer profile to sign the container images")
}
