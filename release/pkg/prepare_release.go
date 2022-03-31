package pkg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/pkg/aws/ecrpublic"
	"github.com/aws/eks-anywhere/release/pkg/aws/s3"
	"github.com/aws/eks-anywhere/release/pkg/git"
	"github.com/aws/eks-anywhere/release/pkg/images"
	"github.com/aws/eks-anywhere/release/pkg/retrier"
	"github.com/aws/eks-anywhere/release/pkg/utils"
)

type EksAReleases []anywherev1alpha1.EksARelease

func (r *ReleaseConfig) SetRepoHeads() error {
	fmt.Println("\n==========================================================")
	fmt.Println("                    Local Repository Setup")
	fmt.Println("==========================================================")

	// Get the repos from env var
	if r.CliRepoUrl == "" || r.BuildRepoUrl == "" {
		return fmt.Errorf("One or both clone URLs are empty")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Cause(err)
	}
	parentSourceDir := filepath.Join(homeDir, "eks-a-source")

	// Clone the CLI repository
	fmt.Println("Cloning CLI repository")
	r.CliRepoSource = filepath.Join(parentSourceDir, "eks-a-cli")
	out, err := git.CloneRepo(r.CliRepoUrl, r.CliRepoSource)
	fmt.Println(out)
	if err != nil {
		return errors.Cause(err)
	}

	// Clone the build-tooling repository
	fmt.Println("Cloning build-tooling repository")
	r.BuildRepoSource = filepath.Join(parentSourceDir, "eks-a-build")
	out, err = git.CloneRepo(r.BuildRepoUrl, r.BuildRepoSource)
	fmt.Println(out)
	if err != nil {
		return errors.Cause(err)
	}

	if r.BuildRepoBranchName != "main" {
		fmt.Printf("Checking out build-tooling repo at branch %s\n", r.BuildRepoBranchName)
		out, err = git.CheckoutRepo(r.BuildRepoSource, r.BuildRepoBranchName)
		fmt.Println(out)
		if err != nil {
			return errors.Cause(err)
		}
	}

	if r.CliRepoBranchName != "main" {
		fmt.Printf("Checking out CLI repo at branch %s\n", r.CliRepoBranchName)
		out, err = git.CheckoutRepo(r.CliRepoSource, r.CliRepoBranchName)
		fmt.Println(out)
		if err != nil {
			return errors.Cause(err)
		}
	}

	// Set HEADs of the repos
	r.CliRepoHead, err = git.GetHead(r.CliRepoSource)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Printf("Head of cli repo: %s\n", r.CliRepoHead)

	r.BuildRepoHead, err = git.GetHead(r.BuildRepoSource)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Printf("Head of build repo: %s\n", r.BuildRepoHead)

	fmt.Printf("%s Successfully completed local repository setup\n", SuccessIcon)

	return nil
}

