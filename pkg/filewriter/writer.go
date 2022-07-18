package filewriter

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type writer struct {
	dir     string
	tempDir string
}

func NewWriter(dir string) (FileWriter, error) {
	newFolder := filepath.Join(dir, DefaultTmpFolder)
	if _, err := os.Stat(newFolder); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(newFolder, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("creating directory [%s]: %v", dir, err)
		}
	}
	return &writer{dir: dir, tempDir: newFolder}, nil
}

func (t *writer) Write(fileName string, content []byte, f ...FileOptionsFunc) (string, error) {
	op := defaultFileOptions() // Default file options. -->> temporary file with default permissions
	for _, optionFunc := range f {
		optionFunc(op)
	}
	var currentDir string
	if op.IsTemp {
		currentDir = t.tempDir
	} else {
		currentDir = t.dir
	}
	filePath := filepath.Join(currentDir, fileName)
	err := ioutil.WriteFile(filePath, content, op.Permissions)
	if err != nil {
		return "", fmt.Errorf("writing to file [%s]: %v", filePath, err)
	}

	return filePath, nil
}

func (w *writer) WithDir(dir string) (FileWriter, error) {
	return NewWriter(filepath.Join(w.dir, dir))
}

func (t *writer) Dir() string {
	return t.dir
}

func (t *writer) TempDir() string {
	return t.tempDir
}

func (t *writer) CleanUp() {
	_, err := os.Stat(t.dir)
	if err == nil {
		os.RemoveAll(t.dir)
	}
}

func (t *writer) CleanUpTemp() {
	_, err := os.Stat(t.tempDir)
	if err == nil {
		os.RemoveAll(t.tempDir)
	}
}
