package pkg

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
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
	out, err := cloneRepo(r.CliRepoUrl, r.CliRepoSource)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

	// Clone the build-tooling repository
	fmt.Println("Cloning build-tooling repository")
	r.BuildRepoSource = filepath.Join(parentSourceDir, "eks-a-build")
	out, err = cloneRepo(r.BuildRepoUrl, r.BuildRepoSource)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

	if r.BuildRepoBranchName != "main" {
		fmt.Printf("Checking out build-tooling repo at branch %s\n", r.BuildRepoBranchName)
		out, err = checkoutRepo(r.BuildRepoSource, r.BuildRepoBranchName)
		if err != nil {
			return errors.Cause(err)
		}
		fmt.Println(out)
	}

	if r.CliRepoBranchName != "main" {
		fmt.Printf("Checking out CLI repo at branch %s\n", r.CliRepoBranchName)
		out, err = checkoutRepo(r.CliRepoSource, r.CliRepoBranchName)
		if err != nil {
			return errors.Cause(err)
		}
		fmt.Println(out)
	}

	// Set HEADs of the repos
	r.CliRepoHead, err = GetHead(r.CliRepoSource)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Printf("Head of cli repo: %s\n", r.CliRepoHead)

	r.BuildRepoHead, err = GetHead(r.BuildRepoSource)
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
				if r.DryRun && strings.HasSuffix(artifact.Archive.SourceS3Key, ".ova") {
					fmt.Println("Skipping OVA renames in dry-run mode")
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
	s3Retrier := NewRetrier(60*time.Minute, WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if r.BuildRepoBranchName == "main" && IsObjectNotFoundError(err) && totalRetries < 60 {
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
				if r.DryRun && strings.HasSuffix(sourceS3Key, ".ova") {
					fmt.Println("Skipping OVA downloads in dry-run mode")
					continue
				}

				err := s3Retrier.Retry(func() error {
					if !ExistsInS3(r.SourceBucket, objectKey) {
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

				err = downloadFileFromS3(objectLocalFilePath, r.SourceBucket, objectKey)
				if err != nil {
					return errors.Cause(err)
				}

				// Download checksum files for the archive
				checksumExtensions := []string{
					".sha256",
					".sha512",
				}
				for _, extension := range checksumExtensions {
					objectShasumFileName := fmt.Sprintf("%s%s", sourceS3Key, extension)
					objectShasumFileKey := filepath.Join(sourceS3Prefix, objectShasumFileName)
					objectShasumFileLocalFilePath := filepath.Join(artifactPath, objectShasumFileName)
					fmt.Printf("Checksum file - %s\n", objectShasumFileKey)

					err := s3Retrier.Retry(func() error {
						if !ExistsInS3(r.SourceBucket, objectShasumFileKey) {
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

					err = downloadFileFromS3(objectShasumFileLocalFilePath, r.SourceBucket, objectShasumFileKey)
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
					if !ExistsInS3(r.SourceBucket, objectKey) {
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

				err = downloadFileFromS3(objectLocalFilePath, r.SourceBucket, objectKey)
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
				err := r.UploadFileToS3(archiveFile, aws.String(key))
				if err != nil {
					return errors.Cause(err)
				}

				checksumExtensions := []string{".sha256", ".sha512"}
				for _, extension := range checksumExtensions {
					checksumFile := filepath.Join(artifact.Archive.ArtifactPath, artifact.Archive.ReleaseName) + extension
					fmt.Printf("Checksum - %s\n", checksumFile)
					key := filepath.Join(artifact.Archive.ReleaseS3Path, artifact.Archive.ReleaseName) + extension
					err := r.UploadFileToS3(checksumFile, aws.String(key))
					if err != nil {
						return errors.Cause(err)
					}
				}
			}

			if artifact.Manifest != nil {
				manifestFile := filepath.Join(artifact.Manifest.ArtifactPath, artifact.Manifest.ReleaseName)
				fmt.Printf("Manifest - %s\n", manifestFile)
				key := filepath.Join(artifact.Manifest.ReleaseS3Path, artifact.Manifest.ReleaseName)
				err := r.UploadFileToS3(manifestFile, aws.String(key))
				if err != nil {
					return errors.Cause(err)
				}
			}

			if artifact.Image != nil {
				sourceImageUri := artifact.Image.SourceImageURI
				releaseImageUri := artifact.Image.ReleaseImageURI
				fmt.Printf("Source Image - %s\n", sourceImageUri)
				fmt.Printf("Destination Image - %s\n", releaseImageUri)
				err := copyImageFromSourceToDest(sourceEcrAuthConfig, releaseEcrAuthConfig, sourceImageUri, releaseImageUri)
				if err != nil {
					return errors.Cause(err)
				}
			}
		}
	}
	fmt.Printf("%s Successsfully uploaded artifacts\n", SuccessIcon)

	return nil
}

// downloadFileFromS3 downloads a file from S3 and writes it to a local destination
func downloadFileFromS3(objectLocalFilePath, bucketName, key string) error {
	objectURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, key)

	if err := os.MkdirAll(filepath.Dir(objectLocalFilePath), 0o755); err != nil {
		return errors.Cause(err)
	}

	fd, err := os.Create(objectLocalFilePath)
	if err != nil {
		return errors.Cause(err)
	}
	defer fd.Close()

	// Get the data
	resp, err := http.Get(objectURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(fd, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// UploadFileToS3 uploads the file to S3 with ACL public-read
func (r *ReleaseConfig) UploadFileToS3(filePath string, key *string) error {
	s3Uploader := r.ReleaseClients.S3.Uploader
	bucketName := aws.String(r.ReleaseBucket)
	f, err := os.Open(filePath)
	if err != nil {
		return errors.Cause(err)
	}

	result, err := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: bucketName,
		Key:    key,
		Body:   f,
		ACL:    aws.String("public-read"),
	})
	f.Close()
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Printf("Artifact file uploaded to %s\n", result.Location)
	return nil
}

func cloneRepo(cloneUrl, cloneFolder string) (string, error) {
	cmd := exec.Command("git", "clone", cloneUrl, cloneFolder)
	return execCommand(cmd)
}

func checkoutRepo(cloneFolder, branch string) (string, error) {
	cmd := exec.Command("git", "-C", cloneFolder, "checkout", branch)
	return execCommand(cmd)
}

// Gets the head commit has of the repo in the path provided
func GetHead(path string) (string, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", errors.Cause(err)
	}
	ref, err := repo.Head()
	if err != nil {
		return "", errors.Cause(err)
	}

	headData := strings.Split(ref.String(), " ")
	if len(headData) == 0 {
		return "", fmt.Errorf("error getting head data for repo at %s", path)
	}
	return headData[0], nil
}

func execCommand(cmd *exec.Cmd) (string, error) {
	stdout, err := cmd.Output()
	if err != nil {
		return "", errors.Cause(err)
	}
	return string(stdout), nil
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
					imageDigestStr, err = r.GetECRPublicImageDigest(artifact.Image.ReleaseImageURI)
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
	eksAReleaseManifestKey := r.GetManifestFilepaths(anywherev1alpha1.ReleaseKind)
	eksAReleaseManifestUrl := fmt.Sprintf("%s/%s", r.CDN, eksAReleaseManifestKey)

	if ExistsInS3(r.ReleaseBucket, eksAReleaseManifestKey) {
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
	for i, release := range releases {
		if release.Version == r.Version {
			releases[i] = r
			fmt.Println("Updating existing release in releases manifest")
			return releases
		}
	}
	releases = append(releases, r)
	fmt.Println("Adding new release to releases manifest")
	return releases
}

func IsObjectNotFoundError(err error) bool {
	return err.Error() == "Requested object not found"
}

func IsImageNotFoundError(err error) bool {
	return err.Error() == "Requested image not found"
}

func getLatestUploadDestination(sourcedFromBranch string) string {
	if sourcedFromBranch == "main" {
		return "latest"
	} else {
		return sourcedFromBranch
	}
}

func (r *ReleaseConfig) waitForSourceImage(sourceAuthConfig *docker.AuthConfiguration, sourceImageUri string) error {
	sourceImageUriSplit := strings.Split(sourceImageUri, ":")
	sourceImageName := strings.Replace(sourceImageUriSplit[0], r.SourceContainerRegistry+"/", "", -1)
	sourceImageTag := sourceImageUriSplit[1]

	var requestUrl string
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		requestUrl = fmt.Sprintf("https://%s:%s@%s/v2/%s/manifests/%s", sourceAuthConfig.Username, sourceAuthConfig.Password, r.SourceContainerRegistry, sourceImageName, sourceImageTag)
	} else {
		requestUrl = fmt.Sprintf("https://%s:%s@public.ecr.aws/v2/%s/%s/manifests/%s", sourceAuthConfig.Username, sourceAuthConfig.Password, filepath.Base(r.SourceContainerRegistry), sourceImageName, sourceImageTag)
	}

	// Creating new GET request
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return errors.Cause(err)
	}

	// Retrier for downloading source ECR images. This retrier has a max timeout of 60 minutes. It
	// checks whether the error occured during download is an ImageNotFound error and retries the
	// download operation for a maximum of 60 retries, with a wait time of 30 seconds per retry.
	retrier := NewRetrier(60*time.Minute, WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if r.BuildRepoBranchName == "main" && IsImageNotFoundError(err) && totalRetries < 60 {
			return true, 30 * time.Second
		}
		return false, 0
	}))

	err = retrier.Retry(func() error {
		var err error
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		bodyStr := string(body)
		if strings.Contains(bodyStr, "MANIFEST_UNKNOWN") {
			return fmt.Errorf("Requested image not found")
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("retries exhausted waiting for source image to be available for copy: %v", err)
	}

	return nil
}

func copyImageFromSourceToDest(sourceAuthConfig, releaseAuthConfig *docker.AuthConfiguration, sourceImageUri, releaseImageUri string) error {
	sourceRegistryUsername := sourceAuthConfig.Username
	sourceRegistryPassword := sourceAuthConfig.Password
	releaseRegistryUsername := releaseAuthConfig.Username
	releaseRegistryPassword := releaseAuthConfig.Password
	cmd := exec.Command("skopeo", "copy", "--src-creds", fmt.Sprintf("%s:%s", sourceRegistryUsername, sourceRegistryPassword), "--dest-creds", fmt.Sprintf("%s:%s", releaseRegistryUsername, releaseRegistryPassword), fmt.Sprintf("docker://%s", sourceImageUri), fmt.Sprintf("docker://%s", releaseImageUri), "-f", "oci", "--all")
	out, err := execCommand(cmd)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

	return nil
}

func (r *ReleaseConfig) GetManifestFilepaths(kind string) string {
	var manifestFilepath string
	switch kind {
	case anywherev1alpha1.BundlesKind:
		if r.DevRelease {
			if r.BuildRepoBranchName != "main" {
				manifestFilepath = fmt.Sprintf("%s/bundle-release.yaml", r.BuildRepoBranchName)
			} else {
				manifestFilepath = "bundle-release.yaml"
			}
		} else {
			manifestFilepath = fmt.Sprintf("releases/bundles/%d/manifest.yaml", r.BundleNumber)
		}
	case anywherev1alpha1.ReleaseKind:
		if r.DevRelease {
			if r.BuildRepoBranchName != "main" {
				manifestFilepath = fmt.Sprintf("%s/eks-a-release.yaml", r.BuildRepoBranchName)
			} else {
				manifestFilepath = "eks-a-release.yaml"
			}
			manifestFilepath = "eks-a-release.yaml"
		} else {
			manifestFilepath = "releases/eks-a/manifest.yaml"
		}
	}
	return manifestFilepath
}

func (r *ReleaseConfig) GetECRImageDigest(sourceImageUri string) (string, error) {
	ecrClient := r.SourceClients.ECR.EcrClient
	sourceUriSplit := strings.Split(sourceImageUri, ":")
	repoName := strings.Replace(sourceUriSplit[0], r.SourceContainerRegistry+"/", "", -1)
	imageTag := sourceUriSplit[1]
	describeImagesOutput, err := ecrClient.DescribeImages(
		&ecr.DescribeImagesInput{
			ImageIds: []*ecr.ImageIdentifier{
				{
					ImageTag: aws.String(imageTag),
				},
			},
			RepositoryName: aws.String(repoName),
		},
	)
	if err != nil {
		return "", errors.Cause(err)
	}

	imageDigest := describeImagesOutput.ImageDetails[0].ImageDigest
	imageDigestStr := *imageDigest
	return imageDigestStr, nil
}

func (r *ReleaseConfig) GetECRPublicImageDigest(releaseImageUri string) (string, error) {
	ecrPublicClient := r.ReleaseClients.ECRPublic.Client
	releaseUriSplit := strings.Split(releaseImageUri, ":")
	repoName := strings.Replace(releaseUriSplit[0], r.ReleaseContainerRegistry+"/", "", -1)
	imageTag := releaseUriSplit[1]
	describeImagesOutput, err := ecrPublicClient.DescribeImages(
		&ecrpublic.DescribeImagesInput{
			ImageIds: []*ecrpublic.ImageIdentifier{
				{
					ImageTag: aws.String(imageTag),
				},
			},
			RepositoryName: aws.String(repoName),
		},
	)
	if err != nil {
		return "", errors.Cause(err)
	}

	imageDigest := describeImagesOutput.ImageDetails[0].ImageDigest
	imageDigestStr := *imageDigest
	return imageDigestStr, nil
}
