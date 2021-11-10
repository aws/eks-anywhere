package filewriter

import "os"

type FileWriter interface {
	Write(fileName string, content []byte, f ...FileOptionsFunc) (path string, err error)
	WithDir(dir string) (FileWriter, error)
	CleanUp()
	CleanUpTemp()
	Dir() string
	Copy(from, to string) error
}

type FileOptions struct {
	IsTemp      bool
	Permissions os.FileMode
}

type FileOptionsFunc func(op *FileOptions)
