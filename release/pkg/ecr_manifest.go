package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type ManifestMeta struct {
	MediaType string
	Digest    string
	Size      int
}

type Platform struct {
	Architecture string
	Os           string
}

type Manifest struct {
	ManifestMeta
	Platform Platform
}

type ImageManifest struct {
	MediaType     string
	SchemaVersion int
	Config        ManifestMeta
	Layers        []ManifestMeta
}

type ManifestList struct {
	MediaType     string
	SchemaVersion int
	Manifests     []Manifest
}

func (r *ReleaseConfig) getImageUriMap(sourceClients *SourceClients, sourceImageUri string) (map[string]string, error) {
	var imageManifest ImageManifest
	var manifestList ManifestList
	archImageUriMap := map[string]string{}

	sourceImageUriSplit := strings.Split(sourceImageUri, ":")
	sourceImageName := strings.Replace(sourceImageUriSplit[0], r.SourceContainerRegistry+"/", "", -1)
	sourceImageTag := sourceImageUriSplit[1]

	var authToken, authHeader, requestUrl string
	var err error
	if r.DevRelease || r.ReleaseEnvironment == "development" {
		// Get ECR authorization token
		authToken, err = getEcrAuthToken(sourceClients.ECR.EcrClient)
		if err != nil {
			return nil, errors.Cause(err)
		}
		authHeader = fmt.Sprintf("Basic %s", authToken)
		requestUrl = fmt.Sprintf("https://%s/v2/%s/manifests/%s", r.SourceContainerRegistry, sourceImageName, sourceImageTag)
	} else {
		// Get ECR Public authorization token
		authToken, err = getEcrPublicAuthToken(sourceClients.ECR.EcrPublicClient)
		if err != nil {
			return nil, errors.Cause(err)
		}
		authHeader = fmt.Sprintf("Bearer %s", authToken)
		requestUrl = fmt.Sprintf("https://public.ecr.aws/v2/%s/%s/manifests/%s", filepath.Base(r.SourceContainerRegistry), sourceImageName, sourceImageTag)
	}

	// Creating new GET request with
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, errors.Cause(err)
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Cause(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Cause(err)
	}
	fmt.Println(string(body))

	if err = json.Unmarshal(body, &imageManifest); err != nil {
		return nil, errors.Cause(err)
	}
	if imageManifest.Layers == nil {
		if err := json.Unmarshal(body, &manifestList); err != nil {
			return nil, errors.Cause(err)
		}
		for _, manifest := range manifestList.Manifests {
			archImageUriMap[manifest.Platform.Architecture] = fmt.Sprintf("%s/%s@%s", r.SourceContainerRegistry, sourceImageName, manifest.Digest)
		}
	} else {
		archImageUriMap["amd64"] = fmt.Sprintf("%s/%s@%s", r.SourceContainerRegistry, sourceImageName, imageManifest.Config.Digest)
	}

	return archImageUriMap, nil
}

func createManifest(releaseImageUri string, archImageUriMap map[string]string) error {
	cmd := exec.Command("docker", "manifest", "create", releaseImageUri, archImageUriMap["amd64"], archImageUriMap["arm64"])
	out, err := execCommand(cmd)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

	return nil
}

func pushManifest(releaseImageUri string) error {
	cmd := exec.Command("docker", "manifest", "push", releaseImageUri)
	out, err := execCommand(cmd)
	if err != nil {
		return errors.Cause(err)
	}
	fmt.Println(out)

	return nil
}
