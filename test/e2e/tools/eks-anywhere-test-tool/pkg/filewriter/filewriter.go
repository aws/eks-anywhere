package filewriter

import "os"

type FileWriter interface {
	Write(fileName string, content []byte, f ...FileOptionsFunc) (path string, err error)
	WithDir(dir string) (FileWriter, error)
	WriteTestArtifactsS3ToFile(key string, data []byte) error
	CleanUp()
	CleanUpTemp()
	Dir() string
}

type FileOptions struct {
	IsTemp      bool
	Permissions os.FileMode
}

type FileOptionsFunc func(op *FileOptions)
