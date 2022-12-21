package files

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
	untar "github.com/aws/eks-anywhere/pkg/tar"
)

// GzipFileDownloadExtract downloads and extracts a specific file to destination.
func GzipFileDownloadExtract(uri, fileToExtract, destination string) error {
	targetFile := filepath.Join(destination, fileToExtract)
	if _, err := os.Stat(targetFile); err == nil {
		logger.V(4).Info("File already exists", "file", targetFile)
		return nil
	}

	client := &http.Client{}
	resp, err := client.Get(uri)
	if err != nil {
		return fmt.Errorf("getting download: %v", err)
	}
	defer resp.Body.Close()
	gzf, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("initializing gzip: %v", err)
	}
	defer gzf.Close()

	router := singleFileRouter{
		fileName: fileToExtract,
		folder:   destination,
	}

	return untar.Untar(gzf, router)
}

type singleFileRouter struct {
	folder   string
	fileName string
}

func (s singleFileRouter) ExtractPath(header *tar.Header) string {
	if strings.TrimPrefix(header.Name, "./") != s.fileName {
		return ""
	}

	return filepath.Join(s.folder, header.FileInfo().Name())
}
