package filewriter

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type writer struct {
	dir string
}

func NewWriter(dir string) (FileWriter, error) {
	newFolder := filepath.Join(dir, DefaultTmpFolder)
	if _, err := os.Stat(newFolder); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(newFolder, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("creating directory [%s]: %v", dir, err)
		}
	}
	return &writer{dir: dir}, nil
}

func (t *writer) Write(fileName string, content []byte, f ...FileOptionsFunc) (string, error) {
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

// This method writes the e2e test artifacts from S3 to files in a directory named after the e2e test name.
// Since OIDC tests have an additional folder within the S3 directory, we must take an additional step to
// ensure that the OIDC files for the associated e2e test are placed in the correct directory
func (t *writer) WriteTestArtifactsS3ToFile(key string, data []byte) error {
	i := strings.LastIndex(key, "/")
	filePath := path.Base(key[:i])
	if filePath == "oidc" {
		j := strings.LastIndex(key[:i], "/")
		filePath = path.Base(key[:j])
	}
	d := path.Join(t.dir, filePath)

	i = strings.LastIndex(key, filePath)
	f := path.Join(t.dir, key[i:])

	err := os.MkdirAll(d, os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(f, data, os.ModePerm)
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
