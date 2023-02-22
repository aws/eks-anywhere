package filewriter

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
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

func (w *writer) Write(fileName string, content []byte, opts ...FileOptionsFunc) (string, error) {
	o := buildOptions(w, opts)

	filePath := filepath.Join(o.BasePath, fileName)
	err := os.WriteFile(filePath, content, o.Permissions)
	if err != nil {
		return "", fmt.Errorf("writing to file [%s]: %v", filePath, err)
	}

	return filePath, nil
}

func (w *writer) WithDir(dir string) (FileWriter, error) {
	return NewWriter(filepath.Join(w.dir, dir))
}

func (w *writer) Dir() string {
	return w.dir
}

func (w *writer) TempDir() string {
	return w.tempDir
}

func (w *writer) CleanUp() {
	_, err := os.Stat(w.dir)
	if err == nil {
		os.RemoveAll(w.dir)
	}
}

func (w *writer) CleanUpTemp() {
	_, err := os.Stat(w.tempDir)
	if err == nil {
		os.RemoveAll(w.tempDir)
	}
}

// Create creates a file with the given name rooted at w's base directory.
func (w *writer) Create(name string, opts ...FileOptionsFunc) (_ io.WriteCloser, path string, _ error) {
	o := buildOptions(w, opts)

	path = filepath.Join(o.BasePath, name)
	fh, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, o.Permissions)
	return fh, path, err
}

type options struct {
	BasePath    string
	Permissions fs.FileMode
}

// buildOptions converts a set of FileOptionsFunc's to a single options struct.
func buildOptions(w *writer, opts []FileOptionsFunc) options {
	op := defaultFileOptions()
	for _, fn := range opts {
		fn(op)
	}

	var basePath string
	if op.IsTemp {
		basePath = w.tempDir
	} else {
		basePath = w.dir
	}

	return options{
		BasePath:    basePath,
		Permissions: op.Permissions,
	}
}
