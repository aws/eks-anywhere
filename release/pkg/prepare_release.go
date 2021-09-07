package pkg

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	v2aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
)

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

	// Clone cli repository
	// TODO: replace these exec calls with go-git sdk calls using PlainClone
	fmt.Println("Cloning cli repository")
	parentSourceDir := filepath.Join(homeDir, "eks-a-source")
	r.CliRepoSource = filepath.Join(parentSourceDir, "eks-a-cli")
	cmd := exec.Command("git", "clone", cliRepoUrl, r.CliRepoSource)
	out, err := execCommand(cmd)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

	// Clone the build repository
	fmt.Println("Cloning build repository")
	r.BuildRepoSource = filepath.Join(parentSourceDir, "eks-a-build")
	cmd = exec.Command("git", "clone", buildRepoUrl, r.BuildRepoSource)
	out, err = execCommand(cmd)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

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
				fmt.Printf("Renaming artifact - %s\n", newArtifactFile)
				err := os.Rename(oldArtifactFile, newArtifactFile)
				if err != nil {
					return errors.Cause(err)
				}

				// Change the names of the checksum files
				checksumExtensions := []string{".sha256", ".sha512"}
				for _, extension := range checksumExtensions {
					oldChecksumFile := oldArtifactFile + extension
					newChecksumFile := newArtifactFile + extension
					fmt.Printf("Renaming artifact - %s\n", newChecksumFile)
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
					compiledRegex, err := regexp.Compile(regex)
					if err != nil {
						return errors.Cause(err)
					}

					updatedManifestFileContents := compiledRegex.ReplaceAllString(string(manifestFileContents), imageTagOverride.ReleaseUri)
					err = ioutil.WriteFile(newArtifactFile, []byte(updatedManifestFileContents), 0644)
					if err != nil {
						return errors.Cause(err)
					}
				}
			}

			// Rename the image name/tag to the release names
			dockClient := sourceClients.Docker.Client
			if artifact.Image != nil {
				imageArtifact := artifact.Image
				fmt.Printf("Renaming artifact - %s\n", imageArtifact.ReleaseImageURI)
				err := dockClient.TagImage(imageArtifact.SourceImageURI, docker.TagImageOptions{
					Repo:    imageArtifact.ReleaseImageURI,
					Context: context.Background(),
				})
				if err != nil {
					return errors.Cause(err)
				}
			}
		}
	}
	return nil
}

func downloadArtifacts(sourceClients *SourceClients, r *ReleaseConfig, eksArtifacts map[string][]Artifact) error {
	// Get s3 client and docker clients
	dockClient := sourceClients.Docker.Client
	s3Downloader := sourceClients.S3.Downloader
	s3Client := sourceClients.S3.Client
	ecrAuthConfig := sourceClients.Docker.AuthConfig
	fmt.Println("============================================================")
	fmt.Println("                 Downloading Artifacts                      ")
	fmt.Println("============================================================")

	for _, artifacts := range eksArtifacts {
		for _, artifact := range artifacts {

			// Check if there is an archive to be downloaded
			if artifact.Archive != nil {
				err := s3Client.ListObjectsPages(&s3.ListObjectsInput{
					Bucket: aws.String(r.SourceBucket),
					Prefix: aws.String(artifact.Archive.SourceS3Prefix),
				}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
					for _, obj := range page.Contents {
						objName := filepath.Base(v2aws.ToString(obj.Key))
						file := filepath.Join(artifact.Archive.ArtifactPath, objName)
						if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
							return false
						}

						fd, err := os.Create(file)
						if err != nil {
							return false
						}
						defer fd.Close()
						fmt.Printf("Archive - %s\n", v2aws.ToString(obj.Key))
						_, err = s3Downloader.Download(fd, &s3.GetObjectInput{Bucket: &r.SourceBucket, Key: obj.Key})
						if err != nil {
							return false
						}
					}
					return true
				})
				if err != nil {
					return errors.Cause(err)
				}
			}

			// Check if there is a manifest to be downloaded
			if artifact.Manifest != nil {
				err := s3Client.ListObjectsPages(&s3.ListObjectsInput{
					Bucket: aws.String(r.SourceBucket),
					Prefix: aws.String(artifact.Manifest.SourceS3Prefix),
				}, func(page *s3.ListObjectsOutput, lastPage bool) bool {
					for _, obj := range page.Contents {
						objName := filepath.Base(v2aws.ToString(obj.Key))
						file := filepath.Join(artifact.Manifest.ArtifactPath, objName)
						if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
							return false
						}

						fd, err := os.Create(file)
						if err != nil {
							return false
						}
						defer fd.Close()
						fmt.Printf("Manifest - %s\n", v2aws.ToString(obj.Key))
						_, err = s3Downloader.Download(fd, &s3.GetObjectInput{Bucket: &r.SourceBucket, Key: obj.Key})
						if err != nil {
							return false
						}
					}
					return true
				})
				if err != nil {
					return errors.Cause(err)
				}
			}

			// Check if there is image to be pulled to local
			if artifact.Image != nil {
				fmt.Printf("Image   - %s\n", artifact.Image.SourceImageURI)
				// TODO: replace background context with proper timeouts
				err := dockClient.PullImage(docker.PullImageOptions{
					Repository: artifact.Image.SourceImageURI,
					Context:    context.Background(),
				}, *ecrAuthConfig)
				if err != nil {
					return errors.Cause(err)
				}
			}
		}
	}
	return nil
}

func UploadArtifacts(releaseClients *ReleaseClients, r *ReleaseConfig, eksArtifacts map[string][]Artifact) error {
	// Get clients
	s3Uploader := releaseClients.S3.Uploader
	dockClient := releaseClients.Docker.Client
	ecrAuthConfig := releaseClients.Docker.AuthConfig
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
				fmt.Printf("Image   - %s\n", artifact.Image.ReleaseImageURI)
				err := dockClient.PushImage(docker.PushImageOptions{
					Name:    artifact.Image.ReleaseImageURI,
					Context: context.Background(),
				}, *ecrAuthConfig)
				if err != nil {
					return errors.Cause(err)
				}
			}
		}
	}

	return nil
}

// UploadFileToS3 uploads the file to s3 with ACL public-read
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
			}
		}
	}

	return imageDigests, nil
}
