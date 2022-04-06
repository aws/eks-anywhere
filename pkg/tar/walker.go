package tar

import (
	"archive/tar"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func NewFolderWalker(folder string) FolderWalker {
	return FolderWalker{
		folder:       folder,
		folderPrefix: fmt.Sprintf("%s/", folder),
	}
}

type FolderWalker struct {
	folder, folderPrefix string
}

func (f FolderWalker) Walk(fn TarFunc) error {
	return filepath.Walk(f.folder, f.trimFolder(fn))
}

func (f FolderWalker) trimFolder(fn TarFunc) filepath.WalkFunc {
	return func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if file == f.folder {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(file, f.folderPrefix)
		return fn(file, info, header)
	}
}
