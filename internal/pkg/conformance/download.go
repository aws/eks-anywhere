package conformance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/aws/eks-anywhere/internal/pkg/files"
)

const (
	destinationFile   = "sonobuoy"
	sonobouyGitHubAPI = "https://api.github.com/repos/vmware-tanzu/sonobuoy/releases/latest"
)

type githubRelease struct {
	Assets []asset `json:"assets"`
}

type asset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
}

func Download() error {
	resp, err := http.Get(sonobouyGitHubAPI)
	if err != nil {
		return fmt.Errorf("getting latest sonobouy version from GitHub: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading the response body for sonobouy release: %v", err)
	}

	sonobouyRelease := githubRelease{}
	if err := json.Unmarshal(body, &sonobouyRelease); err != nil {
		return fmt.Errorf("unmarshalling the response body for sonobouy release: %v", err)
	}

	downloadURL := ""
	for _, asset := range sonobouyRelease.Assets {
		if strings.Contains(asset.BrowserDownloadURL, runtime.GOOS) && strings.Contains(asset.BrowserDownloadURL, runtime.GOARCH) {
			downloadURL = asset.BrowserDownloadURL
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binaries found for sonobouy for OS %s and ARCH %s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading sonobuoy from %s\n", downloadURL)
	err = files.GzipFileDownloadExtract(downloadURL, destinationFile, "")
	if err != nil {
		return fmt.Errorf("failed to download sonobouy: %v", err)
	}
	return nil
}
