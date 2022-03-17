package hardware

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultManifestDir is the default directory for writing Tinkerbell hardware files.
const DefaultManifestDir = "hardware-manifests"

// DefaultJsonDir is the default directory for writing hardware json files.
const DefaultJsonDir = "json"

// CreateDefaultManifestDir creates the defaut directory where hardware files are written returning it as the string parameter.
func CreateManifestDir(path string) (string, error) {
	if path == "" {
		path = DefaultManifestDir
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", fmt.Errorf(
			"could not create manifest directory: %v: %v",
			path,
			err,
		)
	}

	return path, nil
}

// CreateDefaultJsonDir creates the defaut directory where hardware json files are written returning it as the string
// parameter.
func CreateDefaultJsonDir(basepath string) (string, error) {
	path := filepath.Join(basepath, DefaultJsonDir)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", fmt.Errorf(
			"could not create json manifest directory: %v: %v",
			path,
			err,
		)
	}
	return path, nil
}
