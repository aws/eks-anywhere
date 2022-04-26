package tar

import (
	"compress/gzip"
	"fmt"
	"os"
)

func GzipTarFolder(sourceFolder, dstFile string) error {
	tarfile, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("creating dst tar file: %v", err)
	}
	defer tarfile.Close()
	gw := gzip.NewWriter(tarfile)
	defer gw.Close()

	if err := tarFolderToWriter(sourceFolder, gw); err != nil {
		return fmt.Errorf("gzip taring folder [%s] to [%s]: %v", sourceFolder, dstFile, err)
	}

	return nil
}

func UnGzipTarFile(tarFile, dstFolder string) error {
	tarball, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer tarball.Close()

	gr, err := gzip.NewReader(tarball)
	if err != nil {
		return err
	}
	defer gr.Close()

	return Untar(gr, NewFolderRouter(dstFolder))
}
