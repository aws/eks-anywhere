package images

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/pkg/retrier"
	"github.com/aws/eks-anywhere/release/pkg/utils"
)

func PollForExistence(devRelease bool, authConfig *docker.AuthConfiguration, imageUri, imageContainerRegistry, releaseEnvironment, branchName string) error {
	repository, tag := utils.SplitImageUri(imageUri, imageContainerRegistry)

	var requestUrl string
	if devRelease || releaseEnvironment == "development" {
		requestUrl = fmt.Sprintf("https://%s:%s@%s/v2/%s/manifests/%s", authConfig.Username, authConfig.Password, imageContainerRegistry, repository, tag)
	} else {
		requestUrl = fmt.Sprintf("https://%s:%s@public.ecr.aws/v2/%s/%s/manifests/%s", authConfig.Username, authConfig.Password, filepath.Base(imageContainerRegistry), repository, tag)
	}

	// Creating new GET request
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return errors.Cause(err)
	}

	// Retrier for downloading source ECR images. This retrier has a max timeout of 60 minutes. It
	// checks whether the error occured during download is an ImageNotFound error and retries the
	// download operation for a maximum of 60 retries, with a wait time of 30 seconds per retry.
	retrier := retrier.NewRetrier(60*time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		if branchName == "main" && utils.IsImageNotFoundError(err) && totalRetries < 60 {
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
		return fmt.Errorf("retries exhausted waiting for source image %s to be available for copy: %v", imageUri, err)
	}

	return nil
}

func CopyToDestination(sourceAuthConfig, releaseAuthConfig *docker.AuthConfiguration, sourceImageUri, releaseImageUri string) error {
	sourceRegistryUsername := sourceAuthConfig.Username
	sourceRegistryPassword := sourceAuthConfig.Password
	releaseRegistryUsername := releaseAuthConfig.Username
	releaseRegistryPassword := releaseAuthConfig.Password
	cmd := exec.Command("skopeo", "copy", "--src-creds", fmt.Sprintf("%s:%s", sourceRegistryUsername, sourceRegistryPassword), "--dest-creds", fmt.Sprintf("%s:%s", releaseRegistryUsername, releaseRegistryPassword), fmt.Sprintf("docker://%s", sourceImageUri), fmt.Sprintf("docker://%s", releaseImageUri), "-f", "oci", "--all")
	out, err := utils.ExecCommand(cmd)
	fmt.Println(out)
	if err != nil {
		return fmt.Errorf("executing skopeo copy command: %v", err)
	}

	return nil
}
