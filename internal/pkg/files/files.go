package files

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// GzipFileDownloadExtract downloads and extracts a specific file to destination
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

	tarReader := tar.NewReader(gzf)
	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return fmt.Errorf("%s not found: %v", fileToExtract, err)
		case err != nil:
			return fmt.Errorf("reading archive: %v", err)
		case header == nil:
			continue
		}
		switch header.Typeflag {
		case tar.TypeReg:
			name := header.FileInfo().Name()
			if strings.TrimPrefix(name, "./") == fileToExtract {
				out, err := os.OpenFile(targetFile, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
				if err != nil {
					return fmt.Errorf("opening %s file: %v", fileToExtract, err)
				}
				defer out.Close()

				_, err = io.Copy(out, tarReader)
				if err != nil {
					return fmt.Errorf("writing %s file: %v", fileToExtract, err)
				}
				logger.V(4).Info("Downloaded", "file", fileToExtract, "uri", uri)
				return nil
			}
		}
	}
}
