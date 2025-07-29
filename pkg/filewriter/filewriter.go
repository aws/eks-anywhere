package filewriter

import (
	"io"
	"os"
)

type FileWriter interface {
	Write(fileName string, content []byte, f ...FileOptionsFunc) (path string, err error)
	WithDir(dir string) (FileWriter, error)
	CleanUp()
	CleanUpTemp()
	Dir() string
	TempDir() string
	Create(name string, f ...FileOptionsFunc) (_ io.WriteCloser, path string, _ error)
	Delete(fileName string) error
}

type FileOptions struct {
	IsTemp      bool
	Permissions os.FileMode
}

type FileOptionsFunc func(op *FileOptions)