func (r *ReleaseConfig) PrepareBundleRelease() error {
	fmt.Println("\n==========================================================")
	fmt.Println("                  Bundle Release Preparation")
	fmt.Println("==========================================================")
	err := r.downloadArtifacts(r.BundleArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	err = r.renameArtifacts(r.BundleArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	return nil
}

func (r *ReleaseConfig) PrepareEksARelease() error {
	fmt.Println("\n==========================================================")
	fmt.Println("                 EKS-A CLI Release Preparation")
	fmt.Println("==========================================================")
	err := r.downloadArtifacts(r.EksAArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	err = r.renameArtifacts(r.EksAArtifactsTable)
	if err != nil {
		return errors.Cause(err)
	}

	return nil
}

func (r *ReleaseConfig) renameArtifacts(artifacts map[string][]Artifact) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                    Artifacts Rename")
	fmt.Println("==========================================================")
	for _, artifactsList := range artifacts {
		for _, artifact := range artifactsList {

			// Change the name of the archive along with the checksum files
			if artifact.Archive != nil {
				if r.DryRun && artifact.Archive.ImageFormat != "" {
					fmt.Println("Skipping OS image renames in dry-run mode")
					continue
				}
				archiveArtifact := artifact.Archive
				oldArtifactFile := filepath.Join(archiveArtifact.ArtifactPath, archiveArtifact.SourceS3Key)
				newArtifactFile := filepath.Join(archiveArtifact.ArtifactPath, archiveArtifact.ReleaseName)
				fmt.Printf("Renaming archive - %s\n", newArtifactFile)
				err := os.Rename(oldArtifactFile, newArtifactFile)
				if err != nil {
					return errors.Cause(err)
				}

				// Change the names of the checksum files
				checksumExtensions := []string{".sha256", ".sha512"}

				// Adding a special case for tinkerbell/hook project.
				// The project builds linux kernel files that are not stored as tarballs and currently do not have SHA checksums.
				// TODO(pokearu): Add logic to generate SHA for hook project
				if artifact.Archive.ProjectPath == hookProjectPath {
					checksumExtensions = []string{}
				}

				for _, extension := range checksumExtensions {
					oldChecksumFile := oldArtifactFile + extension
					newChecksumFile := newArtifactFile + extension
					fmt.Printf("Renaming checksum file - %s\n", newChecksumFile)
					err = os.Rename(oldChecksumFile, newChecksumFile)
					if err != nil {
						return errors.Cause(err)
					}
				}
			}

			// Override images in the manifest with release URIs
			if artifact.Manifest != nil {
				manifestArtifact := artifact.Manifest
				oldArtifactFile := filepath.Join(manifestArtifact.ArtifactPath, manifestArtifact.SourceS3Key)
				newArtifactFile := filepath.Join(manifestArtifact.ArtifactPath, manifestArtifact.ReleaseName)
				fmt.Printf("Renaming manifest - %s\n", newArtifactFile)
				err := os.Rename(oldArtifactFile, newArtifactFile)
				if err != nil {
					return errors.Cause(err)
				}

				for _, imageTagOverride := range manifestArtifact.ImageTagOverrides {
					manifestFileContents, err := ioutil.ReadFile(newArtifactFile)
					if err != nil {
						return errors.Cause(err)
					}
					regex := fmt.Sprintf("%s/%s.*", r.SourceContainerRegistry, imageTagOverride.Repository)
					compiledRegex := regexp.MustCompile(regex)
					fmt.Printf("Overriding image to %s in manifest %s\n", imageTagOverride.ReleaseUri, newArtifactFile)
					updatedManifestFileContents := compiledRegex.ReplaceAllString(string(manifestFileContents), imageTagOverride.ReleaseUri)
					err = ioutil.WriteFile(newArtifactFile, []byte(updatedManifestFileContents), 0o644)
					if err != nil {
						return errors.Cause(err)
					}
				}
			}
		}
	}
	fmt.Printf("%s Successfully renamed artifacts\n", SuccessIcon)

	return nil
}

func (r *ReleaseConfig) downloadArtifacts(eksArtifacts map[string][]Artifact) error {
	// Retrier for downloading source S3 objects. This retrier has a max timeout of 60 minutes. It
	// checks whether the error occured during download is an ObjectNotFound error and retries the
	// download operation for a maximum of 60 retries, with a wait time of 30 seconds per retry.
	s3Retrier := retrier.NewRetrier(60*time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if r.BuildRepoBranchName == "main" && utils.IsObjectNotFoundError(err) && totalRetries < 60 {
			return true, 30 * time.Second
		}
		return false, 0
	}))
	fmt.Println("==========================================================")
	fmt.Println("                  Artifacts Download")
	fmt.Println("==========================================================")

	for _, artifacts := range eksArtifacts {
		for _, artifact := range artifacts {
			// Check if there is an archive to be downloaded
			if artifact.Archive != nil {
				sourceS3Prefix := artifact.Archive.SourceS3Prefix
				sourceS3Key := artifact.Archive.SourceS3Key
				artifactPath := artifact.Archive.ArtifactPath
				objectKey := filepath.Join(sourceS3Prefix, sourceS3Key)
				objectLocalFilePath := filepath.Join(artifactPath, sourceS3Key)
				fmt.Printf("Archive - %s\n", objectKey)
				if r.DryRun && artifact.Archive.ImageFormat != "" {
					fmt.Println("Skipping OS image downloads in dry-run mode")
					continue
				}

				err := s3Retrier.Retry(func() error {
					if !s3.KeyExists(r.SourceBucket, objectKey) {
						return fmt.Errorf("Requested object not found")
					}
					return nil
				})
				if err != nil {
					if r.BuildRepoBranchName != "main" {
						fmt.Printf("Artifact corresponding to %s branch not found for %s archive. Using artifact from main\n", r.BuildRepoBranchName, sourceS3Key)
						gitTagFromMain, err := r.readGitTag(artifact.Archive.ProjectPath, "main")
						if err != nil {
							return errors.Cause(err)
						}
						latestSourceS3PrefixFromMain := strings.NewReplacer(r.BuildRepoBranchName, "latest", artifact.Archive.GitTag, gitTagFromMain).Replace(sourceS3Prefix)
						objectKey = filepath.Join(latestSourceS3PrefixFromMain, sourceS3Key)
					} else {
						return fmt.Errorf("retries exhausted waiting for archive to be uploaded to source location: %v", err)
					}
				}

				err = s3.DownloadFile(objectLocalFilePath, r.SourceBucket, objectKey)
				if err != nil {
					return errors.Cause(err)
				}

				// Download checksum files for the archive
				checksumExtensions := []string{
					".sha256",
					".sha512",
				}

				// Adding a special case for tinkerbell/hook project.
				// The project builds linux kernel files that are not stored as tarballs and currently do not have SHA checksums.
				// TODO(pokearu): Add logic to generate SHA for hook project
				if artifact.Archive.ProjectPath == hookProjectPath {
					checksumExtensions = []string{}
				}

				for _, extension := range checksumExtensions {
					objectShasumFileName := fmt.Sprintf("%s%s", sourceS3Key, extension)
					objectShasumFileKey := filepath.Join(sourceS3Prefix, objectShasumFileName)
					objectShasumFileLocalFilePath := filepath.Join(artifactPath, objectShasumFileName)
					fmt.Printf("Checksum file - %s\n", objectShasumFileKey)

					err := s3Retrier.Retry(func() error {
						if !s3.KeyExists(r.SourceBucket, objectShasumFileKey) {
							return fmt.Errorf("Requested object not found")
						}
						return nil
					})
					if err != nil {
						if r.BuildRepoBranchName != "main" {
							fmt.Printf("Artifact corresponding to %s branch not found for %s checksum file. Using artifact from main\n", r.BuildRepoBranchName, sourceS3Key)
							gitTagFromMain, err := r.readGitTag(artifact.Archive.ProjectPath, "main")
							if err != nil {
								return errors.Cause(err)
							}
							latestSourceS3PrefixFromMain := strings.NewReplacer(r.BuildRepoBranchName, "latest", artifact.Archive.GitTag, gitTagFromMain).Replace(sourceS3Prefix)
							objectShasumFileKey = filepath.Join(latestSourceS3PrefixFromMain, objectShasumFileName)
						} else {
							return fmt.Errorf("retries exhausted waiting for checksum file to be uploaded to source location: %v", err)
						}
					}

					err = s3.DownloadFile(objectShasumFileLocalFilePath, r.SourceBucket, objectShasumFileKey)
					if err != nil {
						return errors.Cause(err)
					}
				}
			}

			// Check if there is a manifest to be downloaded
			if artifact.Manifest != nil {
				sourceS3Prefix := artifact.Manifest.SourceS3Prefix
				sourceS3Key := artifact.Manifest.SourceS3Key
				artifactPath := artifact.Manifest.ArtifactPath
				objectKey := filepath.Join(sourceS3Prefix, sourceS3Key)
				objectLocalFilePath := filepath.Join(artifactPath, sourceS3Key)
				fmt.Printf("Manifest - %s\n", objectKey)

				err := s3Retrier.Retry(func() error {
					if !s3.KeyExists(r.SourceBucket, objectKey) {
						return fmt.Errorf("Requested object not found")
					}
					return nil
				})
				if err != nil {
					if r.BuildRepoBranchName != "main" {
						fmt.Printf("Artifact corresponding to %s branch not found for %s manifest. Using artifact from main\n", r.BuildRepoBranchName, sourceS3Key)
						gitTagFromMain, err := r.readGitTag(artifact.Manifest.ProjectPath, "main")
						if err != nil {
							return errors.Cause(err)
						}
						latestSourceS3PrefixFromMain := strings.NewReplacer(r.BuildRepoBranchName, "latest", artifact.Manifest.GitTag, gitTagFromMain).Replace(sourceS3Prefix)
						objectKey = filepath.Join(latestSourceS3PrefixFromMain, sourceS3Key)
					} else {
						return fmt.Errorf("retries exhausted waiting for archive to be uploaded to source location: %v", err)
					}
				}

				err = s3.DownloadFile(objectLocalFilePath, r.SourceBucket, objectKey)
				if err != nil {
					return errors.Cause(err)
				}
			}
		}
	}
	fmt.Printf("%s Successfully downloaded artifacts\n", SuccessIcon)

	return nil
}

