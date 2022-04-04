package tar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Packager struct{}

func NewPackager() Packager {
	return Packager{}
}

func (Packager) Package(sourceFolder, dstFile string) error {
	return TarFolder(sourceFolder, dstFile)
}

func TarFolder(sourceFolder, dstFile string) error {
	tarfile, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("creating dst tar file: %v", err)
	}
	defer tarfile.Close()

	walker := NewFolderWalker(sourceFolder)

	if err := Tar(walker, tarfile); err != nil {
		return fmt.Errorf("taring folder [%s] to [%s]: %v", sourceFolder, dstFile, err)
	}

	return nil
}

type Walker interface {
	Walk(fn filepath.WalkFunc) error
}

func Tar(source Walker, dst io.Writer) error {
	tw := tar.NewWriter(dst)
	defer tw.Close()

	if err := source.Walk(addToTar(tw)); err != nil {
		return fmt.Errorf("taring: %v", err)
	}

	return nil
}

func addToTar(tw *tar.Writer) filepath.WalkFunc {
	return func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		if err = tw.WriteHeader(header); err != nil {
			return err
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
