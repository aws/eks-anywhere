package images

import (
	"encoding/json"
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

const (
	realmKey   = "realm="
	serviceKey = "service="
	scopeKey   = "scope="
)

type tokenResponse struct {
	Token string `json:"token"`
}

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

func PollForExistenceV2(imageUri string, authHeader string) (string, error) {
	registry, repository, tag := utils.SplitImageUriV2(imageUri)
	requestUrl := fmt.Sprintf("https://%s/v2/%s/manifests/%s", registry, repository, tag)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return "", errors.Cause(err)
	}
	req.Header.Add("Authorization", authHeader)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("requested image not found")
	} else if resp.StatusCode == http.StatusUnauthorized && len(authHeader) == 0 {
		splits := strings.Split(resp.Header.Get("www-authenticate"), ",")
		var realm, service, scope string
		for _, split := range splits {
			if strings.Contains(split, realmKey) {
				startIndex := strings.Index(split, realmKey) + len(realmKey)
				realm = strings.Trim(split[startIndex:], "\"")
			} else if strings.Contains(split, serviceKey) {
				startIndex := strings.Index(split, serviceKey) + len(serviceKey)
				service = strings.Trim(split[startIndex:], "\"")
			} else if strings.Contains(split, scopeKey) {
				startIndex := strings.Index(split, scopeKey) + len(scopeKey)
				scope = strings.Trim(split[startIndex:], "\"")
			}
		}
		token, err := GetRegistryToken(realm, service, scope)
		if err != nil {
			return "", err
		}
		return PollForExistenceV2(imageUri, "Bearer "+token)
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unknown response: %s", resp.Status)
	}
	return authHeader, nil
}

func GetRegistryToken(realm, service, scope string) (string, error) {
	requestUrl := fmt.Sprintf("%s?service=\"%s\"&scope=\"%s\"", realm, service, scope)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return "", errors.Cause(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to token from %s", requestUrl)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	tokenResp := tokenResponse{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	return tokenResp.Token, nil
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
