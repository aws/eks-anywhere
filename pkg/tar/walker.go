package tar

import "path/filepath"

func NewFolderWalker(folder string) FolderWalker {
	return FolderWalker{
		folder: folder,
	}
}

type FolderWalker struct {
	folder string
}

func (f FolderWalker) Walk(fn filepath.WalkFunc) error {
	return filepath.Walk(f.folder, fn)
}
