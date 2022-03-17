package hardware

import (
	"fmt"
	"os"
)

// DefaultDir is the default directory for writing Tinkerbell hardware files.
const DefaultDir = "hardware-manifests"

// DefaultJsonDir is the default directory for writing hardware json files.
const DefaultJsonDir = "hardware-manifests/json"

// CreateDefaultDir creates the defaut directory where hardware files are written returning it as the string parameter.
func CreateDefaultDir() (string, error) {
	if err := os.MkdirAll(DefaultDir, os.ModePerm); err != nil {
		return "", fmt.Errorf(
			"could not create default hardware directory: %v: %v",
			DefaultDir,
			err,
		)
	}

	return DefaultDir, nil
}

// CreateDefaultJsonDir creates the defaut directory where hardware json files are written returning it as the string
// parameter.
func CreateDefaultJsonDir() (string, error) {
	if err := os.MkdirAll(DefaultJsonDir, os.ModePerm); err != nil {
		return "", fmt.Errorf(
			"could not create default tinkerbell hardware json directory: %v: %v",
			DefaultJsonDir,
			err,
		)
	}

	return DefaultJsonDir, nil
}
