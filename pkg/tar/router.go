package tar

import (
	"archive/tar"
	"path/filepath"
)

// Router instructs where to extract a file.
type Router interface {
	// ExtractPath instructs the path where a file should be extracted.
	// Empty strings instructs to omit the file extraction
	ExtractPath(header *tar.Header) string
}

type FolderRouter struct {
	folder string
}

func NewFolderRouter(folder string) FolderRouter {
	return FolderRouter{folder: folder}
}

func (f FolderRouter) ExtractPath(header *tar.Header) string {
	return filepath.Join(f.folder, header.Name)
}
