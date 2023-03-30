package filewriter

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
)

type writer struct {
	dir string
}

func NewWriter(dir string) FileWriter {
	return &writer{dir: dir}
}

func (t *writer) Write(fileName string, content []byte, f ...FileOptionsFunc) (string, error) {
	if strings.Contains(fileName, "|") {
		count := strings.Count(fileName, "|") - 1
		fileName = fmt.Sprintf("%s+%dTests", fileName[:strings.Index(fileName, "|")], count)
	}
	newFolder := filepath.Join(t.dir, DefaultTmpFolder)
	if _, err := os.Stat(newFolder); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(newFolder, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	op := defaultFileOptions() // Default file options. -->> temporary file with default permissions
	for _, optionFunc := range f {
		optionFunc(op)
	}
	var currentDir string
	if op.IsTemp {
		currentDir = t.dir + "/" + DefaultTmpFolder
	} else {
		currentDir = t.dir
	}
	filePath := filepath.Join(currentDir, fileName)
	err := os.WriteFile(filePath, content, op.Permissions)
	if err != nil {
		return "", fmt.Errorf("writing to file [%s]: %v", filePath, err)
	}

	return filePath, nil
}

func (w *writer) WithDir(dir string) (FileWriter, error) {
	return NewWriter(filepath.Join(w.dir, dir)), nil
}

func (t *writer) Dir() string {
	return t.dir
}

// This method writes the e2e test artifacts from S3 to files in a directory named after the e2e test name.
func (t *writer) WriteTestArtifactsS3ToFile(key string, data []byte) error {
	i := strings.LastIndex(key, "/Test")
	if i == -1 {
		logger.Info("Failed writing object to file", "key", key)
		return nil
	}

	p := path.Join(t.dir, key[i:])

	err := os.MkdirAll(path.Dir(p), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.WriteFile(p, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (t *writer) CleanUp() {
	_, err := os.Stat(t.dir)
	if err == nil {
		os.RemoveAll(t.dir)
	}
}

func (t *writer) CleanUpTemp() {
	currentDir := filepath.Join(t.dir, DefaultTmpFolder)
	_, err := os.Stat(currentDir)
	if err == nil {
		os.RemoveAll(currentDir)
	}
}
