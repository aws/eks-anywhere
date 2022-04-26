package tar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

func TarFolder(sourceFolder, dstFile string) error {
	tarfile, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("creating dst tar file: %v", err)
	}
	defer tarfile.Close()

	if err := tarFolderToWriter(sourceFolder, tarfile); err != nil {
		return fmt.Errorf("taring folder [%s] to [%s]: %v", sourceFolder, dstFile, err)
	}

	return nil
}

func tarFolderToWriter(sourceFolder string, dst io.Writer) error {
	walker := NewFolderWalker(sourceFolder)
	return Tar(walker, dst)
}

type TarFunc func(file string, info os.FileInfo, header *tar.Header) error

type Walker interface {
	Walk(TarFunc) error
}

func Tar(source Walker, dst io.Writer) error {
	tw := tar.NewWriter(dst)
	defer tw.Close()

	if err := source.Walk(addToTar(tw)); err != nil {
		return fmt.Errorf("taring: %v", err)
	}

	return nil
}

func addToTar(tw *tar.Writer) TarFunc {
	return func(file string, info os.FileInfo, header *tar.Header) error {
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	}
}