func (r *ReleaseConfig) UploadArtifacts(eksArtifacts map[string][]Artifact) error {
	fmt.Println("\n==========================================================")
	fmt.Println("                  Artifacts Upload")
	fmt.Println("==========================================================")
	if r.DryRun {
		fmt.Println("Skipping artifacts upload in dry-run mode")
		return nil
	}

	sourceEcrAuthConfig := r.SourceClients.ECR.AuthConfig
	releaseEcrAuthConfig := r.ReleaseClients.ECRPublic.AuthConfig

	for _, artifacts := range eksArtifacts {
		for _, artifact := range artifacts {
			if artifact.Archive != nil {
				archiveFile := filepath.Join(artifact.Archive.ArtifactPath, artifact.Archive.ReleaseName)
				fmt.Printf("Archive - %s\n", archiveFile)
				key := filepath.Join(artifact.Archive.ReleaseS3Path, artifact.Archive.ReleaseName)
				err := s3.UploadFile(archiveFile, aws.String(r.ReleaseBucket), aws.String(key), r.ReleaseClients.S3.Uploader)
				if err != nil {
					return errors.Cause(err)
				}

				checksumExtensions := []string{".sha256", ".sha512"}
				// Adding a special case for tinkerbell/hook project.
				// The project builds linux kernel files that are not stored as tarballs and currently do not have SHA checksums.
				// TODO(pokearu): Add logic to generate SHA for hook project
				if artifact.Archive.ProjectPath == hookProjectPath {
					checksumExtensions = []string{}
				}

				for _, extension := range checksumExtensions {
					checksumFile := filepath.Join(artifact.Archive.ArtifactPath, artifact.Archive.ReleaseName) + extension
					fmt.Printf("Checksum - %s\n", checksumFile)
					key := filepath.Join(artifact.Archive.ReleaseS3Path, artifact.Archive.ReleaseName) + extension
					err := s3.UploadFile(checksumFile, aws.String(r.ReleaseBucket), aws.String(key), r.ReleaseClients.S3.Uploader)
					if err != nil {
						return errors.Cause(err)
					}
				}
			}

			if artifact.Manifest != nil {
				manifestFile := filepath.Join(artifact.Manifest.ArtifactPath, artifact.Manifest.ReleaseName)
				fmt.Printf("Manifest - %s\n", manifestFile)
				key := filepath.Join(artifact.Manifest.ReleaseS3Path, artifact.Manifest.ReleaseName)
				err := s3.UploadFile(manifestFile, aws.String(r.ReleaseBucket), aws.String(key), r.ReleaseClients.S3.Uploader)
				if err != nil {
					return errors.Cause(err)
				}
			}

			if artifact.Image != nil {
				sourceImageUri := artifact.Image.SourceImageURI
				releaseImageUri := artifact.Image.ReleaseImageURI
				fmt.Printf("Source Image - %s\n", sourceImageUri)
				fmt.Printf("Destination Image - %s\n", releaseImageUri)
				exists, err := ecrpublic.CheckImageExistence(releaseImageUri, r.ReleaseContainerRegistry, r.ReleaseClients.ECRPublic.Client)
				if err != nil {
					return fmt.Errorf("checking for image existence in ECR Public: %v", err)
				}
				if !exists {
					err := images.CopyToDestination(sourceEcrAuthConfig, releaseEcrAuthConfig, sourceImageUri, releaseImageUri)
					if err != nil {
						return fmt.Errorf("copying image from source to destination: %v", err)
					}
				}
			}
		}
	}
	fmt.Printf("%s Successsfully uploaded artifacts\n", SuccessIcon)

	return nil
}

