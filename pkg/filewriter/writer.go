package filewriter

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type writer struct {
	dir string
}

func NewWriter(dir string) (FileWriter, error) {
	newFolder := filepath.Join(dir, DefaultTmpFolder)
	if _, err := os.Stat(newFolder); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(newFolder, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating directory [%s]: %v", dir, err)
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
		return "", fmt.Errorf("error writing to file [%s]: %v", filePath, err)
	}

	return filePath, nil
}

func (t *writer) Copy(from, to string) error {
	fromPath := filepath.Join(t.dir, from)
	f, err := os.Stat(fromPath)
	if err != nil {
		return err
	}
	if f.IsDir() {
		return fmt.Errorf("error copying file: [%s] is a directory", from)
	}

	content, err := ioutil.ReadFile(fromPath)
	if err != nil {
		return fmt.Errorf("error reading file content: %v", err)
	}

	nw, err := t.WithDir(path.Dir(to))
	if err != nil {
		return err
	}
	_, err = nw.Write(path.Base(to), content, PersistentFile)
	if err != nil {
		return fmt.Errorf("error write content to [%s]: %v", to, err)
	}
	nw.CleanUpTemp()
	return nil
}

func (w *writer) WithDir(dir string) (FileWriter, error) {
	return NewWriter(filepath.Join(w.dir, dir))
}

func (t *writer) Dir() string {
	return t.dir
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
