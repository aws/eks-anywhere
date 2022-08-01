package tar

import (
	"archive/tar"
	"io"
	"os"
)

func UntarFile(tarFile, dstFolder string) error {
	reader, err := os.Open(tarFile)
	if err != nil {
		return err
	}

	defer reader.Close()
	return Untar(reader, NewFolderRouter(dstFolder))
}

func Untar(source io.Reader, router Router) error {
	tarReader := tar.NewReader(source)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := router.ExtractPath(header)
		if path == "" {
			continue
		}

		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}