func (r *ReleaseConfig) GenerateImageDigestsTable(eksArtifacts map[string][]Artifact) (map[string]string, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("                 Image Digests Table Generation")
	fmt.Println("==========================================================")
	imageDigests := make(map[string]string)

	for _, artifacts := range eksArtifacts {
		for _, artifact := range artifacts {
			if artifact.Image != nil {
				var imageDigestStr string
				var err error
				if r.DryRun {
					sha256sum, err := GenerateRandomSha(256)
					if err != nil {
						return nil, errors.Cause(err)
					}
					imageDigestStr = fmt.Sprintf("sha256:%s", sha256sum)
				} else {
					imageDigestStr, err = ecrpublic.GetImageDigest(artifact.Image.ReleaseImageURI, r.ReleaseContainerRegistry, r.ReleaseClients.ECRPublic.Client)
					if err != nil {
						return nil, errors.Cause(err)
					}
				}

				imageDigests[artifact.Image.ReleaseImageURI] = imageDigestStr
				fmt.Printf("Image digest for %s - %s\n", artifact.Image.ReleaseImageURI, imageDigestStr)
			}
		}
	}
	fmt.Printf("%s Successfully generated image digests table\n", SuccessIcon)

	return imageDigests, nil
}

func (r *ReleaseConfig) GetPreviousReleaseIfExists() (*anywherev1alpha1.Release, error) {
	emptyRelease := &anywherev1alpha1.Release{
		Spec: anywherev1alpha1.ReleaseSpec{
			Releases: []anywherev1alpha1.EksARelease{},
		},
	}
	if r.DryRun {
		return emptyRelease, nil
	}

	release := &anywherev1alpha1.Release{}
	eksAReleaseManifestKey := utils.GetManifestFilepaths(r.DevRelease, r.BundleNumber, anywherev1alpha1.ReleaseKind, r.BuildRepoBranchName)
	eksAReleaseManifestUrl := fmt.Sprintf("%s/%s", r.CDN, eksAReleaseManifestKey)

	if s3.KeyExists(r.ReleaseBucket, eksAReleaseManifestKey) {
		contents, err := ReadHttpFile(eksAReleaseManifestUrl)
		if err != nil {
			return nil, fmt.Errorf("Error reading releases manifest from S3: %v", err)
		}

		if err = yaml.Unmarshal(contents, release); err != nil {
			return nil, fmt.Errorf("Error unmarshaling releases manifest from [%s]: %v", eksAReleaseManifestUrl, err)
		}

		return release, nil
	}

	return emptyRelease, nil
}

func (releases EksAReleases) AppendOrUpdateRelease(r anywherev1alpha1.EksARelease) EksAReleases {
	currentReleaseSemver := strings.Split(r.Version, "+")[0]
	for i, release := range releases {
		existingReleaseSemver := strings.Split(release.Version, "+")[0]
		if currentReleaseSemver == existingReleaseSemver {
			releases[i] = r
			fmt.Println("Updating existing release in releases manifest")
			return releases
		}
	}
	releases = append(releases, r)
	fmt.Println("Adding new release to releases manifest")
	return releases
}

func getLatestUploadDestination(sourcedFromBranch string) string {
	if sourcedFromBranch == "main" {
		return "latest"
	} else {
		return sourcedFromBranch
	}
}

func sortArtifactsMap(m map[string][]Artifact) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
