package filewriter

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/logger"
)

type writer struct {
	dir       string
	tmpFolder string
}

func NewWriter(dir string) (FileWriter, error) {
	tmpFolder := DefaultTmpFolder
	newFolder := filepath.Join(dir, tmpFolder)
	if _, err := os.Stat(newFolder); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(newFolder, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating directory [%s]: %v", dir, err)
		}
	}
	return &writer{dir: dir, tmpFolder: tmpFolder}, nil
}

func (t *writer) Write(fileName string, content []byte, f ...FileOptionsFunc) (string, error) {
	op := defaultFileOptions() // Default file options. -->> temporary file with default permissions
	for _, optionFunc := range f {
		optionFunc(op)
	}
	var currentDir string
	if op.IsTemp {
		currentDir = t.TmpDir()
	} else {
		currentDir = t.dir
	}
	filePath := filepath.Join(currentDir, fileName)
	err := ioutil.WriteFile(filePath, content, op.Permissions)
	if err != nil {
		return "", fmt.Errorf("error writing to file [%s]: %v", filePath, err)
	}

	return filePath, nil
}

func (w *writer) WithDir(dir string) (FileWriter, error) {
	return NewWriter(filepath.Join(w.dir, dir))
}

func (t *writer) Dir() string {
	return t.dir
}

func (t *writer) TmpDir() string {
	return filepath.Join(t.dir, t.tmpFolder)
}

func (t *writer) CleanUp() {
	_, err := os.Stat(t.dir)
	if err == nil {
		os.RemoveAll(t.dir)
	}
}

func (t *writer) CleanUpTemp() {
	currentDir := t.TmpDir()
	if _, err := os.Stat(currentDir); err == nil {
		if err = os.RemoveAll(currentDir); err != nil {
			logger.Error(err, "failed deleting tmp dir")
		}
	}
}
