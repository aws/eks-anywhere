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
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type EksAReleases []anywherev1alpha1.EksARelease

func (r *ReleaseConfig) SetRepoHeads() error {
	// Get the repos from env var
	cliRepoUrl := os.Getenv("CLI_REPO_URL")
	buildRepoUrl := os.Getenv("BUILD_REPO_URL")
	if cliRepoUrl == "" || buildRepoUrl == "" {
		return fmt.Errorf("clone env urls not set")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return errors.Cause(err)
	}

	// Clone the CLI repository
	fmt.Println("Cloning CLI repository")
	parentSourceDir := filepath.Join(homeDir, "eks-a-source")
	r.CliRepoSource = filepath.Join(parentSourceDir, "eks-a-cli")
	cmd := exec.Command("git", "clone", cliRepoUrl, r.CliRepoSource)
	out, err := execCommand(cmd)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

	// Clone the build-tooling repository
	fmt.Println("Cloning build-tooling repository")
	r.BuildRepoSource = filepath.Join(parentSourceDir, "eks-a-build")
	cmd = exec.Command("git", "clone", buildRepoUrl, r.BuildRepoSource)
	out, err = execCommand(cmd)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

	if r.DevRelease && r.BranchName != "main" {
		fmt.Printf("Checking out build-tooling repo at branch %s", r.BranchName)
		cmd = exec.Command("git", "-C", r.BuildRepoSource, "checkout", r.BranchName)
		out, err = execCommand(cmd)
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
	return nil
}

func (r *ReleaseConfig) PrepareBundleRelease(sourceClients *SourceClients) (map[string][]Artifact, error) {
	artifactsTable, err := r.GetBundleArtifactsData()
	if err != nil {
		return nil, errors.Cause(err)
	}
	fmt.Println("Initialized artifacts data")

	err = downloadArtifacts(sourceClients, r, artifactsTable)
	if err != nil {
		return nil, errors.Cause(err)
	}
	fmt.Println("Artifacts download complete")

	err = r.renameArtifacts(sourceClients, artifactsTable)
	if err != nil {
		return nil, errors.Cause(err)
	}
	fmt.Println("Renaming artifacts complete")

	return artifactsTable, nil
}

func (r *ReleaseConfig) PrepareEksARelease(sourceClients *SourceClients) (map[string][]Artifact, error) {
	artifactsTable, err := r.GetEksAArtifactsData()
	if err != nil {
		return nil, errors.Cause(err)
	}
	fmt.Println("Initialized artifacts data")

	err = downloadArtifacts(sourceClients, r, artifactsTable)
	if err != nil {
		return nil, errors.Cause(err)
	}
	fmt.Println("Artifacts download complete")

	err = r.renameArtifacts(sourceClients, artifactsTable)
	if err != nil {
		return nil, errors.Cause(err)
	}
	fmt.Println("Renaming artifacts complete")

	return artifactsTable, nil
}

func (r *ReleaseConfig) renameArtifacts(sourceClients *SourceClients, artifacts map[string][]Artifact) error {
	fmt.Println("============================================================")
	fmt.Println("                 Renaming Artifacts                         ")
	fmt.Println("============================================================")
	for _, artifactsList := range artifacts {
		for _, artifact := range artifactsList {

			// Change the name of the archive along with the checksum files
			if artifact.Archive != nil {
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
	return nil
}

func downloadArtifacts(sourceClients *SourceClients, r *ReleaseConfig, eksArtifacts map[string][]Artifact) error {
	// Get s3 client and docker clients
	s3Downloader := sourceClients.S3.Downloader
	s3Client := sourceClients.S3.Client
	// Retrier for downloading source S3 objects. This retrier has a max timeout of 60 minutes. It
	// checks whether the error occured during download is an ObjectNotFound error and retries the
	// download operation for a maximum of 60 retries, with a wait time of 30 seconds per retry.
	s3Retrier := NewRetrier(60*time.Minute, WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if IsObjectNotPresentError(err) && totalRetries < 60 {
			return true, 30 * time.Second
		}
		return false, 0
	}))

	fmt.Println("============================================================")
	fmt.Println("                 Downloading Artifacts                      ")
	fmt.Println("============================================================")

	for _, artifacts := range eksArtifacts {
		for _, artifact := range artifacts {
			// Check if there is an archive to be downloaded
			if artifact.Archive != nil {
				objectKey := fmt.Sprintf("%s/%s", artifact.Archive.SourceS3Prefix, artifact.Archive.SourceS3Key)
				err := s3Retrier.Retry(func() error {
					_, err := s3Client.HeadObject(&s3.HeadObjectInput{
						Bucket: aws.String(r.SourceBucket),
						Key:    aws.String(objectKey),
					})
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("retries exhausted waiting for archive to be uploaded to source location: %v", err)
				}

				err = downloadFileFromS3(artifact.Archive.ArtifactPath, "Archive", r.SourceBucket, objectKey, s3Downloader)
				if err != nil {
					return errors.Cause(err)
				}

				// Download checksum files for the archive
				checksumExtensions := []string{
					".sha256",
					".sha512",
				}
				for _, extension := range checksumExtensions {
					objectShasumFileKey := fmt.Sprintf("%s%s", objectKey, extension)
					err = downloadFileFromS3(artifact.Archive.ArtifactPath, "Checksum file", r.SourceBucket, objectShasumFileKey, s3Downloader)
					if err != nil {
						return errors.Cause(err)
					}
				}

			}

			// Check if there is a manifest to be downloaded
			if artifact.Manifest != nil {
				objectKey := fmt.Sprintf("%s/%s", artifact.Manifest.SourceS3Prefix, artifact.Manifest.SourceS3Key)
				err := s3Retrier.Retry(func() error {
					_, err := s3Client.HeadObject(&s3.HeadObjectInput{
						Bucket: aws.String(r.SourceBucket),
						Key:    aws.String(objectKey),
					})
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("retries exhausted waiting for manifest to be uploaded to source location: %v", err)
				}

				err = downloadFileFromS3(artifact.Manifest.ArtifactPath, "Manifest", r.SourceBucket, objectKey, s3Downloader)
				if err != nil {
					return errors.Cause(err)
				}
			}
		}
	}
	return nil
}

func UploadArtifacts(sourceClients *SourceClients, releaseClients *ReleaseClients, r *ReleaseConfig, eksArtifacts map[string][]Artifact) error {
	// Get clients
	s3Uploader := releaseClients.S3.Uploader
	sourceEcrAuthConfig := sourceClients.ECR.AuthConfig
	releaseEcrAuthConfig := releaseClients.ECRPublic.AuthConfig

	fmt.Println("============================================================")
	fmt.Println("                 Uploading Artifacts                      ")
	fmt.Println("============================================================")

	for _, artifacts := range eksArtifacts {
		for _, artifact := range artifacts {
			if artifact.Archive != nil {
				archiveFile := filepath.Join(artifact.Archive.ArtifactPath, artifact.Archive.ReleaseName)
				fmt.Printf("Archive - %s\n", archiveFile)
				key := filepath.Join(artifact.Archive.ReleaseS3Path, artifact.Archive.ReleaseName)
				err := UploadFileToS3(archiveFile, aws.String(r.ReleaseBucket), aws.String(key), s3Uploader)
				if err != nil {
					return errors.Cause(err)
				}

				checksumExtensions := []string{".sha256", ".sha512"}
				for _, extension := range checksumExtensions {
					checksumFile := filepath.Join(artifact.Archive.ArtifactPath, artifact.Archive.ReleaseName) + extension
					fmt.Printf("Checksum - %s\n", checksumFile)
					key := filepath.Join(artifact.Archive.ReleaseS3Path, artifact.Archive.ReleaseName) + extension
					err := UploadFileToS3(checksumFile, aws.String(r.ReleaseBucket), aws.String(key), s3Uploader)
					if err != nil {
						return errors.Cause(err)
					}
				}
			}

			if artifact.Manifest != nil {
				manifestFile := filepath.Join(artifact.Manifest.ArtifactPath, artifact.Manifest.ReleaseName)
				fmt.Printf("Manifest - %s\n", manifestFile)
				key := filepath.Join(artifact.Manifest.ReleaseS3Path, artifact.Manifest.ReleaseName)
				err := UploadFileToS3(manifestFile, aws.String(r.ReleaseBucket), aws.String(key), s3Uploader)
				if err != nil {
					return errors.Cause(err)
				}
			}

			if artifact.Image != nil {
				fmt.Printf("Source Image - %s\n", artifact.Image.SourceImageURI)
				fmt.Printf("Destination Image - %s\n", artifact.Image.ReleaseImageURI)
				err := r.waitForSourceImage(sourceEcrAuthConfig, artifact.Image.SourceImageURI)
				if err != nil {
					return errors.Cause(err)
				}
				err = copyImageFromSourceToDest(sourceEcrAuthConfig, releaseEcrAuthConfig, artifact.Image.SourceImageURI, artifact.Image.ReleaseImageURI)
				if err != nil {
					return errors.Cause(err)
				}
			}
		}
	}

	return nil
}

// downloadFileFromS3 downloads a file from S3 and writes it to a local destination
func downloadFileFromS3(artifactPath, artifactType, bucketName, key string, s3Downloader *s3manager.Downloader) error {
	objectName := filepath.Base(key)
	fileName := filepath.Join(artifactPath, objectName)
	if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
		return errors.Cause(err)
	}

	fd, err := os.Create(fileName)
	if err != nil {
		return errors.Cause(err)
	}
	defer fd.Close()
	fmt.Printf("%s - %s\n", artifactType, key)
	_, err = s3Downloader.Download(fd, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return errors.Cause(err)
	}

	return nil
}

// UploadFileToS3 uploads the file to S3 with ACL public-read
func UploadFileToS3(filePath string, bucketName, key *string, s3Uploader *s3manager.Uploader) error {
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

func UpdateImageDigests(releaseClients *ReleaseClients, r *ReleaseConfig, eksArtifacts map[string][]Artifact) (map[string]string, error) {
	// Get clients
	ecrPublicClient := releaseClients.ECRPublic.Client
	fmt.Println("============================================================")
	fmt.Println("                 Updating Image Digests                      ")
	fmt.Println("============================================================")

	imageDigests := make(map[string]string)
	for _, artifacts := range eksArtifacts {
		for _, artifact := range artifacts {
			if artifact.Image != nil {
				var imageTag string
				releaseUriSplit := strings.Split(artifact.Image.ReleaseImageURI, ":")
				repoName := strings.Replace(releaseUriSplit[0], r.ReleaseContainerRegistry+"/", "", -1)
				imageTag = releaseUriSplit[1]
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
					return nil, errors.Cause(err)
				}

				imageDigest := describeImagesOutput.ImageDetails[0].ImageDigest

				imageDigests[artifact.Image.ReleaseImageURI] = *imageDigest
				fmt.Printf("Image digest for %s - %s\n", artifact.Image.ReleaseImageURI, *imageDigest)
			}
		}
	}

	return imageDigests, nil
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

func IsObjectNotPresentError(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case "NotFound":
			return true
		default:
			return false
		}
	}
	return false
}

func IsImageNotFoundError(err error) bool {
	return err.Error() == "Requested image not found"
}

func (r *ReleaseConfig) getLatestUploadDestination() string {
	if r.BranchName == "main" {
		return "latest"
	} else {
		return r.BranchName
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
		if IsImageNotFoundError(err) && totalRetries < 60 {
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
