package filewriter

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Writer is a filesystem context aware io abstraction.
type Writer struct {
	dir     string
	tempDir string
}

// NewWriter creates a new Writer istance configured with a root directory at dir.
func NewWriter(dir string) (*Writer, error) {
	newFolder := filepath.Join(dir, DefaultTmpFolder)
	if _, err := os.Stat(newFolder); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(newFolder, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("creating directory [%s]: %v", dir, err)
		}
	}
	return &Writer{dir: dir, tempDir: newFolder}, nil
}

// Writes a file called fileName containing content in the directory configured in w.
func (w *Writer) Write(fileName string, content []byte, f ...FileOptionsFunc) (string, error) {
	op := defaultFileOptions() // Default file options. -->> temporary file with default permissions
	for _, optionFunc := range f {
		optionFunc(op)
	}
	var currentDir string
	if op.IsTemp {
		currentDir = w.tempDir
	} else {
		currentDir = w.dir
	}
	filePath := filepath.Join(currentDir, fileName)
	err := ioutil.WriteFile(filePath, content, op.Permissions)
	if err != nil {
		return "", fmt.Errorf("writing to file [%s]: %v", filePath, err)
	}

	return filePath, nil
}

// Create creates a file called fileName in the directory configured in w. If the file already
// exists it is truncated.
func (w *Writer) Create(fileName string) (fh io.WriteCloser, absPath string, err error) {
	absPath = filepath.Join(w.dir, fileName)
	fh, err = os.Create(absPath)
	if err != nil {
		return nil, "", fmt.Errorf("creating file [%s]: %v", absPath, err)
	}

	return fh, absPath, nil
}

// WithDir creates a new Writer instance with dir appended to w's configured root directory.
func (w *Writer) WithDir(dir string) (FileWriter, error) {
	return NewWriter(filepath.Join(w.dir, dir))
}

// Dir retrieves w's configured root directory.
func (w *Writer) Dir() string {
	return w.dir
}

// TempDir retrieves w's configured temp directory.
func (w *Writer) TempDir() string {
	return w.tempDir
}

// CleanUp removes all files and directories in w's configured root directory.
func (w *Writer) CleanUp() {
	_, err := os.Stat(w.dir)
	if err == nil {
		os.RemoveAll(w.dir)
	}
}

// CleanUpTemp removes all files and directories in w's configured temp directory.
func (w *Writer) CleanUpTemp() {
	_, err := os.Stat(w.tempDir)
	if err == nil {
		os.RemoveAll(w.tempDir)
	}
}
