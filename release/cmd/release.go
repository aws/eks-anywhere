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
		branchName := viper.GetString("branch-name")
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
			CliRepoSource:                  cliRepoDir,
			BuildRepoSource:                buildRepoDir,
			BranchName:                     branchName,
			ArtifactDir:                    artifactDir,
			SourceBucket:                   sourceBucket,
			ReleaseBucket:                  releaseBucket,
			SourceContainerRegistry:        sourceContainerRegistry,
			ReleaseContainerRegistry:       releaseContainerRegistry,
			CDN:                            cdn,
			BundleNumber:                   bundleNumber,
			ReleaseNumber:                  releaseNumber,
			ReleaseVersion:                 releaseVersion,
			ReleaseDate:                    releaseTime,
			DevRelease:                     devRelease,
			ReleaseEnvironment:             releaseEnvironment,
			GenerateComponentBundleVersion: pkg.GetBuildComponentVersionFunc(devRelease),
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

		if devRelease {
			releaseVersion, err = releaseConfig.GetCurrentEksADevReleaseVersion(sourceClients, releaseVersion)
			if err != nil {
				fmt.Printf("Error getting previous eks a dev release number: %v\n", err)
				os.Exit(1)
			}
			releaseConfig.ReleaseVersion = releaseVersion
			releaseConfig.DevReleaseUriVersion = strings.ReplaceAll(releaseVersion, "+", "-")
		}

		if devRelease || bundleRelease {
			bundle := &anywherev1alpha1.Bundles{
				Spec: anywherev1alpha1.BundlesSpec{
					Number:        bundleNumber,
					CliMinVersion: cliMinVersion,
					CliMaxVersion: cliMaxVersion,
				},
			}
			bundle.APIVersion = "anywhere.eks.amazonaws.com/v1alpha1"
			bundle.Kind = "Bundles"
			bundle.CreationTimestamp = v1.Time{Time: releaseTime}

			// Prepare release - Download artifacts, ECR images + S3 artifacts
			// Rename them to the proper release name for manifest generation.
			artifactsTable, err := releaseConfig.PrepareBundleRelease(sourceClients)
			if err != nil {
				fmt.Printf("Error preparing release: %v\n", err)
				os.Exit(1)
			}

			err = pkg.UploadArtifacts(sourceClients, releaseClients, releaseConfig, artifactsTable)
			if err != nil {
				fmt.Printf("Error uploading release artifacts: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("All artifacts have been uploaded")

			imageDigests, err := pkg.UpdateImageDigests(releaseClients, releaseConfig, artifactsTable)
			if err != nil {
				fmt.Printf("Error updating image digest in bundle manifest: %+v\n", err)
				os.Exit(1)
			}

			err = releaseConfig.GenerateBundleSpec(bundle, imageDigests)
			if err != nil {
				fmt.Printf("Error generating bundle manifest: %+v\n", err)
				os.Exit(1)
			}

			bundleManifest, err := yaml.Marshal(bundle)
			if err != nil {
				fmt.Printf("Error marshaling bundle manifest: %+v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Generated bundle manifest:\n%s", string(bundleManifest))

			err = ioutil.WriteFile(bundleReleaseManifestFile, bundleManifest, 0o755)
			if err != nil {
				fmt.Printf("Error writing bundle manifest file to disk: %v\n", err)
				os.Exit(1)
			}

			var bundleReleaseManifestKey string
			if devRelease {
				bundleReleaseManifestKey = bundleReleaseManifestFile
				if releaseConfig.BranchName != "main" {
					bundleReleaseManifestKey = fmt.Sprintf("/%s/%s", releaseConfig.BranchName, bundleReleaseManifestFile)
				}
			} else {
				bundleReleaseManifestKey = fmt.Sprintf("/releases/bundles/%d/manifest.yaml", releaseConfig.BundleNumber)
			}
			err = pkg.UploadFileToS3(bundleReleaseManifestFile, aws.String(releaseConfig.ReleaseBucket), aws.String(bundleReleaseManifestKey), releaseClients.S3.Uploader)
			if err != nil {
				fmt.Printf("Error uploading bundle manifest to release bucket: %+v", err)
				os.Exit(1)
			}
		}

		if devRelease || !bundleRelease {
			var eksAReleaseManifestKey string
			release := &anywherev1alpha1.Release{}

			if devRelease {
				eksAReleaseManifestKey = eksAReleaseManifestFile
			} else {
				eksAReleaseManifestKey = "/releases/eks-a/manifest.yaml"
			}
			eksAReleaseManifestUrl := fmt.Sprintf("%s%s", releaseConfig.CDN, eksAReleaseManifestKey)

			exists, err := pkg.ExistsInS3(releaseClients.S3.Client, releaseConfig.ReleaseBucket, eksAReleaseManifestKey)
			if err != nil {
				fmt.Printf("Error checking if release manifest exists in S3: %v", err)
				os.Exit(1)
			}
			if exists {
				contents, err := pkg.ReadHttpFile(eksAReleaseManifestUrl)
				if err != nil {
					fmt.Printf("Error reading release manifest from S3: %v", err)
					os.Exit(1)
				}
				if err = yaml.Unmarshal(contents, release); err != nil {
					fmt.Printf("Error unmarshaling releases manifest from [%s]: %v", eksAReleaseManifestUrl, err)
					os.Exit(1)
				}
			} else {
				release = &anywherev1alpha1.Release{
					Spec: anywherev1alpha1.ReleaseSpec{
						Releases: []anywherev1alpha1.EksARelease{},
					},
				}
			}

			release.Name = "eks-anywhere"
			release.APIVersion = "anywhere.eks.amazonaws.com/v1alpha1"
			release.Kind = "Release"
			release.CreationTimestamp = v1.Time{Time: releaseTime}
			release.Spec.LatestVersion = releaseVersion

			artifactsTable, err := releaseConfig.PrepareEksARelease(sourceClients)
			if err != nil {
				fmt.Printf("Error preparing release: %v\n", err)
				os.Exit(1)
			}

			err = pkg.UploadArtifacts(sourceClients, releaseClients, releaseConfig, artifactsTable)
			if err != nil {
				fmt.Printf("Error uploading release artifacts: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("All artifacts have been uploaded")

			eksARelease, err := releaseConfig.GetEksARelease()
			if err != nil {
				fmt.Printf("Error getting eks-a release asset: %v\n", err)
				os.Exit(1)
			}

			currentReleases := pkg.EksAReleases(release.Spec.Releases)
			release.Spec.Releases = currentReleases.AppendOrUpdateRelease(eksARelease)

			output, err := yaml.Marshal(release)
			if err != nil {
				fmt.Printf("Error marshaling release: %+v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(output))
			if dryRun {
				fmt.Println("Dry-run completed.")
				os.Exit(0)
			}

			// Push the manifest file and other artifacts to release locations
			err = ioutil.WriteFile(eksAReleaseManifestFile, output, 0o755)
			if err != nil {
				fmt.Printf("Error writing manifest file to disk: %v\n", err)
				os.Exit(1)
			}

			err = pkg.UploadFileToS3(eksAReleaseManifestFile, aws.String(releaseConfig.ReleaseBucket), aws.String(eksAReleaseManifestKey), releaseClients.S3.Uploader)
			if err != nil {
				fmt.Printf("Error uploading bundle manifest to release bucket: %+v", err)
				os.Exit(1)
			}

			if devRelease {
				err = releaseConfig.PutEksAReleaseVersion(releaseClients, releaseVersion)
				if err != nil {
					fmt.Printf("Error uploading latest release version to S3: %v\n", err)
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
	releaseCmd.Flags().String("cli-repo-source", "", "The eks-anywhere-cli source")
	releaseCmd.Flags().String("build-repo-source", "", "The eks-anywhere-build-tooling source")
	releaseCmd.Flags().String("branch-name", "main", "The branch name to build bundles from")
	releaseCmd.Flags().String("artifact-dir", "", "The base directory for artifacts")
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
